package mapper

import (
	"github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
)

func ToAssetResponse(asset *models.EmployeeAsset) dto.EmployeeAssetResponse {
	resp := dto.EmployeeAssetResponse{
		ID:              asset.ID,
		EmployeeID:      asset.EmployeeID,
		AssetName:       asset.AssetName,
		AssetCode:       asset.AssetCode,
		AssetCategory:   asset.AssetCategory,
		BorrowDate:      asset.BorrowDate.Format("2006-01-02"),
		BorrowCondition: string(asset.BorrowCondition),
		AssetImage:      asset.AssetImage,
		Notes:           asset.Notes,
		Status:          string(asset.GetStatus()),
		DaysBorrowed:    asset.DaysBorrowed(),
		CreatedAt:       &asset.CreatedAt,
		UpdatedAt:       &asset.UpdatedAt,
	}

	if asset.ReturnDate != nil {
		returnStr := asset.ReturnDate.Format("2006-01-02")
		resp.ReturnDate = &returnStr
	}

	if asset.ReturnCondition != nil {
		returnCond := string(*asset.ReturnCondition)
		resp.ReturnCondition = &returnCond
	}

	return resp
}

func ToAssetResponseList(assets []*models.EmployeeAsset) []dto.EmployeeAssetResponse {
	responses := make([]dto.EmployeeAssetResponse, len(assets))
	for i, asset := range assets {
		responses[i] = ToAssetResponse(asset)
	}
	return responses
}

func ToAssetBriefResponse(asset *models.EmployeeAsset) *dto.EmployeeAssetBriefResponse {
	if asset == nil {
		return nil
	}

	resp := &dto.EmployeeAssetBriefResponse{
		ID:            asset.ID,
		AssetName:     asset.AssetName,
		AssetCode:     asset.AssetCode,
		AssetCategory: asset.AssetCategory,
		BorrowDate:    asset.BorrowDate.Format("2006-01-02"),
		AssetImage:    asset.AssetImage,
		Status:        string(asset.GetStatus()),
		DaysBorrowed:  asset.DaysBorrowed(),
	}

	if asset.ReturnDate != nil {
		returnStr := asset.ReturnDate.Format("2006-01-02")
		resp.ReturnDate = &returnStr
	}

	return resp
}
