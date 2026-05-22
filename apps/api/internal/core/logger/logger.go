package logger

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

var Logger *log.Logger

func Init() {
	Logger = log.New(os.Stdout, "[API] ", log.LstdFlags|log.Lshortfile)
}

// LogRequest logs HTTP request
func LogRequest(c *gin.Context) {
	Logger.Printf(
		"[%s] %s %s - %s - RequestID: %s",
		c.Request.Method,
		c.Request.URL.Path,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		getRequestID(c),
	)
}

// LogError logs error with context
func LogError(err error, context map[string]interface{}) {
	Logger.Printf("ERROR: %v - Context: %+v", err, context)
}

// LogInfo logs info message
func LogInfo(message string, context map[string]interface{}) {
	Logger.Printf("INFO: %s - Context: %+v", message, context)
}

func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return "unknown"
}
