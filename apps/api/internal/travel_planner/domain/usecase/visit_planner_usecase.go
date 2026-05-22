package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	travelWS "github.com/gilabs/gims/api/internal/core/infrastructure/ws"
	crmModels "github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/travel_planner/domain/dto"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const defaultOSRMBaseURL = "https://router.project-osrm.org"

var (
	ErrPermissionScopeDenied     = errors.New("permission scope denied")
	ErrEmployeeContextMissing    = errors.New("employee context missing")
	ErrInvalidCoordinate         = errors.New("invalid coordinate value")
	ErrVisitLogNotFound          = errors.New("visit log not found")
	ErrVisitLogAlreadyCheckedIn  = errors.New("visit already checked in")
	ErrVisitLogNotCheckedIn      = errors.New("visit is not checked in")
	ErrVisitLogAlreadyCheckedOut = errors.New("visit already checked out")
	ErrVisitLocationRequired     = errors.New("location is required")
)

type resolvedNavigationCheckpoint struct {
	CheckpointID string
	Type         string
	RefID        *string
	Lat          float64
	Lng          float64
	Label        string
}

type visitRouteBucket struct {
	route       dto.ActiveVisitRouteResponse
	coordinates [][2]float64
	routeDate   time.Time
}

type routeMetrics struct {
	Polyline       string
	LegDistancesM  []float64
	LegDurationsS  []float64
	TotalDistanceM float64
	TotalDurationS float64
}

func (uc *travelPlanUsecase) OptimizeRouteForVisit(ctx context.Context, req *dto.OptimizeNavigationRequest) (*dto.OptimizeNavigationResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}
	if len(req.Checkpoints) == 0 {
		return nil, errors.New("checkpoints are required")
	}

	effectiveEmployeeID, err := uc.resolveEffectiveEmployeeID(ctx, req.EmployeeID)
	if err != nil {
		return nil, err
	}

	scope := getPermissionScope(ctx)
	warnings := make([]string, 0)
	resolved := make([]resolvedNavigationCheckpoint, 0, len(req.Checkpoints))

	for idx := range req.Checkpoints {
		normalized, warningText, resolveErr := uc.resolveNavigationCheckpoint(
			ctx,
			req.Checkpoints[idx],
			effectiveEmployeeID,
			scope,
			idx,
		)
		if resolveErr != nil {
			return nil, resolveErr
		}
		if warningText != "" {
			warnings = append(warnings, warningText)
		}
		if normalized != nil {
			resolved = append(resolved, *normalized)
		}
	}

	if len(resolved) == 0 {
		return &dto.OptimizeNavigationResponse{
			OrderedCheckpoints: []dto.OptimizedNavigationCheckpoint{},
			Polyline:           "",
			Summary:            dto.OptimizeNavigationSummary{},
			Warnings:           warnings,
		}, nil
	}

	ordered := optimizeVisitCheckpoints(resolved)

	mode := dto.NavigationOptimizeModeDriving
	if req.Options != nil && req.Options.Mode != "" {
		mode = req.Options.Mode
	}

	orderedResponse := make([]dto.OptimizedNavigationCheckpoint, 0, len(ordered))
	polylinePoints := make([][2]float64, 0, len(ordered))
	for idx := range ordered {
		polylinePoints = append(polylinePoints, [2]float64{ordered[idx].Lat, ordered[idx].Lng})
	}

	metrics := routeMetrics{}
	if len(polylinePoints) > 1 {
		osrmMode := "driving"
		if mode == dto.NavigationOptimizeModeWalking {
			osrmMode = "foot"
		}

		resolvedMetrics, routeErr := uc.requestOSRMRouteMetrics(ctx, polylinePoints, osrmMode)
		if routeErr != nil {
			warnings = append(warnings, "OSRM route unavailable, using straight-line fallback")
		} else {
			metrics = resolvedMetrics
		}
	}

	if len(metrics.LegDistancesM) != len(ordered)-1 || len(metrics.LegDurationsS) != len(ordered)-1 {
		metrics = fallbackRouteMetrics(ordered, mode)
	}

	totalDistance := 0.0
	totalDuration := 0.0

	for idx := range ordered {
		legDistance := 0.0
		if idx > 0 {
			legDistance = metrics.LegDistancesM[idx-1]
		}
		legDuration := 0.0
		if idx > 0 {
			legDuration = metrics.LegDurationsS[idx-1]
		}

		totalDistance += legDistance
		totalDuration += legDuration

		orderedResponse = append(orderedResponse, dto.OptimizedNavigationCheckpoint{
			CheckpointID: ordered[idx].CheckpointID,
			Type:         ordered[idx].Type,
			RefID:        ordered[idx].RefID,
			Lat:          ordered[idx].Lat,
			Lng:          ordered[idx].Lng,
			Sequence:     idx + 1,
			LegDistanceM: int64(math.Round(legDistance)),
			LegDurationS: int64(math.Round(legDuration)),
		})
	}

	polyline := metrics.Polyline
	if polyline == "" {
		polyline = encodePolyline(polylinePoints)
	}

	return &dto.OptimizeNavigationResponse{
		OrderedCheckpoints: orderedResponse,
		Polyline:           polyline,
		Summary: dto.OptimizeNavigationSummary{
			TotalDistanceM: int64(math.Round(totalDistance)),
			TotalDurationS: int64(math.Round(totalDuration)),
		},
		Warnings: warnings,
	}, nil
}

func (uc *travelPlanUsecase) UpsertVisitLog(ctx context.Context, req *dto.UpsertVisitLogRequest) (*dto.VisitLogResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	effectiveEmployeeID, err := uc.resolveEffectiveEmployeeID(ctx, req.EmployeeID)
	if err != nil {
		return nil, err
	}

	action := string(req.Action)
	if action == "" {
		return nil, errors.New("action is required")
	}

	now := apptime.NowForEmployee(effectiveEmployeeID)
	actorID := strings.TrimSpace(getActorID(ctx))

	var savedVisit crmModels.VisitReport
	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		visit, visitErr := uc.findOrCreateVisitForAction(tx, ctx, effectiveEmployeeID, req, now, actorID)
		if visitErr != nil {
			return visitErr
		}

		switch req.Action {
		case dto.VisitActionCheckIn:
			if visit.CheckInAt != nil {
				return ErrVisitLogAlreadyCheckedIn
			}
			if req.Location == nil {
				return ErrVisitLocationRequired
			}

			locationJSON := buildLocationJSON(req.Location.Lat, req.Location.Lng, nil)
			visit.CheckInAt = &now
			visit.CheckInLocation = &locationJSON
			visit.ActualTime = &now
		case dto.VisitActionCheckOut:
			if visit.CheckInAt == nil {
				return ErrVisitLogNotCheckedIn
			}
			if visit.CheckOutAt != nil {
				return ErrVisitLogAlreadyCheckedOut
			}
			if req.Location == nil {
				return ErrVisitLocationRequired
			}

			locationJSON := buildLocationJSON(req.Location.Lat, req.Location.Lng, nil)
			visit.CheckOutAt = &now
			visit.CheckOutLocation = &locationJSON
		case dto.VisitActionSubmitVisit:
			visit.Status = crmModels.VisitReportStatusSubmitted
			if strings.TrimSpace(req.Notes) != "" {
				visit.Notes = strings.TrimSpace(req.Notes)
			}
			if strings.TrimSpace(req.Outcome) != "" {
				visit.Outcome = strings.TrimSpace(req.Outcome)
			}
			if strings.TrimSpace(req.ActivityType) != "" {
				visit.Result = strings.TrimSpace(req.ActivityType)
			}
			if len(req.Photos) > 0 {
				photosJSON, marshalErr := json.Marshal(req.Photos)
				if marshalErr != nil {
					return marshalErr
				}
				photos := string(photosJSON)
				visit.Photos = &photos
			}
			if err := uc.upsertVisitProductInterests(tx, visit.ID, req.ProductInterests); err != nil {
				return err
			}
		default:
			return errors.New("unsupported action")
		}

		if err := tx.Save(visit).Error; err != nil {
			return err
		}

		if visit.LeadID != nil {
			if err := tx.Table("crm_leads").Where("id = ?", *visit.LeadID).Update("updated_at", now).Error; err != nil {
				return err
			}
		}
		if visit.DealID != nil {
			if err := tx.Table("crm_deals").Where("id = ?", *visit.DealID).Update("updated_at", now).Error; err != nil {
				return err
			}
		}

		if err := tx.
			Preload("Customer").
			Preload("Employee").
			Preload("Employee.User").
			Preload("Details").
			Where("id = ?", visit.ID).
			First(&savedVisit).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	statusValue := action
	if req.Action == dto.VisitActionCheckIn {
		statusValue = "checked_in"
	}
	if req.Action == dto.VisitActionCheckOut {
		statusValue = "checked_out"
	}
	if req.Action == dto.VisitActionSubmitVisit {
		statusValue = "submitted"
	}

	travelWS.DefaultLocationHub().PublishRouteStatus(travelWS.RouteStatusUpdate{
		EmployeeID:   effectiveEmployeeID,
		RouteID:      savedVisit.TravelPlanID,
		CheckpointID: req.CheckpointID,
		Status:       statusValue,
		Timestamp:    now,
	})

	visitResponse := mapTravelPlanVisitToResponse(&savedVisit)
	return &dto.VisitLogResponse{
		Action: action,
		Visit:  visitResponse,
	}, nil
}

func (uc *travelPlanUsecase) UpsertLocation(ctx context.Context, req *dto.LocationUpdateRequest) (*dto.LocationUpdateResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}
	if req.Lat < -90 || req.Lat > 90 || req.Lng < -180 || req.Lng > 180 {
		return nil, ErrInvalidCoordinate
	}

	effectiveEmployeeID, err := uc.resolveEffectiveEmployeeID(ctx, req.EmployeeID)
	if err != nil {
		return nil, err
	}

	now := apptime.NowForEmployee(effectiveEmployeeID)

	// Apply throttle: suppress WS broadcast if the employee published within the last 5 s.
	// The REST response is always returned so the mobile client knows the request succeeded.
	if !travelWS.ShouldThrottleLocation(effectiveEmployeeID) {
		name, avatar := uc.resolveEmployeeDisplayInfo(ctx, effectiveEmployeeID)
		travelWS.DefaultLocationHub().PublishLocationUpdate(travelWS.LocationUpdate{
			EmployeeID:        effectiveEmployeeID,
			RouteID:           req.RouteID,
			CheckpointID:      req.CheckpointID,
			Lat:               req.Lat,
			Lng:               req.Lng,
			Heading:           req.Heading,
			Timestamp:         now,
			NavigationStatus:  "navigating",
			EmployeeName:      name,
			EmployeeAvatarURL: avatar,
		})
	}

	return &dto.LocationUpdateResponse{
		EmployeeID:       effectiveEmployeeID,
		RouteID:          req.RouteID,
		CheckpointID:     req.CheckpointID,
		Lat:              req.Lat,
		Lng:              req.Lng,
		Heading:          req.Heading,
		NavigationStatus: "navigating",
		Timestamp:        now.Format(time.RFC3339),
	}, nil
}

// StartNavigation broadcasts a navigation_started WebSocket event to all scope-visible
// supervisors so they can see the sales employee begin their route in real time.
func (uc *travelPlanUsecase) StartNavigation(ctx context.Context, req *dto.StartNavigationRequest) (*dto.NavigationStatusResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}
	if req.Lat < -90 || req.Lat > 90 || req.Lng < -180 || req.Lng > 180 {
		return nil, ErrInvalidCoordinate
	}

	effectiveEmployeeID, err := uc.resolveEffectiveEmployeeID(ctx, req.EmployeeID)
	if err != nil {
		return nil, err
	}

	now := apptime.NowForEmployee(effectiveEmployeeID)
	name, avatar := uc.resolveEmployeeDisplayInfo(ctx, effectiveEmployeeID)

	travelWS.DefaultLocationHub().PublishNavigationUpdate(travelWS.NavigationUpdate{
		EmployeeID:        effectiveEmployeeID,
		RouteID:           req.RouteID,
		Lat:               req.Lat,
		Lng:               req.Lng,
		Heading:           req.Heading,
		Status:            "navigating",
		Timestamp:         now,
		EmployeeName:      name,
		EmployeeAvatarURL: avatar,
	})

	lat := req.Lat
	lng := req.Lng
	return &dto.NavigationStatusResponse{
		EmployeeID: effectiveEmployeeID,
		RouteID:    req.RouteID,
		Lat:        &lat,
		Lng:        &lng,
		Status:     "navigating",
		Timestamp:  now.Format(time.RFC3339),
	}, nil
}

// StopNavigation broadcasts a navigation_stopped WebSocket event, clearing the live
// navigation indicator for supervisors in the visit-planner map view.
func (uc *travelPlanUsecase) StopNavigation(ctx context.Context, req *dto.StopNavigationRequest) (*dto.NavigationStatusResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	effectiveEmployeeID, err := uc.resolveEffectiveEmployeeID(ctx, req.EmployeeID)
	if err != nil {
		return nil, err
	}

	now := apptime.NowForEmployee(effectiveEmployeeID)
	name, avatar := uc.resolveEmployeeDisplayInfo(ctx, effectiveEmployeeID)

	travelWS.DefaultLocationHub().PublishNavigationUpdate(travelWS.NavigationUpdate{
		EmployeeID:        effectiveEmployeeID,
		RouteID:           req.RouteID,
		Lat:               0,
		Lng:               0,
		Status:            "idle",
		Timestamp:         now,
		EmployeeName:      name,
		EmployeeAvatarURL: avatar,
	})

	return &dto.NavigationStatusResponse{
		EmployeeID: effectiveEmployeeID,
		RouteID:    req.RouteID,
		Status:     "idle",
		Timestamp:  now.Format(time.RFC3339),
	}, nil
}

// resolveEmployeeDisplayInfo fetches the employee's display name and avatar URL for
// enriching WebSocket events.  A best-effort query is used — empty strings are returned
// on any error so the caller never fails due to display enrichment.
func (uc *travelPlanUsecase) resolveEmployeeDisplayInfo(ctx context.Context, employeeID string) (name string, avatarURL string) {
	if strings.TrimSpace(employeeID) == "" {
		return "", ""
	}

	var row struct {
		Name      string  `gorm:"column:name"`
		AvatarURL *string `gorm:"column:profile_photo_url"`
	}

	if err := uc.db.WithContext(ctx).
		Table("employees").
		Select("COALESCE(name, '') AS name, profile_photo_url").
		Where("id = ? AND deleted_at IS NULL", employeeID).
		Limit(1).
		Scan(&row).Error; err != nil {
		return "", ""
	}

	if row.AvatarURL != nil {
		avatarURL = *row.AvatarURL
	}
	return row.Name, avatarURL
}

func (uc *travelPlanUsecase) upsertVisitProductInterests(tx *gorm.DB, visitID string, items []dto.VisitProductInterestInput) error {
	if strings.TrimSpace(visitID) == "" {
		return nil
	}

	if err := tx.Where("visit_report_id = ?", visitID).Delete(&crmModels.VisitReportDetail{}).Error; err != nil {
		return err
	}

	if len(items) == 0 {
		return nil
	}

	for idx := range items {
		item := items[idx]
		productID := strings.TrimSpace(item.ProductID)
		if productID == "" {
			continue
		}

		var exists bool
		if err := tx.Table("products").
			Select("COUNT(1) > 0").
			Where("id = ? AND deleted_at IS NULL", productID).
			Scan(&exists).Error; err != nil {
			return err
		}
		if !exists {
			continue
		}

		detail := crmModels.VisitReportDetail{
			VisitReportID: visitID,
			ProductID:     productID,
			InterestLevel: item.InterestLevel,
			Notes:         strings.TrimSpace(item.Notes),
			Quantity:      item.Quantity,
			Price:         item.Price,
		}

		if err := tx.Create(&detail).Error; err != nil {
			return err
		}
	}

	return nil
}

func (uc *travelPlanUsecase) ListVisitPlannerRoutes(ctx context.Context, req *dto.ListVisitPlannerRoutesRequest) ([]dto.ActiveVisitRouteResponse, int64, int, int, error) {
	if req == nil {
		req = &dto.ListVisitPlannerRoutesRequest{}
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	query := uc.db.WithContext(ctx).
		Model(&crmModels.VisitReport{}).
		Preload("Employee").
		Preload("Employee.User").
		Preload("Customer").
		Preload("Lead").
		Preload("Deal").
		Preload("Deal.Customer").
		Preload("Details")

	query = security.ApplyScopeFilter(query, ctx, security.HRDScopeQueryOptions())

	if req.EmployeeID != nil && strings.TrimSpace(*req.EmployeeID) != "" {
		if err := uc.validateEmployeeScope(ctx, strings.TrimSpace(*req.EmployeeID)); err != nil {
			return nil, 0, page, perPage, err
		}
		query = query.Where("employee_id = ?", strings.TrimSpace(*req.EmployeeID))
	}
	if req.DivisionID != nil && strings.TrimSpace(*req.DivisionID) != "" {
		query = query.Where("employee_id IN (SELECT id FROM employees WHERE division_id = ? AND deleted_at IS NULL)", strings.TrimSpace(*req.DivisionID))
	}

	routeDateRaw := strings.TrimSpace(req.RouteDate)
	if routeDateRaw != "" {
		routeDate, parseErr := time.Parse("2006-01-02", routeDateRaw)
		if parseErr != nil {
			return nil, 0, page, perPage, parseErr
		}
		startDate := time.Date(routeDate.Year(), routeDate.Month(), routeDate.Day(), 0, 0, 0, 0, routeDate.Location())
		endDate := startDate.AddDate(0, 0, 1)
		query = query.Where("visit_date >= ? AND visit_date < ?", startDate, endDate)
	}

	visits := make([]crmModels.VisitReport, 0)
	if err := query.Order("visit_date ASC").Order("created_at ASC").Limit(1000).Find(&visits).Error; err != nil {
		return nil, 0, page, perPage, err
	}

	if len(visits) == 0 {
		return []dto.ActiveVisitRouteResponse{}, 0, page, perPage, nil
	}

	buckets := make(map[string]*visitRouteBucket)
	for idx := range visits {
		visit := visits[idx]
		visitDateKey := visit.VisitDate.Format("2006-01-02")
		bucketKey := visit.EmployeeID + ":" + visitDateKey

		bucket, exists := buckets[bucketKey]
		if !exists {
			employeeName := ""
			employeeAvatar := ""
			if visit.Employee != nil {
				employeeName = strings.TrimSpace(visit.Employee.Name)
				if visit.Employee.User != nil {
					employeeAvatar = strings.TrimSpace(visit.Employee.User.AvatarURL)
				}
			}
			bucket = &visitRouteBucket{
				route: dto.ActiveVisitRouteResponse{
					RouteID:           visit.ID,
					PlanCode:          "VP-" + visit.VisitDate.Format("20060102"),
					PlanTitle:         fmt.Sprintf("Visit Navigation %s", visitDateKey),
					EmployeeID:        visit.EmployeeID,
					EmployeeName:      employeeName,
					EmployeeAvatarURL: employeeAvatar,
					Checkpoints:       make([]dto.ActiveVisitRouteCheckpoint, 0),
				},
				coordinates: make([][2]float64, 0),
				routeDate:   visit.VisitDate,
			}
			buckets[bucketKey] = bucket
		}

		checkpointType := "customer"
		var refID *string
		label := "Checkpoint"
		lat, lng, warningText := resolveVisitReferenceLocation(&visit)

		if visit.LeadID != nil {
			checkpointType = "lead"
			refID = visit.LeadID
		}
		if visit.DealID != nil {
			checkpointType = "deal"
			refID = visit.DealID
		}
		if visit.CustomerID != nil {
			checkpointType = "customer"
			refID = visit.CustomerID
		}

		if visit.Lead != nil {
			label = strings.TrimSpace(visit.Lead.CompanyName)
			if label == "" {
				label = strings.TrimSpace(strings.TrimSpace(visit.Lead.FirstName) + " " + strings.TrimSpace(visit.Lead.LastName))
			}
		}
		if visit.Deal != nil {
			if strings.TrimSpace(visit.Deal.Title) != "" {
				label = strings.TrimSpace(visit.Deal.Title)
			}
		}
		if visit.Customer != nil && strings.TrimSpace(visit.Customer.Name) != "" {
			label = strings.TrimSpace(visit.Customer.Name)
		}

		status := "pending"
		if visit.CheckOutAt != nil {
			status = "completed"
			bucket.route.CompletedTotal++
		} else if visit.CheckInAt != nil {
			status = "in_progress"
			bucket.route.InProgressTotal++
		}

		checkpoint := dto.ActiveVisitRouteCheckpoint{
			VisitID:              visit.ID,
			CheckpointID:         visit.ID,
			Type:                 checkpointType,
			RefID:                refID,
			Label:                label,
			Status:               status,
			Warning:              warningText,
			ProductInterestCount: len(visit.Details),
			DocumentationCount:   countVisitPhotos(visit.Photos),
		}
		if lat != nil && lng != nil {
			checkpoint.Lat = lat
			checkpoint.Lng = lng
			bucket.coordinates = append(bucket.coordinates, [2]float64{*lat, *lng})
		}

		bucket.route.Checkpoints = append(bucket.route.Checkpoints, checkpoint)
		bucket.route.CheckpointTotal++
	}

	routes := make([]dto.ActiveVisitRouteResponse, 0, len(buckets))
	for _, bucket := range buckets {
		if bucket.route.CheckpointTotal == 0 {
			continue
		}
		if bucket.route.CompletedTotal >= bucket.route.CheckpointTotal {
			continue
		}
		metrics, routeErr := uc.requestOSRMRouteMetrics(ctx, bucket.coordinates, "driving")
		if routeErr != nil || metrics.Polyline == "" {
			bucket.route.Polyline = encodePolyline(bucket.coordinates)
		} else {
			bucket.route.Polyline = metrics.Polyline
			bucket.route.CurrentETAS = int64(math.Round(metrics.TotalDurationS))
		}
		routes = append(routes, bucket.route)
	}

	sort.SliceStable(routes, func(i, j int) bool {
		if routes[i].EmployeeID == routes[j].EmployeeID {
			return routes[i].PlanCode < routes[j].PlanCode
		}
		return routes[i].EmployeeName < routes[j].EmployeeName
	})

	total := int64(len(routes))
	start := (page - 1) * perPage
	if start >= len(routes) {
		return []dto.ActiveVisitRouteResponse{}, total, page, perPage, nil
	}
	end := start + perPage
	if end > len(routes) {
		end = len(routes)
	}

	return routes[start:end], total, page, perPage, nil
}

func (uc *travelPlanUsecase) GetVisitPlannerFormData(ctx context.Context, req *dto.VisitPlannerFormDataRequest) (*dto.VisitPlannerFormDataResponse, error) {
	if req == nil {
		req = &dto.VisitPlannerFormDataRequest{}
	}

	search := strings.TrimSpace(req.Search)
	scope := getPermissionScope(ctx)
	scopeEmployeeID := strings.TrimSpace(getScopeEmployeeID(ctx))
	scopeDivisionID := strings.TrimSpace(getScopeDivisionID(ctx))

	employeeFilter, err := uc.resolveEmployeeFilterByScope(ctx, req.EmployeeID)
	if err != nil {
		return nil, err
	}

	employees, err := uc.listVisibleEmployees(ctx, scope, scopeEmployeeID, scopeDivisionID, search)
	if err != nil {
		return nil, err
	}

	warnings := make([]string, 0)

	leadRows := make([]struct {
		ID          string   `gorm:"column:id"`
		AssignedTo  *string  `gorm:"column:assigned_to"`
		CompanyName string   `gorm:"column:company_name"`
		FirstName   string   `gorm:"column:first_name"`
		LastName    string   `gorm:"column:last_name"`
		Latitude    *float64 `gorm:"column:latitude"`
		Longitude   *float64 `gorm:"column:longitude"`
	}, 0)

	leadQuery := uc.db.WithContext(ctx).
		Table("crm_leads AS l").
		Select("l.id, l.assigned_to, l.company_name, l.first_name, l.last_name, l.latitude, l.longitude").
		Where("l.deleted_at IS NULL")

	if search != "" {
		like := search + "%"
		leadQuery = leadQuery.Where("l.company_name ILIKE ? OR l.first_name ILIKE ? OR l.last_name ILIKE ?", like, like, like)
	}
	leadQuery = applyAssignmentScopeFilter(leadQuery, scope, scopeEmployeeID, scopeDivisionID, employeeFilter)
	if err := leadQuery.Order("l.updated_at DESC").Limit(100).Find(&leadRows).Error; err != nil {
		warnings = append(warnings, "Lead options are temporarily unavailable")
		leadRows = leadRows[:0]
	}

	leads := make([]dto.VisitPlannerCandidate, 0, len(leadRows))
	for idx := range leadRows {
		label := strings.TrimSpace(leadRows[idx].CompanyName)
		if label == "" {
			label = strings.TrimSpace(strings.TrimSpace(leadRows[idx].FirstName) + " " + strings.TrimSpace(leadRows[idx].LastName))
		}
		if label == "" {
			label = "Lead"
		}
		hasLocation := leadRows[idx].Latitude != nil && leadRows[idx].Longitude != nil
		warningText := ""
		if !hasLocation {
			warningText = fmt.Sprintf("%s has no location", label)
			warnings = append(warnings, warningText)
		}

		leads = append(leads, dto.VisitPlannerCandidate{
			ID:          leadRows[idx].ID,
			Type:        "lead",
			Label:       label,
			AssignedTo:  leadRows[idx].AssignedTo,
			Lat:         leadRows[idx].Latitude,
			Lng:         leadRows[idx].Longitude,
			HasLocation: hasLocation,
			Warning:     warningText,
		})
	}

	dealRows := make([]struct {
		ID         string   `gorm:"column:id"`
		AssignedTo *string  `gorm:"column:assigned_to"`
		Title      string   `gorm:"column:title"`
		Latitude   *float64 `gorm:"column:latitude"`
		Longitude  *float64 `gorm:"column:longitude"`
	}, 0)

	dealQuery := uc.db.WithContext(ctx).
		Table("crm_deals AS d").
		Select("d.id, d.assigned_to, d.title, COALESCE(c.latitude, l.latitude) AS latitude, COALESCE(c.longitude, l.longitude) AS longitude").
		Joins("LEFT JOIN customers AS c ON c.id = d.customer_id AND c.deleted_at IS NULL").
		Joins("LEFT JOIN crm_leads AS l ON l.id = d.lead_id AND l.deleted_at IS NULL").
		Where("d.deleted_at IS NULL")

	if search != "" {
		like := search + "%"
		dealQuery = dealQuery.Where("d.title ILIKE ?", like)
	}
	dealQuery = applyAssignmentScopeFilter(dealQuery, scope, scopeEmployeeID, scopeDivisionID, employeeFilter)
	if err := dealQuery.Order("d.updated_at DESC").Limit(100).Find(&dealRows).Error; err != nil {
		warnings = append(warnings, "Deal options are temporarily unavailable")
		dealRows = dealRows[:0]
	}

	deals := make([]dto.VisitPlannerCandidate, 0, len(dealRows))
	for idx := range dealRows {
		label := strings.TrimSpace(dealRows[idx].Title)
		if label == "" {
			label = "Deal"
		}
		hasLocation := dealRows[idx].Latitude != nil && dealRows[idx].Longitude != nil
		warningText := ""
		if !hasLocation {
			warningText = fmt.Sprintf("%s has no location", label)
			warnings = append(warnings, warningText)
		}

		deals = append(deals, dto.VisitPlannerCandidate{
			ID:          dealRows[idx].ID,
			Type:        "deal",
			Label:       label,
			AssignedTo:  dealRows[idx].AssignedTo,
			Lat:         dealRows[idx].Latitude,
			Lng:         dealRows[idx].Longitude,
			HasLocation: hasLocation,
			Warning:     warningText,
		})
	}

	customerRows := make([]struct {
		ID        string   `gorm:"column:id"`
		Name      string   `gorm:"column:name"`
		Latitude  *float64 `gorm:"column:latitude"`
		Longitude *float64 `gorm:"column:longitude"`
	}, 0)
	customerQuery := uc.db.WithContext(ctx).
		Table("customers").
		Select("id, name, latitude, longitude").
		Where("deleted_at IS NULL")
	if search != "" {
		like := search + "%"
		customerQuery = customerQuery.Where("name ILIKE ?", like)
	}
	if err := customerQuery.Order("updated_at DESC").Limit(100).Find(&customerRows).Error; err != nil {
		warnings = append(warnings, "Customer options are temporarily unavailable")
		customerRows = customerRows[:0]
	}

	customers := make([]dto.VisitPlannerCandidate, 0, len(customerRows))
	for idx := range customerRows {
		label := strings.TrimSpace(customerRows[idx].Name)
		if label == "" {
			label = "Customer"
		}
		hasLocation := customerRows[idx].Latitude != nil && customerRows[idx].Longitude != nil
		warningText := ""
		if !hasLocation {
			warningText = fmt.Sprintf("%s has no location", label)
			warnings = append(warnings, warningText)
		}
		customers = append(customers, dto.VisitPlannerCandidate{
			ID:          customerRows[idx].ID,
			Type:        "customer",
			Label:       label,
			Lat:         customerRows[idx].Latitude,
			Lng:         customerRows[idx].Longitude,
			HasLocation: hasLocation,
			Warning:     warningText,
		})
	}

	productRows := make([]struct {
		ID           string  `gorm:"column:id"`
		Code         string  `gorm:"column:code"`
		Name         string  `gorm:"column:name"`
		SellingPrice float64 `gorm:"column:selling_price"`
	}, 0)

	productQuery := uc.db.WithContext(ctx).
		Table("products").
		Select("id, code, name, COALESCE(selling_price, 0) AS selling_price").
		Where("deleted_at IS NULL")

	if search != "" {
		like := search + "%"
		productQuery = productQuery.Where("name ILIKE ? OR code ILIKE ?", like, like)
	}

	if err := productQuery.Order("updated_at DESC").Limit(100).Find(&productRows).Error; err != nil {
		warnings = append(warnings, "Product options are temporarily unavailable")
		productRows = productRows[:0]
	}

	products := make([]dto.VisitPlannerProductOption, 0, len(productRows))
	for idx := range productRows {
		products = append(products, dto.VisitPlannerProductOption{
			ID:           productRows[idx].ID,
			Code:         productRows[idx].Code,
			Name:         productRows[idx].Name,
			SellingPrice: productRows[idx].SellingPrice,
		})
	}

	return &dto.VisitPlannerFormDataResponse{
		Employees: employees,
		Leads:     leads,
		Deals:     deals,
		Customers: customers,
		Products:  products,
		Warnings:  dedupeStrings(warnings),
	}, nil
}

func (uc *travelPlanUsecase) CreateVisitPlannerPlan(ctx context.Context, req *dto.CreateVisitPlannerPlanRequest) (*dto.CreateVisitPlannerPlanResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}
	if len(req.Checkpoints) == 0 {
		return nil, errors.New("checkpoints are required")
	}

	effectiveEmployeeID, err := uc.resolveEffectiveEmployeeID(ctx, req.EmployeeID)
	if err != nil {
		return nil, err
	}

	routeDate, err := parseDate(req.RouteDate)
	if err != nil {
		return nil, err
	}

	scope := getPermissionScope(ctx)
	resolved := make([]resolvedNavigationCheckpoint, 0, len(req.Checkpoints))
	for idx := range req.Checkpoints {
		normalized, _, resolveErr := uc.resolveNavigationCheckpoint(ctx, req.Checkpoints[idx], effectiveEmployeeID, scope, idx)
		if resolveErr != nil {
			return nil, resolveErr
		}
		if normalized != nil {
			resolved = append(resolved, *normalized)
		}
	}
	if len(resolved) == 0 {
		return nil, ErrInvalidCheckpoint
	}

	ordered := optimizeVisitCheckpoints(resolved)

	planTitle := strings.TrimSpace(req.Title)
	if planTitle == "" {
		planTitle = fmt.Sprintf("Visit Navigation %s", routeDate.Format("2006-01-02"))
	}
	planCode := "VP-" + routeDate.Format("20060102")

	actorID := strings.TrimSpace(getActorID(ctx))
	createdBy := pointerIfNotEmpty(actorID)
	visitIDs := make([]string, 0, len(ordered))

	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for idx := range ordered {
			checkpoint := ordered[idx]

			visitCode, codeErr := uc.generateVisitCode(ctx, apptime.Now().Add(time.Duration(idx)*time.Second))
			if codeErr != nil {
				return codeErr
			}

			visit := crmModels.VisitReport{
				Code:       visitCode,
				EmployeeID: effectiveEmployeeID,
				VisitDate:  routeDate,
				Purpose:    "Visit navigation checkpoint",
				Status:     crmModels.VisitReportStatusDraft,
				CreatedBy:  createdBy,
				Latitude:   &checkpoint.Lat,
				Longitude:  &checkpoint.Lng,
			}

			switch checkpoint.Type {
			case "lead":
				visit.LeadID = checkpoint.RefID
			case "deal":
				visit.DealID = checkpoint.RefID
			case "customer":
				visit.CustomerID = checkpoint.RefID
			}

			if err := tx.Create(&visit).Error; err != nil {
				return err
			}

			visitIDs = append(visitIDs, visit.ID)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	routeID := ""
	if len(visitIDs) > 0 {
		routeID = visitIDs[0]
	}

	return &dto.CreateVisitPlannerPlanResponse{
		RouteID:         routeID,
		PlanCode:        planCode,
		PlanTitle:       planTitle,
		EmployeeID:      effectiveEmployeeID,
		CheckpointTotal: len(ordered),
		VisitIDs:        visitIDs,
	}, nil
}

func (uc *travelPlanUsecase) ResolveVisibleEmployeeIDs(ctx context.Context, requested []string) ([]string, error) {
	scope := getPermissionScope(ctx)
	scopeEmployeeID := strings.TrimSpace(getScopeEmployeeID(ctx))
	scopeDivisionID := strings.TrimSpace(getScopeDivisionID(ctx))

	if len(requested) == 0 {
		switch scope {
		case "OWN":
			if scopeEmployeeID == "" {
				return nil, ErrEmployeeContextMissing
			}
			return []string{scopeEmployeeID}, nil
		case "DIVISION", "AREA":
			if scopeDivisionID == "" {
				if scopeEmployeeID == "" {
					return nil, ErrEmployeeContextMissing
				}
				return []string{scopeEmployeeID}, nil
			}
			ids, err := uc.listEmployeeIDsByDivision(ctx, scopeDivisionID)
			if err != nil {
				return nil, err
			}
			return ids, nil
		default:
			return []string{}, nil
		}
	}

	clean := dedupeStrings(requested)
	if len(clean) == 0 {
		return []string{}, nil
	}

	switch scope {
	case "OWN":
		if scopeEmployeeID == "" {
			return nil, ErrEmployeeContextMissing
		}
		for _, employeeID := range clean {
			if employeeID != scopeEmployeeID {
				return nil, ErrPermissionScopeDenied
			}
		}
		return []string{scopeEmployeeID}, nil
	case "DIVISION", "AREA":
		allowed := make(map[string]struct{})
		if scopeDivisionID != "" {
			ids, err := uc.listEmployeeIDsByDivision(ctx, scopeDivisionID)
			if err != nil {
				return nil, err
			}
			for _, id := range ids {
				allowed[id] = struct{}{}
			}
		}
		if scopeEmployeeID != "" {
			allowed[scopeEmployeeID] = struct{}{}
		}
		filtered := make([]string, 0, len(clean))
		for _, employeeID := range clean {
			if _, exists := allowed[employeeID]; !exists {
				return nil, ErrPermissionScopeDenied
			}
			filtered = append(filtered, employeeID)
		}
		return filtered, nil
	default:
		return clean, nil
	}
}

func (uc *travelPlanUsecase) resolveEffectiveEmployeeID(ctx context.Context, requested *string) (string, error) {
	scope := getPermissionScope(ctx)
	scopeEmployeeID := strings.TrimSpace(getScopeEmployeeID(ctx))
	requestedEmployeeID := strings.TrimSpace(valueOrEmpty(requested))

	switch scope {
	case "OWN":
		if scopeEmployeeID == "" {
			return "", ErrEmployeeContextMissing
		}
		if requestedEmployeeID != "" && requestedEmployeeID != scopeEmployeeID {
			return "", ErrPermissionScopeDenied
		}
		return scopeEmployeeID, nil
	case "DIVISION", "AREA":
		if requestedEmployeeID == "" {
			if scopeEmployeeID == "" {
				return "", ErrEmployeeContextMissing
			}
			return scopeEmployeeID, nil
		}
		if err := uc.validateEmployeeScope(ctx, requestedEmployeeID); err != nil {
			return "", err
		}
		return requestedEmployeeID, nil
	default:
		if requestedEmployeeID != "" {
			return requestedEmployeeID, nil
		}
		if scopeEmployeeID != "" {
			return scopeEmployeeID, nil
		}
		return "", ErrEmployeeContextMissing
	}
}

func (uc *travelPlanUsecase) validateEmployeeScope(ctx context.Context, targetEmployeeID string) error {
	scope := getPermissionScope(ctx)
	scopeEmployeeID := strings.TrimSpace(getScopeEmployeeID(ctx))
	scopeDivisionID := strings.TrimSpace(getScopeDivisionID(ctx))
	targetEmployeeID = strings.TrimSpace(targetEmployeeID)

	if targetEmployeeID == "" {
		return ErrEmployeeContextMissing
	}

	switch scope {
	case "OWN":
		if scopeEmployeeID == "" || scopeEmployeeID != targetEmployeeID {
			return ErrPermissionScopeDenied
		}
	case "DIVISION", "AREA":
		if scopeDivisionID == "" {
			if scopeEmployeeID != targetEmployeeID {
				return ErrPermissionScopeDenied
			}
			return nil
		}

		var exists bool
		err := uc.db.WithContext(ctx).
			Table("employees").
			Select("COUNT(1) > 0").
			Where("id = ? AND division_id = ? AND deleted_at IS NULL", targetEmployeeID, scopeDivisionID).
			Scan(&exists).Error
		if err != nil {
			return err
		}
		if !exists {
			return ErrPermissionScopeDenied
		}
	}

	return nil
}

func (uc *travelPlanUsecase) resolveNavigationCheckpoint(
	ctx context.Context,
	input dto.NavigationCheckpointInput,
	effectiveEmployeeID string,
	scope string,
	index int,
) (*resolvedNavigationCheckpoint, string, error) {
	checkpointID := strings.TrimSpace(valueOrEmpty(input.ID))
	if checkpointID == "" {
		checkpointID = fmt.Sprintf("checkpoint-%d", index+1)
	}

	hasCoords := input.Lat != nil && input.Lng != nil
	if hasCoords {
		if *input.Lat < -90 || *input.Lat > 90 || *input.Lng < -180 || *input.Lng > 180 {
			return nil, "", ErrInvalidCoordinate
		}
	}

	refID := ""
	if input.RefID != nil {
		refID = strings.TrimSpace(*input.RefID)
	}
	typeName := normalizeCheckpointType(input.Type)

	if refID == "" {
		if !hasCoords {
			return nil, "", fmt.Errorf("%w: checkpoint %d requires ref_id or coordinates", ErrInvalidCheckpoint, index+1)
		}
		return &resolvedNavigationCheckpoint{
			CheckpointID: checkpointID,
			Type:         typeName,
			RefID:        nil,
			Lat:          *input.Lat,
			Lng:          *input.Lng,
			Label:        typeName,
		}, "", nil
	}

	switch typeName {
	case "lead":
		row := struct {
			ID         string   `gorm:"column:id"`
			AssignedTo *string  `gorm:"column:assigned_to"`
			Company    string   `gorm:"column:company_name"`
			FirstName  string   `gorm:"column:first_name"`
			LastName   string   `gorm:"column:last_name"`
			Latitude   *float64 `gorm:"column:latitude"`
			Longitude  *float64 `gorm:"column:longitude"`
		}{}

		err := uc.db.WithContext(ctx).
			Table("crm_leads").
			Select("id, assigned_to, company_name, first_name, last_name, latitude, longitude").
			Where("id = ? AND deleted_at IS NULL", refID).
			First(&row).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, "", fmt.Errorf("%w: lead not found", ErrInvalidCheckpoint)
			}
			return nil, "", err
		}

		if scope != "ALL" && effectiveEmployeeID != "" {
			if row.AssignedTo != nil && strings.TrimSpace(*row.AssignedTo) != effectiveEmployeeID {
				return nil, "", fmt.Errorf("%w: lead is assigned to another employee", ErrInvalidCheckpoint)
			}
		}

		lat := 0.0
		lng := 0.0
		if hasCoords {
			lat = *input.Lat
			lng = *input.Lng
		} else if row.Latitude != nil && row.Longitude != nil {
			lat = *row.Latitude
			lng = *row.Longitude
		}

		if !hasCoords && row.Latitude == nil && row.Longitude == nil {
			return nil, "", fmt.Errorf("%w: lead location is missing", ErrInvalidCheckpoint)
		}

		return &resolvedNavigationCheckpoint{
			CheckpointID: checkpointID,
			Type:         typeName,
			RefID:        &refID,
			Lat:          lat,
			Lng:          lng,
			Label:        strings.TrimSpace(row.Company),
		}, "", nil
	case "deal":
		row := struct {
			ID         string   `gorm:"column:id"`
			AssignedTo *string  `gorm:"column:assigned_to"`
			Title      string   `gorm:"column:title"`
			Latitude   *float64 `gorm:"column:latitude"`
			Longitude  *float64 `gorm:"column:longitude"`
		}{}

		err := uc.db.WithContext(ctx).
			Table("crm_deals AS d").
			Select("d.id, d.assigned_to, d.title, COALESCE(c.latitude, l.latitude) AS latitude, COALESCE(c.longitude, l.longitude) AS longitude").
			Joins("LEFT JOIN customers AS c ON c.id = d.customer_id AND c.deleted_at IS NULL").
			Joins("LEFT JOIN crm_leads AS l ON l.id = d.lead_id AND l.deleted_at IS NULL").
			Where("d.id = ? AND d.deleted_at IS NULL", refID).
			First(&row).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, "", fmt.Errorf("%w: deal not found", ErrInvalidCheckpoint)
			}
			return nil, "", err
		}

		if scope != "ALL" && effectiveEmployeeID != "" {
			if row.AssignedTo != nil && strings.TrimSpace(*row.AssignedTo) != effectiveEmployeeID {
				return nil, "", fmt.Errorf("%w: deal is assigned to another employee", ErrInvalidCheckpoint)
			}
		}

		lat := 0.0
		lng := 0.0
		if hasCoords {
			lat = *input.Lat
			lng = *input.Lng
		} else if row.Latitude != nil && row.Longitude != nil {
			lat = *row.Latitude
			lng = *row.Longitude
		}

		if !hasCoords && row.Latitude == nil && row.Longitude == nil {
			return nil, "", fmt.Errorf("%w: deal location is missing", ErrInvalidCheckpoint)
		}

		return &resolvedNavigationCheckpoint{
			CheckpointID: checkpointID,
			Type:         typeName,
			RefID:        &refID,
			Lat:          lat,
			Lng:          lng,
			Label:        strings.TrimSpace(row.Title),
		}, "", nil
	case "customer":
		row := struct {
			ID        string   `gorm:"column:id"`
			Name      string   `gorm:"column:name"`
			Latitude  *float64 `gorm:"column:latitude"`
			Longitude *float64 `gorm:"column:longitude"`
		}{}

		err := uc.db.WithContext(ctx).
			Table("customers").
			Select("id, name, latitude, longitude").
			Where("id = ? AND deleted_at IS NULL", refID).
			First(&row).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, "", fmt.Errorf("%w: customer not found", ErrInvalidCheckpoint)
			}
			return nil, "", err
		}

		lat := 0.0
		lng := 0.0
		if hasCoords {
			lat = *input.Lat
			lng = *input.Lng
		} else if row.Latitude != nil && row.Longitude != nil {
			lat = *row.Latitude
			lng = *row.Longitude
		}

		if !hasCoords && row.Latitude == nil && row.Longitude == nil {
			return nil, "", fmt.Errorf("%w: customer location is missing", ErrInvalidCheckpoint)
		}

		return &resolvedNavigationCheckpoint{
			CheckpointID: checkpointID,
			Type:         typeName,
			RefID:        &refID,
			Lat:          lat,
			Lng:          lng,
			Label:        strings.TrimSpace(row.Name),
		}, "", nil
	default:
		return nil, "", fmt.Errorf("%w: unsupported checkpoint type %q", ErrInvalidCheckpoint, strings.TrimSpace(input.Type))
	}
}

func normalizeCheckpointType(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "pipeline" {
		return "deal"
	}

	return normalized
}

func (uc *travelPlanUsecase) findOrCreateVisitForAction(
	tx *gorm.DB,
	ctx context.Context,
	effectiveEmployeeID string,
	req *dto.UpsertVisitLogRequest,
	now time.Time,
	actorID string,
) (*crmModels.VisitReport, error) {
	if req.VisitID != nil && strings.TrimSpace(*req.VisitID) != "" {
		visit := &crmModels.VisitReport{}
		query := tx.WithContext(ctx).Model(&crmModels.VisitReport{})
		query = security.ApplyScopeFilter(query, ctx, security.HRDScopeQueryOptions())
		if err := query.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", strings.TrimSpace(*req.VisitID)).
			First(visit).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrVisitLogNotFound
			}
			return nil, err
		}
		if visit.EmployeeID != effectiveEmployeeID {
			return nil, ErrPermissionScopeDenied
		}
		return visit, nil
	}

	query := tx.WithContext(ctx).
		Model(&crmModels.VisitReport{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("employee_id = ?", effectiveEmployeeID)

	if req.LeadID != nil && strings.TrimSpace(*req.LeadID) != "" {
		query = query.Where("lead_id = ?", strings.TrimSpace(*req.LeadID))
	}
	if req.DealID != nil && strings.TrimSpace(*req.DealID) != "" {
		query = query.Where("deal_id = ?", strings.TrimSpace(*req.DealID))
	}
	if req.CustomerID != nil && strings.TrimSpace(*req.CustomerID) != "" {
		query = query.Where("customer_id = ?", strings.TrimSpace(*req.CustomerID))
	}

	query = query.Where("check_out_at IS NULL")

	visit := &crmModels.VisitReport{}
	if err := query.Order("created_at DESC").First(visit).Error; err == nil {
		return visit, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if req.Action == dto.VisitActionCheckOut {
		return nil, ErrVisitLogNotFound
	}

	if err := uc.validateLeadAndDealOwnership(tx, ctx, effectiveEmployeeID, req.LeadID, req.DealID); err != nil {
		return nil, err
	}

	code, err := uc.generateVisitCode(ctx, now)
	if err != nil {
		return nil, err
	}

	visitDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	newVisit := &crmModels.VisitReport{
		Code:       code,
		VisitDate:  visitDate,
		EmployeeID: effectiveEmployeeID,
		LeadID:     req.LeadID,
		DealID:     req.DealID,
		CustomerID: req.CustomerID,
		Status:     crmModels.VisitReportStatusDraft,
		Purpose:    "Visit checkpoint",
		Notes:      strings.TrimSpace(req.Notes),
		CreatedBy:  pointerIfNotEmpty(actorID),
	}

	if err := tx.Create(newVisit).Error; err != nil {
		return nil, err
	}

	return newVisit, nil
}

func (uc *travelPlanUsecase) validateLeadAndDealOwnership(
	tx *gorm.DB,
	ctx context.Context,
	effectiveEmployeeID string,
	leadID *string,
	dealID *string,
) error {
	scope := getPermissionScope(ctx)
	if scope == "ALL" {
		return nil
	}

	if leadID != nil && strings.TrimSpace(*leadID) != "" {
		var assignedTo *string
		err := tx.WithContext(ctx).
			Table("crm_leads").
			Select("assigned_to").
			Where("id = ? AND deleted_at IS NULL", strings.TrimSpace(*leadID)).
			Scan(&assignedTo).Error
		if err != nil {
			return err
		}
		if assignedTo == nil || strings.TrimSpace(*assignedTo) != effectiveEmployeeID {
			return ErrPermissionScopeDenied
		}
	}

	if dealID != nil && strings.TrimSpace(*dealID) != "" {
		var assignedTo *string
		err := tx.WithContext(ctx).
			Table("crm_deals").
			Select("assigned_to").
			Where("id = ? AND deleted_at IS NULL", strings.TrimSpace(*dealID)).
			Scan(&assignedTo).Error
		if err != nil {
			return err
		}
		if assignedTo == nil || strings.TrimSpace(*assignedTo) != effectiveEmployeeID {
			return ErrPermissionScopeDenied
		}
	}

	return nil
}

func (uc *travelPlanUsecase) resolveEmployeeFilterByScope(ctx context.Context, requested *string) (*string, error) {
	scope := getPermissionScope(ctx)
	scopeEmployeeID := strings.TrimSpace(getScopeEmployeeID(ctx))
	requestedEmployeeID := strings.TrimSpace(valueOrEmpty(requested))

	switch scope {
	case "OWN":
		if scopeEmployeeID == "" {
			return nil, ErrEmployeeContextMissing
		}
		return &scopeEmployeeID, nil
	case "DIVISION", "AREA":
		if requestedEmployeeID == "" {
			return nil, nil
		}
		if err := uc.validateEmployeeScope(ctx, requestedEmployeeID); err != nil {
			return nil, err
		}
		return &requestedEmployeeID, nil
	default:
		if requestedEmployeeID == "" {
			return nil, nil
		}
		return &requestedEmployeeID, nil
	}
}

func (uc *travelPlanUsecase) listVisibleEmployees(
	ctx context.Context,
	scope string,
	scopeEmployeeID string,
	scopeDivisionID string,
	search string,
) ([]dto.EmployeeFormOption, error) {
	rows := make([]struct {
		ID           string `gorm:"column:id"`
		EmployeeCode string `gorm:"column:employee_code"`
		Name         string `gorm:"column:name"`
		AvatarURL    string `gorm:"column:avatar_url"`
	}, 0)

	query := uc.db.WithContext(ctx).
		Table("employees AS e").
		Select("e.id, e.employee_code, e.name, COALESCE(u.avatar_url, '') AS avatar_url").
		Joins("LEFT JOIN users AS u ON u.id = e.user_id").
		Where("e.deleted_at IS NULL").
		Where("e.is_active = ?", true)

	if search != "" {
		like := search + "%"
		query = query.Where("e.name ILIKE ? OR e.employee_code ILIKE ?", like, like)
	}

	switch scope {
	case "OWN":
		if scopeEmployeeID == "" {
			return []dto.EmployeeFormOption{}, nil
		}
		query = query.Where("e.id = ?", scopeEmployeeID)
	case "DIVISION", "AREA":
		if scopeDivisionID != "" {
			query = query.Where("e.division_id = ?", scopeDivisionID)
		} else if scopeEmployeeID != "" {
			query = query.Where("e.id = ?", scopeEmployeeID)
		}
	}

	if err := query.Order("e.name ASC").Limit(100).Find(&rows).Error; err != nil {
		fallbackRows := make([]struct {
			ID           string `gorm:"column:id"`
			EmployeeCode string `gorm:"column:employee_code"`
			Name         string `gorm:"column:name"`
		}, 0)

		fallbackQuery := uc.db.WithContext(ctx).
			Table("employees AS e").
			Select("e.id, e.employee_code, e.name").
			Where("e.deleted_at IS NULL").
			Where("e.is_active = ?", true)

		if search != "" {
			like := search + "%"
			fallbackQuery = fallbackQuery.Where("e.name ILIKE ? OR e.employee_code ILIKE ?", like, like)
		}

		switch scope {
		case "OWN":
			if scopeEmployeeID == "" {
				return []dto.EmployeeFormOption{}, nil
			}
			fallbackQuery = fallbackQuery.Where("e.id = ?", scopeEmployeeID)
		case "DIVISION", "AREA":
			if scopeDivisionID != "" {
				fallbackQuery = fallbackQuery.Where("e.division_id = ?", scopeDivisionID)
			} else if scopeEmployeeID != "" {
				fallbackQuery = fallbackQuery.Where("e.id = ?", scopeEmployeeID)
			}
		}

		if fallbackErr := fallbackQuery.Order("e.name ASC").Limit(100).Find(&fallbackRows).Error; fallbackErr != nil {
			return nil, fallbackErr
		}

		items := make([]dto.EmployeeFormOption, 0, len(fallbackRows))
		for idx := range fallbackRows {
			items = append(items, dto.EmployeeFormOption{
				ID:           fallbackRows[idx].ID,
				EmployeeCode: fallbackRows[idx].EmployeeCode,
				Name:         fallbackRows[idx].Name,
				AvatarURL:    "",
			})
		}

		return items, nil
	}

	items := make([]dto.EmployeeFormOption, 0, len(rows))
	for idx := range rows {
		items = append(items, dto.EmployeeFormOption{
			ID:           rows[idx].ID,
			EmployeeCode: rows[idx].EmployeeCode,
			Name:         rows[idx].Name,
			AvatarURL:    rows[idx].AvatarURL,
		})
	}

	return items, nil
}

func (uc *travelPlanUsecase) listEmployeeIDsByDivision(ctx context.Context, divisionID string) ([]string, error) {
	ids := make([]string, 0)
	if err := uc.db.WithContext(ctx).
		Table("employees").
		Where("division_id = ? AND deleted_at IS NULL", divisionID).
		Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return dedupeStrings(ids), nil
}

// countVisitPhotos parses the JSONB photos field and returns the number of photo URLs.
func countVisitPhotos(photos *string) int {
	if photos == nil {
		return 0
	}
	raw := strings.TrimSpace(*photos)
	if raw == "" || raw == "null" {
		return 0
	}
	var urls []string
	if err := json.Unmarshal([]byte(raw), &urls); err != nil {
		return 0
	}
	return len(urls)
}

func resolveVisitReferenceLocation(visit *crmModels.VisitReport) (*float64, *float64, string) {
	if visit == nil {
		return nil, nil, ""
	}
	if visit.Latitude != nil && visit.Longitude != nil {
		return visit.Latitude, visit.Longitude, ""
	}
	if visit.Customer != nil && visit.Customer.Latitude != nil && visit.Customer.Longitude != nil {
		return visit.Customer.Latitude, visit.Customer.Longitude, ""
	}
	if visit.Lead != nil && visit.Lead.Latitude != nil && visit.Lead.Longitude != nil {
		return visit.Lead.Latitude, visit.Lead.Longitude, ""
	}
	if visit.Deal != nil && visit.Deal.Customer != nil && visit.Deal.Customer.Latitude != nil && visit.Deal.Customer.Longitude != nil {
		return visit.Deal.Customer.Latitude, visit.Deal.Customer.Longitude, ""
	}
	return nil, nil, "missing location"
}

func optimizeVisitCheckpoints(input []resolvedNavigationCheckpoint) []resolvedNavigationCheckpoint {
	if len(input) <= 2 {
		result := append([]resolvedNavigationCheckpoint{}, input...)
		return result
	}

	remaining := append([]resolvedNavigationCheckpoint{}, input...)
	optimized := make([]resolvedNavigationCheckpoint, 0, len(input))

	optimized = append(optimized, remaining[0])
	remaining = remaining[1:]

	for len(remaining) > 0 {
		last := optimized[len(optimized)-1]
		nearestIndex := 0
		nearestDistance := haversineMeters(last.Lat, last.Lng, remaining[0].Lat, remaining[0].Lng)

		for idx := 1; idx < len(remaining); idx++ {
			distance := haversineMeters(last.Lat, last.Lng, remaining[idx].Lat, remaining[idx].Lng)
			if distance < nearestDistance {
				nearestDistance = distance
				nearestIndex = idx
			}
		}

		optimized = append(optimized, remaining[nearestIndex])
		remaining = append(remaining[:nearestIndex], remaining[nearestIndex+1:]...)
	}

	return optimized
}

func haversineMeters(lat1 float64, lng1 float64, lat2 float64, lng2 float64) float64 {
	const earthRadiusM = 6371000.0

	toRadians := func(value float64) float64 {
		return value * math.Pi / 180
	}

	dLat := toRadians(lat2 - lat1)
	dLng := toRadians(lng2 - lng1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusM * c
}

func fallbackRouteMetrics(points []resolvedNavigationCheckpoint, mode dto.NavigationOptimizeMode) routeMetrics {
	metrics := routeMetrics{
		LegDistancesM: make([]float64, 0, max(0, len(points)-1)),
		LegDurationsS: make([]float64, 0, max(0, len(points)-1)),
	}

	speedMetersPerSecond := 11.11
	if mode == dto.NavigationOptimizeModeWalking {
		speedMetersPerSecond = 1.39
	}

	for idx := 1; idx < len(points); idx++ {
		distance := haversineMeters(points[idx-1].Lat, points[idx-1].Lng, points[idx].Lat, points[idx].Lng)
		duration := 0.0
		if distance > 0 {
			duration = distance / speedMetersPerSecond
		}

		metrics.LegDistancesM = append(metrics.LegDistancesM, distance)
		metrics.LegDurationsS = append(metrics.LegDurationsS, duration)
		metrics.TotalDistanceM += distance
		metrics.TotalDurationS += duration
	}

	coords := make([][2]float64, 0, len(points))
	for idx := range points {
		coords = append(coords, [2]float64{points[idx].Lat, points[idx].Lng})
	}
	metrics.Polyline = encodePolyline(coords)

	return metrics
}

func (uc *travelPlanUsecase) requestOSRMRouteMetrics(ctx context.Context, coordinates [][2]float64, mode string) (routeMetrics, error) {
	if len(coordinates) < 2 {
		return routeMetrics{}, errors.New("at least two coordinates are required")
	}

	osrmBase := strings.TrimSpace(os.Getenv("OSRM_BASE_URL"))
	if osrmBase == "" {
		osrmBase = defaultOSRMBaseURL
	}
	osrmBase = strings.TrimSuffix(osrmBase, "/")

	coordParts := make([]string, 0, len(coordinates))
	for idx := range coordinates {
		coordParts = append(coordParts, fmt.Sprintf("%.6f,%.6f", coordinates[idx][1], coordinates[idx][0]))
	}

	requestURL := fmt.Sprintf(
		"%s/route/v1/%s/%s?overview=full&geometries=polyline&annotations=distance,duration&steps=false",
		osrmBase,
		mode,
		strings.Join(coordParts, ";"),
	)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return routeMetrics{}, err
	}

	response, err := uc.httpClient.Do(request)
	if err != nil {
		return routeMetrics{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return routeMetrics{}, fmt.Errorf("osrm route status %d", response.StatusCode)
	}

	var payload struct {
		Code   string `json:"code"`
		Routes []struct {
			Distance float64 `json:"distance"`
			Duration float64 `json:"duration"`
			Geometry string  `json:"geometry"`
			Legs     []struct {
				Distance float64 `json:"distance"`
				Duration float64 `json:"duration"`
			} `json:"legs"`
		} `json:"routes"`
	}

	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return routeMetrics{}, err
	}

	if len(payload.Routes) == 0 {
		return routeMetrics{}, errors.New("osrm route not found")
	}

	best := payload.Routes[0]
	metrics := routeMetrics{
		Polyline:       strings.TrimSpace(best.Geometry),
		LegDistancesM:  make([]float64, 0, len(best.Legs)),
		LegDurationsS:  make([]float64, 0, len(best.Legs)),
		TotalDistanceM: best.Distance,
		TotalDurationS: best.Duration,
	}

	for idx := range best.Legs {
		metrics.LegDistancesM = append(metrics.LegDistancesM, best.Legs[idx].Distance)
		metrics.LegDurationsS = append(metrics.LegDurationsS, best.Legs[idx].Duration)
	}

	return metrics, nil
}

func encodePolyline(points [][2]float64) string {
	if len(points) == 0 {
		return ""
	}

	var encoded strings.Builder
	prevLat := 0
	prevLng := 0

	for _, point := range points {
		lat := int(math.Round(point[0] * 1e5))
		lng := int(math.Round(point[1] * 1e5))

		encoded.WriteString(encodePolylineComponent(lat - prevLat))
		encoded.WriteString(encodePolylineComponent(lng - prevLng))

		prevLat = lat
		prevLng = lng
	}

	return encoded.String()
}

func encodePolylineComponent(value int) string {
	shifted := value << 1
	if value < 0 {
		shifted = ^shifted
	}

	var builder strings.Builder
	for shifted >= 0x20 {
		builder.WriteByte(byte((0x20 | (shifted & 0x1f)) + 63))
		shifted >>= 5
	}
	builder.WriteByte(byte(shifted + 63))

	return builder.String()
}

func applyAssignmentScopeFilter(
	query *gorm.DB,
	scope string,
	scopeEmployeeID string,
	scopeDivisionID string,
	employeeFilter *string,
) *gorm.DB {
	if employeeFilter != nil && strings.TrimSpace(*employeeFilter) != "" {
		return query.Where("assigned_to = ?", strings.TrimSpace(*employeeFilter))
	}

	switch scope {
	case "OWN":
		if scopeEmployeeID != "" {
			return query.Where("assigned_to = ?", scopeEmployeeID)
		}
	case "DIVISION", "AREA":
		if scopeDivisionID != "" {
			return query.Where("assigned_to IN (SELECT id FROM employees WHERE division_id = ? AND deleted_at IS NULL)", scopeDivisionID)
		}
		if scopeEmployeeID != "" {
			return query.Where("assigned_to = ?", scopeEmployeeID)
		}
	}

	return query
}

func getPermissionScope(ctx context.Context) string {
	scope, _ := ctx.Value("permission_scope").(string)
	normalized := strings.ToUpper(strings.TrimSpace(scope))
	if normalized == "" {
		return "ALL"
	}
	return normalized
}

func getScopeEmployeeID(ctx context.Context) string {
	employeeID, _ := ctx.Value("scope_employee_id").(string)
	return normalizeUUIDOrEmpty(employeeID)
}

func getScopeDivisionID(ctx context.Context) string {
	divisionID, _ := ctx.Value("scope_division_id").(string)
	return normalizeUUIDOrEmpty(divisionID)
}

func normalizeUUIDOrEmpty(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if _, err := uuid.Parse(trimmed); err != nil {
		return ""
	}
	return trimmed
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func pointerIfNotEmpty(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func dedupeStrings(input []string) []string {
	if len(input) == 0 {
		return []string{}
	}

	result := make([]string, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for _, value := range input {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}

	return result
}

func buildLocationJSON(lat float64, lng float64, accuracy *float64) string {
	payload := map[string]interface{}{
		"lat": lat,
		"lng": lng,
	}
	if accuracy != nil {
		payload["accuracy"] = *accuracy
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}

	return string(encoded)
}
