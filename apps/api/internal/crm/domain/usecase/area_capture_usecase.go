package usecase

import (
	"context"
	"errors"

	"github.com/gilabs/gims/api/internal/core/utils"
	"github.com/gilabs/gims/api/internal/crm/data/repositories"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	"github.com/gilabs/gims/api/internal/crm/domain/mapper"
)

var (
	ErrAreaCaptureNotFound = errors.New("area capture not found")
)

// AreaCaptureUsecase defines the interface for area capture business logic
type AreaCaptureUsecase interface {
	Capture(ctx context.Context, req *dto.CreateAreaCaptureRequest) (*dto.AreaCaptureResponse, error)
	List(ctx context.Context, req *dto.ListAreaCapturesRequest) ([]dto.AreaCaptureResponse, *utils.PaginationResult, error)
	GetHeatmap(ctx context.Context, areaID string) (*dto.HeatmapResponse, error)
	GetCoverage(ctx context.Context) ([]dto.AreaCoverageResponse, error)
}

type areaCaptureUsecase struct {
	captureRepo repositories.AreaCaptureRepository
}

// NewAreaCaptureUsecase creates a new AreaCaptureUsecase
func NewAreaCaptureUsecase(captureRepo repositories.AreaCaptureRepository) AreaCaptureUsecase {
	return &areaCaptureUsecase{
		captureRepo: captureRepo,
	}
}

func (u *areaCaptureUsecase) Capture(ctx context.Context, req *dto.CreateAreaCaptureRequest) (*dto.AreaCaptureResponse, error) {
	capture := mapper.AreaCaptureFromCreateRequest(req)

	if err := u.captureRepo.Create(ctx, capture); err != nil {
		return nil, err
	}

	return mapper.ToAreaCaptureResponse(capture), nil
}

func (u *areaCaptureUsecase) List(ctx context.Context, req *dto.ListAreaCapturesRequest) ([]dto.AreaCaptureResponse, *utils.PaginationResult, error) {
	captures, total, err := u.captureRepo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := mapper.ToAreaCaptureResponses(captures)

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

	pagination := &utils.PaginationResult{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	}

	return responses, pagination, nil
}

func (u *areaCaptureUsecase) GetHeatmap(ctx context.Context, areaID string) (*dto.HeatmapResponse, error) {
	points, err := u.captureRepo.GetHeatmapData(ctx, areaID)
	if err != nil {
		return nil, err
	}

	maxIntensity := 0
	for _, p := range points {
		if p.Intensity > maxIntensity {
			maxIntensity = p.Intensity
		}
	}

	return &dto.HeatmapResponse{
		Points:       points,
		MaxIntensity: maxIntensity,
	}, nil
}

func (u *areaCaptureUsecase) GetCoverage(ctx context.Context) ([]dto.AreaCoverageResponse, error) {
	return u.captureRepo.GetCoverageByArea(ctx)
}
