package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SupportTicket struct {
	ID            string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TicketNumber  string         `gorm:"type:varchar(80);not null;uniqueIndex" json:"ticket_number"`
	ReporterType  string         `gorm:"type:varchar(40);not null;index:idx_ticket_reporter" json:"reporter_type"`
	ReporterID    string         `gorm:"type:uuid;not null;index:idx_ticket_reporter" json:"reporter_id"`
	Category      string         `gorm:"type:varchar(80);not null;index" json:"category"`
	Subject       string         `gorm:"type:varchar(255);not null;index" json:"subject"`
	Description   string         `gorm:"type:text;not null" json:"description"`
	Priority      string         `gorm:"type:varchar(40);not null;default:'normal';index" json:"priority"`
	Status        string         `gorm:"type:varchar(40);not null;default:'open';index" json:"status"`
	AssignedTo    string         `gorm:"type:uuid;index" json:"assigned_to"`
	SLADeadlineAt *time.Time     `json:"sla_deadline_at"`
	ClosedAt      *time.Time     `json:"closed_at"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SupportTicket) TableName() string {
	return "support_tickets"
}

func (s *SupportTicket) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.Priority == "" {
		s.Priority = "normal"
	}
	if s.Status == "" {
		s.Status = "open"
	}
	return nil
}

type SupportTicketMessage struct {
	ID              string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupportTicketID string         `gorm:"type:uuid;not null;index" json:"support_ticket_id"`
	SenderType      string         `gorm:"type:varchar(40);not null;index" json:"sender_type"`
	SenderID        string         `gorm:"type:uuid;not null;index" json:"sender_id"`
	Body            string         `gorm:"type:text;not null" json:"body"`
	IsInternalNote  bool           `gorm:"not null;default:false;index" json:"is_internal_note"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SupportTicketMessage) TableName() string {
	return "support_ticket_messages"
}

func (s *SupportTicketMessage) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

type SupportTicketAttachment struct {
	ID              string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupportTicketID string         `gorm:"type:uuid;not null;index" json:"support_ticket_id"`
	MessageID       string         `gorm:"type:uuid;index" json:"message_id"`
	FileURL         string         `gorm:"type:text;not null" json:"file_url"`
	FileName        string         `gorm:"type:varchar(255);not null" json:"file_name"`
	MimeType        string         `gorm:"type:varchar(120)" json:"mime_type"`
	FileSize        int64          `gorm:"not null;default:0" json:"file_size"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SupportTicketAttachment) TableName() string {
	return "support_ticket_attachments"
}

func (s *SupportTicketAttachment) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

type FAQArticle struct {
	ID        string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Slug      string         `gorm:"type:varchar(180);not null;uniqueIndex" json:"slug"`
	Title     string         `gorm:"type:varchar(255);not null;index" json:"title"`
	Body      string         `gorm:"type:text;not null" json:"body"`
	Topic     string         `gorm:"type:varchar(120);index" json:"topic"`
	Status    string         `gorm:"type:varchar(40);not null;default:'draft';index" json:"status"`
	SortOrder int            `gorm:"not null;default:0;index" json:"sort_order"`
	CreatedBy string         `gorm:"type:uuid;index" json:"created_by"`
	UpdatedBy string         `gorm:"type:uuid;index" json:"updated_by"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FAQArticle) TableName() string {
	return "faq_articles"
}

func (f *FAQArticle) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	if f.Status == "" {
		f.Status = "draft"
	}
	return nil
}

type AbuseReport struct {
	ID           string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ReporterType string         `gorm:"type:varchar(40);not null;index:idx_abuse_reporter" json:"reporter_type"`
	ReporterID   string         `gorm:"type:uuid;not null;index:idx_abuse_reporter" json:"reporter_id"`
	ReportedType string         `gorm:"type:varchar(40);not null;index:idx_abuse_reported" json:"reported_type"`
	ReportedID   string         `gorm:"type:uuid;not null;index:idx_abuse_reported" json:"reported_id"`
	Reason       string         `gorm:"type:varchar(160);not null;index" json:"reason"`
	Description  string         `gorm:"type:text" json:"description"`
	Status       string         `gorm:"type:varchar(40);not null;default:'open';index" json:"status"`
	AssignedTo   string         `gorm:"type:uuid;index" json:"assigned_to"`
	Resolution   string         `gorm:"type:text" json:"resolution"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AbuseReport) TableName() string {
	return "abuse_reports"
}

func (a *AbuseReport) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	if a.Status == "" {
		a.Status = "open"
	}
	return nil
}
