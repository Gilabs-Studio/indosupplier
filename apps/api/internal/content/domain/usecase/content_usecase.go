package usecase

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/content/data/models"
	"github.com/gilabs/indosupplier/api/internal/content/data/repositories"
	"github.com/gilabs/indosupplier/api/internal/content/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/core/apptime"
)

var (
	ErrContentArticleNotFound = errors.New("content article not found")
	ErrContentSlugExists      = errors.New("content slug already exists")
)

type ContentUsecase interface {
	ListPublic(ctx context.Context, req dto.ListContentArticlesRequest) ([]dto.ContentArticleResponse, int64, error)
	GetPublicBySlug(ctx context.Context, slug string, locale string) (*dto.ContentArticleResponse, error)
	ListAdmin(ctx context.Context, req dto.ListContentArticlesRequest) ([]dto.ContentArticleResponse, int64, error)
	GetAdminByID(ctx context.Context, id string) (*dto.ContentArticleResponse, error)
	Create(ctx context.Context, req dto.CreateContentArticleRequest) (*dto.ContentArticleResponse, error)
	Update(ctx context.Context, id string, req dto.UpdateContentArticleRequest) (*dto.ContentArticleResponse, error)
	Delete(ctx context.Context, id string) error
}

type contentUsecase struct {
	repo repositories.ContentArticleRepository
}

func NewContentUsecase(repo repositories.ContentArticleRepository) ContentUsecase {
	return &contentUsecase{repo: repo}
}

func (u *contentUsecase) ListPublic(ctx context.Context, req dto.ListContentArticlesRequest) ([]dto.ContentArticleResponse, int64, error) {
	req.Status = "published"
	if req.Locale == "" {
		req.Locale = "id"
	}

	articles, total, err := u.repo.List(ctx, repositories.ContentArticleFilters{
		Type:       req.Type,
		Locale:     req.Locale,
		Search:     req.Search,
		PublicOnly: true,
		Page:       req.Page,
		PerPage:    req.PerPage,
	})
	if err != nil {
		return nil, 0, err
	}

	return toContentArticleResponses(articles), total, nil
}

func (u *contentUsecase) GetPublicBySlug(ctx context.Context, slug string, locale string) (*dto.ContentArticleResponse, error) {
	if locale == "" {
		locale = "id"
	}

	article, err := u.repo.FindPublishedBySlug(ctx, slug, locale)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContentArticleNotFound
		}
		return nil, err
	}

	response := toContentArticleResponse(*article)
	return &response, nil
}

func (u *contentUsecase) ListAdmin(ctx context.Context, req dto.ListContentArticlesRequest) ([]dto.ContentArticleResponse, int64, error) {
	articles, total, err := u.repo.List(ctx, repositories.ContentArticleFilters{
		Type:    req.Type,
		Status:  req.Status,
		Search:  req.Search,
		Page:    req.Page,
		PerPage: req.PerPage,
	})
	if err != nil {
		return nil, 0, err
	}

	return toContentArticleResponses(articles), total, nil
}

func (u *contentUsecase) GetAdminByID(ctx context.Context, id string) (*dto.ContentArticleResponse, error) {
	article, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContentArticleNotFound
		}
		return nil, err
	}

	response := toContentArticleResponse(*article)
	return &response, nil
}

func (u *contentUsecase) Create(ctx context.Context, req dto.CreateContentArticleRequest) (*dto.ContentArticleResponse, error) {
	locale := req.Locale
	if locale == "" {
		locale = "id"
	}
	status := req.Status
	if status == "" {
		status = "draft"
	}
	slug := req.Slug
	if slug == "" {
		slug = req.Title
	}

	normalizedSlug, err := u.uniqueSlug(ctx, slug, "")
	if err != nil {
		return nil, err
	}

	var publishedAt *time.Time
	if status == "published" {
		now := apptime.Now()
		publishedAt = &now
	}

	article := &models.ContentArticle{
		Type:              req.Type,
		Locale:            locale,
		Title:             strings.TrimSpace(req.Title),
		Slug:              normalizedSlug,
		Excerpt:           strings.TrimSpace(req.Excerpt),
		Body:              strings.TrimSpace(req.Body),
		AuthorName:        strings.TrimSpace(req.AuthorName),
		ImageURL:          strings.TrimSpace(req.ImageURL),
		VideoURL:          strings.TrimSpace(req.VideoURL),
		Duration:          strings.TrimSpace(req.Duration),
		ViewCount:         req.ViewCount,
		SupplierProfileID: req.SupplierProfileID,
		SupplierProductID: req.SupplierProductID,
		Status:            status,
		IsFeatured:        req.IsFeatured,
		SortOrder:         req.SortOrder,
		PublishedAt:       publishedAt,
	}

	if err := u.repo.Create(ctx, article); err != nil {
		return nil, err
	}

	response := toContentArticleResponse(*article)
	return &response, nil
}

func (u *contentUsecase) Update(ctx context.Context, id string, req dto.UpdateContentArticleRequest) (*dto.ContentArticleResponse, error) {
	article, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContentArticleNotFound
		}
		return nil, err
	}

	if req.Type != nil {
		article.Type = *req.Type
	}
	if req.Locale != nil {
		article.Locale = *req.Locale
	}
	if req.Title != nil {
		article.Title = strings.TrimSpace(*req.Title)
	}
	if req.Slug != nil {
		nextSlug, err := u.uniqueSlug(ctx, *req.Slug, article.ID)
		if err != nil {
			return nil, err
		}
		article.Slug = nextSlug
	}
	if req.Excerpt != nil {
		article.Excerpt = strings.TrimSpace(*req.Excerpt)
	}
	if req.Body != nil {
		article.Body = strings.TrimSpace(*req.Body)
	}
	if req.AuthorName != nil {
		article.AuthorName = strings.TrimSpace(*req.AuthorName)
	}
	if req.ImageURL != nil {
		article.ImageURL = strings.TrimSpace(*req.ImageURL)
	}
	if req.VideoURL != nil {
		article.VideoURL = strings.TrimSpace(*req.VideoURL)
	}
	if req.Duration != nil {
		article.Duration = strings.TrimSpace(*req.Duration)
	}
	if req.ViewCount != nil {
		article.ViewCount = *req.ViewCount
	}
	if req.SupplierProfileID != nil {
		article.SupplierProfileID = req.SupplierProfileID
	}
	if req.SupplierProductID != nil {
		article.SupplierProductID = req.SupplierProductID
	}
	if req.IsFeatured != nil {
		article.IsFeatured = *req.IsFeatured
	}
	if req.SortOrder != nil {
		article.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		if article.Status != "published" && *req.Status == "published" && article.PublishedAt == nil {
			now := apptime.Now()
			article.PublishedAt = &now
		}
		article.Status = *req.Status
	}

	if err := u.repo.Update(ctx, article); err != nil {
		return nil, err
	}

	response := toContentArticleResponse(*article)
	return &response, nil
}

func (u *contentUsecase) Delete(ctx context.Context, id string) error {
	if _, err := u.repo.FindByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrContentArticleNotFound
		}
		return err
	}

	return u.repo.Delete(ctx, id)
}

func (u *contentUsecase) uniqueSlug(ctx context.Context, input string, excludeID string) (string, error) {
	slug := slugify(input)
	if slug == "" {
		slug = "content-" + uuid.NewString()[:8]
	}

	exists, err := u.repo.SlugExists(ctx, slug, excludeID)
	if err != nil {
		return "", err
	}
	if !exists {
		return slug, nil
	}

	return slug + "-" + uuid.NewString()[:8], nil
}

func toContentArticleResponses(articles []models.ContentArticle) []dto.ContentArticleResponse {
	responses := make([]dto.ContentArticleResponse, 0, len(articles))
	for _, article := range articles {
		responses = append(responses, toContentArticleResponse(article))
	}
	return responses
}

func toContentArticleResponse(article models.ContentArticle) dto.ContentArticleResponse {
	publishedAt := ""
	if article.PublishedAt != nil {
		publishedAt = article.PublishedAt.Format(time.RFC3339)
	}

	return dto.ContentArticleResponse{
		ID:                article.ID,
		Type:              article.Type,
		Locale:            article.Locale,
		Title:             article.Title,
		Slug:              article.Slug,
		Excerpt:           article.Excerpt,
		Body:              article.Body,
		AuthorName:        article.AuthorName,
		ImageURL:          article.ImageURL,
		VideoURL:          article.VideoURL,
		Duration:          article.Duration,
		ViewCount:         article.ViewCount,
		SupplierProfileID: article.SupplierProfileID,
		SupplierProductID: article.SupplierProductID,
		Status:            article.Status,
		IsFeatured:        article.IsFeatured,
		SortOrder:         article.SortOrder,
		PublishedAt:       publishedAt,
		CreatedAt:         article.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         article.UpdatedAt.Format(time.RFC3339),
	}
}

func slugify(value string) string {
	lower := strings.ToLower(strings.TrimSpace(value))
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug := reg.ReplaceAllString(lower, "-")
	return strings.Trim(slug, "-")
}
