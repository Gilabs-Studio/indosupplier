package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gilabs/gims/api/internal/core/apptime"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	"gorm.io/gorm"
)

var (
	ErrAdjustmentApprovalNotFound     = errors.New("adjustment approval history not found")
	ErrAdjustmentAlreadySubmitted     = errors.New("adjustment journal already submitted")
	ErrAdjustmentAlreadyApproved      = errors.New("adjustment journal already approved")
	ErrAdjustmentMustBeSubmittedFirst = errors.New("adjustment journal must be submitted first")
	ErrAdjustmentAlreadyRejected      = errors.New("adjustment journal already rejected")
	ErrAdjustmentNeedsApproval        = errors.New("adjustment journal must be approved before posting")
	ErrJournalTemplateNotFound        = errors.New("journal template not found")
)

func (uc *journalEntryUsecase) SubmitAdjustmentJournalForApproval(ctx context.Context, id string, notes string) (*dto.JournalEntryResponse, error) {
	entry, latest, actorID, err := uc.validateAdjustmentWorkflowAction(ctx, id)
	if err != nil {
		return nil, err
	}

	if latest != nil {
		switch latest.Action {
		case financeModels.AdjustmentJournalApprovalActionSubmitted:
			return nil, ErrAdjustmentAlreadySubmitted
		case financeModels.AdjustmentJournalApprovalActionApproved:
			return nil, ErrAdjustmentAlreadyApproved
		}
	}

	if err := uc.adjustmentApprovalRepo.Create(ctx, &financeModels.AdjustmentJournalApproval{
		JournalID: entry.ID,
		Action:    financeModels.AdjustmentJournalApprovalActionSubmitted,
		ActorID:   actorID,
		Notes:     strings.TrimSpace(notes),
	}); err != nil {
		return nil, err
	}

	return uc.GetByID(ctx, entry.ID)
}

func (uc *journalEntryUsecase) ApproveAdjustmentJournal(ctx context.Context, id string, notes string) (*dto.JournalEntryResponse, error) {
	entry, latest, actorID, err := uc.validateAdjustmentWorkflowAction(ctx, id)
	if err != nil {
		return nil, err
	}
	if latest == nil || latest.Action != financeModels.AdjustmentJournalApprovalActionSubmitted {
		return nil, ErrAdjustmentMustBeSubmittedFirst
	}

	if err := uc.adjustmentApprovalRepo.Create(ctx, &financeModels.AdjustmentJournalApproval{
		JournalID: entry.ID,
		Action:    financeModels.AdjustmentJournalApprovalActionApproved,
		ActorID:   actorID,
		Notes:     strings.TrimSpace(notes),
	}); err != nil {
		return nil, err
	}

	return uc.GetByID(ctx, entry.ID)
}

func (uc *journalEntryUsecase) RejectAdjustmentJournal(ctx context.Context, id string, notes string) (*dto.JournalEntryResponse, error) {
	entry, latest, actorID, err := uc.validateAdjustmentWorkflowAction(ctx, id)
	if err != nil {
		return nil, err
	}
	if latest == nil || latest.Action != financeModels.AdjustmentJournalApprovalActionSubmitted {
		return nil, ErrAdjustmentMustBeSubmittedFirst
	}

	if err := uc.adjustmentApprovalRepo.Create(ctx, &financeModels.AdjustmentJournalApproval{
		JournalID: entry.ID,
		Action:    financeModels.AdjustmentJournalApprovalActionRejected,
		ActorID:   actorID,
		Notes:     strings.TrimSpace(notes),
	}); err != nil {
		return nil, err
	}

	return uc.GetByID(ctx, entry.ID)
}

func (uc *journalEntryUsecase) GetAdjustmentApprovalHistory(ctx context.Context, id string) ([]dto.AdjustmentJournalApprovalResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}

	entry, err := uc.repo.FindByID(ctx, id, false)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrJournalNotFound
		}
		return nil, err
	}
	if !isManualAdjustmentEntry(entry) {
		return nil, errors.New("can only view approval history for manual adjustment journals")
	}

	history, err := uc.adjustmentApprovalRepo.ListByJournalID(ctx, id)
	if err != nil {
		return nil, err
	}

	out := make([]dto.AdjustmentJournalApprovalResponse, 0, len(history))
	for _, item := range history {
		out = append(out, dto.AdjustmentJournalApprovalResponse{
			ID:        item.ID,
			JournalID: item.JournalID,
			Action:    string(item.Action),
			ActorID:   item.ActorID,
			Notes:     item.Notes,
			CreatedAt: item.CreatedAt,
		})
	}
	return out, nil
}

func (uc *journalEntryUsecase) ListJournalTemplates(ctx context.Context, req *dto.ListJournalTemplatesRequest) ([]dto.JournalTemplateResponse, int64, error) {
	if req == nil {
		req = &dto.ListJournalTemplatesRequest{}
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

	items, total, err := uc.journalTemplateRepo.List(ctx, repositories.ListJournalTemplateParams{
		CompanyID: strings.TrimSpace(req.CompanyID),
		Search:    strings.TrimSpace(req.Search),
		Limit:     perPage,
		Offset:    (page - 1) * perPage,
	})
	if err != nil {
		return nil, 0, err
	}

	out := make([]dto.JournalTemplateResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapJournalTemplateToResponse(item))
	}
	return out, total, nil
}

func (uc *journalEntryUsecase) CreateJournalTemplate(ctx context.Context, req *dto.CreateJournalTemplateRequest) (*dto.JournalTemplateResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}
	if _, _, err := validateLines(req.Lines); err != nil {
		return nil, err
	}

	payload, err := json.Marshal(req.Lines)
	if err != nil {
		return nil, err
	}

	actorID := strings.TrimSpace(getActorIDFromContext(ctx))
	if actorID == "" {
		return nil, errors.New("actor is required")
	}

	jt := &financeModels.JournalTemplate{
		CompanyID:    strings.TrimSpace(req.CompanyID),
		TemplateName: strings.TrimSpace(req.TemplateName),
		JournalType:  financeModels.JournalType(strings.ToUpper(strings.TrimSpace(req.JournalType))),
		Description:  strings.TrimSpace(req.Description),
		Lines:        string(payload),
		CreatedBy:    &actorID,
	}

	if err := uc.journalTemplateRepo.Create(ctx, jt); err != nil {
		return nil, err
	}
	res := mapJournalTemplateToResponse(*jt)
	return &res, nil
}

func (uc *journalEntryUsecase) UseJournalTemplate(ctx context.Context, id string) (*dto.UseJournalTemplateResponse, error) {
	item, err := uc.journalTemplateRepo.FindByID(ctx, strings.TrimSpace(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrJournalTemplateNotFound
		}
		return nil, err
	}

	var lines []dto.JournalLineRequest
	if err := json.Unmarshal([]byte(item.Lines), &lines); err != nil {
		return nil, err
	}

	_ = uc.journalTemplateRepo.TouchLastUsed(ctx, item.ID)

	res := mapJournalTemplateToResponse(*item)
	return &dto.UseJournalTemplateResponse{
		Template: res,
		Prefill: dto.CreateAdjustmentJournalRequest{
			CompanyID:    item.CompanyID,
			EntryDate:    apptime.Now().Format("2006-01-02"),
			Description:  item.Description,
			CurrencyCode: "IDR",
			Lines:        lines,
		},
	}, nil
}

func mapJournalTemplateToResponse(item financeModels.JournalTemplate) dto.JournalTemplateResponse {
	return dto.JournalTemplateResponse{
		ID:           item.ID,
		CompanyID:    item.CompanyID,
		TemplateName: item.TemplateName,
		JournalType:  string(item.JournalType),
		Description:  item.Description,
		Lines:        item.Lines,
		CreatedBy:    item.CreatedBy,
		LastUsedAt:   item.LastUsedAt,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}

func (uc *journalEntryUsecase) validateAdjustmentWorkflowAction(ctx context.Context, id string) (*financeModels.JournalEntry, *financeModels.AdjustmentJournalApproval, string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, nil, "", errors.New("id is required")
	}

	entry, err := uc.repo.FindByID(ctx, id, false)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, "", ErrJournalNotFound
		}
		return nil, nil, "", err
	}
	if !isManualAdjustmentEntry(entry) {
		return nil, nil, "", errors.New("can only perform this action on manual adjustment journals")
	}
	if entry.Status != financeModels.JournalStatusDraft {
		return nil, nil, "", errors.New("only draft adjustment journals can be processed")
	}

	latest, err := uc.adjustmentApprovalRepo.GetLatestByJournalID(ctx, entry.ID)
	if err != nil {
		return nil, nil, "", err
	}

	actorID := strings.TrimSpace(getActorIDFromContext(ctx))
	if actorID == "" {
		return nil, nil, "", errors.New("actor is required")
	}

	return entry, latest, actorID, nil
}

func isManualAdjustmentEntry(entry *financeModels.JournalEntry) bool {
	if entry == nil || entry.ReferenceType == nil {
		return false
	}
	return reference.Normalize(*entry.ReferenceType) == reference.RefTypeManualAdjustment
}

func getActorIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	userID, _ := ctx.Value("user_id").(string)
	return strings.TrimSpace(userID)
}
