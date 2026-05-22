package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/gilabs/gims/api/seeders"
	"gorm.io/gorm"
)

type tenantAreaDefaultSeed struct {
	Name        string
	Description string
	Code        string
	Color       string
	Province    string
}

type tenantBankDefaultSeed struct {
	Name      string
	Code      string
	SwiftCode string
}

var tenantAreaDefaults = []tenantAreaDefaultSeed{
	{Name: "Jabodetabek", Description: "Jakarta, Bogor, Depok, Tangerang, Bekasi", Code: "JABODETABEK", Color: "#3b82f6", Province: "DKI Jakarta"},
	{Name: "Jawa Barat", Description: "Bandung and surrounding areas", Code: "JAWA-BARAT", Color: "#10b981", Province: "Jawa Barat"},
	{Name: "Jawa Tengah", Description: "Semarang, Solo and surrounding areas", Code: "JAWA-TENGAH", Color: "#f59e0b", Province: "Jawa Tengah"},
	{Name: "Jawa Timur", Description: "Surabaya, Malang and surrounding areas", Code: "JAWA-TIMUR", Color: "#ef4444", Province: "Jawa Timur"},
	{Name: "Bali", Description: "Bali region", Code: "BALI", Color: "#8b5cf6", Province: "Bali"},
	{Name: "Banten", Description: "Serang, Cilegon, Tangerang Selatan", Code: "BANTEN", Color: "#06b6d4", Province: "Banten"},
	{Name: "DI Yogyakarta", Description: "Yogyakarta and surrounding areas", Code: "DIY", Color: "#ec4899", Province: "Daerah Istimewa Yogyakarta"},
	{Name: "Sumatera Utara", Description: "Medan and surrounding areas", Code: "SUMUT", Color: "#14b8a6", Province: "Sumatera Utara"},
	{Name: "Sulawesi Selatan", Description: "Makassar and surrounding areas", Code: "SULSEL", Color: "#f97316", Province: "Sulawesi Selatan"},
	{Name: "Kalimantan Timur", Description: "Balikpapan, Samarinda", Code: "KALTIM", Color: "#84cc16", Province: "Kalimantan Timur"},
}

var tenantBankDefaults = []tenantBankDefaultSeed{
	{Name: "Bank Central Asia", Code: "BCA", SwiftCode: "CENAIDJA"},
	{Name: "Bank Mandiri", Code: "MANDIRI", SwiftCode: "BMRIIDJA"},
	{Name: "Bank Negara Indonesia", Code: "BNI", SwiftCode: "BNIAIDJAXXX"},
	{Name: "Bank Rakyat Indonesia", Code: "BRI", SwiftCode: "BRINIDJA"},
	{Name: "Bank CIMB Niaga", Code: "CIMB", SwiftCode: "BNIAIDJA"},
}

// ensurePOSGrowthAccountingDefaults keeps legacy billing flows compatible by
// ensuring canonical tenant-1 seed data remains available.
func (u *authUsecase) ensurePOSGrowthAccountingDefaults(ctx context.Context, planSlug string) error {
	_ = ctx
	_ = planSlug

	if err := seeders.SeedChartOfAccounts(); err != nil {
		return fmt.Errorf("failed to ensure default chart of accounts from seeder: %w", err)
	}

	if err := seeders.SeedSystemAccountMappings(); err != nil {
		return fmt.Errorf("failed to ensure default system account mappings from seeder: %w", err)
	}

	return nil
}

// ensureTenantAccountingDefaults provisions tenant-scoped COA and mapping
// defaults during tenant registration/backfill.
func (u *authUsecase) ensureTenantAccountingDefaults(tx *gorm.DB, tenantID string) error {
	if tx == nil {
		return gorm.ErrInvalidDB
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return gorm.ErrInvalidData
	}

	if err := seeders.EnsureTenantAccountingDefaults(tx, tenantID); err != nil {
		return fmt.Errorf("failed to ensure tenant accounting defaults: %w", err)
	}

	return nil
}

func (u *authUsecase) ensureTenantAreaDefaults(tx *gorm.DB, tenantID string) error {
	if tx == nil {
		return gorm.ErrInvalidDB
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return gorm.ErrInvalidData
	}

	for _, areaSeed := range tenantAreaDefaults {
		if err := tx.Exec(`
			INSERT INTO areas (
				id, tenant_id, name, description, is_active, code, color, province, created_at, updated_at
			)
			SELECT
				gen_random_uuid(),
				?,
				?,
				?,
				true,
				?,
				?,
				?,
				NOW(),
				NOW()
			WHERE NOT EXISTS (
				SELECT 1
				FROM areas
				WHERE tenant_id = ?
				  AND lower(name) = lower(?)
				  AND deleted_at IS NULL
			)
		`, tenantID, areaSeed.Name, areaSeed.Description, areaSeed.Code, areaSeed.Color, areaSeed.Province, tenantID, areaSeed.Name).Error; err != nil {
			return fmt.Errorf("failed to ensure tenant default area %s: %w", areaSeed.Name, err)
		}
	}

	return nil
}

func (u *authUsecase) ensureTenantBankDefaults(tx *gorm.DB, tenantID string) error {
	if tx == nil {
		return gorm.ErrInvalidDB
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return gorm.ErrInvalidData
	}

	for _, bankSeed := range tenantBankDefaults {
		if err := tx.Exec(`
			INSERT INTO banks (
				id, tenant_id, name, code, swift_code, is_active, created_at, updated_at
			)
			SELECT
				gen_random_uuid(),
				?,
				?,
				?,
				?,
				true,
				NOW(),
				NOW()
			WHERE NOT EXISTS (
				SELECT 1
				FROM banks
				WHERE tenant_id = ?
				  AND lower(code) = lower(?)
				  AND deleted_at IS NULL
			)
		`, tenantID, bankSeed.Name, bankSeed.Code, bankSeed.SwiftCode, tenantID, bankSeed.Code).Error; err != nil {
			return fmt.Errorf("failed to ensure tenant default bank %s: %w", bankSeed.Code, err)
		}
	}

	return nil
}
