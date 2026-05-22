package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/data/repositories"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	"github.com/gilabs/gims/api/internal/crm/domain/mapper"
	orgRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	"github.com/google/uuid"
)

// ScheduleUsecase defines business logic for CRM schedules
type ScheduleUsecase interface {
	Create(ctx context.Context, req dto.CreateScheduleRequest, createdBy string) (dto.ScheduleResponse, error)
	GetByID(ctx context.Context, id string) (dto.ScheduleResponse, error)
	List(ctx context.Context, params repositories.ScheduleListParams) ([]dto.ScheduleResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateScheduleRequest) (dto.ScheduleResponse, error)
	Delete(ctx context.Context, id string) error
	GetFormData(ctx context.Context) (*dto.ScheduleFormDataResponse, error)
}

type scheduleUsecase struct {
	scheduleRepo repositories.ScheduleRepository
	taskRepo     repositories.TaskRepository
	employeeRepo orgRepos.EmployeeRepository
}

// NewScheduleUsecase creates a new schedule usecase
func NewScheduleUsecase(
	scheduleRepo repositories.ScheduleRepository,
	taskRepo repositories.TaskRepository,
	employeeRepo orgRepos.EmployeeRepository,
) ScheduleUsecase {
	return &scheduleUsecase{
		scheduleRepo: scheduleRepo,
		taskRepo:     taskRepo,
		employeeRepo: employeeRepo,
	}
}

func (u *scheduleUsecase) Create(ctx context.Context, req dto.CreateScheduleRequest, createdBy string) (dto.ScheduleResponse, error) {
	// Validate employee
	if _, err := u.employeeRepo.FindByID(ctx, req.EmployeeID); err != nil {
		return dto.ScheduleResponse{}, errors.New("employee not found")
	}

	// Validate task if provided
	if req.TaskID != nil && *req.TaskID != "" {
		if _, err := u.taskRepo.FindByID(ctx, *req.TaskID); err != nil {
			return dto.ScheduleResponse{}, errors.New("task not found")
		}
	}

	// Parse scheduled_at
	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		return dto.ScheduleResponse{}, errors.New("invalid scheduled_at format, use ISO 8601")
	}

	// Parse end_at
	var endAt *time.Time
	if req.EndAt != nil && *req.EndAt != "" {
		t, err := time.Parse(time.RFC3339, *req.EndAt)
		if err != nil {
			return dto.ScheduleResponse{}, errors.New("invalid end_at format, use ISO 8601")
		}
		if t.Before(scheduledAt) {
			return dto.ScheduleResponse{}, errors.New("end_at must be after scheduled_at")
		}
		endAt = &t
	}

	reminderMinutes := 30
	if req.ReminderMinutesBefore != nil {
		reminderMinutes = *req.ReminderMinutesBefore
	}

	schedule := &models.Schedule{
		ID:                    uuid.New().String(),
		TaskID:                req.TaskID,
		EmployeeID:            req.EmployeeID,
		Title:                 req.Title,
		Description:           req.Description,
		ScheduledAt:           scheduledAt,
		EndAt:                 endAt,
		Status:                "pending",
		ReminderMinutesBefore: reminderMinutes,
		CreatedBy:             &createdBy,
	}

	if err := u.scheduleRepo.Create(ctx, schedule); err != nil {
		return dto.ScheduleResponse{}, fmt.Errorf("failed to create schedule: %w", err)
	}

	created, err := u.scheduleRepo.FindByID(ctx, schedule.ID)
	if err != nil {
		return dto.ScheduleResponse{}, err
	}
	return mapper.ToScheduleResponse(created), nil
}

func (u *scheduleUsecase) GetByID(ctx context.Context, id string) (dto.ScheduleResponse, error) {
	schedule, err := u.scheduleRepo.FindByID(ctx, id)
	if err != nil {
		return dto.ScheduleResponse{}, errors.New("schedule not found")
	}
	return mapper.ToScheduleResponse(schedule), nil
}

func (u *scheduleUsecase) List(ctx context.Context, params repositories.ScheduleListParams) ([]dto.ScheduleResponse, int64, error) {
	schedules, total, err := u.scheduleRepo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToScheduleResponseList(schedules), total, nil
}

func (u *scheduleUsecase) Update(ctx context.Context, id string, req dto.UpdateScheduleRequest) (dto.ScheduleResponse, error) {
	schedule, err := u.scheduleRepo.FindByID(ctx, id)
	if err != nil {
		return dto.ScheduleResponse{}, errors.New("schedule not found")
	}

	if req.TaskID != nil {
		if *req.TaskID != "" {
			if _, err := u.taskRepo.FindByID(ctx, *req.TaskID); err != nil {
				return dto.ScheduleResponse{}, errors.New("task not found")
			}
		}
		schedule.TaskID = req.TaskID
	}
	if req.EmployeeID != nil {
		if _, err := u.employeeRepo.FindByID(ctx, *req.EmployeeID); err != nil {
			return dto.ScheduleResponse{}, errors.New("employee not found")
		}
		schedule.EmployeeID = *req.EmployeeID
	}
	if req.Title != nil {
		schedule.Title = *req.Title
	}
	if req.Description != nil {
		schedule.Description = *req.Description
	}
	if req.ScheduledAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ScheduledAt)
		if err != nil {
			return dto.ScheduleResponse{}, errors.New("invalid scheduled_at format, use ISO 8601")
		}
		schedule.ScheduledAt = t
	}
	if req.EndAt != nil {
		if *req.EndAt == "" {
			schedule.EndAt = nil
		} else {
			t, err := time.Parse(time.RFC3339, *req.EndAt)
			if err != nil {
				return dto.ScheduleResponse{}, errors.New("invalid end_at format, use ISO 8601")
			}
			schedule.EndAt = &t
		}
	}
	if req.Status != nil {
		schedule.Status = *req.Status
	}
	if req.ReminderMinutesBefore != nil {
		schedule.ReminderMinutesBefore = *req.ReminderMinutesBefore
	}

	if err := u.scheduleRepo.Update(ctx, schedule); err != nil {
		return dto.ScheduleResponse{}, fmt.Errorf("failed to update schedule: %w", err)
	}

	updated, err := u.scheduleRepo.FindByID(ctx, schedule.ID)
	if err != nil {
		return dto.ScheduleResponse{}, err
	}
	return mapper.ToScheduleResponse(updated), nil
}

func (u *scheduleUsecase) Delete(ctx context.Context, id string) error {
	if _, err := u.scheduleRepo.FindByID(ctx, id); err != nil {
		return errors.New("schedule not found")
	}
	return u.scheduleRepo.Delete(ctx, id)
}

func (u *scheduleUsecase) GetFormData(ctx context.Context) (*dto.ScheduleFormDataResponse, error) {
	// Fetch employees
	employees, _, err := u.employeeRepo.List(ctx, orgRepos.EmployeeListParams{
		Page: 1, PerPage: 500, SortBy: "name", SortDir: "ASC",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch employees: %w", err)
	}
	employeeOptions := make([]dto.ScheduleEmployeeOption, 0, len(employees))
	for _, emp := range employees {
		employeeOptions = append(employeeOptions, dto.ScheduleEmployeeOption{
			ID: emp.ID, EmployeeCode: emp.EmployeeCode, Name: emp.Name,
		})
	}

	// Fetch open tasks for linking
	tasks, _, err := u.taskRepo.List(ctx, repositories.TaskListParams{
		Limit: 500, SortBy: "created_at", SortDir: "DESC",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks: %w", err)
	}
	taskOptions := make([]dto.ScheduleTaskOption, 0, len(tasks))
	for _, t := range tasks {
		taskOptions = append(taskOptions, dto.ScheduleTaskOption{
			ID: t.ID, Title: t.Title, Status: t.Status, Priority: t.Priority,
		})
	}

	return &dto.ScheduleFormDataResponse{
		Employees: employeeOptions,
		Tasks:     taskOptions,
	}, nil
}
