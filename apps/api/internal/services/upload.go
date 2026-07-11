package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"xians/api/internal/config"
	"xians/api/internal/models"
	"xians/api/internal/storage"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrFileTooLarge       = errors.New("file exceeds maximum size")
	ErrBatchTooLarge      = errors.New("batch exceeds maximum total size")
	ErrTooManyFiles       = errors.New("too many files in batch")
	ErrUnsupportedType    = errors.New("unsupported file type")
	ErrEmptyFile          = errors.New("file is empty")
	ErrBatchNotFound      = errors.New("batch not found")
	ErrBatchNotPending    = errors.New("batch is not in pending state")
	ErrSlugCollision      = errors.New("slug collision after retries")
	ErrUploadNotVerified  = errors.New("upload could not be verified")
	ErrGalleryNotActive   = errors.New("gallery is not available")
	ErrPasswordRequired   = errors.New("password required")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrDownloadLimitMet   = errors.New("download limit reached")
	ErrOwnerTokenRequired = errors.New("owner token required")
	ErrInvalidOwnerToken  = errors.New("invalid owner token")
	ErrFileNotFound       = errors.New("file not found in batch")

	supportedImageTypes = map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
		"image/gif":  ".gif",
	}
	supportedVideoTypes = map[string]string{
		"video/mp4":          ".mp4",
		"video/quicktime":    ".mov",
		"video/webm":         ".webm",
	}
)

func getAllSupportedTypes() map[string]string {
	m := make(map[string]string)
	for k, v := range supportedImageTypes {
		m[k] = v
	}
	for k, v := range supportedVideoTypes {
		m[k] = v
	}
	return m
}

type UploadService struct {
	cfg    *config.Config
	repo   *models.BatchRepository
	store  *storage.R2Client
}

func NewUploadService(cfg *config.Config, repo *models.BatchRepository, store *storage.R2Client) *UploadService {
	return &UploadService{cfg: cfg, repo: repo, store: store}
}

func (s *UploadService) InitBatch(ctx context.Context, files []models.FileManifest, ip string) (*models.InitResponse, error) {
	supported := getAllSupportedTypes()

	if len(files) > s.cfg.MaxFilesPerBatch {
		return nil, ErrTooManyFiles
	}

	var totalSize int64
	for _, f := range files {
		if f.SizeBytes == 0 {
			return nil, fmt.Errorf("%w: %s", ErrEmptyFile, f.Name)
		}
		if f.SizeBytes > s.cfg.MaxFileSizeBytes() {
			return nil, fmt.Errorf("%w: %s (%dMB)", ErrFileTooLarge, f.Name, s.cfg.MaxFileSizeMB)
		}
		if _, ok := supported[f.ContentType]; !ok {
			return nil, fmt.Errorf("%w: %s (%s)", ErrUnsupportedType, f.Name, f.ContentType)
		}
		totalSize += f.SizeBytes
	}
	if totalSize > s.cfg.MaxBatchSizeBytes() {
		return nil, ErrBatchTooLarge
	}

	ipHash := hashIP(ip)
	slug, err := generateSlug()
	if err != nil {
		return nil, fmt.Errorf("generate slug: %w", err)
	}
	ownerToken, err := generateOwnerToken()
	if err != nil {
		return nil, fmt.Errorf("generate owner token: %w", err)
	}

	batchFiles := make([]models.File, len(files))
	presignedFiles := make([]models.FilePresigned, len(files))

	for i, f := range files {
		fileID := shortID()
		ext := supported[f.ContentType]
		batchFiles[i] = models.File{
			ID:           fileID,
			OriginalName: f.Name,
			ContentType:  f.ContentType,
			SizeBytes:    f.SizeBytes,
		}

		url, err := s.store.PresignPutURL(ctx, slug, fileID, ext)
		if err != nil {
			return nil, fmt.Errorf("presign put %s: %w", fileID, err)
		}
		presignedFiles[i] = models.FilePresigned{
			FileID: fileID,
			URL:    url,
			Method: "PUT",
		}
	}

	batch := &models.Batch{
		ID:             generateBatchID(),
		Slug:           slug,
		Files:          batchFiles,
		CreatedAt:      time.Now(),
		Status:         models.BatchStatusPending,
		OwnerToken:     ownerToken,
		UploaderIPHash: ipHash,
	}

	if err := s.repo.Insert(ctx, batch); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, ErrSlugCollision
		}
		return nil, fmt.Errorf("insert batch: %w", err)
	}

	return &models.InitResponse{
		BatchID: batch.ID,
		Files:   presignedFiles,
	}, nil
}

func (s *UploadService) UploadFile(ctx context.Context, batchID, fileID string, body io.Reader) error {
	batch, err := s.repo.FindByID(ctx, batchID)
	if err != nil {
		return fmt.Errorf("find batch: %w", err)
	}
	if batch == nil {
		return ErrBatchNotFound
	}
	if batch.Status != models.BatchStatusPending {
		return ErrBatchNotPending
	}

	var ext string
	for _, f := range batch.Files {
		if f.ID == fileID {
			ext = extensionForContentType(f.ContentType)
			break
		}
	}
	if ext == "" {
		return ErrFileNotFound
	}

	err = s.store.PutObject(ctx, batch.Slug, fileID, ext, body)
	if err != nil {
		log.Error().Err(err).Str("slug", batch.Slug).Str("fileID", fileID).Str("ext", ext).Msg("PutObject failed")
		return err
	}
	return nil
}

func (s *UploadService) CompleteBatch(ctx context.Context, batchID string, req models.CompleteRequest) (*models.CompleteResponse, error) {
	batch, err := s.repo.FindByID(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("find batch: %w", err)
	}
	if batch == nil {
		return nil, ErrBatchNotFound
	}
	if batch.Status != models.BatchStatusPending {
		return nil, ErrBatchNotPending
	}

	supported := getAllSupportedTypes()

	for _, f := range batch.Files {
		ext := supported[f.ContentType]
		exists, err := s.store.HeadObject(ctx, batch.Slug, f.ID, ext)
		if err != nil {
			return nil, fmt.Errorf("verify %s: %w", f.ID, err)
		}
		if !exists {
			return nil, fmt.Errorf("%w: %s", ErrUploadNotVerified, f.OriginalName)
		}
	}

	expiresAt := computeExpiry(req.TTL)
	var passwordHash *string
	if req.Password != nil && *req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), 12)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		h := string(hash)
		passwordHash = &h
	}

	if err := s.repo.UpdateStatus(ctx, batch.ID, models.BatchStatusActive, expiresAt, passwordHash, req.MaxDownloads); err != nil {
		return nil, fmt.Errorf("update batch: %w", err)
	}

	return &models.CompleteResponse{
		Slug:       batch.Slug,
		OwnerToken: batch.OwnerToken,
	}, nil
}

type GalleryService struct {
	repo  *models.BatchRepository
	store *storage.R2Client
}

func NewGalleryService(repo *models.BatchRepository, store *storage.R2Client) *GalleryService {
	return &GalleryService{repo: repo, store: store}
}

func (s *GalleryService) GetGallery(ctx context.Context, slug string, hasAccess bool) (*models.GalleryResponse, error) {
	batch, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("find batch: %w", err)
	}
	if batch == nil {
		return nil, ErrBatchNotFound
	}
	if batch.Status != models.BatchStatusActive {
		return nil, ErrGalleryNotActive
	}

	// check status in handler layer too (defense in depth)
	if batch.Status == models.BatchStatusReported || batch.Status == models.BatchStatusExpired || batch.Status == models.BatchStatusDeleted {
		return nil, ErrGalleryNotActive
	}

	if batch.PasswordHash != nil && !hasAccess {
		return &models.GalleryResponse{
			Slug:             batch.Slug,
			RequiresPassword: true,
			Status:           string(batch.Status),
		}, nil
	}

	var expiresAt *string
	if batch.ExpiresAt != nil {
		e := batch.ExpiresAt.Format(time.RFC3339)
		expiresAt = &e
	}

	fileInfos := make([]models.FileInfo, len(batch.Files))
	for i, f := range batch.Files {
		fileInfos[i] = models.FileInfo{
			ID:           f.ID,
			OriginalName: f.OriginalName,
			ContentType:  f.ContentType,
			SizeBytes:    f.SizeBytes,
			Width:        f.Width,
			Height:       f.Height,
			DurationMs:   f.DurationMs,
		}
	}

	return &models.GalleryResponse{
		Slug:             batch.Slug,
		Files:            fileInfos,
		CreatedAt:        batch.CreatedAt.Format(time.RFC3339),
		ExpiresAt:        expiresAt,
		RequiresPassword: false,
		DownloadCount:    batch.DownloadCount,
		MaxDownloads:     batch.MaxDownloads,
		Status:           string(batch.Status),
	}, nil
}

func (s *GalleryService) Unlock(ctx context.Context, slug, password string) (string, error) {
	batch, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return "", fmt.Errorf("find batch: %w", err)
	}
	if batch == nil || batch.PasswordHash == nil {
		return "", ErrPasswordRequired
	}
	if batch.Status != models.BatchStatusActive {
		return "", ErrGalleryNotActive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*batch.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidPassword
	}

	return batch.Slug, nil
}

func (s *GalleryService) GetFileDownloadURL(ctx context.Context, slug, fileID string, hasAccess bool) (string, error) {
	batch, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return "", fmt.Errorf("find batch: %w", err)
	}
	if batch == nil || batch.Status != models.BatchStatusActive {
		return "", ErrGalleryNotActive
	}

	if batch.PasswordHash != nil && !hasAccess {
		return "", ErrPasswordRequired
	}

	var file *models.File
	for i := range batch.Files {
		if batch.Files[i].ID == fileID {
			file = &batch.Files[i]
			break
		}
	}
	if file == nil {
		return "", ErrBatchNotFound
	}

	if batch.MaxDownloads != nil {
		ok, err := s.repo.AtomicIncrementDownload(ctx, slug, *batch.MaxDownloads)
		if err != nil {
			return "", fmt.Errorf("increment download: %w", err)
		}
		if !ok {
			return "", ErrDownloadLimitMet
		}
	} else {
		// increment without limit check
		_, err := s.repo.AtomicIncrementDownload(ctx, slug, 0)
		if err != nil {
			return "", fmt.Errorf("increment download: %w", err)
		}
	}

	supported := getAllSupportedTypes()
	ext, ok := supported[file.ContentType]
	if !ok {
		ext = ""
	}

	url, err := s.store.PresignGetURL(ctx, slug, fileID, ext)
	if err != nil {
		return "", fmt.Errorf("presign get: %w", err)
	}

	return url, nil
}

func (s *GalleryService) StreamFile(ctx context.Context, slug, fileID string, hasAccess bool) (io.ReadCloser, string, int64, error) {
	batch, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, "", 0, fmt.Errorf("find batch: %w", err)
	}
	if batch == nil || batch.Status != models.BatchStatusActive {
		return nil, "", 0, ErrGalleryNotActive
	}

	if batch.PasswordHash != nil && !hasAccess {
		return nil, "", 0, ErrPasswordRequired
	}

	var file *models.File
	for i := range batch.Files {
		if batch.Files[i].ID == fileID {
			file = &batch.Files[i]
			break
		}
	}
	if file == nil {
		return nil, "", 0, ErrBatchNotFound
	}

	supported := getAllSupportedTypes()
	ext, ok := supported[file.ContentType]
	if !ok {
		ext = ""
	}

	body, contentType, contentLength, err := s.store.GetObjectStream(ctx, slug, fileID, ext)
	if err != nil {
		return nil, "", 0, fmt.Errorf("get object stream: %w", err)
	}

	return body, *contentType, contentLength, nil
}

func (s *GalleryService) StreamZip(ctx context.Context, slug string, w http.ResponseWriter, hasAccess bool) error {
	batch, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return fmt.Errorf("find batch: %w", err)
	}
	if batch == nil || batch.Status != models.BatchStatusActive {
		return ErrGalleryNotActive
	}

	if batch.PasswordHash != nil && !hasAccess {
		return ErrPasswordRequired
	}

	supported := getAllSupportedTypes()

	var entries []storage.ZipFileEntry
	for _, f := range batch.Files {
		ext := supported[f.ContentType]
		entries = append(entries, storage.ZipFileEntry{
			ID:          f.ID,
			Name:        f.OriginalName,
			Ext:         ext,
			ContentType: f.ContentType,
			Size:        f.SizeBytes,
		})
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.zip"`, batch.Slug))

	if err := s.store.StreamZip(ctx, batch.Slug, entries, w); err != nil {
		return fmt.Errorf("stream zip: %w", err)
	}
	return nil
}

func (s *GalleryService) Report(ctx context.Context, slug string) error {
	batch, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return fmt.Errorf("find batch: %w", err)
	}
	if batch == nil {
		return ErrBatchNotFound
	}

	return s.repo.UpdateStatus(ctx, batch.ID, models.BatchStatusReported, nil, nil, nil)
}

func (s *GalleryService) Delete(ctx context.Context, slug, ownerToken string) error {
	batch, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return fmt.Errorf("find batch: %w", err)
	}
	if batch == nil {
		return ErrBatchNotFound
	}
	if batch.OwnerToken != ownerToken {
		return ErrInvalidOwnerToken
	}

	if err := s.store.ListAndDeleteBatchObjects(ctx, batch.Slug); err != nil {
		return fmt.Errorf("delete r2 objects: %w", err)
	}

	if err := s.repo.UpdateStatus(ctx, batch.ID, models.BatchStatusDeleted, nil, nil, nil); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	return nil
}

type CleanupService struct {
	repo  *models.BatchRepository
	store *storage.R2Client
}

func NewCleanupService(repo *models.BatchRepository, store *storage.R2Client) *CleanupService {
	return &CleanupService{repo: repo, store: store}
}

func (s *CleanupService) CleanupExpired(ctx context.Context) error {
	now := time.Now()

	expired, err := s.repo.FindExpiredActive(ctx, now)
	if err != nil {
		return fmt.Errorf("find expired: %w", err)
	}

	for _, b := range expired {
		log.Ctx(ctx).Info().Str("slug", b.Slug).Msg("cleaning expired batch")
		if err := s.store.ListAndDeleteBatchObjects(ctx, b.Slug); err != nil {
			log.Ctx(ctx).Error().Err(err).Str("slug", b.Slug).Msg("failed to delete R2 objects, skipping mongo delete")
			continue
		}
		if err := s.repo.DeleteByID(ctx, b.ID); err != nil {
			log.Ctx(ctx).Error().Err(err).Str("slug", b.Slug).Msg("failed to delete mongo doc")
		}
	}

	stale, err := s.repo.FindStalePending(ctx, now.Add(-24*time.Hour))
	if err != nil {
		return fmt.Errorf("find stale pending: %w", err)
	}

	for _, b := range stale {
		log.Ctx(ctx).Info().Str("slug", b.Slug).Msg("cleaning stale pending batch")
		// best-effort delete R2 objects
		if err := s.store.ListAndDeleteBatchObjects(ctx, b.Slug); err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("slug", b.Slug).Msg("failed to delete R2 objects for stale batch")
		}
		if err := s.repo.DeleteByID(ctx, b.ID); err != nil {
			log.Ctx(ctx).Error().Err(err).Str("slug", b.Slug).Msg("failed to delete stale mongo doc")
		}
	}

	return nil
}

func computeExpiry(ttl string) *time.Time {
	switch ttl {
	case "24h":
		t := time.Now().Add(24 * time.Hour)
		return &t
	case "7d":
		t := time.Now().Add(7 * 24 * time.Hour)
		return &t
	case "30d":
		t := time.Now().Add(30 * 24 * time.Hour)
		return &t
	case "never":
		return nil
	default:
		t := time.Now().Add(7 * 24 * time.Hour)
		return &t
	}
}

func generateSlug() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 21)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[n.Int64()]
	}
	return string(b), nil
}

func generateOwnerToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func shortID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateBatchID() string {
	b := make([]byte, 12)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func extensionForContentType(contentType string) string {
	all := getAllSupportedTypes()
	return all[contentType]
}

func hashIP(ip string) string {
	// In production this should include a daily salt stored in env/config
	h := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(h[:])
}
