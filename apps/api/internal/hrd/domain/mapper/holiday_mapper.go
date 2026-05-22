package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
)

// HolidayMapper handles mapping between Holiday model and DTOs
type HolidayMapper struct{}

// NewHolidayMapper creates a new HolidayMapper
func NewHolidayMapper() *HolidayMapper {
	return &HolidayMapper{}
}

// ToModel converts CreateHolidayRequest to Holiday model
func (m *HolidayMapper) ToModel(req *dto.CreateHolidayRequest) (*models.Holiday, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, err
	}

	h := &models.Holiday{
		Date:              date,
		Name:              req.Name,
		Description:       req.Description,
		Type:              models.HolidayType(req.Type),
		Year:              date.Year(),
		IsCollectiveLeave: req.IsCollectiveLeave,
		CutsAnnualLeave:   req.CutsAnnualLeave,
		IsRecurring:       req.IsRecurring,
		IsActive:          req.IsActive,
		CompanyID:         req.CompanyID,
	}
	return h, nil
}

// ApplyUpdate applies UpdateHolidayRequest to existing model
func (m *HolidayMapper) ApplyUpdate(h *models.Holiday, req *dto.UpdateHolidayRequest) error {
	if req.Date != nil {
		date, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			return err
		}
		h.Date = date
		h.Year = date.Year()
	}
	if req.Name != nil {
		h.Name = *req.Name
	}
	if req.Description != nil {
		h.Description = *req.Description
	}
	if req.Type != nil {
		h.Type = models.HolidayType(*req.Type)
	}
	if req.IsCollectiveLeave != nil {
		h.IsCollectiveLeave = *req.IsCollectiveLeave
	}
	if req.CutsAnnualLeave != nil {
		h.CutsAnnualLeave = *req.CutsAnnualLeave
	}
	if req.IsRecurring != nil {
		h.IsRecurring = *req.IsRecurring
	}
	if req.IsActive != nil {
		h.IsActive = *req.IsActive
	}
	if req.CompanyID != nil {
		h.CompanyID = req.CompanyID
	}
	return nil
}

// ToResponse converts Holiday model to response DTO
func (m *HolidayMapper) ToResponse(h *models.Holiday) *dto.HolidayResponse {
	resp := &dto.HolidayResponse{
		ID:                h.ID,
		Date:              h.Date.Format("2006-01-02"),
		Name:              h.Name,
		Description:       h.Description,
		Type:              string(h.Type),
		Year:              h.Year,
		IsCollectiveLeave: h.IsCollectiveLeave,
		CutsAnnualLeave:   h.CutsAnnualLeave,
		IsRecurring:       h.IsRecurring,
		IsActive:          h.IsActive,
		CompanyID:         h.CompanyID,
		CreatedAt:         h.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         h.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	return resp
}

// ToResponseList converts a list of Holiday models to response DTOs
func (m *HolidayMapper) ToResponseList(holidays []models.Holiday) []dto.HolidayResponse {
	responses := make([]dto.HolidayResponse, len(holidays))
	for i, h := range holidays {
		responses[i] = *m.ToResponse(&h)
	}
	return responses
}

// ToCalendarResponse groups holidays by month for calendar view
func (m *HolidayMapper) ToCalendarResponse(holidays []models.Holiday, year int) *dto.HolidayCalendarResponse {
	resp := &dto.HolidayCalendarResponse{
		Year:     year,
		Holidays: make(map[string][]dto.HolidayResponse),
	}

	for _, h := range holidays {
		monthKey := h.Date.Format("2006-01")
		resp.Holidays[monthKey] = append(resp.Holidays[monthKey], *m.ToResponse(&h))
	}

	return resp
}

// CSVRowToModel converts a CSV row to Holiday model
func (m *HolidayMapper) CSVRowToModel(row *dto.HolidayCSVRow, year int) (*models.Holiday, error) {
	date, err := time.Parse("2006-01-02", row.Date)
	if err != nil {
		// Try alternate format
		date, err = time.Parse("02/01/2006", row.Date)
		if err != nil {
			return nil, err
		}
	}

	holidayType := models.HolidayTypeNational
	if row.Type != "" {
		holidayType = models.HolidayType(row.Type)
	}

	h := &models.Holiday{
		Date:              date,
		Name:              row.Name,
		Description:       row.Description,
		Type:              holidayType,
		Year:              date.Year(),
		IsCollectiveLeave: row.IsCollectiveLeave == "true",
		CutsAnnualLeave:   row.CutsAnnualLeave == "true",
		IsActive:          true,
	}
	return h, nil
}
