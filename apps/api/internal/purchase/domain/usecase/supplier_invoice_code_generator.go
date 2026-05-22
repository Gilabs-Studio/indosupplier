package usecase

import (
	"fmt"
	"strings"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/purchase/data/models"
	"gorm.io/gorm"
)

// Code generator (prefix + date + seq) with advisory transaction lock.
func getNextSupplierInvoiceCodeLocked(tx *gorm.DB, prefix string) (string, error) {
	// Lock scope: per-prefix per-day.
	lockKey := int64(0)
	for _, ch := range prefix {
		lockKey += int64(ch)
	}
	lockKey = lockKey*100000 + int64(apptime.Now().YearDay())

	if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", lockKey).Error; err != nil {
		return "", err
	}

	datePart := apptime.Now().Format("20060102")
	like := fmt.Sprintf("%s-%s-%%", prefix, datePart)

	var last string
	_ = tx.Model(&models.SupplierInvoice{}).
		Unscoped().
		Where("code LIKE ?", like).
		Order("code DESC").
		Limit(1).
		Pluck("code", &last).Error

	seq := 0
	if strings.TrimSpace(last) != "" {
		parts := strings.Split(last, "-")
		if len(parts) >= 3 {
			fmt.Sscanf(parts[len(parts)-1], "%d", &seq)
		}
	}
	seq++
	return fmt.Sprintf("%s-%s-%04d", prefix, datePart, seq), nil
}
