package mapper

import (
	"encoding/json"

	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
)

// ToActivityResponse converts an Activity model to ActivityResponse
func ToActivityResponse(activity *models.Activity) dto.ActivityResponse {
	resp := dto.ActivityResponse{
		ID:             activity.ID,
		Type:           activity.Type,
		ActivityTypeID: activity.ActivityTypeID,
		CustomerID:     activity.CustomerID,
		ContactID:      activity.ContactID,
		DealID:         activity.DealID,
		LeadID:         activity.LeadID,
		VisitReportID:  activity.VisitReportID,
		EmployeeID:     activity.EmployeeID,
		Description:    activity.Description,
		Timestamp:      activity.Timestamp.Format("2006-01-02T15:04:05+07:00"),
		CreatedAt:      activity.CreatedAt.Format("2006-01-02T15:04:05+07:00"),
	}

	// Inline metadata as a JSON object, not a double-encoded string
	if activity.Metadata != nil {
		resp.Metadata = json.RawMessage(*activity.Metadata)
	}

	if activity.ActivityType != nil {
		resp.ActivityType = &dto.ActivityTypeInfo{
			ID:         activity.ActivityType.ID,
			Name:       activity.ActivityType.Name,
			Code:       activity.ActivityType.Code,
			Icon:       activity.ActivityType.Icon,
			BadgeColor: activity.ActivityType.BadgeColor,
		}
	}

	if activity.Employee != nil {
		resp.Employee = &dto.ActivityEmployeeInfo{
			ID:           activity.Employee.ID,
			EmployeeCode: activity.Employee.EmployeeCode,
			Name:         activity.Employee.Name,
		}
	}

	return resp
}

// ToActivityResponseList converts a slice of Activity models to responses
func ToActivityResponseList(activities []models.Activity) []dto.ActivityResponse {
	result := make([]dto.ActivityResponse, 0, len(activities))
	for i := range activities {
		result = append(result, ToActivityResponse(&activities[i]))
	}
	return result
}
