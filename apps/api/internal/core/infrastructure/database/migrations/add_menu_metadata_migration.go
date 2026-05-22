package migrations

import (
	"fmt"

	"gorm.io/gorm"
)

// MenuMetadataMigration is a temporary struct used for menu metadata migration.
type MenuMetadataMigration struct {
	ID          string  `gorm:"type:uuid;primaryKey"`
	Name        string  `gorm:"type:varchar(255);not null"`
	Slug        string  `gorm:"type:varchar(255);index"`
	Module      string  `gorm:"type:varchar(100);index;default:''"`
	URL         string  `gorm:"type:varchar(255);not null"`
	IsActive    bool    `gorm:"type:boolean;not null;default:true"`
	IsClickable bool    `gorm:"type:boolean;not null;default:true"`
	Status      string  `gorm:"type:varchar(20);not null;default:'active'"`
	ParentID    *string `gorm:"type:uuid;index"`
}

func (MenuMetadataMigration) TableName() string {
	return "menus"
}

// AddMenuMetadataMigration adds canonical menu metadata columns used by finance refactor.
func AddMenuMetadataMigration(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	if !db.Migrator().HasColumn(&MenuMetadataMigration{}, "slug") {
		if err := db.Migrator().AddColumn(&MenuMetadataMigration{}, "slug"); err != nil {
			return fmt.Errorf("failed to add menus.slug: %w", err)
		}
	}
	if !db.Migrator().HasColumn(&MenuMetadataMigration{}, "module") {
		if err := db.Migrator().AddColumn(&MenuMetadataMigration{}, "module"); err != nil {
			return fmt.Errorf("failed to add menus.module: %w", err)
		}
	}
	if !db.Migrator().HasColumn(&MenuMetadataMigration{}, "is_active") {
		if err := db.Migrator().AddColumn(&MenuMetadataMigration{}, "is_active"); err != nil {
			return fmt.Errorf("failed to add menus.is_active: %w", err)
		}
	}
	if !db.Migrator().HasColumn(&MenuMetadataMigration{}, "is_clickable") {
		if err := db.Migrator().AddColumn(&MenuMetadataMigration{}, "is_clickable"); err != nil {
			return fmt.Errorf("failed to add menus.is_clickable: %w", err)
		}
	}

	// Keep legacy status + new boolean flag in sync.
	if err := db.Exec(`
		UPDATE menus
		SET is_active = CASE WHEN status = 'active' THEN TRUE ELSE FALSE END
		WHERE is_active IS DISTINCT FROM (status = 'active')
	`).Error; err != nil {
		return fmt.Errorf("failed to sync menus.is_active from status: %w", err)
	}

	if err := db.Exec(`
		UPDATE menus
		SET is_clickable = CASE WHEN COALESCE(url, '') IN ('', '#') THEN FALSE ELSE TRUE END
		WHERE is_clickable IS DISTINCT FROM (CASE WHEN COALESCE(url, '') IN ('', '#') THEN FALSE ELSE TRUE END)
	`).Error; err != nil {
		return fmt.Errorf("failed to backfill menus.is_clickable: %w", err)
	}

	// Backfill module for finance root and descendants.
	if err := db.Exec(`
		UPDATE menus
		SET module = 'finance'
		WHERE module = '' AND (url = '/finance' OR url LIKE '/finance/%')
	`).Error; err != nil {
		return fmt.Errorf("failed to backfill menus.module for finance urls: %w", err)
	}

	for i := 0; i < 5; i++ {
		if err := db.Exec(`
			UPDATE menus child
			SET module = parent.module
			FROM menus parent
			WHERE child.parent_id = parent.id
			  AND COALESCE(child.module, '') = ''
			  AND COALESCE(parent.module, '') <> ''
		`).Error; err != nil {
			return fmt.Errorf("failed to propagate menus.module from parent: %w", err)
		}
	}

	// Backfill slug from URL for finance menus where possible.
	if err := db.Exec(`
		UPDATE menus
		SET slug = TRIM(BOTH '.' FROM REPLACE(REPLACE(url, '/finance', 'finance'), '/', '.'))
		WHERE module = 'finance'
		  AND COALESCE(slug, '') = ''
		  AND COALESCE(url, '') NOT IN ('', '#')
	`).Error; err != nil {
		return fmt.Errorf("failed to backfill menus.slug from url: %w", err)
	}

	if err := db.Exec(`
		UPDATE menus
		SET slug = module || '.' || REGEXP_REPLACE(LOWER(name), '[^a-z0-9]+', '-', 'g') || '.' || LEFT(id::text, 8)
		WHERE COALESCE(slug, '') = ''
		  AND COALESCE(module, '') <> ''
	`).Error; err != nil {
		return fmt.Errorf("failed to backfill menus.slug from name: %w", err)
	}

	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_menus_slug ON menus(slug)`).Error; err != nil {
		return fmt.Errorf("failed to create idx_menus_slug: %w", err)
	}

	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_menus_module ON menus(module)`).Error; err != nil {
		return fmt.Errorf("failed to create idx_menus_module: %w", err)
	}

	return nil
}
