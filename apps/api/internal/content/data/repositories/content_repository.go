package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/content/data/models"
)

type ContentArticleFilters struct {
	Type       string
	Locale     string
	Status     string
	Search     string
	PublicOnly bool
	Page       int
	PerPage    int
}

type ContentArticleRepository interface {
	List(ctx context.Context, filters ContentArticleFilters) ([]models.ContentArticle, int64, error)
	FindByID(ctx context.Context, id string) (*models.ContentArticle, error)
	FindPublishedBySlug(ctx context.Context, slug string, locale string) (*models.ContentArticle, error)
	SlugExists(ctx context.Context, slug string, excludeID string) (bool, error)
	Create(ctx context.Context, article *models.ContentArticle) error
	Update(ctx context.Context, article *models.ContentArticle) error
	Delete(ctx context.Context, id string) error
}

type contentArticleRepository struct {
	db *gorm.DB
}

func NewContentArticleRepository(db *gorm.DB) ContentArticleRepository {
	return &contentArticleRepository{db: db}
}

func (r *contentArticleRepository) List(ctx context.Context, filters ContentArticleFilters) ([]models.ContentArticle, int64, error) {
	page := filters.Page
	if page <= 0 {
		page = 1
	}
	perPage := filters.PerPage
	if perPage <= 0 {
		perPage = 12
	}
	if perPage > 100 {
		perPage = 100
	}

	query := r.db.WithContext(ctx).Model(&models.ContentArticle{})
	if filters.PublicOnly {
		query = query.Where("status = ?", "published")
		if filters.Locale != "" {
			query = query.Where("(locale = ? OR locale = ?)", filters.Locale, "both")
		}
	} else if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Type != "" {
		query = query.Where("type = ?", filters.Type)
	}
	if filters.Search != "" {
		like := "%" + filters.Search + "%"
		query = query.Where("title ILIKE ? OR excerpt ILIKE ? OR body ILIKE ? OR author_name ILIKE ?", like, like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var articles []models.ContentArticle
	err := query.
		Order("is_featured DESC").
		Order("sort_order ASC").
		Order("published_at DESC NULLS LAST").
		Order("created_at DESC").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&articles).Error
	if err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

func (r *contentArticleRepository) FindByID(ctx context.Context, id string) (*models.ContentArticle, error) {
	var article models.ContentArticle
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&article).Error; err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *contentArticleRepository) FindPublishedBySlug(ctx context.Context, slug string, locale string) (*models.ContentArticle, error) {
	var article models.ContentArticle
	query := r.db.WithContext(ctx).Where("slug = ? AND status = ?", slug, "published")
	if locale != "" {
		query = query.Where("(locale = ? OR locale = ?)", locale, "both")
	}
	if err := query.First(&article).Error; err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *contentArticleRepository) SlugExists(ctx context.Context, slug string, excludeID string) (bool, error) {
	query := r.db.WithContext(ctx).Model(&models.ContentArticle{}).Where("slug = ?", slug)
	if excludeID != "" {
		query = query.Where("id <> ?", excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *contentArticleRepository) Create(ctx context.Context, article *models.ContentArticle) error {
	return r.db.WithContext(ctx).Create(article).Error
}

func (r *contentArticleRepository) Update(ctx context.Context, article *models.ContentArticle) error {
	return r.db.WithContext(ctx).Save(article).Error
}

func (r *contentArticleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.ContentArticle{}).Error
}
