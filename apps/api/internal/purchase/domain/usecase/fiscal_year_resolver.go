package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	financeRepositories "github.com/gilabs/gims/api/internal/finance/data/repositories"
	"gorm.io/gorm"
)

type purchaseFiscalYearResolver struct {
	db             *gorm.DB
	fiscalYearRepo financeRepositories.FiscalYearRepository
}

func newPurchaseFiscalYearResolver(db *gorm.DB, fiscalYearRepo financeRepositories.FiscalYearRepository) *purchaseFiscalYearResolver {
	return &purchaseFiscalYearResolver{db: db, fiscalYearRepo: fiscalYearRepo}
}

func (r *purchaseFiscalYearResolver) Resolve(ctx context.Context, referenceDate string) (string, string, error) {
	companyID, err := r.resolveCompanyID(ctx)
	if err != nil {
		return "", "", err
	}

	fiscalYearID, err := r.resolveFiscalYearID(ctx, companyID, referenceDate)
	if err != nil {
		return "", "", err
	}

	return companyID, fiscalYearID, nil
}

func (r *purchaseFiscalYearResolver) resolveCompanyID(ctx context.Context) (string, error) {
	userID, _ := ctx.Value("user_id").(string)
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", errors.New("user not authenticated")
	}

	var companyID string
	if err := database.GetDB(ctx, r.db).
		Table("employees").
		Select("company_id").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Scan(&companyID).Error; err != nil {
		return "", fmt.Errorf("failed to resolve purchase company: %w", err)
	}
	companyID = strings.TrimSpace(companyID)
	if companyID == "" {
		return "", errors.New("purchase company is required")
	}

	return companyID, nil
}

func (r *purchaseFiscalYearResolver) resolveFiscalYearID(ctx context.Context, companyID string, referenceDate string) (string, error) {
	companyID = strings.TrimSpace(companyID)
	if companyID == "" {
		return "", errors.New("company id is required")
	}

	resolvedDate := strings.TrimSpace(referenceDate)
	if resolvedDate == "" {
		resolvedDate = apptime.Now().Format("2006-01-02")
	} else if _, err := time.Parse("2006-01-02", resolvedDate); err != nil {
		resolvedDate = apptime.Now().Format("2006-01-02")
	}

	fiscalYear, err := r.fiscalYearRepo.FindActiveByCompany(ctx, companyID)
	if err != nil {
		return "", fmt.Errorf("active fiscal year not found for company %s on %s: %w", companyID, resolvedDate, err)
	}
	if fiscalYear == nil || strings.TrimSpace(fiscalYear.ID) == "" {
		return "", errors.New("active fiscal year is required")
	}

	return strings.TrimSpace(fiscalYear.ID), nil
}

func resolvePurchaseJournalScope(
	ctx context.Context,
	db *gorm.DB,
	companyID *string,
	fiscalYearID *string,
	referenceDate string,
) (string, *string, error) {
	if db == nil {
		return "", nil, errors.New("db is nil")
	}

	resolvedCompanyID := ""
	if companyID != nil {
		resolvedCompanyID = strings.TrimSpace(*companyID)
	}

	var resolvedFiscalYearID *string
	if fiscalYearID != nil {
		trimmedFiscalYearID := strings.TrimSpace(*fiscalYearID)
		if trimmedFiscalYearID != "" {
			resolvedFiscalYearID = &trimmedFiscalYearID
		}
	}

	resolver := newPurchaseFiscalYearResolver(db, financeRepositories.NewFiscalYearRepository(db))

	if resolvedCompanyID == "" {
		companyFromContext, fiscalYearFromContext, err := resolver.Resolve(ctx, referenceDate)
		if err != nil {
			return "", nil, err
		}

		resolvedCompanyID = strings.TrimSpace(companyFromContext)
		if resolvedFiscalYearID == nil {
			trimmedFiscalYearID := strings.TrimSpace(fiscalYearFromContext)
			if trimmedFiscalYearID != "" {
				resolvedFiscalYearID = &trimmedFiscalYearID
			}
		}
	} else if resolvedFiscalYearID == nil {
		fiscalYearFromCompany, err := resolver.resolveFiscalYearID(ctx, resolvedCompanyID, referenceDate)
		if err != nil {
			return "", nil, err
		}

		trimmedFiscalYearID := strings.TrimSpace(fiscalYearFromCompany)
		if trimmedFiscalYearID != "" {
			resolvedFiscalYearID = &trimmedFiscalYearID
		}
	}

	if resolvedCompanyID == "" {
		return "", nil, errors.New("purchase company is required")
	}

	return resolvedCompanyID, resolvedFiscalYearID, nil
}
