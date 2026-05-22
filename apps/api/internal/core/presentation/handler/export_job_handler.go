package handler

import (
	"net/http"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/exportjob"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gin-gonic/gin"
)

type ExportJobHandler struct {
	mgr *exportjob.Manager
}

func NewExportJobHandler(mgr *exportjob.Manager) *ExportJobHandler {
	return &ExportJobHandler{mgr: mgr}
}

func (h *ExportJobHandler) Get(c *gin.Context) {
	jobID := strings.TrimSpace(c.Param("id"))
	if jobID == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "job id is required", nil, nil)
		return
	}

	job, err := h.mgr.Get(jobID, getUserID(c))
	if err != nil {
		handleExportJobError(c, err)
		return
	}

	response.SuccessResponse(c, job, nil)
}

func (h *ExportJobHandler) Download(c *gin.Context) {
	jobID := strings.TrimSpace(c.Param("id"))
	if jobID == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "job id is required", nil, nil)
		return
	}

	job, err := h.mgr.ResolveDownload(jobID, getUserID(c))
	if err != nil {
		handleExportJobError(c, err)
		return
	}

	c.Redirect(http.StatusFound, job.FileURL)
}

func getUserID(c *gin.Context) string {
	if value, ok := c.Get("user_id"); ok {
		if userID, ok := value.(string); ok {
			return userID
		}
	}
	return ""
}

func handleExportJobError(c *gin.Context, err error) {
	switch err {
	case exportjob.ErrJobNotFound:
		response.ErrorResponse(c, http.StatusNotFound, "EXPORT_JOB_NOT_FOUND", err.Error(), nil, nil)
	case exportjob.ErrJobForbidden:
		response.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", err.Error(), nil, nil)
	case exportjob.ErrJobNotReady:
		response.ErrorResponse(c, http.StatusConflict, "EXPORT_JOB_NOT_READY", err.Error(), nil, nil)
	default:
		response.ErrorResponse(c, http.StatusInternalServerError, "EXPORT_JOB_FAILED", err.Error(), nil, nil)
	}
}
