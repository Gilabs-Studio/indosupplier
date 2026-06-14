package usecase

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/supplier/data/models"
	"github.com/gilabs/indosupplier/api/internal/supplier/domain/dto"
)

type DiscoveryUsecase interface {
	List(ctx context.Context, q string, categoryID string, region string, verifiedOnly bool) ([]dto.PublicSupplierDto, error)
	GetBySlug(ctx context.Context, slug string) (*dto.PublicSupplierDto, error)
}

type discoveryUsecase struct {
	db *gorm.DB
}

func NewDiscoveryUsecase(db *gorm.DB) DiscoveryUsecase {
	return &discoveryUsecase{
		db: db,
	}
}

func (u *discoveryUsecase) List(ctx context.Context, q string, categoryID string, region string, verifiedOnly bool) ([]dto.PublicSupplierDto, error) {
	db := u.db.WithContext(ctx).Model(&models.SupplierProfile{}).Where("status = ?", "active")

	if q != "" {
		db = db.Where("company_name ILIKE ? OR description ILIKE ?", "%"+q+"%", "%"+q+"%")
	}
	if region != "" {
		db = db.Where("city_id ILIKE ? OR province_id ILIKE ?", "%"+region+"%", "%"+region+"%")
	}
	if verifiedOnly {
		db = db.Where("verification_level >= ?", 2)
	}
	if categoryID != "" {
		db = db.Joins("JOIN supplier_categories ON supplier_categories.supplier_profile_id = supplier_profiles.id").
			Where("supplier_categories.category_id = ?", categoryID)
	}

	var profiles []models.SupplierProfile
	if err := db.Find(&profiles).Error; err != nil {
		return nil, err
	}

	var responses []dto.PublicSupplierDto
	for _, p := range profiles {
		// Fetch category names
		var catNames []string
		u.db.WithContext(ctx).
			Table("supplier_categories").
			Select("categories.name").
			Joins("join categories on categories.id = supplier_categories.category_id").
			Where("supplier_categories.supplier_profile_id = ?", p.ID).
			Pluck("name", &catNames)

		// Fetch key products
		var keyProds []string
		u.db.WithContext(ctx).
			Model(&models.SupplierProduct{}).
			Where("supplier_profile_id = ?", p.ID).
			Order("sort_order ASC").
			Limit(3).
			Pluck("name", &keyProds)

		// Fetch certifications
		var certNames []string
		u.db.WithContext(ctx).
			Table("supplier_certifications").
			Select("certifications.name").
			Joins("join certifications on certifications.id = supplier_certifications.certification_id").
			Where("supplier_certifications.supplier_profile_id = ? AND supplier_certifications.status = ?", p.ID, "approved").
			Pluck("name", &certNames)

		establishedYear, _ := strconv.Atoi(p.EstablishedYear)
		if establishedYear == 0 {
			establishedYear = 2015
		}

		responses = append(responses, dto.PublicSupplierDto{
			ID:              p.ID,
			Slug:            slugify(p.CompanyName),
			CompanyName:     p.CompanyName,
			BusinessType:    p.CompanyType,
			EstablishedYear: establishedYear,
			EmployeeCount:   p.EmployeesCount,
			Location:        p.CityID,
			Address:         p.Address,
			Description:     p.Description,
			IsVerified:      p.VerificationLevel >= 2,
			Rating:          p.StarRating,
			ReviewCount:     p.ReviewCount,
			KeyProducts:     keyProds,
			Certifications:  certNames,
		})
	}

	return responses, nil
}

func (u *discoveryUsecase) GetBySlug(ctx context.Context, slug string) (*dto.PublicSupplierDto, error) {
	var profiles []models.SupplierProfile
	if err := u.db.WithContext(ctx).Where("status = ?", "active").Find(&profiles).Error; err != nil {
		return nil, err
	}

	var matchedProfile *models.SupplierProfile
	for _, p := range profiles {
		if slugify(p.CompanyName) == slug {
			matchedProfile = &p
			break
		}
	}

	if matchedProfile == nil {
		return nil, errors.New("supplier profile not found")
	}

	// Fetch categories
	var catNames []string
	u.db.WithContext(ctx).
		Table("supplier_categories").
		Select("categories.name").
		Joins("join categories on categories.id = supplier_categories.category_id").
		Where("supplier_categories.supplier_profile_id = ?", matchedProfile.ID).
		Pluck("name", &catNames)

	// Fetch products
	var products []models.SupplierProduct
	if err := u.db.WithContext(ctx).Preload("Photos").Where("supplier_profile_id = ?", matchedProfile.ID).Order("sort_order ASC").Find(&products).Error; err != nil {
		return nil, err
	}

	var productDtos []dto.SupplierProductDto
	for _, pr := range products {
		var photos []string
		for _, ph := range pr.Photos {
			photos = append(photos, ph.FileURL)
		}
		productDtos = append(productDtos, dto.SupplierProductDto{
			ID:          pr.ID,
			Name:        pr.Name,
			Description: pr.Description,
			Price:       pr.StartingPrice,
			MinOrder:    pr.MOQ,
			Photos:      photos,
		})
	}

	// Fetch certifications
	type certInfo struct {
		ID          string
		Name        string
		IssuedBy    string
		IssuedAt    *time.Time
	}
	var certs []certInfo
	u.db.WithContext(ctx).
		Table("supplier_certifications").
		Select("supplier_certifications.id, certifications.name, supplier_certifications.issued_by, supplier_certifications.issued_at").
		Joins("join certifications on certifications.id = supplier_certifications.certification_id").
		Where("supplier_certifications.supplier_profile_id = ? AND supplier_certifications.status = ?", matchedProfile.ID, "approved").
		Scan(&certs)

	var certNames []string
	var certDtos []dto.SupplierCertificationDto
	for _, c := range certs {
		certNames = append(certNames, c.Name)

		year := 0
		if c.IssuedAt != nil {
			year = c.IssuedAt.Year()
		}
		certDtos = append(certDtos, dto.SupplierCertificationDto{
			ID:          c.ID,
			Name:        c.Name,
			Institution: c.IssuedBy,
			Year:        year,
		})
	}

	establishedYear, _ := strconv.Atoi(matchedProfile.EstablishedYear)
	if establishedYear == 0 {
		establishedYear = 2015
	}

	return &dto.PublicSupplierDto{
		ID:                matchedProfile.ID,
		Slug:              slugify(matchedProfile.CompanyName),
		CompanyName:       matchedProfile.CompanyName,
		BusinessType:      matchedProfile.CompanyType,
		EstablishedYear:   establishedYear,
		EmployeeCount:     matchedProfile.EmployeesCount,
		Location:          matchedProfile.CityID,
		Address:           matchedProfile.Address,
		Description:       matchedProfile.Description,
		IsVerified:        matchedProfile.VerificationLevel >= 2,
		Rating:            matchedProfile.StarRating,
		ReviewCount:       matchedProfile.ReviewCount,
		KeyProducts:       catNames,
		Certifications:    certNames,
		Products:          productDtos,
		CertificationList: certDtos,
		Phone:             matchedProfile.Phone,
		Email:             matchedProfile.Email,
		Website:           matchedProfile.Website,
	}, nil
}

func slugify(s string) string {
	s = strings.ToLower(s)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}
