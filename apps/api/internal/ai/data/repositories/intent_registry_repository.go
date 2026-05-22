package repositories

import (
	"context"
	"strings"

	"github.com/gilabs/gims/api/internal/ai/data/models"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"gorm.io/gorm"
)

// IntentRegistryRepository defines the interface for AI intent registry data access
type IntentRegistryRepository interface {
	FindAll(ctx context.Context) ([]models.AIIntentRegistry, error)
	FindActive(ctx context.Context) ([]models.AIIntentRegistry, error)
	FindByIntentCode(ctx context.Context, code string) (*models.AIIntentRegistry, error)
	Update(ctx context.Context, intent *models.AIIntentRegistry) error
	Create(ctx context.Context, intent *models.AIIntentRegistry) error
}

type intentRegistryRepository struct {
	db *gorm.DB
}

// NewIntentRegistryRepository creates a new intent registry repository
func NewIntentRegistryRepository(db *gorm.DB) IntentRegistryRepository {
	return &intentRegistryRepository{db: db}
}

func (r *intentRegistryRepository) FindAll(ctx context.Context) ([]models.AIIntentRegistry, error) {
	db := database.GetDB(ctx, r.db)
	var intents []models.AIIntentRegistry
	if err := db.Order("module ASC, intent_code ASC").Find(&intents).Error; err != nil {
		return nil, err
	}
	return intents, nil
}

func (r *intentRegistryRepository) FindActive(ctx context.Context) ([]models.AIIntentRegistry, error) {
	db := database.GetDB(ctx, r.db)
	var intents []models.AIIntentRegistry
	if err := db.Where("is_active = ?", true).Order("module ASC, intent_code ASC").Find(&intents).Error; err != nil {
		return nil, err
	}
	return intents, nil
}

func (r *intentRegistryRepository) FindByIntentCode(ctx context.Context, code string) (*models.AIIntentRegistry, error) {
	db := database.GetDB(ctx, r.db)
	var intent models.AIIntentRegistry
	if err := db.Where("UPPER(intent_code) = ? AND is_active = ?", strings.ToUpper(strings.TrimSpace(code)), true).First(&intent).Error; err != nil {
		return nil, err
	}
	return &intent, nil
}

func (r *intentRegistryRepository) Update(ctx context.Context, intent *models.AIIntentRegistry) error {
	db := database.GetDB(ctx, r.db)
	return db.Save(intent).Error
}

func (r *intentRegistryRepository) Create(ctx context.Context, intent *models.AIIntentRegistry) error {
	db := database.GetDB(ctx, r.db)
	return db.Create(intent).Error
}
