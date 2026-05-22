package response

import "net/http"

// Standard Error Codes
const (
	ErrCodeValidationError        = "VALIDATION_ERROR"
	ErrCodeUnauthorized           = "UNAUTHORIZED"
	ErrCodeForbidden              = "FORBIDDEN"
	ErrCodeNotFound               = "NOT_FOUND"
	ErrCodeConflict               = "CONFLICT"
	ErrCodeConcurrentLockConflict = "CONCURRENT_LOCK_CONFLICT"
	ErrCodeAccountNotPostable     = "ACCOUNT_NOT_POSTABLE"
	ErrCodeAccountInactive        = "ACCOUNT_INACTIVE"
	ErrCodePeriodClosed           = "PERIOD_CLOSED"
	ErrCodeMappingNotConfigured   = "MAPPING_NOT_CONFIGURED"
	ErrCodeInternalServerError    = "INTERNAL_SERVER_ERROR"
	ErrCodeServiceUnavailable     = "SERVICE_UNAVAILABLE"
	ErrCodeBadRequest             = "BAD_REQUEST"
	ErrCodeQueryTimeout           = "QUERY_TIMEOUT"
)

// ErrorInfo structure for error definition
type ErrorInfo struct {
	HTTPStatus int
	Message    string
	Code       string
}

// ErrorCodeMap maps internal error codes to HTTP status and messages
var ErrorCodeMap = map[string]ErrorInfo{
	ErrCodeValidationError: {
		HTTPStatus: http.StatusBadRequest,
		Message:    "Validation failed",
		Code:       ErrCodeValidationError,
	},
	ErrCodeUnauthorized: {
		HTTPStatus: http.StatusUnauthorized,
		Message:    "Unauthorized access",
		Code:       ErrCodeUnauthorized,
	},
	ErrCodeForbidden: {
		HTTPStatus: http.StatusForbidden,
		Message:    "Access forbidden",
		Code:       ErrCodeForbidden,
	},
	ErrCodeNotFound: {
		HTTPStatus: http.StatusNotFound,
		Message:    "Resource not found",
		Code:       ErrCodeNotFound,
	},
	ErrCodeConflict: {
		HTTPStatus: http.StatusConflict,
		Message:    "Resource conflict",
		Code:       ErrCodeConflict,
	},
	ErrCodeConcurrentLockConflict: {
		HTTPStatus: http.StatusConflict,
		Message:    "Concurrent update conflict",
		Code:       ErrCodeConcurrentLockConflict,
	},
	ErrCodeAccountNotPostable: {
		HTTPStatus: http.StatusUnprocessableEntity,
		Message:    "Selected account is not postable",
		Code:       ErrCodeAccountNotPostable,
	},
	ErrCodeAccountInactive: {
		HTTPStatus: http.StatusUnprocessableEntity,
		Message:    "Selected account is inactive",
		Code:       ErrCodeAccountInactive,
	},
	ErrCodePeriodClosed: {
		HTTPStatus: http.StatusUnprocessableEntity,
		Message:    "Accounting period is closed",
		Code:       ErrCodePeriodClosed,
	},
	ErrCodeMappingNotConfigured: {
		HTTPStatus: http.StatusUnprocessableEntity,
		Message:    "Required system account mapping is not configured",
		Code:       ErrCodeMappingNotConfigured,
	},
	ErrCodeInternalServerError: {
		HTTPStatus: http.StatusInternalServerError,
		Message:    "Internal server error",
		Code:       ErrCodeInternalServerError,
	},
	ErrCodeServiceUnavailable: {
		HTTPStatus: http.StatusServiceUnavailable,
		Message:    "Service unavailable",
		Code:       ErrCodeServiceUnavailable,
	},
	ErrCodeQueryTimeout: {
		HTTPStatus: http.StatusGatewayTimeout,
		Message:    "Query timeout",
		Code:       ErrCodeQueryTimeout,
	},
}

// GetErrorInfo returns standard error info
func GetErrorInfo(code string) ErrorInfo {
	if info, ok := ErrorCodeMap[code]; ok {
		return info
	}
	return ErrorCodeMap[ErrCodeInternalServerError]
}
