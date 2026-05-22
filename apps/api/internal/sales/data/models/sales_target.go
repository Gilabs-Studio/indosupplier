package models

import (
	"time"

	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SalesTarget represents a sales target assigned to a specific employee (sales rep)
type SalesTarget struct {
	ID          string  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    string  `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	EmployeeID  string  `gorm:"type:uuid;not null;index" json:"employee_id"`
	Year        int     `gorm:"not null;index" json:"year"`
	TotalTarget float64 `gorm:"type:decimal(20,2);not null" json:"total_target"`
	Notes       string  `gorm:"type:text" json:"notes"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Employee       *orgModels.Employee    `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	MonthlyTargets []MonthlySalesTarget   `gorm:"foreignKey:SalesTargetID;constraint:OnDelete:CASCADE" json:"monthly_targets,omitempty"`
}

// TableName specifies the table name for SalesTarget
func (SalesTarget) TableName() string {
	return "sales_targets"
}

// BeforeCreate hook to generate UUID
func (st *SalesTarget) BeforeCreate(tx *gorm.DB) error {
	if st.ID == "" {
		st.ID = uuid.New().String()
	}
	return nil
}

// CalculateAchievements calculates total actual and achievement percent
func (st *SalesTarget) CalculateAchievements() (totalActual float64, achievementPercent float64) {
	for i := range st.MonthlyTargets {
		mt := &st.MonthlyTargets[i]
		mt.CalculateSalesAchievement()
		totalActual += mt.ActualAmount
	}

	if st.TotalTarget > 0 {
		achievementPercent = (totalActual / st.TotalTarget) * 100
	}

	return totalActual, achievementPercent
}

// MonthlySalesTarget represents monthly breakdown of a sales target per employee
type MonthlySalesTarget struct {
	ID               string  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID         string  `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	SalesTargetID    string  `gorm:"type:uuid;not null;index" json:"sales_target_id"`
	Month            int     `gorm:"not null" json:"month"` // 1-12
	TargetAmount     float64 `gorm:"type:decimal(20,2);not null" json:"target_amount"`
	ActualAmount     float64 `gorm:"type:decimal(20,2);default:0" json:"actual_amount"`
	AchievementPercent float64 `gorm:"type:decimal(5,2);default:0" json:"achievement_percent"`
	Notes            string  `gorm:"type:text" json:"notes"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	SalesTarget *SalesTarget `gorm:"foreignKey:SalesTargetID" json:"sales_target,omitempty"`
}

// TableName specifies the table name for MonthlySalesTarget
func (MonthlySalesTarget) TableName() string {
	return "monthly_sales_targets"
}

// BeforeCreate hook to generate UUID
func (mst *MonthlySalesTarget) BeforeCreate(tx *gorm.DB) error {
	if mst.ID == "" {
		mst.ID = uuid.New().String()
	}
	return nil
}

// CalculateSalesAchievement calculates the achievement percentage
func (mst *MonthlySalesTarget) CalculateSalesAchievement() {
	if mst.TargetAmount > 0 {
		mst.AchievementPercent = (mst.ActualAmount / mst.TargetAmount) * 100
	} else {
		mst.AchievementPercent = 0
	}
}
