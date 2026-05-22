package handler

import (
	"net/http"

	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/usecase"
	"github.com/gin-gonic/gin"
)

type UpCountryCostHandler struct {
	uc usecase.UpCountryCostUsecase
}

func NewUpCountryCostHandler(uc usecase.UpCountryCostUsecase) *UpCountryCostHandler {
	return &UpCountryCostHandler{uc: uc}
}

func (h *UpCountryCostHandler) Create(c *gin.Context) {
	var req dto.CreateUpCountryCostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.uc.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *UpCountryCostHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateUpCountryCostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.uc.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *UpCountryCostHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *UpCountryCostHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *UpCountryCostHandler) List(c *gin.Context) {
	var req dto.ListUpCountryCostsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	items, total, err := h.uc.List(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": items,
		"meta": gin.H{
			"total":    total,
			"page":     req.Page,
			"per_page": req.PerPage,
		},
	})
}

func (h *UpCountryCostHandler) GetStats(c *gin.Context) {
	stats, err := h.uc.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}

func (h *UpCountryCostHandler) Submit(c *gin.Context) {
	id := c.Param("id")
	res, err := h.uc.Submit(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *UpCountryCostHandler) ManagerApprove(c *gin.Context) {
	id := c.Param("id")
	res, err := h.uc.ManagerApprove(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *UpCountryCostHandler) ManagerReject(c *gin.Context) {
	id := c.Param("id")
	var req dto.RejectUpCountryCostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.uc.ManagerReject(c.Request.Context(), id, req.Comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *UpCountryCostHandler) FinanceApprove(c *gin.Context) {
	id := c.Param("id")
	res, err := h.uc.FinanceApprove(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *UpCountryCostHandler) MarkPaid(c *gin.Context) {
	id := c.Param("id")
	res, err := h.uc.MarkPaid(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *UpCountryCostHandler) Approve(c *gin.Context) {
	id := c.Param("id")
	res, err := h.uc.Approve(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

