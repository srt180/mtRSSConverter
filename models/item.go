package models

type Item struct {
	GUID   string `gorm:"primaryKey;type:varchar(36);not null" json:"guid" example:"123e4567-e89b-12d3-a456-426614174000"` // GUID
	RawURL string `gorm:"type:varchar(512);not null" json:"raw_url" example:"https://example.com/raw"`                     // 原始URL
}
