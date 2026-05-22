package dto

import "time"

// EmployeeContractInput represents contract data when creating/updating an employee
type EmployeeContractInput struct {
	ContractNumber string `json:"contract_number" binding:"required,max=50"`
	ContractType   string `json:"contract_type" binding:"required,oneof=PKWTT PKWT Intern"`
	StartDate      string `json:"start_date" binding:"required"`
	EndDate        string `json:"end_date,omitempty"`
	DocumentPath   string `json:"document_path" binding:"max=255"`
}

// CreateEmployeeRequest represents the request to create an employee
type CreateEmployeeRequest struct {
	EmployeeCode     string     `json:"employee_code" binding:"omitempty,max=50"`
	Name             string     `json:"name" binding:"required,min=2,max=200"`
	Email            string     `json:"email" binding:"omitempty,email,max=100"`
	Phone            string     `json:"phone" binding:"max=20"`
	UserID           *string    `json:"user_id"`
	DivisionID       *string    `json:"division_id"`
	JobPositionID    *string    `json:"job_position_id"`
	CompanyID        *string    `json:"company_id"`
	DateOfBirth      *time.Time `json:"date_of_birth"`
	PlaceOfBirth     string     `json:"place_of_birth" binding:"max=100"`
	Gender           string     `json:"gender" binding:"omitempty,oneof=male female"`
	Religion         string     `json:"religion" binding:"max=50"`
	Address          string     `json:"address" binding:"max=500"`
	VillageID        *string    `json:"village_id"`
	NIK              string     `json:"nik" binding:"max=20"`
	NPWP             string     `json:"npwp" binding:"max=30"`
	BPJS             string     `json:"bpjs" binding:"max=30"`
	TotalLeaveQuota  int        `json:"total_leave_quota"`
	PTKPStatus       string     `json:"ptkp_status" binding:"max=10"`
	ReplacementForID *string    `json:"replacement_for_id"`
	// AreaIDs assigns the employee as a regular member of the specified areas.
	AreaIDs []string `json:"area_ids"`
	// SupervisedAreaIDs assigns the employee as a supervisor of the specified areas.
	SupervisedAreaIDs []string `json:"supervised_area_ids"`
	// OutletIDs assigns outlets to the employee. Associated warehouses are auto-selected.
	OutletIDs []string `json:"outlet_ids"`
	// WarehouseIDs assigns warehouses to the employee (optional, auto-populated from outlets).
	WarehouseIDs []string `json:"warehouse_ids"`
	IsActive     *bool    `json:"is_active"`
	// InitialContract creates a contract along with the employee (optional)
	InitialContract *EmployeeContractInput `json:"initial_contract,omitempty"`
}

// UpdateEmployeeRequest represents the request to update an employee
type UpdateEmployeeRequest struct {
	EmployeeCode     *string    `json:"employee_code" binding:"omitempty,max=50"`
	Name             *string    `json:"name" binding:"omitempty,min=2,max=200"`
	Email            *string    `json:"email" binding:"omitempty,email,max=100"`
	Phone            *string    `json:"phone" binding:"omitempty,max=20"`
	UserID           *string    `json:"user_id"`
	DivisionID       *string    `json:"division_id"`
	JobPositionID    *string    `json:"job_position_id"`
	CompanyID        *string    `json:"company_id"`
	DateOfBirth      *time.Time `json:"date_of_birth"`
	PlaceOfBirth     *string    `json:"place_of_birth" binding:"omitempty,max=100"`
	Gender           *string    `json:"gender" binding:"omitempty,oneof=male female"`
	Religion         *string    `json:"religion" binding:"omitempty,max=50"`
	Address          *string    `json:"address" binding:"omitempty,max=500"`
	VillageID        *string    `json:"village_id"`
	NIK              *string    `json:"nik" binding:"omitempty,max=20"`
	NPWP             *string    `json:"npwp" binding:"omitempty,max=30"`
	BPJS             *string    `json:"bpjs" binding:"omitempty,max=30"`
	TotalLeaveQuota  *int       `json:"total_leave_quota"`
	PTKPStatus       *string    `json:"ptkp_status" binding:"omitempty,max=10"`
	ReplacementForID *string    `json:"replacement_for_id"`
	// AreaIDs replaces all member area assignments when provided (nil = no change).
	AreaIDs []string `json:"area_ids"`
	// SupervisedAreaIDs replaces all supervisor area assignments when provided (nil = no change).
	SupervisedAreaIDs []string `json:"supervised_area_ids"`
	// OutletIDs replaces all outlet assignments when provided (nil = no change).
	// Associated warehouses are auto-selected from outlets.
	OutletIDs []string `json:"outlet_ids"`
	// WarehouseIDs replaces all warehouse assignments when provided (nil = no change).
	// Can include manual warehouse assignments beyond those auto-selected from outlets.
	WarehouseIDs []string `json:"warehouse_ids"`
	IsActive     *bool    `json:"is_active"`
}

// ApproveEmployeeRequest represents the request to approve/reject an employee
type ApproveEmployeeRequest struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
	Reason string `json:"reason"`
}

// AssignEmployeeAreasRequest represents the request to assign areas to an employee
type AssignEmployeeAreasRequest struct {
	AreaIDs []string `json:"area_ids" binding:"required"`
}

// AssignEmployeeSupervisorAreasRequest represents the request to set an employee's supervised areas
type AssignEmployeeSupervisorAreasRequest struct {
	AreaIDs []string `json:"area_ids" binding:"required"`
}

// AreaAssignment represents a single area assignment with role designation.
type AreaAssignment struct {
	AreaID       string `json:"area_id" binding:"required,uuid"`
	IsSupervisor bool   `json:"is_supervisor"`
}

// BulkUpdateEmployeeAreasRequest replaces all area assignments for an employee atomically.
type BulkUpdateEmployeeAreasRequest struct {
	Assignments []AreaAssignment `json:"assignments" binding:"required,dive"`
}

// EmployeeFormDataResponse provides all dropdown/select options needed by the employee form.
type EmployeeFormDataResponse struct {
	Divisions    []FormOption `json:"divisions"`
	JobPositions []FormOption `json:"job_positions"`
	Companies    []FormOption `json:"companies"`
	Areas        []FormOption `json:"areas"`
	Outlets      []FormOption `json:"outlets"`
	Warehouses   []FormOption `json:"warehouses"`
}

// FormOption is a generic id+label pair for select dropdowns.
type FormOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// EmployeeListParams represents the query parameters for listing employees
type EmployeeListParams struct {
	Page          int    `form:"page"`
	PerPage       int    `form:"per_page"`
	Search        string `form:"search"`
	DivisionID    string `form:"division_id"`
	JobPositionID string `form:"job_position_id"`
	AreaID        string `form:"area_id"`
	CompanyID     string `form:"company_id"`
	IsActive      *bool  `form:"is_active"`
	SortBy        string `form:"sort_by"`
	SortDir       string `form:"sort_dir"`
}

// EmployeeAreaSummary represents an area assignment for an employee, including the supervisor flag
type EmployeeAreaSummary struct {
	AreaID       string `json:"area_id"`
	AreaName     string `json:"area_name"`
	Description  string `json:"description"`
	IsActive     bool   `json:"is_active"`
	IsSupervisor bool   `json:"is_supervisor"`
}

// EmployeeResponse represents the employee response DTO
type EmployeeResponse struct {
	ID              string                `json:"id"`
	EmployeeCode    string                `json:"employee_code"`
	Name            string                `json:"name"`
	Email           string                `json:"email"`
	Phone           string                `json:"phone"`
	UserID          *string               `json:"user_id"`
	User            *UserBriefResponse    `json:"user,omitempty"`
	DivisionID      *string               `json:"division_id"`
	Division        *DivisionResponse     `json:"division,omitempty"`
	JobPositionID   *string               `json:"job_position_id"`
	JobPosition     *JobPositionResponse  `json:"job_position,omitempty"`
	CompanyID       *string               `json:"company_id"`
	Company         *CompanyBriefResponse `json:"company,omitempty"`
	DateOfBirth     *string               `json:"date_of_birth"`
	PlaceOfBirth    string                `json:"place_of_birth"`
	Gender          string                `json:"gender"`
	Religion        string                `json:"religion"`
	Address         string                `json:"address"`
	VillageID       *string               `json:"village_id"`
	Village         *VillageResponse      `json:"village,omitempty"`
	NIK             string                `json:"nik"`
	NPWP            string                `json:"npwp"`
	BPJS            string                `json:"bpjs"`
	TotalLeaveQuota int                   `json:"total_leave_quota"`
	PTKPStatus      string                `json:"ptkp_status"`
	IsDisability    bool                  `json:"is_disability"`
	// CurrentContract is the employee's active contract (if any)
	CurrentContract *EmployeeContractBriefResponse `json:"current_contract,omitempty"`
	// LatestEducation is the employee's ongoing or most recent education (if any)
	LatestEducation *EmployeeEducationBriefResponse `json:"latest_education,omitempty"`
	// LatestCertification is the employee's most recent valid certification (if any)
	LatestCertification *EmployeeCertificationBriefResponse `json:"latest_certification,omitempty"`
	// Signature is the employee's digital signature (if uploaded)
	Signature        *EmployeeSignatureResponse `json:"signature,omitempty"`
	ReplacementForID *string                    `json:"replacement_for_id"`
	ReplacementFor   *EmployeeBriefResponse     `json:"replacement_for,omitempty"`
	// Areas contains all assigned areas with their role (supervisor or member).
	Areas []EmployeeAreaSummary `json:"areas,omitempty"`
	// Outlets contains all assigned outlets.
	Outlets []EmployeeOutletResponse `json:"outlets,omitempty"`
	// Warehouses contains all assigned warehouses (manual and auto-selected).
	Warehouses []EmployeeWarehouseResponse `json:"warehouses,omitempty"`
	// IsAreaSupervisor is true when the employee supervises at least one area.
	IsAreaSupervisor bool    `json:"is_area_supervisor"`
	CreatedBy        *string `json:"created_by"`
	IsActive         bool    `json:"is_active"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// EmployeeSignatureResponse represents an employee's digital signature response
type EmployeeSignatureResponse struct {
	ID         string `json:"id"`
	FilePath   string `json:"file_path"`
	FileURL    string `json:"file_url"`
	FileName   string `json:"file_name"` // Original filename (e.g., "signature.jpg")
	FileSize   int64  `json:"file_size"`
	FileHash   string `json:"file_hash"`
	MimeType   string `json:"mime_type"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	UploadedBy string `json:"uploaded_by"`
	UploadedAt string `json:"uploaded_at"`
}

// UploadEmployeeSignatureRequest represents the request to upload a signature
type UploadEmployeeSignatureRequest struct {
	// File will be handled via multipart form, not JSON
}

// EmployeeBriefResponse represents a brief employee response (for nested refs)
type EmployeeBriefResponse struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	Name         string `json:"name"`
}

// EmployeeContractBriefResponse represents a brief employee contract response
type EmployeeContractBriefResponse struct {
	ID             string `json:"id"`
	ContractNumber string `json:"contract_number"`
	ContractType   string `json:"contract_type"`
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	DocumentPath   string `json:"document_path"`
	Status         string `json:"status"`
}

// UserBriefResponse represents a brief user response
type UserBriefResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CompanyBriefResponse represents a brief company response
type CompanyBriefResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// VillageResponse represents village with nested hierarchy
type VillageResponse struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	District *DistrictResponse `json:"district,omitempty"`
}

// DistrictResponse represents district with nested hierarchy
type DistrictResponse struct {
	ID   string        `json:"id"`
	Name string        `json:"name"`
	City *CityResponse `json:"city,omitempty"`
}

// CityResponse represents city with nested hierarchy
type CityResponse struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Province *ProvinceResponse `json:"province,omitempty"`
}

// ProvinceResponse represents province
type ProvinceResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// EmployeeListItemResponse represents a brief employee response for lists (no PII)
type EmployeeListItemResponse struct {
	ID              string                         `json:"id"`
	EmployeeCode    string                         `json:"employee_code"`
	Name            string                         `json:"name"`
	Email           string                         `json:"email"`
	Phone           string                         `json:"phone"`
	Division        *DivisionResponse              `json:"division,omitempty"`
	JobPosition     *JobPositionResponse           `json:"job_position,omitempty"`
	Company         *CompanyBriefResponse          `json:"company,omitempty"`
	CurrentContract *EmployeeContractBriefResponse `json:"current_contract,omitempty"`
	Status          string                         `json:"status"`
	IsActive        bool                           `json:"is_active"`
}

// OutletAssignment represents a single outlet assignment.
type OutletAssignment struct {
	OutletID string `json:"outlet_id" binding:"required,uuid"`
}

// WarehouseAssignment represents a single warehouse assignment.
type WarehouseAssignment struct {
	WarehouseID string `json:"warehouse_id" binding:"required,uuid"`
	IsAuto      bool   `json:"is_auto"` // true if auto-created from outlet assignment
}

// AssignEmployeeOutletsRequest represents the request to assign outlets to an employee.
// When outlets are assigned, their associated warehouses are auto-selected.
type AssignEmployeeOutletsRequest struct {
	OutletIDs []string `json:"outlet_ids" binding:"required,dive,uuid"`
}

// BulkUpdateEmployeeOutletsRequest replaces all outlet assignments for an employee atomically.
type BulkUpdateEmployeeOutletsRequest struct {
	Assignments []OutletAssignment `json:"assignments" binding:"required,dive"`
}

// AssignEmployeeWarehousesRequest represents the request to assign warehouses to an employee.
type AssignEmployeeWarehousesRequest struct {
	WarehouseIDs []string `json:"warehouse_ids" binding:"required,dive,uuid"`
}

// BulkUpdateEmployeeWarehousesRequest replaces all warehouse assignments for an employee atomically.
type BulkUpdateEmployeeWarehousesRequest struct {
	Assignments []WarehouseAssignment `json:"assignments" binding:"required,dive"`
}

// EmployeeOutletResponse represents an outlet assignment for an employee
type EmployeeOutletResponse struct {
	OutletID string `json:"outlet_id"`
	OutletName string `json:"outlet_name"`
	Code string `json:"code"`
	Address string `json:"address"`
}

// EmployeeWarehouseResponse represents a warehouse assignment for an employee
type EmployeeWarehouseResponse struct {
	WarehouseID string `json:"warehouse_id"`
	WarehouseName string `json:"warehouse_name"`
	Code string `json:"code"`
	Address string `json:"address"`
	IsAuto bool `json:"is_auto"` // true if auto-created from outlet assignment
}
