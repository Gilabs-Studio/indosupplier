package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"
	"sync"
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
)

// FinanceSettingRepository provides data access for finance_settings.
type FinanceSettingRepository interface {
	GetByKey(ctx context.Context, key string) (string, error)
	FindByKey(ctx context.Context, key string) (*financeModels.FinanceSetting, error)
	GetAll(ctx context.Context) ([]financeModels.FinanceSetting, error)
	Upsert(ctx context.Context, key, value, description, category string) error
}

type financeSettingRepository struct {
	db    *gorm.DB
	mu    sync.RWMutex
	cache map[string]cachedSetting
}

type cachedSetting struct {
	value     string
	fetchedAt time.Time
}

const settingCacheTTL = 5 * time.Minute

func NewFinanceSettingRepository(db *gorm.DB) FinanceSettingRepository {
	return &financeSettingRepository{
		db:    db,
		cache: make(map[string]cachedSetting),
	}
}

// GetByKey returns the value for a given setting key.
// Values are cached in memory for settingCacheTTL to avoid repeated DB lookups
// during request processing.
func (r *financeSettingRepository) GetByKey(ctx context.Context, key string) (string, error) {
	// Check cache first
	r.mu.RLock()
	if cs, ok := r.cache[key]; ok && time.Since(cs.fetchedAt) < settingCacheTTL {
		r.mu.RUnlock()
		return cs.value, nil
	}
	r.mu.RUnlock()

	// Cache miss — read from DB
	var setting financeModels.FinanceSetting
	if err := database.GetDB(ctx, r.db).Where("setting_key = ?", key).First(&setting).Error; err != nil {
		return "", err
	}

	// Update cache
	r.mu.Lock()
	r.cache[key] = cachedSetting{value: setting.Value, fetchedAt: time.Now()}
	r.mu.Unlock()

	return setting.Value, nil
}

// FindByKey returns the full FinanceSetting object for a given key.
// Used for validation and audit purposes.
func (r *financeSettingRepository) FindByKey(ctx context.Context, key string) (*financeModels.FinanceSetting, error) {
	var setting financeModels.FinanceSetting
	if err := database.GetDB(ctx, r.db).Where("setting_key = ?", key).First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &setting, nil
}

// GetAll returns all finance settings.
func (r *financeSettingRepository) GetAll(ctx context.Context) ([]financeModels.FinanceSetting, error) {
	var settings []financeModels.FinanceSetting
	if err := database.GetDB(ctx, r.db).Order("category, setting_key").Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

// Upsert creates or updates a finance setting.
// On conflict, it updates value, description, and category.
func (r *financeSettingRepository) Upsert(ctx context.Context, key, value, description, category string) error {
	setting := financeModels.FinanceSetting{
		SettingKey:  key,
		Value:       value,
		Description: description,
		Category:    category,
	}

	result := database.GetDB(ctx, r.db).
		Where("setting_key = ?", key).
		First(&financeModels.FinanceSetting{})

	if result.Error == gorm.ErrRecordNotFound {
		if err := database.GetDB(ctx, r.db).Create(&setting).Error; err != nil {
			return err
		}
	} else if result.Error == nil {
		if err := database.GetDB(ctx, r.db).
			Model(&financeModels.FinanceSetting{}).
			Where("setting_key = ?", key).
			Updates(map[string]interface{}{
				"value":       value,
				"description": description,
				"category":    category,
			}).Error; err != nil {
			return err
		}
	} else {
		return result.Error
	}

	// Invalidate cache for the key
	r.mu.Lock()
	delete(r.cache, key)
	r.mu.Unlock()

	return nil
}
