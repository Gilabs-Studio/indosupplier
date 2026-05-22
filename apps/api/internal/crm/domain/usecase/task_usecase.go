package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/data/repositories"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	"github.com/gilabs/gims/api/internal/crm/domain/mapper"
	customerRepos "github.com/gilabs/gims/api/internal/customer/data/repositories"
	orgRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	"github.com/google/uuid"
)

// TaskUsecase defines business logic for CRM tasks
type TaskUsecase interface {
	Create(ctx context.Context, req dto.CreateTaskRequest, createdBy string) (dto.TaskResponse, error)
	GetByID(ctx context.Context, id string) (dto.TaskResponse, error)
	List(ctx context.Context, params repositories.TaskListParams) ([]dto.TaskResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateTaskRequest) (dto.TaskResponse, error)
	Delete(ctx context.Context, id string) error
	Assign(ctx context.Context, id string, req dto.AssignTaskRequest, assignedFrom string) (dto.TaskResponse, error)
	Complete(ctx context.Context, id string) (dto.TaskResponse, error)
	MarkInProgress(ctx context.Context, id string) (dto.TaskResponse, error)
	Cancel(ctx context.Context, id string) (dto.TaskResponse, error)
	GetFormData(ctx context.Context) (*dto.TaskFormDataResponse, error)
	// Reminder nested CRUD
	ListReminders(ctx context.Context, taskID string) ([]dto.ReminderResponse, error)
	GetReminderByID(ctx context.Context, taskID string, reminderID string) (dto.ReminderResponse, error)
	CreateReminder(ctx context.Context, taskID string, req dto.CreateReminderRequest, createdBy string) (dto.ReminderResponse, error)
	UpdateReminder(ctx context.Context, taskID string, reminderID string, req dto.UpdateReminderRequest) (dto.ReminderResponse, error)
	DeleteReminder(ctx context.Context, taskID string, reminderID string) error
}

type taskUsecase struct {
	taskRepo     repositories.TaskRepository
	scheduleRepo repositories.ScheduleRepository
	reminderRepo repositories.ReminderRepository
	contactRepo  repositories.ContactRepository
	dealRepo     repositories.DealRepository
	leadRepo     repositories.LeadRepository
	customerRepo customerRepos.CustomerRepository
	employeeRepo orgRepos.EmployeeRepository
}

// NewTaskUsecase creates a new task usecase
func NewTaskUsecase(
	taskRepo repositories.TaskRepository,
	scheduleRepo repositories.ScheduleRepository,
	reminderRepo repositories.ReminderRepository,
	contactRepo repositories.ContactRepository,
	dealRepo repositories.DealRepository,
	leadRepo repositories.LeadRepository,
	customerRepo customerRepos.CustomerRepository,
	employeeRepo orgRepos.EmployeeRepository,
) TaskUsecase {
	return &taskUsecase{
		taskRepo: taskRepo, scheduleRepo: scheduleRepo,
		reminderRepo: reminderRepo,
		contactRepo: contactRepo, dealRepo: dealRepo,
		leadRepo: leadRepo,
		customerRepo: customerRepo, employeeRepo: employeeRepo,
	}
}

func (u *taskUsecase) Create(ctx context.Context, req dto.CreateTaskRequest, createdBy string) (dto.TaskResponse, error) {
	// Validate FK references
	if req.AssignedTo != nil && *req.AssignedTo != "" {
		if _, err := u.employeeRepo.FindByID(ctx, *req.AssignedTo); err != nil {
			return dto.TaskResponse{}, errors.New("assigned employee not found")
		}
	}
	if req.CustomerID != nil && *req.CustomerID != "" {
		if _, err := u.customerRepo.FindByID(ctx, *req.CustomerID); err != nil {
			return dto.TaskResponse{}, errors.New("customer not found")
		}
	}
	if req.ContactID != nil && *req.ContactID != "" {
		if _, err := u.contactRepo.FindByID(ctx, *req.ContactID); err != nil {
			return dto.TaskResponse{}, errors.New("contact not found")
		}
	}
	if req.DealID != nil && *req.DealID != "" {
		if _, err := u.dealRepo.FindByID(ctx, *req.DealID); err != nil {
			return dto.TaskResponse{}, errors.New("deal not found")
		}
	}
	if req.LeadID != nil && *req.LeadID != "" {
		if _, err := u.leadRepo.FindByID(ctx, *req.LeadID); err != nil {
			return dto.TaskResponse{}, errors.New("lead not found")
		}
	}

	// Parse due date
	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		t, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return dto.TaskResponse{}, errors.New("invalid due_date format, use YYYY-MM-DD")
		}
		dueDate = &t
	}

	taskType := "general"
	if req.Type != "" {
		taskType = req.Type
	}
	priority := "medium"
	if req.Priority != "" {
		priority = req.Priority
	}

	task := &models.Task{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Type:        taskType,
		Status:      string(models.TaskStatusPending),
		Priority:    priority,
		DueDate:     dueDate,
		AssignedTo:  req.AssignedTo,
		CustomerID:  req.CustomerID,
		ContactID:   req.ContactID,
		DealID:      req.DealID,
		LeadID:      req.LeadID,
		CreatedBy:   &createdBy,
	}

	if err := u.taskRepo.Create(ctx, task); err != nil {
		return dto.TaskResponse{}, fmt.Errorf("failed to create task: %w", err)
	}

	// Auto-create schedule when task has due_date + assigned_to
	u.autoCreateSchedule(ctx, task)

	created, err := u.taskRepo.FindByID(ctx, task.ID)
	if err != nil {
		return dto.TaskResponse{}, err
	}
	return mapper.ToTaskResponse(created), nil
}

func (u *taskUsecase) GetByID(ctx context.Context, id string) (dto.TaskResponse, error) {
	if !security.CheckRecordScopeAccess(database.DB, ctx, &models.Task{}, id, security.ScopeQueryOptions{
		OwnerEmployeeIDColumn: "assigned_to",
		DivisionJoinSQL:       "assigned_to IN (SELECT id FROM employees WHERE division_id = ?)",
	}) {
		return dto.TaskResponse{}, errors.New("task not found")
	}

	task, err := u.taskRepo.FindByID(ctx, id)
	if err != nil {
		return dto.TaskResponse{}, errors.New("task not found")
	}
	return mapper.ToTaskResponse(task), nil
}

func (u *taskUsecase) List(ctx context.Context, params repositories.TaskListParams) ([]dto.TaskResponse, int64, error) {
	tasks, total, err := u.taskRepo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToTaskResponseList(tasks), total, nil
}

func (u *taskUsecase) Update(ctx context.Context, id string, req dto.UpdateTaskRequest) (dto.TaskResponse, error) {
	task, err := u.taskRepo.FindByID(ctx, id)
	if err != nil {
		return dto.TaskResponse{}, errors.New("task not found")
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Type != nil {
		task.Type = *req.Type
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.Status != nil {
		task.Status = *req.Status
		if task.Status != string(models.TaskStatusCompleted) {
			task.CompletedAt = nil
		}
		if task.Status == string(models.TaskStatusCompleted) && task.CompletedAt == nil {
			now := apptime.Now()
			task.CompletedAt = &now
		}
	}
	if req.DueDate != nil {
		if *req.DueDate == "" {
			task.DueDate = nil
		} else {
			t, err := time.Parse("2006-01-02", *req.DueDate)
			if err != nil {
				return dto.TaskResponse{}, errors.New("invalid due_date format, use YYYY-MM-DD")
			}
			task.DueDate = &t
		}
	}
	if req.AssignedTo != nil {
		if *req.AssignedTo != "" {
			if _, err := u.employeeRepo.FindByID(ctx, *req.AssignedTo); err != nil {
				return dto.TaskResponse{}, errors.New("assigned employee not found")
			}
		}
		task.AssignedTo = req.AssignedTo
	}
	if req.CustomerID != nil {
		task.CustomerID = req.CustomerID
	}
	if req.ContactID != nil {
		task.ContactID = req.ContactID
	}
	if req.DealID != nil {
		task.DealID = req.DealID
	}
	if req.LeadID != nil {
		task.LeadID = req.LeadID
	}

	if err := u.taskRepo.Update(ctx, task); err != nil {
		return dto.TaskResponse{}, fmt.Errorf("failed to update task: %w", err)
	}

	// Auto-sync linked schedule
	u.autoSyncSchedule(ctx, task)

	updated, err := u.taskRepo.FindByID(ctx, task.ID)
	if err != nil {
		return dto.TaskResponse{}, err
	}
	return mapper.ToTaskResponse(updated), nil
}

func (u *taskUsecase) Delete(ctx context.Context, id string) error {
	if _, err := u.taskRepo.FindByID(ctx, id); err != nil {
		return errors.New("task not found")
	}
	// Auto-delete linked schedule
	if err := u.scheduleRepo.DeleteByTaskID(ctx, id); err != nil {
		fmt.Printf("[WARN] failed to delete schedule for task %s: %v\n", id, err)
	}
	return u.taskRepo.Delete(ctx, id)
}

func (u *taskUsecase) Assign(ctx context.Context, id string, req dto.AssignTaskRequest, assignedFrom string) (dto.TaskResponse, error) {
	task, err := u.taskRepo.FindByID(ctx, id)
	if err != nil {
		return dto.TaskResponse{}, errors.New("task not found")
	}

	if task.Status == string(models.TaskStatusCancelled) {
		return dto.TaskResponse{}, errors.New("cannot assign a cancelled task")
	}

	if _, err := u.employeeRepo.FindByID(ctx, req.AssignedTo); err != nil {
		return dto.TaskResponse{}, errors.New("assigned employee not found")
	}

	task.AssignedTo = &req.AssignedTo
	task.AssignedFrom = &assignedFrom

	if err := u.taskRepo.Update(ctx, task); err != nil {
		return dto.TaskResponse{}, fmt.Errorf("failed to assign task: %w", err)
	}

	updated, err := u.taskRepo.FindByID(ctx, task.ID)
	if err != nil {
		return dto.TaskResponse{}, err
	}
	return mapper.ToTaskResponse(updated), nil
}

func (u *taskUsecase) Complete(ctx context.Context, id string) (dto.TaskResponse, error) {
	task, err := u.taskRepo.FindByID(ctx, id)
	if err != nil {
		return dto.TaskResponse{}, errors.New("task not found")
	}

	if task.Status == string(models.TaskStatusCancelled) {
		return dto.TaskResponse{}, errors.New("cannot complete a cancelled task")
	}
	if task.Status == string(models.TaskStatusCompleted) {
		return dto.TaskResponse{}, errors.New("task is already completed")
	}

	now := apptime.Now()
	task.Status = string(models.TaskStatusCompleted)
	task.CompletedAt = &now

	if err := u.taskRepo.Update(ctx, task); err != nil {
		return dto.TaskResponse{}, fmt.Errorf("failed to complete task: %w", err)
	}

	// Auto-complete linked schedule
	u.autoCompleteSchedule(ctx, task.ID)

	updated, err := u.taskRepo.FindByID(ctx, task.ID)
	if err != nil {
		return dto.TaskResponse{}, err
	}
	return mapper.ToTaskResponse(updated), nil
}

func (u *taskUsecase) MarkInProgress(ctx context.Context, id string) (dto.TaskResponse, error) {
	task, err := u.taskRepo.FindByID(ctx, id)
	if err != nil {
		return dto.TaskResponse{}, errors.New("task not found")
	}

	if task.Status == string(models.TaskStatusCancelled) {
		return dto.TaskResponse{}, errors.New("cannot reopen a cancelled task")
	}
	if task.Status == string(models.TaskStatusCompleted) {
		return dto.TaskResponse{}, errors.New("cannot reopen a completed task")
	}

	task.Status = string(models.TaskStatusInProgress)

	if err := u.taskRepo.Update(ctx, task); err != nil {
		return dto.TaskResponse{}, fmt.Errorf("failed to mark task in progress: %w", err)
	}

	updated, err := u.taskRepo.FindByID(ctx, task.ID)
	if err != nil {
		return dto.TaskResponse{}, err
	}
	return mapper.ToTaskResponse(updated), nil
}

func (u *taskUsecase) Cancel(ctx context.Context, id string) (dto.TaskResponse, error) {
	task, err := u.taskRepo.FindByID(ctx, id)
	if err != nil {
		return dto.TaskResponse{}, errors.New("task not found")
	}

	if task.Status == string(models.TaskStatusCancelled) {
		return dto.TaskResponse{}, errors.New("task is already cancelled")
	}
	if task.Status == string(models.TaskStatusCompleted) {
		return dto.TaskResponse{}, errors.New("cannot cancel a completed task")
	}

	task.Status = string(models.TaskStatusCancelled)

	if err := u.taskRepo.Update(ctx, task); err != nil {
		return dto.TaskResponse{}, fmt.Errorf("failed to cancel task: %w", err)
	}

	updated, err := u.taskRepo.FindByID(ctx, task.ID)
	if err != nil {
		return dto.TaskResponse{}, err
	}
	return mapper.ToTaskResponse(updated), nil
}

func (u *taskUsecase) GetFormData(ctx context.Context) (*dto.TaskFormDataResponse, error) {
	// Fetch employees (OWN scope: return only current scoped employee)
	employeeOptions := make([]dto.TaskEmployeeOption, 0)
	if scope, _ := ctx.Value("permission_scope").(string); scope == "OWN" {
		scopedEmployeeID, _ := ctx.Value("scope_employee_id").(string)
		if scopedEmployeeID != "" {
			employee, err := u.employeeRepo.FindByID(ctx, scopedEmployeeID)
			if err == nil && employee != nil {
				employeeOptions = append(employeeOptions, dto.TaskEmployeeOption{
					ID: employee.ID, EmployeeCode: employee.EmployeeCode, Name: employee.Name,
				})
			}
		}
	} else {
		employees, _, err := u.employeeRepo.List(ctx, orgRepos.EmployeeListParams{
			Page: 1, PerPage: 500, SortBy: "name", SortDir: "ASC",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch employees: %w", err)
		}

		employeeOptions = make([]dto.TaskEmployeeOption, 0, len(employees))
		for _, emp := range employees {
			employeeOptions = append(employeeOptions, dto.TaskEmployeeOption{
				ID: emp.ID, EmployeeCode: emp.EmployeeCode, Name: emp.Name,
			})
		}
	}

	// Fetch deals
	deals, _, err := u.dealRepo.List(ctx, repositories.DealListParams{Limit: 500, SortBy: "created_at", SortDir: "DESC"})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deals: %w", err)
	}
	dealOptions := make([]dto.TaskDealOption, 0, len(deals))
	for _, d := range deals {
		dealOptions = append(dealOptions, dto.TaskDealOption{
			ID: d.ID, Code: d.Code, Name: d.Title,
		})
	}

	// Fetch leads
	leads, _, err := u.leadRepo.List(ctx, repositories.LeadListParams{Limit: 500, SortBy: "created_at", SortDir: "DESC"})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch leads: %w", err)
	}
	leadOptions := make([]dto.TaskLeadOption, 0, len(leads))
	for _, l := range leads {
		leadOptions = append(leadOptions, dto.TaskLeadOption{
			ID: l.ID, Code: l.Code, Name: l.FirstName + " " + l.LastName,
		})
	}

	return &dto.TaskFormDataResponse{
		Employees: employeeOptions,
		Deals:     dealOptions,
		Leads:     leadOptions,
	}, nil
}

// --- Reminder nested CRUD ---

func (u *taskUsecase) ListReminders(ctx context.Context, taskID string) ([]dto.ReminderResponse, error) {
	if _, err := u.taskRepo.FindByID(ctx, taskID); err != nil {
		return nil, errors.New("task not found")
	}
	reminders, err := u.reminderRepo.FindByTaskID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return mapper.ToReminderResponseList(reminders), nil
}

func (u *taskUsecase) GetReminderByID(ctx context.Context, taskID string, reminderID string) (dto.ReminderResponse, error) {
	if _, err := u.taskRepo.FindByID(ctx, taskID); err != nil {
		return dto.ReminderResponse{}, errors.New("task not found")
	}
	reminder, err := u.reminderRepo.FindByID(ctx, reminderID)
	if err != nil {
		return dto.ReminderResponse{}, errors.New("reminder not found")
	}
	if reminder.TaskID != taskID {
		return dto.ReminderResponse{}, errors.New("reminder does not belong to this task")
	}
	return mapper.ToReminderResponse(reminder), nil
}

func (u *taskUsecase) CreateReminder(ctx context.Context, taskID string, req dto.CreateReminderRequest, createdBy string) (dto.ReminderResponse, error) {
	if _, err := u.taskRepo.FindByID(ctx, taskID); err != nil {
		return dto.ReminderResponse{}, errors.New("task not found")
	}

	remindAt, err := time.Parse(time.RFC3339, req.RemindAt)
	if err != nil {
		return dto.ReminderResponse{}, errors.New("invalid remind_at format, use ISO 8601")
	}

	reminderType := "in_app"
	if req.ReminderType != "" {
		reminderType = req.ReminderType
	}

	reminder := &models.Reminder{
		ID:           uuid.New().String(),
		TaskID:       taskID,
		RemindAt:     remindAt,
		ReminderType: reminderType,
		Message:      req.Message,
		CreatedBy:    &createdBy,
	}

	if err := u.reminderRepo.Create(ctx, reminder); err != nil {
		return dto.ReminderResponse{}, fmt.Errorf("failed to create reminder: %w", err)
	}

	return mapper.ToReminderResponse(reminder), nil
}

func (u *taskUsecase) UpdateReminder(ctx context.Context, taskID string, reminderID string, req dto.UpdateReminderRequest) (dto.ReminderResponse, error) {
	if _, err := u.taskRepo.FindByID(ctx, taskID); err != nil {
		return dto.ReminderResponse{}, errors.New("task not found")
	}

	reminder, err := u.reminderRepo.FindByID(ctx, reminderID)
	if err != nil {
		return dto.ReminderResponse{}, errors.New("reminder not found")
	}
	if reminder.TaskID != taskID {
		return dto.ReminderResponse{}, errors.New("reminder does not belong to this task")
	}

	if req.RemindAt != nil {
		t, err := time.Parse(time.RFC3339, *req.RemindAt)
		if err != nil {
			return dto.ReminderResponse{}, errors.New("invalid remind_at format, use ISO 8601")
		}
		reminder.RemindAt = t
	}
	if req.ReminderType != nil {
		reminder.ReminderType = *req.ReminderType
	}
	if req.Message != nil {
		reminder.Message = *req.Message
	}

	if err := u.reminderRepo.Update(ctx, reminder); err != nil {
		return dto.ReminderResponse{}, fmt.Errorf("failed to update reminder: %w", err)
	}

	return mapper.ToReminderResponse(reminder), nil
}

func (u *taskUsecase) DeleteReminder(ctx context.Context, taskID string, reminderID string) error {
	if _, err := u.taskRepo.FindByID(ctx, taskID); err != nil {
		return errors.New("task not found")
	}

	reminder, err := u.reminderRepo.FindByID(ctx, reminderID)
	if err != nil {
		return errors.New("reminder not found")
	}
	if reminder.TaskID != taskID {
		return errors.New("reminder does not belong to this task")
	}

	return u.reminderRepo.Delete(ctx, reminderID)
}

// --- Schedule auto-sync helpers ---

// autoCreateSchedule creates a linked schedule when a task has both due_date and assigned_to
func (u *taskUsecase) autoCreateSchedule(ctx context.Context, task *models.Task) {
	if task.DueDate == nil || task.AssignedTo == nil {
		return
	}

	schedule := &models.Schedule{
		TaskID:                &task.ID,
		EmployeeID:            *task.AssignedTo,
		Title:                 task.Title,
		Description:           task.Description,
		ScheduledAt:           *task.DueDate,
		Status:                "pending",
		ReminderMinutesBefore: 30,
		CreatedBy:             task.CreatedBy,
	}
	if err := u.scheduleRepo.Create(ctx, schedule); err != nil {
		fmt.Printf("[WARN] failed to auto-create schedule for task %s: %v\n", task.ID, err)
	}
}

// autoSyncSchedule updates or creates/deletes the linked schedule when task is updated
func (u *taskUsecase) autoSyncSchedule(ctx context.Context, task *models.Task) {
	existing, err := u.scheduleRepo.FindByTaskID(ctx, task.ID)

	if task.DueDate == nil || task.AssignedTo == nil || task.Status == string(models.TaskStatusCancelled) {
		// Remove linked schedule if conditions no longer met or task cancelled
		if err == nil && existing != nil {
			if task.Status == string(models.TaskStatusCancelled) {
				existing.Status = "cancelled"
				if updateErr := u.scheduleRepo.Update(ctx, existing); updateErr != nil {
					fmt.Printf("[WARN] failed to cancel schedule for task %s: %v\n", task.ID, updateErr)
				}
			} else {
				if delErr := u.scheduleRepo.DeleteByTaskID(ctx, task.ID); delErr != nil {
					fmt.Printf("[WARN] failed to delete schedule for task %s: %v\n", task.ID, delErr)
				}
			}
		}
		return
	}

	if err != nil {
		// No existing schedule — create one
		u.autoCreateSchedule(ctx, task)
		return
	}

	// Update existing schedule
	existing.Title = task.Title
	existing.Description = task.Description
	existing.EmployeeID = *task.AssignedTo
	existing.ScheduledAt = *task.DueDate
	if updateErr := u.scheduleRepo.Update(ctx, existing); updateErr != nil {
		fmt.Printf("[WARN] failed to update schedule for task %s: %v\n", task.ID, updateErr)
	}
}

// autoCompleteSchedule marks the linked schedule as completed
func (u *taskUsecase) autoCompleteSchedule(ctx context.Context, taskID string) {
	schedule, err := u.scheduleRepo.FindByTaskID(ctx, taskID)
	if err != nil {
		return
	}
	schedule.Status = "completed"
	if updateErr := u.scheduleRepo.Update(ctx, schedule); updateErr != nil {
		fmt.Printf("[WARN] failed to complete schedule for task %s: %v\n", taskID, updateErr)
	}
}
