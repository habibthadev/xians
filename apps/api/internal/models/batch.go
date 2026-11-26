package models

import (
	"time"
)

type BatchStatus string

const (
	BatchStatusActive   BatchStatus = "active"
	BatchStatusPending  BatchStatus = "pending"
	BatchStatusExpired  BatchStatus = "expired"
	BatchStatusDeleted  BatchStatus = "deleted"
	BatchStatusReported BatchStatus = "reported"
)

type Batch struct {
	ID             string    `gorm:"primaryKey;type:text"`
	Slug           string    `gorm:"uniqueIndex;type:text;not null"`
	Files          []File    `gorm:"foreignKey:BatchID;references:ID;constraint:OnDelete:CASCADE"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	ExpiresAt      *time.Time
	PasswordHash   *string `gorm:"type:text"`
	MaxDownloads   *int
	DownloadCount  int         `gorm:"default:0"`
	OwnerToken     string      `gorm:"type:text;not null"`
	UploaderIPHash string      `gorm:"type:text;not null"`
	Status         BatchStatus `gorm:"type:text;not null;default:pending"`
}

type File struct {
	ID           string `gorm:"type:text;not null"`
	BatchID      string `gorm:"type:text;not null;index"`
	OriginalName string `gorm:"type:text;not null"`
	ContentType  string `gorm:"type:text;not null"`
	SizeBytes    int64  `gorm:"not null"`
	Width        *int
	Height       *int
	DurationMs   *int
}

type FileManifest struct {
	Name        string `json:"name"`
	SizeBytes   int64  `json:"sizeBytes"`
	ContentType string `json:"contentType"`
}

type InitResponse struct {
	BatchID string          `json:"batchId"`
	Files   []FilePresigned `json:"files"`
}

type FilePresigned struct {
	FileID string `json:"fileId"`
	URL    string `json:"url"`
	Method string `json:"method"`
	Parts  []Part `json:"parts,omitempty"`
}

type Part struct {
	PartNumber int    `json:"partNumber"`
	URL        string `json:"url"`
}

type CompleteRequest struct {
	TTL          string  `json:"ttl"`
	Password     *string `json:"password,omitempty"`
	MaxDownloads *int    `json:"maxDownloads,omitempty"`
}

type CompleteResponse struct {
	Slug       string `json:"slug"`
	OwnerToken string `json:"ownerToken"`
}

type GalleryResponse struct {
	Slug             string     `json:"slug"`
	Files            []FileInfo `json:"files"`
	CreatedAt        string     `json:"createdAt"`
	ExpiresAt        *string    `json:"expiresAt"`
	RequiresPassword bool       `json:"requiresPassword"`
	DownloadCount    int        `json:"downloadCount"`
	MaxDownloads     *int       `json:"maxDownloads"`
	Status           string     `json:"status"`
}

type FileInfo struct {
	ID           string `json:"id"`
	OriginalName string `json:"originalName"`
	ContentType  string `json:"contentType"`
	SizeBytes    int64  `json:"sizeBytes"`
	Width        *int   `json:"width"`
	Height       *int   `json:"height"`
	DurationMs   *int   `json:"durationMs"`
}
