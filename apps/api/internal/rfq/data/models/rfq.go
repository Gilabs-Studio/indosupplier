package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RFQ struct {
	ID                     string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BuyerProfileID         string         `gorm:"type:uuid;not null;index" json:"buyer_profile_id"`
	Title                  string         `gorm:"type:varchar(255);not null;index" json:"title"`
	ProductDescription     string         `gorm:"type:text;not null" json:"product_description"`
	QuantityValue          float64        `gorm:"not null;default:0" json:"quantity_value"`
	QuantityUnit           string         `gorm:"type:varchar(60)" json:"quantity_unit"`
	DeliveryTimeline       string         `gorm:"type:varchar(160)" json:"delivery_timeline"`
	DestinationLocation    string         `gorm:"type:varchar(255)" json:"destination_location"`
	BudgetMin              float64        `gorm:"not null;default:0" json:"budget_min"`
	BudgetMax              float64        `gorm:"not null;default:0" json:"budget_max"`
	Specifications         string         `gorm:"type:text" json:"specifications"`
	PreferredContactMethod string         `gorm:"type:varchar(60)" json:"preferred_contact_method"`
	Mode                   string         `gorm:"type:varchar(40);not null;default:'specific';index" json:"mode"`
	CategoryID             string         `gorm:"type:uuid;index" json:"category_id"`
	VisibilityStatus       string         `gorm:"type:varchar(40);not null;default:'open';index" json:"visibility_status"`
	CreatedAt              time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	ClosedAt               *time.Time     `json:"closed_at"`
	DeletedAt              gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RFQ) TableName() string {
	return "rfqs"
}

func (r *RFQ) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if r.Mode == "" {
		r.Mode = "specific"
	}
	if r.VisibilityStatus == "" {
		r.VisibilityStatus = "open"
	}
	return nil
}

type RFQRecipient struct {
	ID                string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RFQID             string         `gorm:"type:uuid;not null;index:idx_rfq_supplier,unique" json:"rfq_id"`
	SupplierProfileID string         `gorm:"type:uuid;not null;index:idx_rfq_supplier,unique" json:"supplier_profile_id"`
	Status            string         `gorm:"type:varchar(40);not null;default:'new';index" json:"status"`
	InterestedAt      *time.Time     `json:"interested_at"`
	RespondedAt       *time.Time     `json:"responded_at"`
	DeclinedAt        *time.Time     `json:"declined_at"`
	RankPosition      int            `gorm:"not null;default:0" json:"rank_position"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RFQRecipient) TableName() string {
	return "rfq_recipients"
}

func (r *RFQRecipient) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if r.Status == "" {
		r.Status = "new"
	}
	return nil
}

type RFQMessage struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RFQID       string         `gorm:"type:uuid;not null;index" json:"rfq_id"`
	SenderType  string         `gorm:"type:varchar(40);not null;index" json:"sender_type"`
	SenderID    string         `gorm:"type:uuid;not null;index" json:"sender_id"`
	MessageType string         `gorm:"type:varchar(40);not null;default:'message';index" json:"message_type"`
	Body        string         `gorm:"type:text;not null" json:"body"`
	Metadata    string         `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RFQMessage) TableName() string {
	return "rfq_messages"
}

func (r *RFQMessage) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if r.MessageType == "" {
		r.MessageType = "message"
	}
	if r.Metadata == "" {
		r.Metadata = "{}"
	}
	return nil
}

type RFQAttachment struct {
	ID        string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RFQID     string         `gorm:"type:uuid;not null;index" json:"rfq_id"`
	MessageID string         `gorm:"type:uuid;index" json:"message_id"`
	FileURL   string         `gorm:"type:text;not null" json:"file_url"`
	FileName  string         `gorm:"type:varchar(255);not null" json:"file_name"`
	MimeType  string         `gorm:"type:varchar(120)" json:"mime_type"`
	FileSize  int64          `gorm:"not null;default:0" json:"file_size"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RFQAttachment) TableName() string {
	return "rfq_attachments"
}

func (r *RFQAttachment) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}
