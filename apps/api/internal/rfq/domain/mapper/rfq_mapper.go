package mapper

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gilabs/indosupplier/api/internal/rfq/data/models"
	"github.com/gilabs/indosupplier/api/internal/rfq/domain/dto"
)

func ResolveRFQID(id string) string {
	return id
}

// ResolveRFQImageURL ensures every RFQ has a consistent, valid image URL
func ResolveRFQImageURL(imageURL, title, categoryName string) string {
	if imageURL != "" {
		return imageURL
	}
	lowerTitle := strings.ToLower(title)
	lowerCat := strings.ToLower(categoryName)
	if strings.Contains(lowerTitle, "steel") || strings.Contains(lowerTitle, "besi") || strings.Contains(lowerTitle, "baja") || strings.Contains(lowerCat, "steel") {
		return "/images/categories/cat-bahan-baku.webp"
	}
	if strings.Contains(lowerTitle, "mineral") || strings.Contains(lowerTitle, "powder") || strings.Contains(lowerTitle, "clay") || strings.Contains(lowerTitle, "bentonite") || strings.Contains(lowerTitle, "quartz") || strings.Contains(lowerTitle, "sand") || strings.Contains(lowerCat, "mineral") {
		return "/images/products/prod-mineral-powder.webp"
	}
	if strings.Contains(lowerTitle, "sugar") || strings.Contains(lowerTitle, "ginger") || strings.Contains(lowerTitle, "pepper") || strings.Contains(lowerCat, "agri") {
		return "/images/categories/cat-makanan-minuman.webp"
	}
	return "/images/categories/cat-bahan-baku.webp"
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
		ImageURL:    ResolveRFQImageURL(rfq.ImageURL, rfq.Title, categoryName),
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
