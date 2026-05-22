package mapper

import (
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
)

type AssetLocationMapper struct{}

func NewAssetLocationMapper() *AssetLocationMapper {
	return &AssetLocationMapper{}
}

func (m *AssetLocationMapper) ToResponse(item *financeModels.AssetLocation) dto.AssetLocationResponse {
	if item == nil {
		return dto.AssetLocationResponse{}
	}
	return dto.AssetLocationResponse{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Address:     item.Address,
		Latitude:    item.Latitude,
		Longitude:   item.Longitude,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}
