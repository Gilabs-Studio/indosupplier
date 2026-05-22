package models

import (
	"time"

	"gorm.io/gorm"
)

// EvaluationType represents who performed the evaluation
type EvaluationType string

const (
	EvaluationTypeSelf    EvaluationType = "SELF"
	EvaluationTypeManager EvaluationType = "MANAGER"
)

// EmployeeEvaluation represents a performance evaluation for an employee
type EmployeeEvaluation struct {
	ID                string         `gorm:"type:char(36);primaryKey" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	EmployeeID        string         `gorm:"type:char(36);not null;index" json:"employee_id"`
	EvaluationGroupID string         `gorm:"type:char(36);not null;index" json:"evaluation_group_id"`
	EvaluatorID       string         `gorm:"type:char(36);not null;index" json:"evaluator_id"` // the employee doing the evaluation
	EvaluationType    EvaluationType `gorm:"type:varchar(20);not null" json:"evaluation_type"`
	PeriodStart       time.Time      `gorm:"type:date;not null" json:"period_start"`
	PeriodEnd         time.Time      `gorm:"type:date;not null" json:"period_end"`
	OverallScore      float64        `gorm:"type:decimal(5,2);not null;default:0" json:"overall_score"` // computed: Σ(score × weight / 100)
	Notes             *string        `gorm:"type:text" json:"notes"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	EvaluationGroup EvaluationGroup              `gorm:"foreignKey:EvaluationGroupID;references:ID" json:"evaluation_group,omitempty"`
	CriteriaScores  []EmployeeEvaluationCriteria `gorm:"foreignKey:EmployeeEvaluationID;references:ID" json:"criteria_scores,omitempty"`
}

func (EmployeeEvaluation) TableName() string {
	return "employee_evaluations"
}

// CalculateOverallScore computes the weighted score: Σ(score × weight / 100)
// It both sets the OverallScore field and returns the value
func (ee *EmployeeEvaluation) CalculateOverallScore() float64 {
	var totalScore float64
	for _, cs := range ee.CriteriaScores {
		totalScore += cs.Score * cs.Weight / 100
	}
	ee.OverallScore = totalScore
	return totalScore
}

// EmployeeEvaluationCriteria represents a score for a single criterion in an evaluation
type EmployeeEvaluationCriteria struct {
	ID                   string         `gorm:"type:char(36);primaryKey" json:"id"`
	EmployeeEvaluationID string         `gorm:"type:char(36);not null;index" json:"employee_evaluation_id"`
	EvaluationCriteriaID string         `gorm:"type:char(36);not null;index" json:"evaluation_criteria_id"`
	Score                float64        `gorm:"type:decimal(5,2);not null;default:0" json:"score"`
	Weight               float64        `gorm:"type:decimal(5,2);not null" json:"weight"` // copied from criteria at evaluation time
	Notes                *string        `gorm:"type:text" json:"notes"`
	CreatedAt            time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (EmployeeEvaluationCriteria) TableName() string {
	return "employee_evaluation_criteria"
}
