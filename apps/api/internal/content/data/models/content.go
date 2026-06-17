package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ContentArticle struct {
	ID                string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Type              string         `gorm:"type:varchar(40);not null;index:idx_content_type_status" json:"type"`
	Locale            string         `gorm:"type:varchar(10);not null;default:'id';index" json:"locale"`
	Title             string         `gorm:"type:varchar(255);not null;index" json:"title"`
	Slug              string         `gorm:"type:varchar(180);not null;uniqueIndex" json:"slug"`
	Excerpt           string         `gorm:"type:text" json:"excerpt"`
	Body              string         `gorm:"type:text" json:"body"`
	AuthorName        string         `gorm:"type:varchar(160)" json:"author_name"`
	ImageURL          string         `gorm:"type:text" json:"image_url"`
	VideoURL          string         `gorm:"type:text" json:"video_url"`
	Duration          string         `gorm:"type:varchar(40)" json:"duration"`
	ViewCount         int            `gorm:"not null;default:0" json:"view_count"`
	SupplierProfileID *string        `gorm:"type:uuid;index" json:"supplier_profile_id"`
	SupplierProductID *string        `gorm:"type:uuid;index" json:"supplier_product_id"`
	Status            string         `gorm:"type:varchar(40);not null;default:'draft';index:idx_content_type_status" json:"status"`
	IsFeatured        bool           `gorm:"not null;default:false;index" json:"is_featured"`
	SortOrder         int            `gorm:"not null;default:0;index" json:"sort_order"`
	PublishedAt       *time.Time     `gorm:"index" json:"published_at"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ContentArticle) TableName() string {
	return "content_articles"
}

func (c *ContentArticle) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	if c.Locale == "" {
		c.Locale = "id"
	}
	if c.Status == "" {
		c.Status = "draft"
	}
	return nil
}
