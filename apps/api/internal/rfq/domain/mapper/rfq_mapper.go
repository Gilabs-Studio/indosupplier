package mapper

import (
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/gilabs/indosupplier/api/internal/rfq/data/models"
	"github.com/gilabs/indosupplier/api/internal/rfq/domain/dto"
)

// ResolveRFQID converts potential mock IDs like "RFQ-2026-004" to their deterministic UUIDs
func ResolveRFQID(id string) string {
	if _, err := uuid.Parse(id); err == nil {
		return id
	}
	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte(id)).String()
}

// GetDisplayID converts deterministic mock UUIDs back to their alphanumeric display forms
func GetDisplayID(id string) string {
	for _, num := range []string{"001", "002", "003", "004"} {
		mockStr := fmt.Sprintf("RFQ-2026-%s", num)
		if id == uuid.NewSHA1(uuid.NameSpaceDNS, []byte(mockStr)).String() {
			return mockStr
		}
	}
	return id
}

// ToRFQResponse converts an RFQ model to its DTO response
func ToRFQResponse(rfq *models.RFQ, categoryName string, replies int, attachment *models.RFQAttachment) dto.RFQResponse {
	displayID := GetDisplayID(rfq.ID)
	
	// Format quantity
	var qtyStr string
	if rfq.QuantityValue == float64(int(rfq.QuantityValue)) {
		qtyStr = fmt.Sprintf("%d %s", int(rfq.QuantityValue), rfq.QuantityUnit)
	} else {
		qtyStr = fmt.Sprintf("%.2f %s", rfq.QuantityValue, rfq.QuantityUnit)
	}

	// Status mapping: default is Waiting for Quotes, or Offers Received, or Completed if closed
	status := "Waiting for Quotes"
	if replies > 0 {
		status = "Offers Received"
	}
	if rfq.ClosedAt != nil {
		status = "Completed"
	}

	res := dto.RFQResponse{
		ID:          displayID,
		Product:     rfq.Title,
		Category:    categoryName,
		Quantity:    qtyStr,
		TargetPort:  rfq.DestinationLocation,
		Date:        rfq.CreatedAt.Format("2006-01-02"),
		Status:      status,
		Replies:     replies,
		Description: rfq.ProductDescription,
	}

	if attachment != nil && attachment.FileURL != "" {
		res.AttachmentURL = attachment.FileURL
		res.AttachmentName = attachment.FileName
		if attachment.FileName == "" {
			res.AttachmentName = filepath.Base(attachment.FileURL)
		}
		
		// Human readable file size
		size := attachment.FileSize
		if size > 1024*1024 {
			res.AttachmentSize = fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
		} else if size > 1024 {
			res.AttachmentSize = fmt.Sprintf("%.1f KB", float64(size)/1024)
		} else if size > 0 {
			res.AttachmentSize = fmt.Sprintf("%d Bytes", size)
		} else {
			res.AttachmentSize = "1.4 MB" // default fallback
		}
	}

	return res
}
