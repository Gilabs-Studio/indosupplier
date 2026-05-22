package mapper

import (
	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"github.com/google/uuid"
)

// ---- EvaluationGroup Mappers ----

// ToEvaluationGroupResponse converts EvaluationGroup model to response DTO
func ToEvaluationGroupResponse(group *models.EvaluationGroup) *dto.EvaluationGroupResponse {
	if group == nil {
		return nil
	}

	response := &dto.EvaluationGroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		IsActive:    group.IsActive,
		TotalWeight: group.TotalWeight(),
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}

	// Map criteria if preloaded
	if len(group.Criteria) > 0 {
		criteriaResponses := make([]dto.EvaluationCriteriaResponse, 0, len(group.Criteria))
		for i := range group.Criteria {
			criteriaResponses = append(criteriaResponses, *ToEvaluationCriteriaResponse(&group.Criteria[i]))
		}
		response.Criteria = criteriaResponses
	}

	return response
}

// ToEvaluationGroupResponseList converts slice of EvaluationGroup models to response DTOs
func ToEvaluationGroupResponseList(groups []models.EvaluationGroup) []*dto.EvaluationGroupResponse {
	responses := make([]*dto.EvaluationGroupResponse, 0, len(groups))
	for i := range groups {
		responses = append(responses, ToEvaluationGroupResponse(&groups[i]))
	}
	return responses
}

// ToEvaluationGroupModel converts CreateEvaluationGroupRequest to EvaluationGroup model
func ToEvaluationGroupModel(req *dto.CreateEvaluationGroupRequest, id string) *models.EvaluationGroup {
	group := &models.EvaluationGroup{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true, // Default to active
	}

	if req.IsActive != nil {
		group.IsActive = *req.IsActive
	}

	return group
}

// UpdateEvaluationGroupModel updates EvaluationGroup model from UpdateEvaluationGroupRequest
func UpdateEvaluationGroupModel(group *models.EvaluationGroup, req *dto.UpdateEvaluationGroupRequest) {
	if req.Name != nil {
		group.Name = *req.Name
	}
	if req.Description != nil {
		group.Description = req.Description
	}
	if req.IsActive != nil {
		group.IsActive = *req.IsActive
	}
}

// ---- EvaluationCriteria Mappers ----

// ToEvaluationCriteriaResponse converts EvaluationCriteria model to response DTO
func ToEvaluationCriteriaResponse(criteria *models.EvaluationCriteria) *dto.EvaluationCriteriaResponse {
	if criteria == nil {
		return nil
	}

	return &dto.EvaluationCriteriaResponse{
		ID:                criteria.ID,
		EvaluationGroupID: criteria.EvaluationGroupID,
		Name:              criteria.Name,
		Description:       criteria.Description,
		Weight:            criteria.Weight,
		MaxScore:          criteria.MaxScore,
		SortOrder:         criteria.SortOrder,
		CreatedAt:         criteria.CreatedAt,
		UpdatedAt:         criteria.UpdatedAt,
	}
}

// ToEvaluationCriteriaResponseList converts slice of EvaluationCriteria models to response DTOs
func ToEvaluationCriteriaResponseList(criteria []models.EvaluationCriteria) []*dto.EvaluationCriteriaResponse {
	responses := make([]*dto.EvaluationCriteriaResponse, 0, len(criteria))
	for i := range criteria {
		responses = append(responses, ToEvaluationCriteriaResponse(&criteria[i]))
	}
	return responses
}

// ToEvaluationCriteriaModel converts CreateEvaluationCriteriaRequest to EvaluationCriteria model
func ToEvaluationCriteriaModel(req *dto.CreateEvaluationCriteriaRequest, id string) *models.EvaluationCriteria {
	criteria := &models.EvaluationCriteria{
		ID:                id,
		EvaluationGroupID: req.EvaluationGroupID,
		Name:              req.Name,
		Description:       req.Description,
		Weight:            req.Weight,
		MaxScore:          100, // Default
		SortOrder:         req.SortOrder,
	}

	if req.MaxScore > 0 {
		criteria.MaxScore = req.MaxScore
	}

	return criteria
}

// UpdateEvaluationCriteriaModel updates EvaluationCriteria model from UpdateEvaluationCriteriaRequest
func UpdateEvaluationCriteriaModel(criteria *models.EvaluationCriteria, req *dto.UpdateEvaluationCriteriaRequest) {
	if req.Name != nil {
		criteria.Name = *req.Name
	}
	if req.Description != nil {
		criteria.Description = req.Description
	}
	if req.Weight != nil {
		criteria.Weight = *req.Weight
	}
	if req.MaxScore != nil {
		criteria.MaxScore = *req.MaxScore
	}
	if req.SortOrder != nil {
		criteria.SortOrder = *req.SortOrder
	}
}

// ---- EmployeeEvaluation Mappers ----

// ToEmployeeEvaluationResponse converts EmployeeEvaluation model to response DTO
func ToEmployeeEvaluationResponse(
	eval *models.EmployeeEvaluation,
	employeeMap map[string]dto.EmployeeSimpleResponse,
	criteriaMap map[string]string, // criteriaID → criteriaName
) *dto.EmployeeEvaluationResponse {
	if eval == nil {
		return nil
	}

	response := &dto.EmployeeEvaluationResponse{
		ID:                eval.ID,
		EmployeeID:        eval.EmployeeID,
		EvaluationGroupID: eval.EvaluationGroupID,
		EvaluatorID:       eval.EvaluatorID,
		EvaluationType:    string(eval.EvaluationType),
		PeriodStart:       eval.PeriodStart.Format("2006-01-02"),
		PeriodEnd:         eval.PeriodEnd.Format("2006-01-02"),
		OverallScore:      eval.OverallScore,
		Notes:             eval.Notes,
		CreatedAt:         eval.CreatedAt,
		UpdatedAt:         eval.UpdatedAt,
	}

	// Include employee data if available
	if emp, ok := employeeMap[eval.EmployeeID]; ok {
		response.Employee = &emp
	}

	// Include evaluator data if available
	if evaluator, ok := employeeMap[eval.EvaluatorID]; ok {
		response.Evaluator = &evaluator
	}

	// Include evaluation group if preloaded
	if eval.EvaluationGroup.ID != "" {
		response.EvaluationGroup = &dto.EvaluationGroupSimpleResponse{
			ID:   eval.EvaluationGroup.ID,
			Name: eval.EvaluationGroup.Name,
		}
	}

	// Map criteria scores if preloaded
	if len(eval.CriteriaScores) > 0 {
		scoreResponses := make([]dto.EvaluationCriteriaScoreResponse, 0, len(eval.CriteriaScores))
		for _, score := range eval.CriteriaScores {
			scoreResp := dto.EvaluationCriteriaScoreResponse{
				ID:                   score.ID,
				EvaluationCriteriaID: score.EvaluationCriteriaID,
				Score:                score.Score,
				Weight:               score.Weight,
				WeightedScore:        score.Score * score.Weight / 100,
				Notes:                score.Notes,
			}

			// Include criteria name if available
			if criteriaMap != nil {
				if name, ok := criteriaMap[score.EvaluationCriteriaID]; ok {
					scoreResp.CriteriaName = name
				}
			}

			scoreResponses = append(scoreResponses, scoreResp)
		}
		response.CriteriaScores = scoreResponses
	}

	return response
}

// ToEmployeeEvaluationResponseList converts slice of EmployeeEvaluation models to response DTOs
func ToEmployeeEvaluationResponseList(
	evaluations []models.EmployeeEvaluation,
	employeeMap map[string]dto.EmployeeSimpleResponse,
) []*dto.EmployeeEvaluationResponse {
	responses := make([]*dto.EmployeeEvaluationResponse, 0, len(evaluations))
	for i := range evaluations {
		responses = append(responses, ToEmployeeEvaluationResponse(&evaluations[i], employeeMap, nil))
	}
	return responses
}

// ToEmployeeEvaluationModel converts CreateEmployeeEvaluationRequest to EmployeeEvaluation model
func ToEmployeeEvaluationModel(req *dto.CreateEmployeeEvaluationRequest, id string, periodStart, periodEnd interface{ Format(string) string }) *models.EmployeeEvaluation {
	return nil // Will be constructed in usecase with parsed dates
}

// ToEvaluationCriteriaScoreModels converts score requests to models
func ToEvaluationCriteriaScoreModels(
	evaluationID string,
	scores []dto.CreateEvaluationCriteriaScoreRequest,
	criteriaWeightMap map[string]float64, // criteriaID → weight
) []models.EmployeeEvaluationCriteria {
	scoreModels := make([]models.EmployeeEvaluationCriteria, 0, len(scores))
	for _, s := range scores {
		weight := criteriaWeightMap[s.EvaluationCriteriaID]
		scoreModels = append(scoreModels, models.EmployeeEvaluationCriteria{
			ID:                   uuid.New().String(),
			EmployeeEvaluationID: evaluationID,
			EvaluationCriteriaID: s.EvaluationCriteriaID,
			Score:                s.Score,
			Weight:               weight, // Copy weight from criteria at eval time
			Notes:                s.Notes,
		})
	}
	return scoreModels
}
