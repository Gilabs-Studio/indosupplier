package router

import (
	"net/http"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gin-gonic/gin"
)

func deprecatedRoute(message string, replacement string) gin.HandlerFunc {
	return func(c *gin.Context) {
		details := map[string]interface{}{}
		if replacement != "" {
			details["replacement"] = replacement
		}

		response.ErrorResponse(
			c,
			http.StatusGone,
			"FINANCE_ROUTE_DEPRECATED",
			message,
			details,
			nil,
		)
	}
}

// RegisterLegacyFinanceRouteBridges keeps removed legacy endpoints discoverable while
// clients migrate to the canonical Finance route structure.
func RegisterLegacyFinanceRouteBridges(rg *gin.RouterGroup) {
	rg.Any("/journal-lines", deprecatedRoute("Journal lines endpoint is deprecated.", "/finance/accounting/journal-entries"))
	rg.Any("/journal-lines/*path", deprecatedRoute("Journal lines endpoint is deprecated.", "/finance/accounting/journal-entries"))

	rg.Any("/cash-bank", deprecatedRoute("Legacy cash-bank endpoint is deprecated. Use canonical Cash & Bank routes.", "/finance/cash-bank/accounts"))
	rg.Any("/cash-bank/*path", deprecatedRoute("Legacy cash-bank endpoint is deprecated. Use canonical Cash & Bank routes.", "/finance/cash-bank/accounts"))

	rg.Any("/salary", deprecatedRoute("Salary endpoint has moved from Finance module.", "/hrd/salary"))
	rg.Any("/salary/*path", deprecatedRoute("Salary endpoint has moved from Finance module.", "/hrd/salary"))
	rg.Any("/salary-structures", deprecatedRoute("Salary structures endpoint has moved from Finance module.", "/hrd/salary"))
	rg.Any("/salary-structures/*path", deprecatedRoute("Salary structures endpoint has moved from Finance module.", "/hrd/salary"))
}
