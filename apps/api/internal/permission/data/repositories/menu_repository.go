package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/permission/data/models"
	"gorm.io/gorm"
)

type MenuRepository interface {
	FindByID(ctx context.Context, id string) (*models.Menu, error)
	FindByURL(ctx context.Context, url string) (*models.Menu, error)
	List(ctx context.Context) ([]models.Menu, error)
	GetRootMenus(ctx context.Context) ([]models.Menu, error)
	Create(ctx context.Context, m *models.Menu) error
	Update(ctx context.Context, m *models.Menu) error
	Delete(ctx context.Context, id string) error
}

type menuRepository struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &menuRepository{db: db}
}

func (r *menuRepository) FindByID(ctx context.Context, id string) (*models.Menu, error) {
	var m models.Menu
	err := r.db.WithContext(ctx).Preload("Parent").Preload("Children").Where("id = ?", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *menuRepository) FindByURL(ctx context.Context, url string) (*models.Menu, error) {
	var m models.Menu
	err := r.db.WithContext(ctx).Preload("Parent").Preload("Children").Where("url = ?", url).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *menuRepository) List(ctx context.Context) ([]models.Menu, error) {
	var menus []models.Menu
	err := r.db.WithContext(ctx).Preload("Parent").Preload("Children").Order("\"order\" ASC").Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return menus, nil
}

func (r *menuRepository) GetRootMenus(ctx context.Context) ([]models.Menu, error) {
	var menus []models.Menu
	err := r.db.WithContext(ctx).Where("parent_id IS NULL").Preload("Children").Order("\"order\" ASC").Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return menus, nil
}

func (r *menuRepository) Create(ctx context.Context, m *models.Menu) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *menuRepository) Update(ctx context.Context, m *models.Menu) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *menuRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Menu{}).Error
}
