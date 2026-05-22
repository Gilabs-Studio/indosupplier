package dto

import "time"

// ---- EvaluationGroup DTOs ----

// CreateEvaluationGroupRequest represents the request to create an evaluation group
type CreateEvaluationGroupRequest struct {
	Name        string  `json:"name" binding:"required,max=200"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
	IsActive    *bool   `json:"is_active" binding:"omitempty"`
}

// UpdateEvaluationGroupRequest represents the request to update an evaluation group
type UpdateEvaluationGroupRequest struct {
	Name        *string `json:"name" binding:"omitempty,max=200"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
	IsActive    *bool   `json:"is_active" binding:"omitempty"`
}

// EvaluationGroupResponse represents the evaluation group response
type EvaluationGroupResponse struct {
	ID          string                       `json:"id"`
	Name        string                       `json:"name"`
	Description *string                      `json:"description"`
	IsActive    bool                         `json:"is_active"`
	TotalWeight float64                      `json:"total_weight"`
	Criteria    []EvaluationCriteriaResponse `json:"criteria,omitempty"`
	CreatedAt   time.Time                    `json:"created_at"`
	UpdatedAt   time.Time                    `json:"updated_at"`
}

// EvaluationGroupSimpleResponse represents a minimal evaluation group for dropdowns
type EvaluationGroupSimpleResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ---- EvaluationCriteria DTOs ----

// CreateEvaluationCriteriaRequest represents the request to create an evaluation criteria
type CreateEvaluationCriteriaRequest struct {
	EvaluationGroupID string  `json:"evaluation_group_id" binding:"required,uuid"`
	Name              string  `json:"name" binding:"required,max=200"`
	Description       *string `json:"description" binding:"omitempty,max=1000"`
	Weight            float64 `json:"weight" binding:"required,gt=0,lte=100"`
	MaxScore          float64 `json:"max_score" binding:"omitempty,gt=0"`
	SortOrder         int     `json:"sort_order" binding:"omitempty,gte=0"`
}

// UpdateEvaluationCriteriaRequest represents the request to update an evaluation criteria
type UpdateEvaluationCriteriaRequest struct {
	Name        *string  `json:"name" binding:"omitempty,max=200"`
	Description *string  `json:"description" binding:"omitempty,max=1000"`
	Weight      *float64 `json:"weight" binding:"omitempty,gt=0,lte=100"`
	MaxScore    *float64 `json:"max_score" binding:"omitempty,gt=0"`
	SortOrder   *int     `json:"sort_order" binding:"omitempty,gte=0"`
}

// EvaluationCriteriaResponse represents the evaluation criteria response
type EvaluationCriteriaResponse struct {
	ID                string    `json:"id"`
	EvaluationGroupID string    `json:"evaluation_group_id"`
	Name              string    `json:"name"`
	Description       *string   `json:"description"`
	Weight            float64   `json:"weight"`
	MaxScore          float64   `json:"max_score"`
	SortOrder         int       `json:"sort_order"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ---- EmployeeEvaluation DTOs ----

// CreateEmployeeEvaluationRequest represents the request to create an employee evaluation
type CreateEmployeeEvaluationRequest struct {
	EmployeeID        string                                 `json:"employee_id" binding:"required,uuid"`
	EvaluationGroupID string                                 `json:"evaluation_group_id" binding:"required,uuid"`
	EvaluatorID       string                                 `json:"evaluator_id" binding:"required,uuid"`
	EvaluationType    string                                 `json:"evaluation_type" binding:"required,oneof=SELF MANAGER"`
	PeriodStart       string                                 `json:"period_start" binding:"required"` // YYYY-MM-DD
	PeriodEnd         string                                 `json:"period_end" binding:"required"`   // YYYY-MM-DD
	Notes             *string                                `json:"notes" binding:"omitempty,max=2000"`
	CriteriaScores    []CreateEvaluationCriteriaScoreRequest `json:"criteria_scores" binding:"omitempty,dive"`
}

// UpdateEmployeeEvaluationRequest represents the request to update an employee evaluation
type UpdateEmployeeEvaluationRequest struct {
	EvaluatorID    *string                                `json:"evaluator_id" binding:"omitempty,uuid"`
	EvaluationType *string                                `json:"evaluation_type" binding:"omitempty,oneof=SELF MANAGER"`
	PeriodStart    *string                                `json:"period_start" binding:"omitempty"` // YYYY-MM-DD
	PeriodEnd      *string                                `json:"period_end" binding:"omitempty"`   // YYYY-MM-DD
	Notes          *string                                `json:"notes" binding:"omitempty,max=2000"`
	CriteriaScores []CreateEvaluationCriteriaScoreRequest `json:"criteria_scores" binding:"omitempty,dive"`
}

// CreateEvaluationCriteriaScoreRequest represents a single criteria score input
type CreateEvaluationCriteriaScoreRequest struct {
	EvaluationCriteriaID string  `json:"evaluation_criteria_id" binding:"required,uuid"`
	Score                float64 `json:"score" binding:"gte=0"`
	Notes                *string `json:"notes" binding:"omitempty,max=500"`
}

// EmployeeEvaluationResponse represents the employee evaluation response
type EmployeeEvaluationResponse struct {
	ID                string                            `json:"id"`
	EmployeeID        string                            `json:"employee_id"`
	Employee          *EmployeeSimpleResponse           `json:"employee,omitempty"`
	EvaluationGroupID string                            `json:"evaluation_group_id"`
	EvaluationGroup   *EvaluationGroupSimpleResponse    `json:"evaluation_group,omitempty"`
	EvaluatorID       string                            `json:"evaluator_id"`
	Evaluator         *EmployeeSimpleResponse           `json:"evaluator,omitempty"`
	EvaluationType    string                            `json:"evaluation_type"`
	PeriodStart       string                            `json:"period_start"`
	PeriodEnd         string                            `json:"period_end"`
	OverallScore      float64                           `json:"overall_score"`
	Notes             *string                           `json:"notes"`
	CriteriaScores    []EvaluationCriteriaScoreResponse `json:"criteria_scores,omitempty"`
	CreatedAt         time.Time                         `json:"created_at"`
	UpdatedAt         time.Time                         `json:"updated_at"`
}

// EvaluationCriteriaScoreResponse represents a single criteria score in the response
type EvaluationCriteriaScoreResponse struct {
	ID                   string  `json:"id"`
	EvaluationCriteriaID string  `json:"evaluation_criteria_id"`
	CriteriaName         string  `json:"criteria_name,omitempty"`
	Score                float64 `json:"score"`
	Weight               float64 `json:"weight"`
	WeightedScore        float64 `json:"weighted_score"`
	Notes                *string `json:"notes"`
}

// EmployeeEvaluationFormDataResponse represents the form data for dropdowns
type EmployeeEvaluationFormDataResponse struct {
	Employees        []EmployeeFormOption            `json:"employees"`
	EvaluationGroups []EvaluationGroupSimpleResponse `json:"evaluation_groups"`
	EvaluationTypes  []EvaluationTypeOption          `json:"evaluation_types"`
}

// EvaluationTypeOption represents an evaluation type option for dropdowns
type EvaluationTypeOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// EvaluationAuditTrailUser represents actor info in audit trail rows.
type EvaluationAuditTrailUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// EvaluationAuditTrailEntry represents one audit trail entry for evaluation resources.
type EvaluationAuditTrailEntry struct {
	ID             string                    `json:"id"`
	Action         string                    `json:"action"`
	PermissionCode string                    `json:"permission_code"`
	TargetID       string                    `json:"target_id"`
	Metadata       map[string]interface{}    `json:"metadata"`
	User           *EvaluationAuditTrailUser `json:"user,omitempty"`
	CreatedAt      time.Time                 `json:"created_at"`
}
