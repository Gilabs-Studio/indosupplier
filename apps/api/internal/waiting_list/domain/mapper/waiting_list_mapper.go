package mapper

import (
	"time"

	"github.com/gilabs/indosupplier/api/internal/waiting_list/data/models"
	"github.com/gilabs/indosupplier/api/internal/waiting_list/domain/dto"
)

func ToWaitingListResponse(w *models.WaitingList) dto.WaitingListResponse {
	return dto.WaitingListResponse{
		ID:          w.ID,
		Email:       w.Email,
		Name:        w.Name,
		CompanyName: w.CompanyName,
		CompanyType: w.CompanyType,
		Phone:       w.Phone,
		Notes:       w.Notes,
		Status:      w.Status,
		CreatedAt:   w.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   w.UpdatedAt.Format(time.RFC3339),
	}
}

func ToWaitingListResponses(items []models.WaitingList) []dto.WaitingListResponse {
	res := make([]dto.WaitingListResponse, len(items))
	for i, item := range items {
		res[i] = ToWaitingListResponse(&item)
	}
	return res
}
