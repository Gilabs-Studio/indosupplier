package mapper

import (
	"fmt"
	"path/filepath"

	"github.com/gilabs/indosupplier/api/internal/rfq/data/models"
	"github.com/gilabs/indosupplier/api/internal/rfq/domain/dto"
)

func ResolveRFQID(id string) string {
	return id
}

// ToRFQResponse converts an RFQ model to its DTO response
func ToRFQResponse(rfq *models.RFQ, categoryName string, replies int, attachment *models.RFQAttachment) dto.RFQResponse {
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
		ID:          rfq.ID,
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
		}
	}

	return res
}
