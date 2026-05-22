package handler

import (
	"net/http"

	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"github.com/gilabs/gims/api/internal/hrd/domain/usecase"
	"github.com/gin-gonic/gin"
)

type SalaryStructureHandler struct {
	uc usecase.SalaryStructureUsecase
}

func NewSalaryStructureHandler(uc usecase.SalaryStructureUsecase) *SalaryStructureHandler {
	return &SalaryStructureHandler{uc: uc}
}

func (h *SalaryStructureHandler) Create(c *gin.Context) {
	var req dto.CreateSalaryStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	res, err := h.uc.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": res})
}

func (h *SalaryStructureHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateSalaryStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	res, err := h.uc.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}

func (h *SalaryStructureHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

func (h *SalaryStructureHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}

func (h *SalaryStructureHandler) List(c *gin.Context) {
	var req dto.ListSalaryStructuresRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	items, total, err := h.uc.List(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
		"meta": gin.H{
			"total":    total,
			"page":     req.Page,
			"per_page": req.PerPage,
		},
	})
}

func (h *SalaryStructureHandler) Approve(c *gin.Context) {
	id := c.Param("id")
	res, err := h.uc.Approve(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}

func (h *SalaryStructureHandler) ToggleStatus(c *gin.Context) {
	id := c.Param("id")
	res, err := h.uc.ToggleStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}

func (h *SalaryStructureHandler) GetStats(c *gin.Context) {
	stats, err := h.uc.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

func (h *SalaryStructureHandler) GetFormData(c *gin.Context) {
	formData, err := h.uc.GetFormData(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": formData})
}

func (h *SalaryStructureHandler) ListGrouped(c *gin.Context) {
	var req dto.ListSalaryStructuresRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	groups, total, err := h.uc.ListGrouped(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    groups,
		"meta": gin.H{
			"total":    total,
			"page":     req.Page,
			"per_page": req.PerPage,
		},
	})
}
