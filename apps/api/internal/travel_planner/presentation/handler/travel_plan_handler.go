package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/exportjob"
	travelWS "github.com/gilabs/gims/api/internal/core/infrastructure/ws"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/travel_planner/domain/dto"
	"github.com/gilabs/gims/api/internal/travel_planner/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type TravelPlanHandler struct {
	uc       usecase.TravelPlanUsecase
	upgrader websocket.Upgrader
}

func NewTravelPlanHandler(uc usecase.TravelPlanUsecase) *TravelPlanHandler {
	return &TravelPlanHandler{
		uc: uc,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *TravelPlanHandler) Create(c *gin.Context) {
	var req dto.CreateTravelPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.Create(c.Request.Context(), &req)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponseCreated(c, res, nil)
}

func (h *TravelPlanHandler) Update(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))

	var req dto.UpdateTravelPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.Update(c.Request.Context(), id, &req)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *TravelPlanHandler) UpdateParticipants(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))

	var req dto.UpdateTravelPlanParticipantsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.UpdateParticipants(c.Request.Context(), id, req.ParticipantIDs)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *TravelPlanHandler) Delete(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponseDeleted(c, "travel_plan", id, nil)
}

func (h *TravelPlanHandler) GetByID(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *TravelPlanHandler) List(c *gin.Context) {
	var req dto.ListTravelPlansRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	items, total, page, perPage, err := h.uc.List(c.Request.Context(), &req)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	meta := &response.Meta{
		Pagination: response.NewPaginationMeta(page, perPage, int(total)),
	}
	response.SuccessResponse(c, items, meta)
}

func (h *TravelPlanHandler) GetFormData(c *gin.Context) {
	res, err := h.uc.GetFormData(c.Request.Context())
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *TravelPlanHandler) ListParticipants(c *gin.Context) {
	var req dto.ListTravelPlanParticipantsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	items, total, page, perPage, err := h.uc.ListParticipants(c.Request.Context(), &req)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	meta := &response.Meta{
		Pagination: response.NewPaginationMeta(page, perPage, int(total)),
	}
	response.SuccessResponse(c, items, meta)
}

func (h *TravelPlanHandler) SearchPlaces(c *gin.Context) {
	query := c.Query("query")
	provider := c.Query("provider")

	results, err := h.uc.SearchPlaces(c.Request.Context(), query, provider)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponse(c, results, nil)
}

func (h *TravelPlanHandler) GetVisitPlannerFormData(c *gin.Context) {
	var req dto.VisitPlannerFormDataRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.GetVisitPlannerFormData(c.Request.Context(), &req)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *TravelPlanHandler) CreateVisitPlannerPlan(c *gin.Context) {
	var req dto.CreateVisitPlannerPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.CreateVisitPlannerPlan(c.Request.Context(), &req)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponseCreated(c, res, nil)
}

func (h *TravelPlanHandler) OptimizeNavigationForVisit(c *gin.Context) {
	var req dto.OptimizeNavigationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.OptimizeRouteForVisit(c.Request.Context(), &req)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *TravelPlanHandler) UpsertVisitLog(c *gin.Context) {
	var req dto.UpsertVisitLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.UpsertVisitLog(c.Request.Context(), &req)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *TravelPlanHandler) UpsertLocation(c *gin.Context) {
	var req dto.LocationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.UpsertLocation(c.Request.Context(), &req)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponse(c, res, nil)
}

// StartNavigation handles POST /travel/locations/navigation/start.
// Broadcasts a navigation_started WebSocket event so scope-visible supervisors
// see the sales employee begin their route on the live map.
func (h *TravelPlanHandler) StartNavigation(c *gin.Context) {
	var req dto.StartNavigationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.StartNavigation(c.Request.Context(), &req)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponse(c, res, nil)
}

// StopNavigation handles POST /travel/locations/navigation/stop.
// Broadcasts a navigation_stopped WebSocket event to clear the live indicator.
func (h *TravelPlanHandler) StopNavigation(c *gin.Context) {
	var req dto.StopNavigationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.StopNavigation(c.Request.Context(), &req)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *TravelPlanHandler) ListVisitPlannerRoutes(c *gin.Context) {
	var req dto.ListVisitPlannerRoutesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	items, total, page, perPage, err := h.uc.ListVisitPlannerRoutes(c.Request.Context(), &req)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	meta := &response.Meta{
		Pagination: response.NewPaginationMeta(page, perPage, int(total)),
	}
	response.SuccessResponse(c, items, meta)
}

func (h *TravelPlanHandler) TravelLocationsWebSocket(c *gin.Context) {
	reqEmployeeIDs := parseEmployeeIDsQuery(c.Query("employee_ids"))
	visibleEmployeeIDs, err := h.uc.ResolveVisibleEmployeeIDs(c.Request.Context(), reqEmployeeIDs)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	bbox, parseErr := parseBBox(c.Query("area_bbox"))
	if parseErr != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", parseErr.Error(), nil, nil)
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "failed to upgrade websocket", nil, nil)
		return
	}

	filter := travelWS.SubscriptionFilter{BBox: bbox}
	if len(visibleEmployeeIDs) > 0 {
		filter.EmployeeIDs = make(map[string]struct{}, len(visibleEmployeeIDs))
		for _, employeeID := range visibleEmployeeIDs {
			filter.EmployeeIDs[employeeID] = struct{}{}
		}
	}

	hub := travelWS.DefaultLocationHub()
	clientID := hub.Register(conn, filter)
	defer hub.Unregister(clientID)

	for _, snapshot := range hub.Snapshot(filter) {
		if writeErr := conn.WriteJSON(map[string]interface{}{
			"type": "location_update",
			"data": map[string]interface{}{
				"employee_id":   snapshot.EmployeeID,
				"route_id":      snapshot.RouteID,
				"checkpoint_id": snapshot.CheckpointID,
				"lat":           snapshot.Lat,
				"lng":           snapshot.Lng,
				"heading":       snapshot.Heading,
				"timestamp":     snapshot.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			},
		}); writeErr != nil {
			return
		}
	}

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (h *TravelPlanHandler) OptimizeRoute(c *gin.Context) {
	planID := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.OptimizeRoute(c.Request.Context(), planID)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *TravelPlanHandler) GetGoogleMapsLinks(c *gin.Context) {
	planID := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.GetGoogleMapsLinks(c.Request.Context(), planID)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *TravelPlanHandler) ExportPDF(c *gin.Context) {
	planID := strings.TrimSpace(c.Param("id"))
	dayIndexQuery := strings.TrimSpace(c.Query("day_index"))

	var dayIndex *int
	if dayIndexQuery != "" {
		parsedDayIndex, err := strconv.Atoi(dayIndexQuery)
		if err != nil || parsedDayIndex <= 0 {
			response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "day_index must be a positive integer", nil, nil)
			return
		}
		dayIndex = &parsedDayIndex
	}

	generator := func(ctx context.Context) (*exportjob.GeneratedFile, error) {
		pdfBytes, filename, err := h.uc.ExportPDF(ctx, planID, dayIndex)
		if err != nil {
			return nil, err
		}
		return &exportjob.GeneratedFile{
			FileName:    filename,
			ContentType: "application/pdf",
			Bytes:       pdfBytes,
		}, nil
	}

	if exportjob.QueueIfRequested(c, generator) {
		return
	}

	file, err := generator(c.Request.Context())
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	c.Header("Cache-Control", "no-store")
	exportjob.WriteSyncFile(c, file)
}

func (h *TravelPlanHandler) ExportReportHTML(c *gin.Context) {
	planID := strings.TrimSpace(c.Param("id"))

	generator := func(ctx context.Context) (*exportjob.GeneratedFile, error) {
		htmlBytes, filename, err := h.uc.ExportReportHTML(ctx, planID)
		if err != nil {
			return nil, err
		}
		return &exportjob.GeneratedFile{
			FileName:    filename,
			ContentType: "text/html; charset=utf-8",
			Bytes:       htmlBytes,
		}, nil
	}

	if exportjob.QueueIfRequested(c, generator) {
		return
	}

	file, err := generator(c.Request.Context())
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	c.Header("Cache-Control", "no-store")
	exportjob.WriteSyncFile(c, file)
}

func (h *TravelPlanHandler) ListExpenses(c *gin.Context) {
	planID := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.ListExpenses(c.Request.Context(), planID)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *TravelPlanHandler) CreateExpense(c *gin.Context) {
	planID := strings.TrimSpace(c.Param("id"))

	var req dto.CreateTravelExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.CreateExpense(c.Request.Context(), planID, &req)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponseCreated(c, res, nil)
}

func (h *TravelPlanHandler) DeleteExpense(c *gin.Context) {
	planID := strings.TrimSpace(c.Param("id"))
	expenseID := strings.TrimSpace(c.Param("expenseId"))

	if err := h.uc.DeleteExpense(c.Request.Context(), planID, expenseID); err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponseDeleted(c, "travel_expense", expenseID, nil)
}

func (h *TravelPlanHandler) ListVisits(c *gin.Context) {
	planID := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.ListVisits(c.Request.Context(), planID)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *TravelPlanHandler) ListAvailableVisits(c *gin.Context) {
	search := strings.TrimSpace(c.Query("search"))
	res, err := h.uc.ListAvailableVisits(c.Request.Context(), search)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *TravelPlanHandler) LinkVisits(c *gin.Context) {
	planID := strings.TrimSpace(c.Param("id"))

	var req dto.LinkTravelPlanVisitsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	linkedCount, err := h.uc.LinkVisits(c.Request.Context(), planID, &req)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponse(c, map[string]interface{}{"linked_count": linkedCount}, nil)
}

func (h *TravelPlanHandler) UnlinkVisit(c *gin.Context) {
	planID := strings.TrimSpace(c.Param("id"))
	visitID := strings.TrimSpace(c.Param("visitId"))

	if err := h.uc.UnlinkVisit(c.Request.Context(), planID, visitID); err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponse(c, map[string]interface{}{"visit_id": visitID, "travel_plan_id": planID}, nil)
}

func (h *TravelPlanHandler) CreateVisitFromTrip(c *gin.Context) {
	planID := strings.TrimSpace(c.Param("id"))

	var req dto.CreateTravelPlanVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.CreateVisitFromTrip(c.Request.Context(), planID, &req)
	if err != nil {
		handleTravelPlanError(c, err)
		return
	}

	response.SuccessResponseCreated(c, res, nil)
}

func handleTravelPlanError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrTravelPlanNotFound):
		response.ErrorResponse(c, http.StatusNotFound, "TRAVEL_PLAN_NOT_FOUND", err.Error(), nil, nil)
	case errors.Is(err, usecase.ErrTravelExpenseNotFound):
		response.ErrorResponse(c, http.StatusNotFound, "TRAVEL_EXPENSE_NOT_FOUND", err.Error(), nil, nil)
	case errors.Is(err, usecase.ErrVisitNotFound):
		response.ErrorResponse(c, http.StatusNotFound, "VISIT_NOT_FOUND", err.Error(), nil, nil)
	case errors.Is(err, usecase.ErrVisitLogNotFound):
		response.ErrorResponse(c, http.StatusNotFound, "VISIT_NOT_FOUND", err.Error(), nil, nil)
	case errors.Is(err, usecase.ErrPermissionScopeDenied):
		response.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", err.Error(), nil, nil)
	case errors.Is(err, usecase.ErrInvalidTravelMode),
		errors.Is(err, usecase.ErrInvalidDateRange),
		errors.Is(err, usecase.ErrInvalidDayDate),
		errors.Is(err, usecase.ErrInvalidStatus),
		errors.Is(err, usecase.ErrInvalidStopCategory),
		errors.Is(err, usecase.ErrInvalidStopSource),
		errors.Is(err, usecase.ErrInvalidExpenseType),
		errors.Is(err, usecase.ErrInvalidBudgetAmount),
		errors.Is(err, usecase.ErrInvalidSearchQuery),
		errors.Is(err, usecase.ErrEmployeeContextMissing),
		errors.Is(err, usecase.ErrInvalidCoordinate),
		errors.Is(err, usecase.ErrInvalidCheckpoint),
		errors.Is(err, usecase.ErrVisitLogAlreadyCheckedIn),
		errors.Is(err, usecase.ErrVisitLogNotCheckedIn),
		errors.Is(err, usecase.ErrVisitLogAlreadyCheckedOut),
		errors.Is(err, usecase.ErrVisitLocationRequired):
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
	default:
		response.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"internal server error",
			map[string]interface{}{"detail": err.Error()},
			nil,
		)
	}
}

func parseEmployeeIDsQuery(value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return []string{}
	}

	parts := strings.Split(trimmed, ",")
	items := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		normalized := strings.TrimSpace(part)
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		items = append(items, normalized)
	}

	return items
}

func parseBBox(value string) (*travelWS.BoundingBox, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	parts := strings.Split(trimmed, ",")
	if len(parts) != 4 {
		return nil, errors.New("area_bbox must contain minLat,minLng,maxLat,maxLng")
	}

	numbers := make([]float64, 0, 4)
	for _, part := range parts {
		parsed, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return nil, errors.New("area_bbox has invalid numeric value")
		}
		numbers = append(numbers, parsed)
	}

	if numbers[0] < -90 || numbers[0] > 90 || numbers[2] < -90 || numbers[2] > 90 {
		return nil, errors.New("area_bbox latitude must be in range -90..90")
	}
	if numbers[1] < -180 || numbers[1] > 180 || numbers[3] < -180 || numbers[3] > 180 {
		return nil, errors.New("area_bbox longitude must be in range -180..180")
	}

	return &travelWS.BoundingBox{
		MinLat: numbers[0],
		MinLng: numbers[1],
		MaxLat: numbers[2],
		MaxLng: numbers[3],
	}, nil
}
