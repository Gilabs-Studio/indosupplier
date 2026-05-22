package response

import "github.com/gin-gonic/gin"

// StandardErrorPayload is a lightweight machine-readable error envelope
// used by modules that require strict {error, code, details} compatibility.
type StandardErrorPayload struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code"`
	Details map[string]interface{} `json:"details"`
}

// StandardErrorResponse writes an error payload in the format:
// {"error":"...","code":"...","details":{...}}
func StandardErrorResponse(c *gin.Context, httpStatus int, code, message string, details map[string]interface{}) {
	if details == nil {
		details = map[string]interface{}{}
	}

	c.AbortWithStatusJSON(httpStatus, StandardErrorPayload{
		Error:   message,
		Code:    code,
		Details: details,
	})
}
