package usecase

import (
	"context"
	"errors"
	"log"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/utils"
	"github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/organization/data/repositories"
	outletRepository "github.com/gilabs/gims/api/internal/organization/data/repositories"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
	"github.com/gilabs/gims/api/internal/organization/domain/mapper"
	"github.com/gilabs/gims/api/internal/organization/domain/service"
	warehouseRepository "github.com/gilabs/gims/api/internal/warehouse/data/repositories"
	"gorm.io/gorm"
)

var (
	ErrCompanyNotFound        = errors.New("company not found")
	ErrCompanyNotPending      = errors.New("company is not pending approval")
	ErrCompanyAlreadyApproved = errors.New("company is already approved")
	ErrInvalidApprovalAction  = errors.New("invalid approval action")
)

// CompanyUsecase defines the interface for company business logic
type CompanyUsecase interface {
	List(ctx context.Context, req *dto.ListCompaniesRequest) ([]dto.CompanyResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.CompanyResponse, error)
	Create(ctx context.Context, req *dto.CreateCompanyRequest, createdBy *string) (*dto.CompanyResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateCompanyRequest) (*dto.CompanyResponse, error)
	Delete(ctx context.Context, id string) error
	SubmitForApproval(ctx context.Context, id string) (*dto.CompanyResponse, error)
	Approve(ctx context.Context, id string, req *dto.ApproveCompanyRequest, approvedBy string) (*dto.CompanyResponse, error)
}

type companyUsecase struct {
	companyRepo     repositories.CompanyRepository
	outletRepo      outletRepository.OutletRepository
	warehouseRepo   warehouseRepository.WarehouseRepository
	timezoneService service.TimezoneService
}

// NewCompanyUsecase creates a new CompanyUsecase
func NewCompanyUsecase(
	companyRepo repositories.CompanyRepository,
	outletRepo outletRepository.OutletRepository,
	warehouseRepo warehouseRepository.WarehouseRepository,
	timezoneService service.TimezoneService,
) CompanyUsecase {
	return &companyUsecase{
		companyRepo:     companyRepo,
		outletRepo:      outletRepo,
		warehouseRepo:   warehouseRepo,
		timezoneService: timezoneService,
	}
}

func (u *companyUsecase) List(ctx context.Context, req *dto.ListCompaniesRequest) ([]dto.CompanyResponse, *utils.PaginationResult, error) {
	companies, total, err := u.companyRepo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := mapper.ToCompanyResponses(companies)

	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	pagination := &utils.PaginationResult{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	}

	return responses, pagination, nil
}

func (u *companyUsecase) GetByID(ctx context.Context, id string) (*dto.CompanyResponse, error) {
	company, err := u.companyRepo.FindByIDWithVillage(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCompanyNotFound
		}
		return nil, err
	}

	return mapper.ToCompanyResponse(company), nil
}

func (u *companyUsecase) Create(ctx context.Context, req *dto.CreateCompanyRequest, createdBy *string) (*dto.CompanyResponse, error) {
	company := mapper.CompanyFromCreateRequest(req, createdBy)

	// Auto-detect timezone from coordinates if not provided
	if company.Timezone == "" || company.Timezone == "Asia/Jakarta" {
		timezone, err := u.timezoneService.GetTimezoneForCompany(ctx, company.Latitude, company.Longitude, company.Timezone)
		if err != nil {
			log.Printf("warning: failed to auto-detect timezone for company: %v", err)
			// Fallback to Asia/Jakarta
			company.Timezone = "Asia/Jakarta"
		} else {
			company.Timezone = timezone
			log.Printf("auto-detected timezone for company %s: %s", company.Name, timezone)
		}
	}

	if err := u.companyRepo.Create(ctx, company); err != nil {
		return nil, err
	}

	// Reload with village
	company, err := u.companyRepo.FindByIDWithVillage(ctx, company.ID)
	if err != nil {
		return nil, err
	}

	return mapper.ToCompanyResponse(company), nil
}

func (u *companyUsecase) Update(ctx context.Context, id string, req *dto.UpdateCompanyRequest) (*dto.CompanyResponse, error) {
	company, err := u.companyRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCompanyNotFound
		}
		return nil, err
	}

	wasActive := company.IsActive

	if req.Name != "" {
		company.Name = req.Name
	}
	if req.Address != "" {
		company.Address = req.Address
	}
	if req.Email != "" {
		company.Email = req.Email
	}
	if req.Phone != "" {
		company.Phone = req.Phone
	}
	if req.NPWP != "" {
		company.NPWP = req.NPWP
	}
	if req.NIB != "" {
		company.NIB = req.NIB
	}
	if req.ProvinceID != nil {
		company.ProvinceID = req.ProvinceID
	}
	if req.CityID != nil {
		company.CityID = req.CityID
	}
	if req.DistrictID != nil {
		company.DistrictID = req.DistrictID
	}
	if req.VillageID != nil {
		company.VillageID = req.VillageID
	}
	if req.VillageName != nil {
		company.VillageName = req.VillageName
	}
	if req.Timezone != "" {
		company.Timezone = req.Timezone
	}
	if req.IsActive != nil {
		company.IsActive = *req.IsActive
	}
	if req.Latitude != nil {
		company.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		company.Longitude = req.Longitude
	}

	// Auto-detect timezone if coordinates changed and timezone not explicitly set
	if (req.Latitude != nil || req.Longitude != nil) && req.Timezone == "" {
		timezone, err := u.timezoneService.GetTimezoneForCompany(ctx, company.Latitude, company.Longitude, company.Timezone)
		if err != nil {
			log.Printf("warning: failed to auto-detect timezone for company update: %v", err)
		} else if timezone != company.Timezone {
			company.Timezone = timezone
			log.Printf("auto-updated timezone for company %s: %s", company.Name, timezone)
		}
	}

	// Cascade deactivation: if company is being deactivated, also deactivate all related outlets and warehouses
	if req.IsActive != nil && !*req.IsActive && wasActive {
		// Company is being deactivated, cascade to outlets and warehouses
		if err := u.outletRepo.UpdateIsActiveByCompanyID(ctx, company.ID, false); err != nil {
			log.Printf("error cascading deactivation to outlets for company %s: %v", company.ID, err)
			return nil, err
		}

		// Find all outlets for this company to get their warehouse connections
		outlets, err := u.outletRepo.FindByCompanyID(ctx, company.ID)
		if err != nil {
			log.Printf("error finding outlets for cascade deactivation: %v", err)
			return nil, err
		}

		if len(outlets) > 0 {
			outletIDs := make([]string, 0, len(outlets))
			for _, outlet := range outlets {
				outletIDs = append(outletIDs, outlet.ID)
			}

			// Deactivate all warehouses linked to these outlets
			if err := u.warehouseRepo.UpdateIsActiveByOutletIDs(ctx, outletIDs, false); err != nil {
				log.Printf("error cascading deactivation to warehouses: %v", err)
				return nil, err
			}
		}
	}

	if err := u.companyRepo.Update(ctx, company); err != nil {
		return nil, err
	}

	// Reload with village
	company, err = u.companyRepo.FindByIDWithVillage(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapper.ToCompanyResponse(company), nil
}

func (u *companyUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.companyRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCompanyNotFound
		}
		return err
	}

	return u.companyRepo.Delete(ctx, id)
}

func (u *companyUsecase) SubmitForApproval(ctx context.Context, id string) (*dto.CompanyResponse, error) {
	company, err := u.companyRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCompanyNotFound
		}
		return nil, err
	}

	if company.IsApproved {
		return nil, ErrCompanyAlreadyApproved
	}

	company.Status = models.CompanyStatusPending
	if err := u.companyRepo.Update(ctx, company); err != nil {
		return nil, err
	}
	actorUserID, _ := ctx.Value("user_id").(string)
	_ = actorUserID
	// Approval workflow disabled for single-company tenant model
	// if err := notificationService.CreateApprovalNotification(ctx, database.DB, notificationService.ApprovalNotificationParams{
	// 	PermissionCode: "company.approve",
	// 	EntityType:     "company",
	// 	EntityID:       company.ID,
	// 	Title:          "Company Approval",
	// 	Message:        "A company record has been submitted and requires your approval.",
	// 	ActorUserID:    actorUserID,
	// }); err != nil {
	// 	log.Printf("warning: failed to create company notification: %v", err)
	// }

	// Reload with village
	company, err = u.companyRepo.FindByIDWithVillage(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapper.ToCompanyResponse(company), nil
}

func (u *companyUsecase) Approve(ctx context.Context, id string, req *dto.ApproveCompanyRequest, approvedBy string) (*dto.CompanyResponse, error) {
	company, err := u.companyRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCompanyNotFound
		}
		return nil, err
	}

	if company.Status != models.CompanyStatusPending {
		return nil, ErrCompanyNotPending
	}

	now := apptime.Now()
	company.ApprovedBy = &approvedBy
	company.ApprovedAt = &now

	switch req.Action {
	case "approve":
		company.Status = models.CompanyStatusApproved
		company.IsApproved = true
	case "reject":
		company.Status = models.CompanyStatusRejected
		company.IsApproved = false
	default:
		return nil, ErrInvalidApprovalAction
	}

	if err := u.companyRepo.Update(ctx, company); err != nil {
		return nil, err
	}

	// Reload with village
	company, err = u.companyRepo.FindByIDWithVillage(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapper.ToCompanyResponse(company), nil
}
