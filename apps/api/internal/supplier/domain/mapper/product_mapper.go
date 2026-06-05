package mapper

import (
	"github.com/gilabs/indosupplier/api/internal/supplier/data/models"
	"github.com/gilabs/indosupplier/api/internal/supplier/domain/dto"
)

func ToCategoryResponse(c *models.Category) *dto.CategoryResponse {
	if c == nil {
		return nil
	}
	return &dto.CategoryResponse{
		ID:          c.ID,
		ParentID:    c.ParentID,
		Slug:        c.Slug,
		Name:        c.Name,
		Description: c.Description,
	}
}

func ToPhotoResponse(p *models.SupplierProductPhoto) dto.PhotoResponse {
	return dto.PhotoResponse{
		ID:        p.ID,
		FileURL:   p.FileURL,
		Caption:   p.Caption,
		SortOrder: p.SortOrder,
		CreatedAt: p.CreatedAt,
	}
}

func ToProductResponse(p *models.SupplierProduct) dto.ProductResponse {
	photos := make([]dto.PhotoResponse, len(p.Photos))
	for i, photo := range p.Photos {
		photos[i] = ToPhotoResponse(&photo)
	}

	resp := dto.ProductResponse{
		ID:                p.ID,
		SupplierProfileID: p.SupplierProfileID,
		CategoryID:        p.CategoryID,
		Category:          ToCategoryResponse(p.Category),
		Name:              p.Name,
		Description:       p.Description,
		MOQ:               p.MOQ,
		StartingPrice:     p.StartingPrice,
		Currency:          p.Currency,
		CapacityText:      p.CapacityText,
		IsFeatured:        p.IsFeatured,
		SortOrder:         p.SortOrder,
		Photos:            photos,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
	return resp
}

func ToProductListResponse(products []models.SupplierProduct) []dto.ProductResponse {
	resp := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		resp[i] = ToProductResponse(&p)
	}
	return resp
}
