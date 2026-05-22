package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"gorm.io/gorm"
)

type recruitmentRequestRepositoryImpl struct {
	db *gorm.DB
}

// NewRecruitmentRequestRepository creates a new instance of RecruitmentRequestRepository
func NewRecruitmentRequestRepository(db *gorm.DB) RecruitmentRequestRepository {
	return &recruitmentRequestRepositoryImpl{db: db}
}

func (r *recruitmentRequestRepositoryImpl) Create(ctx context.Context, req *models.RecruitmentRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := database.GetDB(ctx, r.db).Create(req).Error; err != nil {
		return fmt.Errorf("failed to create recruitment request: %w", err)
	}
	return nil
}

func (r *recruitmentRequestRepositoryImpl) FindByID(ctx context.Context, id string) (*models.RecruitmentRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var req models.RecruitmentRequest
	if err := database.GetDB(ctx, r.db).Where("id = ?", id).First(&req).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find recruitment request: %w", err)
	}
	return &req, nil
}

func (r *recruitmentRequestRepositoryImpl) Update(ctx context.Context, req *models.RecruitmentRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := database.GetDB(ctx, r.db).Save(req).Error; err != nil {
		return fmt.Errorf("failed to update recruitment request: %w", err)
	}
	return nil
}

func (r *recruitmentRequestRepositoryImpl) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result := database.GetDB(ctx, r.db).Where("id = ?", id).Delete(&models.RecruitmentRequest{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete recruitment request: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("recruitment request not found")
	}
	return nil
}

func (r *recruitmentRequestRepositoryImpl) FindAll(ctx context.Context, page, perPage int, search string, status *models.RecruitmentStatus, divisionID, positionID *string, priority *models.RecruitmentPriority) ([]models.RecruitmentRequest, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var requests []models.RecruitmentRequest
	var total int64

	// Use WithContext (not GetDB) because a conditional JOIN on employees is added when
	// searching — both tables have tenant_id and GetDB's unqualified WHERE causes PG ambiguity.
	query := r.db.WithContext(ctx).Model(&models.RecruitmentRequest{})

	// Apply filters
	if search != "" {
		searchPattern := "%" + search + "%"
		// WHY: Join employees table to search by requester name, and search request_code + job_description
		query = query.Joins("LEFT JOIN employees ON employees.id = recruitment_requests.requested_by_id").
			Where("recruitment_requests.request_code ILIKE ? OR recruitment_requests.job_description ILIKE ? OR employees.name ILIKE ?",
				searchPattern, searchPattern, searchPattern)
	}

	if status != nil {
		query = query.Where("recruitment_requests.status = ?", *status)
	}

	if divisionID != nil {
		query = query.Where("recruitment_requests.division_id = ?", *divisionID)
	}

	if positionID != nil {
		query = query.Where("recruitment_requests.position_id = ?", *positionID)
	}

	if priority != nil {
		query = query.Where("recruitment_requests.priority = ?", *priority)
	}

	// Count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count recruitment requests: %w", err)
	}

	// Enforce pagination limits
	if perPage > 100 {
		perPage = 100
	}
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * perPage
	query = query.Select("recruitment_requests.*").
		Offset(offset).Limit(perPage).
		Order("recruitment_requests.created_at DESC")

	if err := query.Find(&requests).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find recruitment requests: %w", err)
	}

	return requests, total, nil
}

// GenerateRequestCode generates a unique request code in format RR-YYYYMM-XXXX
func (r *recruitmentRequestRepositoryImpl) GenerateRequestCode(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	now := apptime.Now()
	prefix := fmt.Sprintf("RR-%s", now.Format("200601"))

	var count int64
	if err := database.GetDB(ctx, r.db).
		Model(&models.RecruitmentRequest{}).
		Where("request_code LIKE ?", prefix+"%").
		Count(&count).Error; err != nil {
		return "", fmt.Errorf("failed to generate request code: %w", err)
	}

	code := fmt.Sprintf("%s-%04d", prefix, count+1)
	return code, nil
}
