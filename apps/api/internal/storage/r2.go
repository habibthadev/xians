package storage

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rs/zerolog/log"
)

type cacheEntry struct {
	url       string
	expiresAt time.Time
}

type R2Client struct {
	client      *s3.Client
	presign     *s3.PresignClient
	presignLong *s3.PresignClient
	bucket      string
	publicHost  string
	zipSem      chan struct{}
	getURLCache sync.Map
}

type ZipFileEntry struct {
	ID          string
	Name        string
	Ext         string
	ContentType string
	Size        int64
}

func NewR2Client(accountID, accessKeyID, secretAccessKey, bucket, publicHost string) (*R2Client, error) {
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID),
			HostnameImmutable: true,
			SigningRegion:     "auto",
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
	})
	presign := s3.NewPresignClient(client, func(o *s3.PresignOptions) {
		o.Expires = 15 * time.Minute
	})
	presignLong := s3.NewPresignClient(client, func(o *s3.PresignOptions) {
		o.Expires = 1 * time.Hour
	})

	return &R2Client{
		client:      client,
		presign:     presign,
		presignLong: presignLong,
		bucket:      bucket,
		publicHost:  publicHost,
		zipSem:      make(chan struct{}, 3),
	}, nil
}

func objectKey(batchID, fileID, ext string) string {
	return fmt.Sprintf("%s/%s%s", batchID, fileID, ext)
}

func (r *R2Client) PutObject(ctx context.Context, batchID, fileID, ext string, body io.Reader) error {
	key := objectKey(batchID, fileID, ext)
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &r.bucket,
		Key:    &key,
		Body:   body,
	})
	return err
}

func (r *R2Client) PresignPutURL(ctx context.Context, batchID, fileID, ext string) (string, error) {
	key := objectKey(batchID, fileID, ext)
	req, err := r.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: &r.bucket,
		Key:    &key,
	})
	if err != nil {
		return "", fmt.Errorf("presign put: %w", err)
	}
	return req.URL, nil
}

func (r *R2Client) InitMultipartUpload(ctx context.Context, batchID, fileID, ext string) (string, error) {
	key := objectKey(batchID, fileID, ext)
	result, err := r.client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket: &r.bucket,
		Key:    &key,
	})
	if err != nil {
		return "", fmt.Errorf("init multipart: %w", err)
	}
	if result.UploadId == nil {
		return "", fmt.Errorf("nil upload id")
	}
	return *result.UploadId, nil
}

func (r *R2Client) PresignUploadPartURL(ctx context.Context, batchID, fileID, ext, uploadID string, partNumber int) (string, error) {
	key := objectKey(batchID, fileID, ext)
	pn := int32(partNumber)
	req, err := r.presign.PresignUploadPart(ctx, &s3.UploadPartInput{
		Bucket:     &r.bucket,
		Key:        &key,
		UploadId:   &uploadID,
		PartNumber: &pn,
	})
	if err != nil {
		return "", fmt.Errorf("presign part: %w", err)
	}
	return req.URL, nil
}

func (r *R2Client) CompleteMultipartUpload(ctx context.Context, batchID, fileID, ext, uploadID string, parts []types.CompletedPart) error {
	key := objectKey(batchID, fileID, ext)
	_, err := r.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   &r.bucket,
		Key:      &key,
		UploadId: &uploadID,
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: parts,
		},
	})
	return err
}

func (r *R2Client) PresignGetURL(ctx context.Context, batchID, fileID, ext string) (string, error) {
	cacheKey := batchID + "/" + fileID + ext
	if cached, ok := r.getURLCache.Load(cacheKey); ok {
		entry := cached.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			return entry.url, nil
		}
		r.getURLCache.Delete(cacheKey)
	}

	key := objectKey(batchID, fileID, ext)
	req, err := r.presignLong.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &r.bucket,
		Key:    &key,
	})
	if err != nil {
		return "", fmt.Errorf("presign get: %w", err)
	}

	r.getURLCache.Store(cacheKey, cacheEntry{
		url:       req.URL,
		expiresAt: time.Now().Add(50 * time.Minute),
	})
	return req.URL, nil
}

func (r *R2Client) HeadObject(ctx context.Context, batchID, fileID, ext string) (bool, error) {
	key := objectKey(batchID, fileID, ext)
	_, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &r.bucket,
		Key:    &key,
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (r *R2Client) DeleteObjects(ctx context.Context, batchID string, files []string) error {
	for _, fileID := range files {
		key := objectKey(batchID, fileID, "")
		_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: &r.bucket,
			Key:    &key,
		})
		if err != nil {
			return fmt.Errorf("delete %s: %w", key, err)
		}
	}
	return nil
}

func (r *R2Client) DeleteObjectByKey(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &r.bucket,
		Key:    &key,
	})
	return err
}

func (r *R2Client) ListAndDeleteBatchObjects(ctx context.Context, batchID string) error {
	prefix := batchID + "/"
	paginator := s3.NewListObjectsV2Paginator(r.client, &s3.ListObjectsV2Input{
		Bucket: &r.bucket,
		Prefix: &prefix,
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("list objects: %w", err)
		}
		for _, obj := range page.Contents {
			if obj.Key == nil {
				continue
			}
			if err := r.DeleteObjectByKey(ctx, *obj.Key); err != nil {
				return fmt.Errorf("delete %s: %w", *obj.Key, err)
			}
		}
	}
	return nil
}

func (r *R2Client) StreamZip(ctx context.Context, batchID string, files []ZipFileEntry, w io.Writer) error {
	r.zipSem <- struct{}{}
	defer func() { <-r.zipSem }()

	zw := zip.NewWriter(w)
	defer zw.Close()

	for _, f := range files {
		key := objectKey(batchID, f.ID, f.Ext)
		result, err := r.client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: &r.bucket,
			Key:    &key,
		})
		if err != nil {
			return fmt.Errorf("get %s for zip: %w", key, err)
		}
		defer result.Body.Close()

		header := &zip.FileHeader{
			Name:   f.Name,
			Method: zip.Store,
		}
		header.SetMode(0644)

		wr, err := zw.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("zip create %s: %w", f.Name, err)
		}

		written, err := io.Copy(wr, result.Body)
		if err != nil {
			return fmt.Errorf("zip copy %s: %w", f.Name, err)
		}
		log.Ctx(ctx).Debug().Str("file", f.Name).Int64("bytes", written).Msg("zipped file")
	}
	return nil
}
