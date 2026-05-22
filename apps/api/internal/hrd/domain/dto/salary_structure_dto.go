package dto

import "time"

type CreateSalaryStructureRequest struct {
	EmployeeID    string  `json:"employee_id" binding:"required,uuid"`
	EffectiveDate string  `json:"effective_date" binding:"required"`
	BasicSalary   float64 `json:"basic_salary" binding:"required,gt=0"`
	Notes         string  `json:"notes"`
}

type UpdateSalaryStructureRequest struct {
	EmployeeID    string  `json:"employee_id" binding:"required,uuid"`
	EffectiveDate string  `json:"effective_date" binding:"required"`
	BasicSalary   float64 `json:"basic_salary" binding:"required,gt=0"`
	Notes         string  `json:"notes"`
}

type ListSalaryStructuresRequest struct {
	Page       int     `form:"page" binding:"omitempty,min=1"`
	PerPage    int     `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search     string  `form:"search"`
	EmployeeID *string `form:"employee_id" binding:"omitempty,uuid"`
	Status     *string `form:"status"`
	SortBy     string  `form:"sort_by"`
	SortDir    string  `form:"sort_dir"`
}

// EmployeeInfo contains minimal employee data for salary response
type EmployeeInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	EmployeeCode string `json:"employee_code"`
	Email        string `json:"email"`
	AvatarURL    string `json:"avatar_url"`
}

type SalaryStructureResponse struct {
	ID            string       `json:"id"`
	EmployeeID    string       `json:"employee_id"`
	Employee      EmployeeInfo `json:"employee"`
	EffectiveDate time.Time    `json:"effective_date"`
	BasicSalary   float64      `json:"basic_salary"`
	Notes         string       `json:"notes"`
	Status        string       `json:"status"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// SalaryEmployeeGroup groups all salary records by employee
type SalaryEmployeeGroup struct {
	EmployeeID  string                    `json:"employee_id"`
	Employee    EmployeeInfo              `json:"employee"`
	SalaryCount int                       `json:"salary_count"`
	Salaries    []SalaryStructureResponse `json:"salaries"`
}

// SalaryStructureStatsResponse contains aggregate stats
type SalaryStructureStatsResponse struct {
	Total               int64                                `json:"total"`
	Active              int64                                `json:"active"`
	Draft               int64                                `json:"draft"`
	Inactive            int64                                `json:"inactive"`
	AverageSalary       float64                              `json:"average_salary"`
	MinSalary           float64                              `json:"min_salary"`
	MaxSalary           float64                              `json:"max_salary"`
	TotalSalaryOverTime []SalaryStructureTotalSalaryOverTime `json:"total_salary_over_time"`
}

type SalaryStructureTotalSalaryOverTime struct {
	Period      string  `json:"period"`
	TotalSalary float64 `json:"total_salary"`
}

// SalaryEmployeeFormOption is a minimal employee item for salary form selection.
type SalaryEmployeeFormOption struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	Name         string `json:"name"`
}

// SalaryFormDataResponse is the response for the /form-data endpoint
type SalaryFormDataResponse struct {
	Employees []SalaryEmployeeFormOption `json:"employees"`
}
