package handler

import (
	"net/http"
	"strings"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gin-gonic/gin"
)

func writeFinanceStandardizedError(c *gin.Context, err error, fallbackStatus int, fallbackCode string) {
	status, code := mapFinanceError(err, fallbackStatus, fallbackCode)
	response.StandardErrorResponse(c, status, code, err.Error(), nil)
}

func mapFinanceError(err error, fallbackStatus int, fallbackCode string) (int, string) {
	message := strings.ToLower(strings.TrimSpace(err.Error()))

	switch {
	case strings.Contains(message, "not postable"):
		return http.StatusUnprocessableEntity, response.ErrCodeAccountNotPostable
	case strings.Contains(message, "inactive") && strings.Contains(message, "account"):
		return http.StatusUnprocessableEntity, response.ErrCodeAccountInactive
	case strings.Contains(message, "period") && strings.Contains(message, "closed"):
		return http.StatusUnprocessableEntity, response.ErrCodePeriodClosed
	case strings.Contains(message, "mapping") && (strings.Contains(message, "not configured") || strings.Contains(message, "belum dikonfigurasi")):
		return http.StatusUnprocessableEntity, response.ErrCodeMappingNotConfigured
	case strings.Contains(message, "lock") || strings.Contains(message, "deadlock") || strings.Contains(message, "concurrent") || strings.Contains(message, "serialization"):
		return http.StatusConflict, response.ErrCodeConcurrentLockConflict
	default:
		return fallbackStatus, fallbackCode
	}
}
