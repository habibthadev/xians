package models

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type BatchRepository struct {
	db *gorm.DB
}

func NewBatchRepository(db *gorm.DB) *BatchRepository {
	return &BatchRepository{db: db}
}

func (r *BatchRepository) Insert(ctx context.Context, batch *Batch) error {
	return r.db.WithContext(ctx).Create(batch).Error
}

func (r *BatchRepository) FindBySlug(ctx context.Context, slug string) (*Batch, error) {
	var batch Batch
	err := r.db.WithContext(ctx).Preload("Files").Where("slug = ?", slug).First(&batch).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &batch, nil
}

func (r *BatchRepository) FindByID(ctx context.Context, id string) (*Batch, error) {
	var batch Batch
	err := r.db.WithContext(ctx).Preload("Files").Where("id = ?", id).First(&batch).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &batch, nil
}

func (r *BatchRepository) UpdateStatus(ctx context.Context, id string, status BatchStatus, expiresAt *time.Time, passwordHash *string, maxDownloads *int) error {
	updates := map[string]interface{}{
		"status": string(status),
	}
	if expiresAt != nil {
		updates["expires_at"] = expiresAt
	}
	if passwordHash != nil {
		updates["password_hash"] = *passwordHash
	}
	if maxDownloads != nil {
		updates["max_downloads"] = *maxDownloads
	}
	return r.db.WithContext(ctx).Model(&Batch{}).Where("id = ?", id).Updates(updates).Error
}

func (r *BatchRepository) AtomicIncrementDownload(ctx context.Context, slug string, maxDownloads int) (bool, error) {
	var result struct {
		Ok bool
	}
	err := r.db.WithContext(ctx).Raw(`
		UPDATE batches
		SET download_count = download_count + 1
		WHERE slug = ?
		  AND status = ?
		  AND (? = 0 OR download_count < ?)
		RETURNING TRUE AS ok
	`, slug, string(BatchStatusActive), maxDownloads, maxDownloads).Scan(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return result.Ok, nil
}

func (r *BatchRepository) FindExpiredActive(ctx context.Context, now time.Time) ([]Batch, error) {
	var batches []Batch
	err := r.db.WithContext(ctx).
		Where("status = ? AND expires_at < ?", string(BatchStatusActive), now).
		Find(&batches).Error
	return batches, err
}

func (r *BatchRepository) FindStalePending(ctx context.Context, cutoff time.Time) ([]Batch, error) {
	var batches []Batch
	err := r.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", string(BatchStatusPending), cutoff).
		Find(&batches).Error
	return batches, err
}

func (r *BatchRepository) DeleteByID(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&Batch{}, "id = ?", id).Error
}

// PruneRateLimits deletes rate-limit counters older than the cutoff.
func (r *BatchRepository) PruneRateLimits(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).Where("window_start < ?", before).Delete(&RateLimit{}).Error
}
