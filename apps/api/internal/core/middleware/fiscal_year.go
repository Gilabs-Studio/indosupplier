package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/response"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CheckActiveFiscalYear ensures referenced fiscal year is active before journal operations proceed.
// This middleware is intentionally lenient when fiscal_year_id is not provided to preserve backward compatibility.
func CheckActiveFiscalYear() gin.HandlerFunc {
	return func(c *gin.Context) {
		if database.DB == nil {
			c.Next()
			return
		}

		fiscalYearID := strings.TrimSpace(c.Query("fiscal_year_id"))
		if fiscalYearID == "" {
			if journalID := strings.TrimSpace(c.Param("id")); journalID != "" {
				var entry financeModels.JournalEntry
				err := database.DB.Select("fiscal_year_id").First(&entry, "id = ?", journalID).Error
				if err == nil && entry.FiscalYearID != nil {
					fiscalYearID = strings.TrimSpace(*entry.FiscalYearID)
				}
			}
		}

		if fiscalYearID == "" {
			c.Next()
			return
		}

		var fy financeModels.FiscalYear
		err := database.DB.First(&fy, "id = ?", fiscalYearID).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.StandardErrorResponse(c, http.StatusNotFound, "FISCAL_YEAR_NOT_FOUND", "Fiscal year not found", nil)
				c.Abort()
				return
			}
			response.StandardErrorResponse(c, http.StatusInternalServerError, response.ErrCodeInternalServerError, "Failed to validate fiscal year", nil)
			c.Abort()
			return
		}

		if fy.Status != financeModels.FiscalYearStatusActive {
			response.StandardErrorResponse(c, http.StatusUnprocessableEntity, "FISCAL_YEAR_NOT_ACTIVE", "Fiscal year must be active", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
