package usecase

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"

	"github.com/gilabs/indosupplier/api/internal/core/utils"
	"github.com/gilabs/indosupplier/api/internal/supplier/data/models"
	"github.com/gilabs/indosupplier/api/internal/supplier/domain/dto"
)

type DiscoveryUsecase interface {
	List(ctx context.Context, q string, categoryID string, region string, verifiedOnly bool) ([]dto.PublicSupplierDto, error)
	GetBySlug(ctx context.Context, slug string) (*dto.PublicSupplierDto, error)
	ListProducts(ctx context.Context, params dto.ListPublicProductsParams) ([]dto.PublicProductDto, error)
	GetProductByID(ctx context.Context, id string) (*dto.PublicProductDetailDto, error)
	LookupSuppliers(ctx context.Context, q string, page int, limit int) ([]dto.PublicSupplierDto, error)
	LookupProducts(ctx context.Context, q string, page int, limit int) ([]dto.PublicProductDto, error)
}

type discoveryUsecase struct {
	db *gorm.DB
}

func NewDiscoveryUsecase(db *gorm.DB) DiscoveryUsecase {
	return &discoveryUsecase{db: db}
}

func (u *discoveryUsecase) List(ctx context.Context, q string, categoryID string, region string, verifiedOnly bool) ([]dto.PublicSupplierDto, error) {
	db := u.db.WithContext(ctx).Model(&models.SupplierProfile{}).Where("supplier_profiles.status = ?", "active")

	if q != "" {
		clauses := make([]string, 0)
		args := make([]interface{}, 0)
		for _, term := range expandedSearchTerms(q) {
			like := "%" + term + "%"
			clauses = append(clauses, "(supplier_profiles.company_name ILIKE ? OR supplier_profiles.description ILIKE ? OR EXISTS (SELECT 1 FROM supplier_products sp WHERE sp.supplier_profile_id = supplier_profiles.id AND (sp.name ILIKE ? OR sp.description ILIKE ?) AND sp.deleted_at IS NULL))")
			args = append(args, like, like, like, like)
		}
		db = db.Where(strings.Join(clauses, " OR "), args...)
	}
	if region != "" {
		like := "%" + region + "%"
		db = db.Where("supplier_profiles.city_id ILIKE ? OR supplier_profiles.province_id ILIKE ?", like, like)
	}
	if verifiedOnly {
		db = db.Where("supplier_profiles.verification_level >= ?", 2)
	}
	if categoryID != "" {
		db = db.Joins("JOIN supplier_categories ON supplier_categories.supplier_profile_id = supplier_profiles.id").
			Where("supplier_categories.category_id = ?", categoryID)
	}

	var profiles []models.SupplierProfile
	if err := db.
		Order("supplier_profiles.verification_level DESC").
		Order("supplier_profiles.response_rate DESC").
		Order("supplier_profiles.star_rating DESC").
		Order("supplier_profiles.updated_at DESC").
		Find(&profiles).Error; err != nil {
		return nil, err
	}

	return u.buildSupplierResponses(ctx, profiles)
}

func (u *discoveryUsecase) GetBySlug(ctx context.Context, slug string) (*dto.PublicSupplierDto, error) {
	var profile models.SupplierProfile
	// Direct indexed lookup by slug (and by ID if input is a valid UUID)
	query := u.db.WithContext(ctx).Where("status = ?", "active")
	if _, errUUID := uuid.Parse(slug); errUUID == nil {
		query = query.Where("slug = ? OR id = ?", slug, slug)
	} else {
		query = query.Where("slug = ?", slug)
	}

	err := query.First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Fallback: match by company_name converted with LIMIT 1 instead of loading all database
			likePattern := "%" + strings.ReplaceAll(slug, "-", "%") + "%"
			if err = u.db.WithContext(ctx).Where("status = ? AND company_name ILIKE ?", "active", likePattern).First(&profile).Error; err != nil {
				return nil, errors.New("supplier profile not found")
			}
		} else {
			return nil, err
		}
	}

	response, err := u.buildSupplierResponse(ctx, profile, true)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (u *discoveryUsecase) ListProducts(ctx context.Context, params dto.ListPublicProductsParams) ([]dto.PublicProductDto, error) {
	db := u.db.WithContext(ctx).
		Model(&models.SupplierProduct{}).
		Preload("Photos").
		Preload("Category").
		Joins("JOIN supplier_profiles ON supplier_profiles.id = supplier_products.supplier_profile_id").
		Where("supplier_profiles.status = ?", "active")

	if params.Query != "" {
		clauses, args := productSearchClauses(params.Query)
		db = db.Where(strings.Join(clauses, " OR "), args...)
	}

	if params.Category != "" {
		db = db.Joins("LEFT JOIN categories ON categories.id = supplier_products.category_id").
			Where("categories.id = ? OR categories.slug = ? OR categories.name ILIKE ?", params.Category, params.Category, "%"+params.Category+"%")
	}

	if params.Location != "" && params.Location != "all" && params.Location != "Semua Lokasi" {
		locLike := "%" + params.Location + "%"
		db = db.Where("supplier_profiles.city_id ILIKE ? OR supplier_profiles.province_id ILIKE ?", locLike, locLike)
	}

	if params.MinPrice != nil && *params.MinPrice > 0 {
		db = db.Where("supplier_products.starting_price >= ?", *params.MinPrice)
	}

	if params.MaxPrice != nil && *params.MaxPrice > 0 {
		db = db.Where("supplier_products.starting_price <= ?", *params.MaxPrice)
	}

	if params.VerifiedOnly {
		db = db.Where("supplier_profiles.verification_level >= ?", 2)
	}

	if params.PowerSupplierOnly {
		db = db.Where("supplier_profiles.is_premium_verified = ?", true)
	}

	if params.SupplierID != "" {
		if !regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`).MatchString(params.SupplierID) {
			return []dto.PublicProductDto{}, nil
		}
		db = db.Where("supplier_products.supplier_profile_id = ?", params.SupplierID)
	}

	// Sorting
	switch params.Sort {
	case "price_asc", "termurah":
		db = db.Order("supplier_products.starting_price ASC")
	case "price_desc", "termahal":
		db = db.Order("supplier_products.starting_price DESC")
	case "rating":
		db = db.Order("supplier_profiles.star_rating DESC")
	case "newest", "terbaru":
		db = db.Order("supplier_products.created_at DESC")
	case "terlaris", "popular":
		fallthrough
	default:
		db = db.Order("supplier_products.is_featured DESC, supplier_profiles.star_rating DESC, supplier_products.sort_order ASC")
	}

	page, limit := utils.NormalizePagination(params.Page, params.Limit, 12)
	db = db.Offset(utils.PaginationOffset(page, limit)).Limit(limit)

	var products []models.SupplierProduct
	if err := db.Preload("Photos").Preload("Category").Find(&products).Error; err != nil {
		return nil, err
	}

	return u.buildProductResponses(ctx, products)
}

func (u *discoveryUsecase) GetProductByID(ctx context.Context, id string) (*dto.PublicProductDetailDto, error) {
	var product models.SupplierProduct
	if err := u.db.WithContext(ctx).
		Preload("Photos").
		Preload("Category").
		Joins("JOIN supplier_profiles ON supplier_profiles.id = supplier_products.supplier_profile_id").
		Where("supplier_products.id = ? AND supplier_profiles.status = ?", id, "active").
		First(&product).Error; err != nil {
		return nil, err
	}

	productResponses, err := u.buildProductResponses(ctx, []models.SupplierProduct{product})
	if err != nil {
		return nil, err
	}
	if len(productResponses) == 0 {
		return nil, errors.New("product not found")
	}

	var supplier models.SupplierProfile
	if err := u.db.WithContext(ctx).Where("id = ? AND status = ?", product.SupplierProfileID, "active").First(&supplier).Error; err != nil {
		return nil, err
	}
	supplierResponse, err := u.buildSupplierResponse(ctx, supplier, true)
	if err != nil {
		return nil, err
	}

	reviews, err := u.getSupplierReviews(ctx, supplier.ID, 8)
	if err != nil {
		return nil, err
	}

	var related []models.SupplierProduct
	if err := u.db.WithContext(ctx).
		Preload("Photos").
		Preload("Category").
		Where("supplier_profile_id = ? AND id <> ?", supplier.ID, product.ID).
		Order("is_featured DESC, sort_order ASC").
		Limit(4).
		Find(&related).Error; err != nil {
		return nil, err
	}
	relatedResponses, err := u.buildProductResponses(ctx, related)
	if err != nil {
		return nil, err
	}

	return &dto.PublicProductDetailDto{
		Product:         productResponses[0],
		Supplier:        supplierResponse,
		Reviews:         reviews,
		RelatedProducts: relatedResponses,
	}, nil
}

func (u *discoveryUsecase) LookupSuppliers(ctx context.Context, q string, page int, limit int) ([]dto.PublicSupplierDto, error) {
	page, limit = utils.NormalizePagination(page, limit, 5)
	offset := utils.PaginationOffset(page, limit)

	db := u.db.WithContext(ctx).Model(&models.SupplierProfile{}).Where("status = ?", "active")
	if q != "" {
		clauses := make([]string, 0)
		args := make([]interface{}, 0)
		for _, term := range expandedSearchTerms(q) {
			like := "%" + term + "%"
			clauses = append(clauses, "(company_name ILIKE ? OR description ILIKE ?)")
			args = append(args, like, like)
		}
		db = db.Where(strings.Join(clauses, " OR "), args...)
	}

	var profiles []models.SupplierProfile
	if err := db.Order("verification_level DESC, company_name ASC").Offset(offset).Limit(limit).Find(&profiles).Error; err != nil {
		return nil, err
	}

	return u.buildSupplierResponses(ctx, profiles)
}

func (u *discoveryUsecase) LookupProducts(ctx context.Context, q string, page int, limit int) ([]dto.PublicProductDto, error) {
	page, limit = utils.NormalizePagination(page, limit, 5)
	offset := utils.PaginationOffset(page, limit)

	db := u.db.WithContext(ctx).
		Model(&models.SupplierProduct{}).
		Preload("Photos").
		Preload("Category").
		Joins("JOIN supplier_profiles ON supplier_profiles.id = supplier_products.supplier_profile_id").
		Where("supplier_profiles.status = ?", "active")
	if q != "" {
		clauses, args := productSearchClauses(q)
		db = db.Where(strings.Join(clauses, " OR "), args...)
	}

	var products []models.SupplierProduct
	if err := db.Order("supplier_products.is_featured DESC, supplier_products.name ASC").Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, err
	}

	return u.buildProductResponses(ctx, products)
}

func (u *discoveryUsecase) buildSupplierResponses(ctx context.Context, profiles []models.SupplierProfile) ([]dto.PublicSupplierDto, error) {
	if len(profiles) == 0 {
		return []dto.PublicSupplierDto{}, nil
	}

	supplierIDs := make([]string, 0, len(profiles))
	userIDs := make([]string, 0, len(profiles))
	for _, p := range profiles {
		supplierIDs = append(supplierIDs, p.ID)
		if p.UserID != "" {
			userIDs = append(userIDs, p.UserID)
		}
	}

	// 1. Batch categories
	type supplierCatRow struct {
		SupplierProfileID string
		Name              string
	}
	var catRows []supplierCatRow
	u.db.WithContext(ctx).
		Table("supplier_categories").
		Select("supplier_categories.supplier_profile_id, categories.name").
		Joins("join categories on categories.id = supplier_categories.category_id").
		Where("supplier_categories.supplier_profile_id IN ?", supplierIDs).
		Order("supplier_categories.is_primary DESC, supplier_categories.created_at ASC").
		Scan(&catRows)

	catMap := make(map[string][]string)
	for _, row := range catRows {
		catMap[row.SupplierProfileID] = append(catMap[row.SupplierProfileID], row.Name)
	}

	// 2. Batch key products (top 3 per supplier)
	var keyProdRows []models.SupplierProduct
	u.db.WithContext(ctx).
		Model(&models.SupplierProduct{}).
		Select("supplier_profile_id, name").
		Where("supplier_profile_id IN ?", supplierIDs).
		Order("is_featured DESC, sort_order ASC").
		Find(&keyProdRows)

	keyProdMap := make(map[string][]string)
	for _, p := range keyProdRows {
		if len(keyProdMap[p.SupplierProfileID]) < 3 {
			keyProdMap[p.SupplierProfileID] = append(keyProdMap[p.SupplierProfileID], p.Name)
		}
	}

	// 3. Batch certifications
	type certRow struct {
		ID                string
		SupplierProfileID string
		Name              string
		IssuedBy          string
		IssuedAt          *time.Time
	}
	var certRows []certRow
	u.db.WithContext(ctx).
		Table("supplier_certifications").
		Select("supplier_certifications.id, supplier_certifications.supplier_profile_id, certifications.name, supplier_certifications.issued_by, supplier_certifications.issued_at").
		Joins("join certifications on certifications.id = supplier_certifications.certification_id").
		Where("supplier_certifications.supplier_profile_id IN ? AND supplier_certifications.status = ?", supplierIDs, "approved").
		Order("certifications.sort_order ASC").
		Scan(&certRows)

	certNameMap := make(map[string][]string)
	certDtoMap := make(map[string][]dto.SupplierCertificationDto)
	for _, c := range certRows {
		certNameMap[c.SupplierProfileID] = append(certNameMap[c.SupplierProfileID], c.Name)
		year := 0
		if c.IssuedAt != nil {
			year = c.IssuedAt.Year()
		}
		certDtoMap[c.SupplierProfileID] = append(certDtoMap[c.SupplierProfileID], dto.SupplierCertificationDto{
			ID:          c.ID,
			Name:        c.Name,
			Institution: c.IssuedBy,
			Year:        year,
		})
	}

	// 4. Batch user avatars
	type userAvatarRow struct {
		ID        string
		AvatarURL string
	}
	var avatarRows []userAvatarRow
	if len(userIDs) > 0 {
		u.db.WithContext(ctx).
			Table("users").
			Select("id, avatar_url").
			Where("id IN ?", userIDs).
			Scan(&avatarRows)
	}
	avatarMap := make(map[string]string)
	for _, a := range avatarRows {
		avatarMap[a.ID] = a.AvatarURL
	}

	// Assemble responses in memory
	responses := make([]dto.PublicSupplierDto, 0, len(profiles))
	for _, p := range profiles {
		catNames := catMap[p.ID]
		if catNames == nil {
			catNames = []string{}
		}

		keyProducts := keyProdMap[p.ID]
		if keyProducts == nil {
			keyProducts = []string{}
		}

		certNames := certNameMap[p.ID]
		if certNames == nil {
			certNames = []string{}
		}

		certDtos := certDtoMap[p.ID]
		if certDtos == nil {
			certDtos = []dto.SupplierCertificationDto{}
		}

		establishedYear, _ := strconv.Atoi(p.EstablishedYear)
		if establishedYear == 0 {
			establishedYear = 2015
		}

		slug := p.Slug
		if slug == "" {
			slug = slugify(p.CompanyName)
		}

		res := dto.PublicSupplierDto{
			ID:                p.ID,
			Slug:              slug,
			CompanyName:       p.CompanyName,
			BusinessType:      p.CompanyType,
			EstablishedYear:   establishedYear,
			EmployeeCount:     p.EmployeesCount,
			Location:          p.CityID,
			Province:          p.ProvinceID,
			Address:           p.Address,
			Description:       p.Description,
			IsVerified:        p.VerificationLevel >= 2,
			VerificationLevel: p.VerificationLevel,
			IsPremiumVerified: p.IsPremiumVerified,
			TaxStatus:         p.TaxStatus,
			ResponseRate:      p.ResponseRate,
			ResponseTime:      responseTimeLabel(p.AvgResponseTimeMinutes),
			Rating:            p.StarRating,
			ReviewCount:       p.ReviewCount,
			KeyProducts:       keyProducts,
			Certifications:    certNames,
			CertificationList: certDtos,
			Logo:              avatarMap[p.UserID],
		}
		if len(res.KeyProducts) == 0 {
			res.KeyProducts = catNames
		}

		responses = append(responses, res)
	}

	return responses, nil
}

func (u *discoveryUsecase) buildSupplierResponse(ctx context.Context, p models.SupplierProfile, includeDetail bool) (dto.PublicSupplierDto, error) {
	catNames, err := u.getSupplierCategories(ctx, p.ID)
	if err != nil {
		return dto.PublicSupplierDto{}, err
	}
	keyProducts, err := u.getSupplierKeyProducts(ctx, p.ID)
	if err != nil {
		return dto.PublicSupplierDto{}, err
	}
	certNames, certDtos, err := u.getSupplierCertifications(ctx, p.ID)
	if err != nil {
		return dto.PublicSupplierDto{}, err
	}

	establishedYear, _ := strconv.Atoi(p.EstablishedYear)
	if establishedYear == 0 {
		establishedYear = 2015
	}

	var avatarURL string
	_ = u.db.WithContext(ctx).
		Table("users").
		Select("avatar_url").
		Where("id = ?", p.UserID).
		Limit(1).
		Row().
		Scan(&avatarURL)

	response := dto.PublicSupplierDto{
		ID:                p.ID,
		Slug:              slugify(p.CompanyName),
		CompanyName:       p.CompanyName,
		BusinessType:      p.CompanyType,
		EstablishedYear:   establishedYear,
		EmployeeCount:     p.EmployeesCount,
		Location:          p.CityID,
		Province:          p.ProvinceID,
		Address:           p.Address,
		Description:       p.Description,
		IsVerified:        p.VerificationLevel >= 2,
		VerificationLevel: p.VerificationLevel,
		IsPremiumVerified: p.IsPremiumVerified,
		TaxStatus:         p.TaxStatus,
		ResponseRate:      p.ResponseRate,
		ResponseTime:      responseTimeLabel(p.AvgResponseTimeMinutes),
		Rating:            p.StarRating,
		ReviewCount:       p.ReviewCount,
		KeyProducts:       keyProducts,
		Certifications:    certNames,
		CertificationList: certDtos,
		Logo:              avatarURL,
	}

	if len(response.KeyProducts) == 0 {
		response.KeyProducts = catNames
	}

	if includeDetail {
		products, err := u.getSupplierProducts(ctx, p.ID)
		if err != nil {
			return dto.PublicSupplierDto{}, err
		}
		reviews, err := u.getSupplierReviews(ctx, p.ID, 8)
		if err != nil {
			return dto.PublicSupplierDto{}, err
		}
		response.Products = products
		response.Reviews = reviews
		response.Phone = p.Phone
		response.WhatsApp = p.WhatsApp
		response.Email = p.Email
		response.Website = p.Website
	}

	return response, nil
}

func (u *discoveryUsecase) getSupplierCategories(ctx context.Context, supplierID string) ([]string, error) {
	var catNames []string
	err := u.db.WithContext(ctx).
		Table("supplier_categories").
		Select("categories.name").
		Joins("join categories on categories.id = supplier_categories.category_id").
		Where("supplier_categories.supplier_profile_id = ?", supplierID).
		Order("supplier_categories.is_primary DESC, supplier_categories.created_at ASC").
		Pluck("name", &catNames).Error
	if err != nil {
		return nil, err
	}
	if catNames == nil {
		catNames = []string{}
	}
	return catNames, nil
}

func (u *discoveryUsecase) getSupplierKeyProducts(ctx context.Context, supplierID string) ([]string, error) {
	var keyProducts []string
	err := u.db.WithContext(ctx).
		Model(&models.SupplierProduct{}).
		Where("supplier_profile_id = ?", supplierID).
		Order("is_featured DESC, sort_order ASC").
		Limit(3).
		Pluck("name", &keyProducts).Error
	if err != nil {
		return nil, err
	}
	if keyProducts == nil {
		keyProducts = []string{}
	}
	return keyProducts, nil
}

func (u *discoveryUsecase) getSupplierProducts(ctx context.Context, supplierID string) ([]dto.SupplierProductDto, error) {
	var products []models.SupplierProduct
	if err := u.db.WithContext(ctx).Preload("Photos").Preload("Category").Where("supplier_profile_id = ?", supplierID).Order("is_featured DESC, sort_order ASC").Find(&products).Error; err != nil {
		return nil, err
	}

	productDtos := make([]dto.SupplierProductDto, 0, len(products))
	for _, product := range products {
		productDtos = append(productDtos, toSupplierProductDto(product))
	}
	return productDtos, nil
}

func (u *discoveryUsecase) getSupplierCertifications(ctx context.Context, supplierID string) ([]string, []dto.SupplierCertificationDto, error) {
	type certInfo struct {
		ID       string
		Name     string
		IssuedBy string
		IssuedAt *time.Time
	}
	var certs []certInfo
	if err := u.db.WithContext(ctx).
		Table("supplier_certifications").
		Select("supplier_certifications.id, certifications.name, supplier_certifications.issued_by, supplier_certifications.issued_at").
		Joins("join certifications on certifications.id = supplier_certifications.certification_id").
		Where("supplier_certifications.supplier_profile_id = ? AND supplier_certifications.status = ?", supplierID, "approved").
		Scan(&certs).Error; err != nil {
		return nil, nil, err
	}

	certNames := make([]string, 0, len(certs))
	certDtos := make([]dto.SupplierCertificationDto, 0, len(certs))
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
	return certNames, certDtos, nil
}

func (u *discoveryUsecase) getSupplierReviews(ctx context.Context, supplierID string, limit int) ([]dto.PublicReviewDto, error) {
	type reviewWithBuyer struct {
		ID                string
		Rating            int
		ReviewText        string
		SupplierReply     string
		SupplierRepliedAt *time.Time
		BuyerName         string
		CreatedAt         time.Time
	}

	query := u.db.WithContext(ctx).
		Table("supplier_reviews").
		Select("supplier_reviews.id, supplier_reviews.rating, supplier_reviews.review_text, supplier_reviews.supplier_reply, supplier_reviews.supplier_replied_at, buyer_profiles.company_name as buyer_name, supplier_reviews.created_at").
		Joins("join buyer_profiles on buyer_profiles.id = supplier_reviews.buyer_profile_id").
		Where("supplier_reviews.supplier_profile_id = ? AND supplier_reviews.status = ?", supplierID, "approved").
		Order("supplier_reviews.created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	var dbReviews []reviewWithBuyer
	if err := query.Scan(&dbReviews).Error; err != nil {
		return nil, err
	}

	reviews := make([]dto.PublicReviewDto, 0, len(dbReviews))
	for _, r := range dbReviews {
		repliedAt := ""
		if r.SupplierRepliedAt != nil {
			repliedAt = r.SupplierRepliedAt.Format(time.RFC3339)
		}
		reviews = append(reviews, dto.PublicReviewDto{
			ID:                r.ID,
			BuyerName:         r.BuyerName,
			Rating:            r.Rating,
			ReviewText:        r.ReviewText,
			SupplierReply:     r.SupplierReply,
			SupplierRepliedAt: repliedAt,
			CreatedAt:         r.CreatedAt.Format(time.RFC3339),
		})
	}
	return reviews, nil
}

func (u *discoveryUsecase) buildProductResponses(ctx context.Context, products []models.SupplierProduct) ([]dto.PublicProductDto, error) {
	if len(products) == 0 {
		return []dto.PublicProductDto{}, nil
	}

	supplierIDSet := make(map[string]struct{})
	var supplierIDs []string
	for _, product := range products {
		if _, exists := supplierIDSet[product.SupplierProfileID]; !exists {
			supplierIDSet[product.SupplierProfileID] = struct{}{}
			supplierIDs = append(supplierIDs, product.SupplierProfileID)
		}
	}

	var suppliers []models.SupplierProfile
	if err := u.db.WithContext(ctx).Where("id IN ? AND status = ?", supplierIDs, "active").Find(&suppliers).Error; err != nil {
		return nil, err
	}

	supplierMap := make(map[string]models.SupplierProfile)
	for _, supplier := range suppliers {
		supplierMap[supplier.ID] = supplier
	}

	responses := make([]dto.PublicProductDto, 0, len(products))
	for _, product := range products {
		supplier, exists := supplierMap[product.SupplierProfileID]
		if !exists {
			continue
		}

		categoryName := ""
		categorySlug := ""
		if product.Category != nil {
			categoryName = product.Category.Name
			categorySlug = product.Category.Slug
		}

		unit := extractUnit(product.MOQ, product.Name)
		tags := []string{"Ready Stock"}
		if product.IsFeatured {
			tags = append(tags, "Pengiriman Cepat")
		} else if supplier.VerificationLevel >= 2 {
			tags = append(tags, "Harga Grosir")
		}

		responses = append(responses, dto.PublicProductDto{
			ID:                  product.ID,
			Name:                product.Name,
			Description:         product.Description,
			Price:               product.StartingPrice,
			Currency:            product.Currency,
			Unit:                unit,
			MinOrder:            product.MOQ,
			CapacityText:        product.CapacityText,
			CategoryName:        categoryName,
			CategorySlug:        categorySlug,
			Photos:              productPhotos(product.Photos),
			SupplierID:          supplier.ID,
			SupplierCompanyName: supplier.CompanyName,
			SupplierSlug:        slugify(supplier.CompanyName),
			SupplierLocation:    supplier.CityID,
			SupplierVerified:    supplier.VerificationLevel >= 2,
			IsPowerSupplier:     supplier.IsPremiumVerified,
			SupplierRating:      supplier.StarRating,
			SupplierReviewCount: supplier.ReviewCount,
			Tags:                tags,
		})
	}

	return responses, nil
}

func extractUnit(moq string, name string) string {
	lower := strings.ToLower(moq + " " + name)
	if strings.Contains(lower, "karton") || strings.Contains(lower, "carton") {
		return "karton"
	}
	if strings.Contains(lower, "box") || strings.Contains(lower, "kotak") {
		return "box"
	}
	if strings.Contains(lower, "ton") {
		return "ton"
	}
	if strings.Contains(lower, "kg") || strings.Contains(lower, "kilo") {
		return "kg"
	}
	if strings.Contains(lower, "meter") || strings.Contains(lower, "m") {
		return "meter"
	}
	if strings.Contains(lower, "rim") {
		return "rim"
	}
	if strings.Contains(lower, "pcs") || strings.Contains(lower, "buah") {
		return "pcs"
	}
	return "unit"
}

func toSupplierProductDto(product models.SupplierProduct) dto.SupplierProductDto {
	categoryName := ""
	if product.Category != nil {
		categoryName = product.Category.Name
	}

	return dto.SupplierProductDto{
		ID:           product.ID,
		Name:         product.Name,
		Description:  product.Description,
		Price:        product.StartingPrice,
		Currency:     product.Currency,
		MinOrder:     product.MOQ,
		CapacityText: product.CapacityText,
		CategoryName: categoryName,
		Photos:       productPhotos(product.Photos),
	}
}

func productPhotos(photos []models.SupplierProductPhoto) []string {
	values := make([]string, 0, len(photos))
	for _, photo := range photos {
		values = append(values, photo.FileURL)
	}
	return values
}

func responseTimeLabel(minutes int) string {
	if minutes <= 0 {
		return "Belum ada data"
	}
	if minutes < 60 {
		return fmt.Sprintf("%d menit", minutes)
	}
	hours := minutes / 60
	if hours < 24 {
		return fmt.Sprintf("%d jam", hours)
	}
	return fmt.Sprintf("%d hari", hours/24)
}

func productSearchClauses(q string) ([]string, []interface{}) {
	clauses := make([]string, 0)
	args := make([]interface{}, 0)
	for _, term := range expandedSearchTerms(q) {
		like := "%" + term + "%"
		clauses = append(clauses, "(supplier_products.name ILIKE ? OR supplier_products.description ILIKE ? OR supplier_profiles.company_name ILIKE ?)")
		args = append(args, like, like, like)
	}
	return clauses, args
}

func expandedSearchTerms(q string) []string {
	base := strings.TrimSpace(strings.ToLower(q))
	if base == "" {
		return []string{}
	}

	terms := []string{base}
	synonyms := map[string][]string{
		"kopi":     {"coffee", "arabica", "gayo"},
		"coffee":   {"kopi"},
		"baja":     {"steel", "metal"},
		"steel":    {"baja", "logam"},
		"kain":     {"fabric", "textile", "cotton"},
		"tekstil":  {"textile", "fabric"},
		"furnitur": {"furniture", "wood", "teak"},
		"mebel":    {"furniture", "wood", "teak"},
		"jahe":     {"ginger"},
		"ginger":   {"jahe"},
	}
	for key, values := range synonyms {
		if strings.Contains(base, key) {
			terms = append(terms, values...)
		}
	}

	seen := make(map[string]struct{}, len(terms))
	unique := make([]string, 0, len(terms))
	for _, term := range terms {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		if _, exists := seen[term]; exists {
			continue
		}
		seen[term] = struct{}{}
		unique = append(unique, term)
	}
	return unique
}

func slugify(s string) string {
	s = strings.ToLower(s)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}
