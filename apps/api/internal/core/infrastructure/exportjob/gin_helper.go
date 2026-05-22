package exportjob

import (
	"context"
	"net/http"
	"strings"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gin-gonic/gin"
)

// contextKeysToPropagate are string keys copied from request context into async job context.
var contextKeysToPropagate = []string{
	"tenant_id",
	"user_id",
	"user_email",
	"user_role",
	"user_permissions",
	"user_permissions_scope",
	"is_system_admin",
}

// propagateRequestContext copies important values from both the request context
// and gin context into a new background context so async jobs keep tenant/auth data
// even when some middleware only sets values on gin.Context.
func propagateRequestContext(c *gin.Context) context.Context {
	ctx := context.Background()
	reqCtx := c.Request.Context()
	for _, key := range contextKeysToPropagate {
		val := reqCtx.Value(key)
		if val == nil {
			if ginVal, exists := c.Get(key); exists {
				val = ginVal
			}
		}
		if val != nil {
			ctx = context.WithValue(ctx, key, val)
		}
	}
	return ctx
}

func IsAsyncRequested(c *gin.Context) bool {
	flag := strings.TrimSpace(strings.ToLower(c.Query("async")))
	return flag == "1" || flag == "true" || flag == "yes"
}

func QueueIfRequested(c *gin.Context, generator Generator) bool {
	return QueueIfRequestedWithProgress(c, func(ctx context.Context, setProgress func(int)) (*GeneratedFile, error) {
		setProgress(15)
		file, err := generator(ctx)
		if err == nil {
			setProgress(95)
		}
		return file, err
	})
}

type ProgressGenerator func(ctx context.Context, setProgress func(int)) (*GeneratedFile, error)

func QueueIfRequestedWithProgress(c *gin.Context, generator ProgressGenerator) bool {
	if !IsAsyncRequested(c) {
		return false
	}

	userID := ""
	if value, ok := c.Get("user_id"); ok {
		if id, ok := value.(string); ok {
			userID = id
		}
	}

	// Capture request context values before the request ends
	baseCtx := propagateRequestContext(c)

	var jobID string
	job := DefaultManager.Enqueue(userID, func(ctx context.Context) (*GeneratedFile, error) {
		// Merge propagated request context values into the job's timeout context
		for _, key := range contextKeysToPropagate {
			if val := baseCtx.Value(key); val != nil {
				ctx = context.WithValue(ctx, key, val)
			}
		}
		setProgress := func(progress int) {
			if jobID == "" {
				return
			}
			DefaultManager.SetProgress(jobID, progress)
		}
		return generator(ctx, setProgress)
	})
	jobID = job.ID
	response.SuccessResponseAccepted(c, job, nil)
	return true
}

func WriteSyncFile(c *gin.Context, file *GeneratedFile) {
	c.Header("Content-Type", file.ContentType)
	c.Header("Content-Disposition", "attachment; filename=\""+file.FileName+"\"")
	c.Data(http.StatusOK, file.ContentType, file.Bytes)
}
