package models

import (
	"time"

	geoModels "github.com/gilabs/gims/api/internal/geographic/data/models"
	userModels "github.com/gilabs/gims/api/internal/user/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// (EmployeeStatus removed entirely)

// Gender represents employee gender
type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

// PTKPStatus represents tax status (Penghasilan Tidak Kena Pajak)
type PTKPStatus string

const (
	PTKPTK0 PTKPStatus = "TK/0"  // Tidak Kawin, 0 tanggungan
	PTKPTK1 PTKPStatus = "TK/1"  // Tidak Kawin, 1 tanggungan
	PTKPTK2 PTKPStatus = "TK/2"  // Tidak Kawin, 2 tanggungan
	PTKPTK3 PTKPStatus = "TK/3"  // Tidak Kawin, 3 tanggungan
	PTKPK0  PTKPStatus = "K/0"   // Kawin, 0 tanggungan
	PTKPK1  PTKPStatus = "K/1"   // Kawin, 1 tanggungan
	PTKPK2  PTKPStatus = "K/2"   // Kawin, 2 tanggungan
	PTKPK3  PTKPStatus = "K/3"   // Kawin, 3 tanggungan
	PTKPKI0 PTKPStatus = "K/I/0" // Kawin, Penghasilan Istri Digabung, 0 tanggungan
	PTKPKI1 PTKPStatus = "K/I/1" // Kawin, Penghasilan Istri Digabung, 1 tanggungan
	PTKPKI2 PTKPStatus = "K/I/2" // Kawin, Penghasilan Istri Digabung, 2 tanggungan
	PTKPKI3 PTKPStatus = "K/I/3" // Kawin, Penghasilan Istri Digabung, 3 tanggungan
)

// Employee represents an employee entity with approval workflow
type Employee struct {
	ID           string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	EmployeeCode string `gorm:"type:varchar(50);index;not null" json:"employee_code"`
	Name         string `gorm:"type:varchar(200);not null;index" json:"name"`
	Email        string `gorm:"type:varchar(100);index" json:"email"`
	Phone        string `gorm:"type:varchar(20)" json:"phone"`

	// User account reference
	UserID *string          `gorm:"type:uuid;index" json:"user_id"`
	User   *userModels.User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Organization references
	DivisionID    *string      `gorm:"type:uuid;index" json:"division_id"`
	Division      *Division    `gorm:"foreignKey:DivisionID" json:"division,omitempty"`
	JobPositionID *string      `gorm:"type:uuid;index" json:"job_position_id"`
	JobPosition   *JobPosition `gorm:"foreignKey:JobPositionID" json:"job_position,omitempty"`
	CompanyID     *string      `gorm:"type:uuid;index" json:"company_id"`
	Company       *Company     `gorm:"foreignKey:CompanyID" json:"company,omitempty"`

	// Personal information
	DateOfBirth  *time.Time `gorm:"type:date" json:"date_of_birth"`
	PlaceOfBirth string     `gorm:"type:varchar(100)" json:"place_of_birth"`
	Gender       Gender     `gorm:"type:varchar(10)" json:"gender"`
	Religion     string     `gorm:"type:varchar(50)" json:"religion"`

	// Address (residence)
	Address   string             `gorm:"type:text" json:"address"`
	VillageID *string            `gorm:"type:uuid;index" json:"village_id"`
	Village   *geoModels.Village `gorm:"foreignKey:VillageID" json:"village,omitempty"`

	// Identity documents
	NIK  string `gorm:"type:varchar(20)" json:"nik"`  // Nomor Induk Kependudukan
	NPWP string `gorm:"type:varchar(30)" json:"npwp"` // Nomor Pokok Wajib Pajak
	BPJS string `gorm:"type:varchar(30)" json:"bpjs"` // BPJS number

	// Leave and benefits
	TotalLeaveQuota int        `gorm:"type:integer;not null;default:12" json:"total_leave_quota"`
	PTKPStatus      PTKPStatus `gorm:"type:varchar(20)" json:"ptkp_status"`

	// Replacement logic (contract takeover)
	ReplacementForID *string   `gorm:"type:uuid;index" json:"replacement_for_id"`
	ReplacementFor   *Employee `gorm:"foreignKey:ReplacementForID" json:"replacement_for,omitempty"`

	// Area assignments (M:N via EmployeeArea — includes both supervisor and member roles)
	Areas []EmployeeArea `gorm:"foreignKey:EmployeeID" json:"areas,omitempty"`

	// IsAreaSupervisor is a derived field (not stored in DB).
	// It is set to true when at least one EmployeeArea record has IsSupervisor=true.
	IsAreaSupervisor bool `gorm:"-" json:"is_area_supervisor"`

	// Outlet assignments (M:N via EmployeeOutlet)
	Outlets []EmployeeOutlet `gorm:"foreignKey:EmployeeID" json:"outlets,omitempty"`

	// Warehouse assignments (M:N via EmployeeWarehouse — includes auto-selected from outlets)
	Warehouses []EmployeeWarehouse `gorm:"foreignKey:EmployeeID" json:"warehouses,omitempty"`

	// Creator tracking
	CreatedBy  *string        `gorm:"type:uuid" json:"created_by"`

	// Standard fields
	IsActive  bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Employee
func (Employee) TableName() string {
	return "employees"
}

// BeforeCreate hook to generate UUID and employee code
func (e *Employee) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}
