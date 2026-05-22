package handler

import (
	"errors"
	"io"
	"log"
	"net/http"

	"encoding/json"

	"github.com/gilabs/gims/api/internal/auth/domain/usecase"
	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gin-gonic/gin"
)

// XenditHandler handles inbound Xendit payment callbacks.
type XenditHandler struct {
	authUC usecase.AuthUsecase
}

func NewXenditHandler(authUC usecase.AuthUsecase) *XenditHandler {
	return &XenditHandler{authUC: authUC}
}

// xenditWebhookBody is the relevant subset of the Xendit invoice callback payload.
type xenditWebhookBody struct {
	ID         string `json:"id"`          // Xendit invoice ID (used for recurring renewal lookup)
	ExternalID string `json:"external_id"` // Set to pending-reg token for new registrations
	Status     string `json:"status"`      // "PAID", "EXPIRED", "SETTLED"
}

// InvoicePaid is called by Xendit when a payment invoice status changes.
// Xendit verifies the request by including the configured webhook token in the
// x-callback-token header. We verify it here before processing the payload.
//
// Route: POST /webhooks/xendit/invoice
func (h *XenditHandler) InvoicePaid(c *gin.Context) {
	// ── 1. Verify Xendit callback token ──────────────────────────────────────
	if config.AppConfig == nil || config.AppConfig.Xendit.WebhookToken == "" {
		coreErrors.InternalServerErrorResponse(c, "webhook not configured")
		return
	}

	incomingToken := c.GetHeader("x-callback-token")
	if incomingToken != config.AppConfig.Xendit.WebhookToken {
		c.Status(http.StatusUnauthorized)
		return
	}

	// ── 2. Parse body ─────────────────────────────────────────────────────────
	bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		coreErrors.InternalServerErrorResponse(c, "cannot read request body")
		return
	}

	var payload xenditWebhookBody
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		coreErrors.ErrorResponse(c, "INVALID_PAYLOAD", nil, nil)
		return
	}

	// Only act on PAID / SETTLED events — ignore EXPIRED and others.
	if payload.Status != "PAID" && payload.Status != "SETTLED" {
		response.SuccessResponse(c, gin.H{"message": "event acknowledged"}, nil)
		return
	}

	if payload.ExternalID == "" {
		coreErrors.ErrorResponse(c, "MISSING_EXTERNAL_ID", nil, nil)
		return
	}

	// ── 3. Route by payload type ──────────────────────────────────────────────
	// Try completing a pending registration first. If Redis has no matching token
	// (e.g. this is a recurring renewal, not a new sign-up), fall back to the
	// renewal path using the Xendit invoice ID from the payload body.
	reqCtx := c.Request.Context()
	if err := h.authUC.CompletePendingRegistration(reqCtx, payload.ExternalID); err != nil {
		if errors.Is(err, usecase.ErrPendingRegistrationDataInvalid) {
			log.Printf("[xendit] invalid pending registration payload for external_id=%s: %v", payload.ExternalID, err)
		}

		if billingErr := h.authUC.CompletePendingBillingChange(reqCtx, payload.ExternalID); billingErr == nil {
			response.SuccessResponse(c, gin.H{"message": "ok"}, nil)
			return
		} else if !errors.Is(billingErr, usecase.ErrPendingBillingChangeNotFound) && !errors.Is(billingErr, usecase.ErrPendingBillingChangeInvalid) {
			log.Printf("[xendit] failed billing change external_id=%s: %v", payload.ExternalID, billingErr)
			response.SuccessResponse(c, gin.H{"message": "event acknowledged"}, nil)
			return
		}

		// No pending registration matched — treat as a recurring subscription renewal.
		if payload.ID != "" {
			if renewErr := h.authUC.HandleRecurringRenewal(reqCtx, payload.ID); renewErr != nil {
				log.Printf("[xendit] failed recurring renewal invoice_id=%s external_id=%s: %v", payload.ID, payload.ExternalID, renewErr)
				response.SuccessResponse(c, gin.H{"message": "event acknowledged"}, nil)
				return
			}
			response.SuccessResponse(c, gin.H{"message": "event acknowledged"}, nil)
			return
		}

		log.Printf("[xendit] failed pending registration external_id=%s: %v", payload.ExternalID, err)
		response.SuccessResponse(c, gin.H{"message": "event acknowledged"}, nil)
		return
	}

	response.SuccessResponse(c, gin.H{"message": "ok"}, nil)
}
