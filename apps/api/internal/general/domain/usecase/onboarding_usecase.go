package usecase

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/general/data/repositories"
	"github.com/gilabs/gims/api/internal/general/domain/dto"
)

// OnboardingUsecase manages tenant onboarding state.
type OnboardingUsecase interface {
	GetState(ctx context.Context) (*dto.OnboardingStateResponse, error)
	SetBusinessType(ctx context.Context, businessType string) (*dto.OnboardingStateResponse, error)
	MarkComplete(ctx context.Context) (*dto.OnboardingStateResponse, error)
}

type onboardingUsecase struct {
	repo      repositories.OnboardingRepository
	publisher OnboardingPublisher
}

type OnboardingPublisher interface {
	Publish(tenantID string, eventType string, payload map[string]interface{})
}

func NewOnboardingUsecase(repo repositories.OnboardingRepository) OnboardingUsecase {
	return &onboardingUsecase{repo: repo}
}

// WithOnboardingPublisher attaches an optional realtime publisher for onboarding state updates.
func WithOnboardingPublisher(uc OnboardingUsecase, publisher OnboardingPublisher) OnboardingUsecase {
	impl, ok := uc.(*onboardingUsecase)
	if !ok {
		return uc
	}
	impl.publisher = publisher
	return impl
}

// stepsFromRepo queries live data and converts it to the DTO shape.
// On error it returns nil steps (non-fatal — caller should still return the base state).
func (u *onboardingUsecase) stepsFromRepo(ctx context.Context) *dto.OnboardingStepsResponse {
	raw, err := u.repo.CheckSteps(ctx)
	if err != nil {
		return nil
	}
	return &dto.OnboardingStepsResponse{
		Company:     raw.Company,
		Outlet:      raw.Outlet,
		FloorLayout: raw.FloorLayout,
		Products:    raw.Products,
		Warehouse:   raw.Warehouse,
		Users:       raw.Users,
		FiscalYear:  raw.FiscalYear,
	}
}

func (u *onboardingUsecase) GetState(ctx context.Context) (*dto.OnboardingStateResponse, error) {
	state, err := u.repo.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.OnboardingStateResponse{
		BusinessType: state.BusinessType,
		Completed:    state.Completed,
		Steps:        u.stepsFromRepo(ctx),
	}, nil
}

func (u *onboardingUsecase) SetBusinessType(ctx context.Context, businessType string) (*dto.OnboardingStateResponse, error) {
	state, err := u.repo.Get(ctx)
	if err != nil {
		return nil, err
	}

	state.BusinessType = businessType

	if err := u.repo.Save(ctx, state); err != nil {
		return nil, err
	}

	u.publishState(ctx, state.BusinessType, state.Completed)

	return &dto.OnboardingStateResponse{
		BusinessType: state.BusinessType,
		Completed:    state.Completed,
		Steps:        u.stepsFromRepo(ctx),
	}, nil
}

func (u *onboardingUsecase) MarkComplete(ctx context.Context) (*dto.OnboardingStateResponse, error) {
	state, err := u.repo.Get(ctx)
	if err != nil {
		return nil, err
	}

	state.Completed = true

	if err := u.repo.Save(ctx, state); err != nil {
		return nil, err
	}

	u.publishState(ctx, state.BusinessType, state.Completed)

	return &dto.OnboardingStateResponse{
		BusinessType: state.BusinessType,
		Completed:    state.Completed,
		Steps:        u.stepsFromRepo(ctx),
	}, nil
}

func (u *onboardingUsecase) publishState(ctx context.Context, businessType string, completed bool) {
	if u.publisher == nil {
		return
	}

	tenantID := middleware.TenantFromContext(ctx)
	if tenantID == "" {
		return
	}

	u.publisher.Publish(tenantID, "general.onboarding.updated", map[string]interface{}{
		"business_type": businessType,
		"completed":     completed,
	})
}
