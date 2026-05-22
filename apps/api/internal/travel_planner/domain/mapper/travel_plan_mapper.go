package mapper

import (
	"sort"
	"time"

	"github.com/gilabs/gims/api/internal/travel_planner/data/models"
	"github.com/gilabs/gims/api/internal/travel_planner/domain/dto"
)

type TravelPlanMapper struct{}

func NewTravelPlanMapper() *TravelPlanMapper {
	return &TravelPlanMapper{}
}

func (m *TravelPlanMapper) ToResponse(plan *models.TravelPlan) dto.TravelPlanResponse {
	if plan == nil {
		return dto.TravelPlanResponse{}
	}

	days := make([]dto.TravelPlanDayResponse, 0, len(plan.Days))
	for _, day := range plan.Days {
		stops := make([]dto.TravelPlanStopResponse, 0, len(day.Stops))
		for _, stop := range day.Stops {
			stops = append(stops, dto.TravelPlanStopResponse{
				ID:         stop.ID,
				PlaceName:  stop.PlaceName,
				Latitude:   stop.Latitude,
				Longitude:  stop.Longitude,
				Category:   string(stop.Category),
				OrderIndex: stop.OrderIndex,
				IsLocked:   stop.IsLocked,
				Source:     string(stop.Source),
				PhotoURL:   stop.PhotoURL,
				Note:       stop.Note,
			})
		}

		sort.SliceStable(stops, func(i, j int) bool {
			return stops[i].OrderIndex < stops[j].OrderIndex
		})

		notes := make([]dto.TravelPlanDayNoteResponse, 0, len(day.Notes))
		for _, note := range day.Notes {
			notes = append(notes, dto.TravelPlanDayNoteResponse{
				ID:         note.ID,
				IconTag:    note.IconTag,
				NoteText:   note.NoteText,
				NoteTime:   note.NoteTime,
				OrderIndex: note.OrderIndex,
			})
		}

		sort.SliceStable(notes, func(i, j int) bool {
			return notes[i].OrderIndex < notes[j].OrderIndex
		})

		days = append(days, dto.TravelPlanDayResponse{
			ID:       day.ID,
			DayIndex: day.DayIndex,
			DayDate:  day.DayDate.Format("2006-01-02"),
			Summary:  day.Summary,
			Stops:    stops,
			Notes:    notes,
		})
	}

	sort.SliceStable(days, func(i, j int) bool {
		return days[i].DayIndex < days[j].DayIndex
	})

	return dto.TravelPlanResponse{
		ID:           plan.ID,
		Code:         plan.Code,
		Title:        plan.Title,
		PlanType:     string(plan.PlanType),
		Mode:         string(plan.Mode),
		StartDate:    plan.StartDate.Format("2006-01-02"),
		EndDate:      plan.EndDate.Format("2006-01-02"),
		Status:       string(plan.Status),
		BudgetAmount: plan.BudgetAmount,
		Notes:        plan.Notes,
		Days:         days,
		CreatedBy:    plan.CreatedBy,
		CreatedAt:    plan.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    plan.UpdatedAt.Format(time.RFC3339),
	}
}

func (m *TravelPlanMapper) ToResponseList(plans []models.TravelPlan) []dto.TravelPlanResponse {
	responses := make([]dto.TravelPlanResponse, 0, len(plans))
	for i := range plans {
		responses = append(responses, m.ToResponse(&plans[i]))
	}
	return responses
}
