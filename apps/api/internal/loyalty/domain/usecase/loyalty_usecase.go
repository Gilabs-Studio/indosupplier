package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/utils"
	customerModels "github.com/gilabs/gims/api/internal/customer/data/models"
	customerRepos "github.com/gilabs/gims/api/internal/customer/data/repositories"
	"github.com/gilabs/gims/api/internal/loyalty/data/models"
	"github.com/gilabs/gims/api/internal/loyalty/data/repositories"
	"github.com/gilabs/gims/api/internal/loyalty/domain/dto"
	"github.com/gilabs/gims/api/internal/loyalty/domain/mapper"
	organizationRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrLoyaltyProgramNotFound = errors.New("LOYALTY_PROGRAM_NOT_FOUND")
	ErrLoyaltyMemberNotFound  = errors.New("LOYALTY_MEMBER_NOT_FOUND")
	ErrAlreadyEnrolled        = errors.New("LOYALTY_ALREADY_ENROLLED")
	ErrInvalidLoyaltyConfig   = errors.New("LOYALTY_INVALID_CONFIG")
	ErrRewardNotFound         = errors.New("LOYALTY_REWARD_NOT_FOUND")
	ErrInsufficientPoints     = errors.New("LOYALTY_INSUFFICIENT_POINTS")
	ErrPointsAlreadyAwarded   = errors.New("LOYALTY_POINTS_ALREADY_AWARDED")
	ErrNoActiveProgram        = errors.New("LOYALTY_NO_ACTIVE_PROGRAM")
	// ErrLoyaltyForbidden is returned when an outlet-scoped user attempts an operation
	// outside their permitted outlets (e.g., creating a global program).
	ErrLoyaltyForbidden = errors.New("LOYALTY_FORBIDDEN")
)

type LoyaltyUsecase interface {
	// Program management
	CreateProgram(ctx context.Context, createdBy string, req *dto.CreateLoyaltyProgramRequest) (*dto.LoyaltyProgramResponse, error)
	UpdateProgram(ctx context.Context, id, updatedBy string, req *dto.UpdateLoyaltyProgramRequest) (*dto.LoyaltyProgramResponse, error)
	DeleteProgram(ctx context.Context, id string) error
	GetProgram(ctx context.Context, id string) (*dto.LoyaltyProgramResponse, error)
	ListPrograms(ctx context.Context, page, perPage int, search string) ([]dto.LoyaltyProgramResponse, *utils.PaginationResult, error)
	ToggleProgramActive(ctx context.Context, id, updatedBy string) (*dto.LoyaltyProgramResponse, error)

	// Member management
	EnrollMember(ctx context.Context, enrolledBy string, req *dto.EnrollMemberRequest) (*dto.LoyaltyMemberResponse, error)
	ChangeProgram(ctx context.Context, id, updatedBy string, req *dto.ChangeProgramRequest) (*dto.LoyaltyMemberResponse, error)
	GetMember(ctx context.Context, id string) (*dto.LoyaltyMemberResponse, error)
	GetMemberByCustomerID(ctx context.Context, customerID string) (*dto.LoyaltyMemberResponse, error)
	LookupMember(ctx context.Context, name, outletID string) (*dto.LookupMemberResponse, error)
	ListMembers(ctx context.Context, params repositories.MemberListParams) ([]dto.LoyaltyMemberResponse, *utils.PaginationResult, error)

	// Points operations
	EarnPoints(ctx context.Context, req *dto.EarnPointsRequest) (int64, error)
	RedeemPoints(ctx context.Context, req *dto.RedeemPointsRequest) (*dto.RedeemPointsResponse, error)
	AdjustPoints(ctx context.Context, req *dto.AdjustPointsRequest) error
	ListLedger(ctx context.Context, params repositories.LedgerListParams) ([]dto.PointLedgerResponse, *utils.PaginationResult, error)

	// Public self-registration
	PublicSelfRegister(ctx context.Context, req *dto.PublicSelfRegisterRequest) (*dto.PublicSelfRegisterResponse, error)
}

type loyaltyUsecase struct {
	db           *gorm.DB
	programRepo  repositories.LoyaltyProgramRepository
	memberRepo   repositories.LoyaltyMemberRepository
	ledgerRepo   repositories.LoyaltyPointLedgerRepository
	customerRepo customerRepos.CustomerRepository
	outletRepo   organizationRepos.OutletRepository
}

func NewLoyaltyUsecase(
	db *gorm.DB,
	programRepo repositories.LoyaltyProgramRepository,
	memberRepo repositories.LoyaltyMemberRepository,
	ledgerRepo repositories.LoyaltyPointLedgerRepository,
	customerRepo customerRepos.CustomerRepository,
	outletRepo organizationRepos.OutletRepository,
) LoyaltyUsecase {
	return &loyaltyUsecase{
		db:           db,
		programRepo:  programRepo,
		memberRepo:   memberRepo,
		ledgerRepo:   ledgerRepo,
		customerRepo: customerRepo,
		outletRepo:   outletRepo,
	}
}

// ─── Program Management ───────────────────────────────────────────────────────

func (u *loyaltyUsecase) CreateProgram(ctx context.Context, createdBy string, req *dto.CreateLoyaltyProgramRequest) (*dto.LoyaltyProgramResponse, error) {
	// Outlet-scoped users may only create outlet-specific programs.
	isScoped := isOutletScopedScope(ctx)
	outletIDs, err := u.resolveScopedOutletIDs(ctx)
	if err != nil {
		return nil, err
	}
	if isScoped {
		if len(outletIDs) == 0 {
			return nil, ErrLoyaltyForbidden
		}
		if req.OutletID == nil || *req.OutletID == "" {
			return nil, ErrLoyaltyForbidden
		}
		if !containsString(outletIDs, *req.OutletID) {
			return nil, ErrLoyaltyForbidden
		}
	}

	configBytes, err := marshalAndValidateConfig(&req.Config)
	if err != nil {
		return nil, err
	}

	program := &models.LoyaltyProgram{
		OutletID:    req.OutletID,
		Name:        req.Name,
		Description: req.Description,
		ConfigJSON:  datatypes.JSON(configBytes),
		IsActive:    req.IsActive,
		CreatedBy:   &createdBy,
		UpdatedBy:   &createdBy,
	}

	if err := u.programRepo.Create(ctx, program); err != nil {
		return nil, err
	}
	return mapper.ToLoyaltyProgramResponse(program), nil
}

func (u *loyaltyUsecase) UpdateProgram(ctx context.Context, id, updatedBy string, req *dto.UpdateLoyaltyProgramRequest) (*dto.LoyaltyProgramResponse, error) {
	program, err := u.programRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if program == nil {
		return nil, ErrLoyaltyProgramNotFound
	}

	// Outlet-scoped users may only modify programs that belong to their outlets.
	isScoped := isOutletScopedScope(ctx)
	outletIDs, err := u.resolveScopedOutletIDs(ctx)
	if err != nil {
		return nil, err
	}
	if isScoped {
		if len(outletIDs) == 0 {
			return nil, ErrLoyaltyForbidden
		}
		if program.OutletID == nil || !containsString(outletIDs, *program.OutletID) {
			return nil, ErrLoyaltyForbidden
		}
	}
	if req.Name != nil {
		program.Name = *req.Name
	}
	if req.Description != nil {
		program.Description = req.Description
	}
	if req.Config != nil {
		configBytes, err := marshalAndValidateConfig(req.Config)
		if err != nil {
			return nil, err
		}
		program.ConfigJSON = datatypes.JSON(configBytes)
	}
	if req.IsActive != nil {
		program.IsActive = *req.IsActive
	}
	program.UpdatedBy = &updatedBy

	if err := u.programRepo.Update(ctx, program); err != nil {
		return nil, err
	}
	return mapper.ToLoyaltyProgramResponse(program), nil
}

func (u *loyaltyUsecase) DeleteProgram(ctx context.Context, id string) error {
	program, err := u.programRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if program == nil {
		return ErrLoyaltyProgramNotFound
	}

	// Outlet-scoped users may only delete programs that belong to their outlets.
	isScoped := isOutletScopedScope(ctx)
	outletIDs, scopeErr := u.resolveScopedOutletIDs(ctx)
	if scopeErr != nil {
		return scopeErr
	}
	if isScoped {
		if len(outletIDs) == 0 {
			return ErrLoyaltyForbidden
		}
		if program.OutletID == nil || !containsString(outletIDs, *program.OutletID) {
			return ErrLoyaltyForbidden
		}
	}

	return u.programRepo.Delete(ctx, id)
}

func (u *loyaltyUsecase) GetProgram(ctx context.Context, id string) (*dto.LoyaltyProgramResponse, error) {
	program, err := u.programRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if program == nil {
		return nil, ErrLoyaltyProgramNotFound
	}
	return mapper.ToLoyaltyProgramResponse(program), nil
}

func (u *loyaltyUsecase) ListPrograms(ctx context.Context, page, perPage int, search string) ([]dto.LoyaltyProgramResponse, *utils.PaginationResult, error) {
	var programs []models.LoyaltyProgram
	var total int64
	var fetchErr error

	// Outlet-scoped users see global + their outlet's programs.
	isScoped := isOutletScopedScope(ctx)
	outletIDs, err := u.resolveScopedOutletIDs(ctx)
	if err != nil {
		return nil, nil, err
	}
	if isScoped {
		programs, total, fetchErr = u.programRepo.ListWithOutletFilter(ctx, outletIDs, page, perPage, search)
	} else {
		programs, total, fetchErr = u.programRepo.List(ctx, page, perPage, search)
	}
	if fetchErr != nil {
		return nil, nil, fetchErr
	}

	resp := make([]dto.LoyaltyProgramResponse, 0, len(programs))
	for i := range programs {
		resp = append(resp, *mapper.ToLoyaltyProgramResponse(&programs[i]))
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	pagination := &utils.PaginationResult{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: totalPages,
	}
	return resp, pagination, nil
}

func (u *loyaltyUsecase) ToggleProgramActive(ctx context.Context, id, updatedBy string) (*dto.LoyaltyProgramResponse, error) {
	program, err := u.programRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if program == nil {
		return nil, ErrLoyaltyProgramNotFound
	}
	program.IsActive = !program.IsActive
	program.UpdatedBy = &updatedBy
	if err := u.programRepo.Update(ctx, program); err != nil {
		return nil, err
	}
	return mapper.ToLoyaltyProgramResponse(program), nil
}

// ─── Member Management ────────────────────────────────────────────────────────

func (u *loyaltyUsecase) EnrollMember(ctx context.Context, enrolledBy string, req *dto.EnrollMemberRequest) (*dto.LoyaltyMemberResponse, error) {
	// Resolve program from outlet if not provided.
	programID := ""
	if req.ProgramID != nil {
		programID = *req.ProgramID
	} else {
		if req.OutletID == nil || *req.OutletID == "" {
			return nil, errors.New("PROGRAM_OR_OUTLET_REQUIRED")
		}
		prog, err := u.programRepo.FindActiveForOutlet(ctx, *req.OutletID)
		if err != nil {
			return nil, err
		}
		if prog == nil {
			return nil, ErrNoActiveProgram
		}
		programID = prog.ID
	}

	// Resolve or create the customer record.
	customerID := ""
	var customerRecord *customerModels.Customer

	if req.CustomerID != nil {
		customerRecord, _ = u.customerRepo.FindByID(ctx, *req.CustomerID)
	}

	if customerRecord == nil {
		// Create a minimal customer record.
		code, err := u.customerRepo.GetNextCode(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate customer code: %w", err)
		}
		newCustomer := &customerModels.Customer{
			ID:   uuid.New().String(),
			Code: code,
			Name: req.Name,
		}
		if req.Phone != nil && *req.Phone != "" {
			newCustomer.ContactPerson = *req.Phone
		}
		if req.Email != nil {
			newCustomer.Email = *req.Email
		}
		if err := u.db.WithContext(ctx).Create(newCustomer).Error; err != nil {
			return nil, fmt.Errorf("failed to create customer: %w", err)
		}
		customerRecord = newCustomer
	}
	customerID = customerRecord.ID

	// Prevent duplicate enrollments.
	existing, err := u.memberRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		// Return existing member rather than erroring — the caller handles the "already enrolled" case.
		return u.buildMemberResponse(ctx, existing, customerRecord)
	}

	program, err := u.programRepo.GetByID(ctx, programID)
	if err != nil || program == nil {
		return nil, ErrLoyaltyProgramNotFound
	}

	cfg, err := parseConfig(program)
	if err != nil {
		return nil, err
	}

	tier := computeTier(0, cfg.Tiers)

	memberCode, err := u.memberRepo.GetNextMemberCode(ctx)
	if err != nil {
		return nil, err
	}

	now := apptime.Now()
	member := &models.LoyaltyMember{
		CustomerID:       customerID,
		ProgramID:        programID,
		MemberCode:       memberCode,
		EnrolledOutletID: req.OutletID,
		CurrentTier:      tier.Name,
		TierBadgeColor:   tier.BadgeColor,
		LifetimePoints:   0,
		PointBalance:     0,
		JoinedAt:         now,
		CreatedBy:        &enrolledBy,
		UpdatedBy:        &enrolledBy,
	}

	if err := u.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	linkedOrderTotalAmount := 0.0
	if req.PosOrderID != nil && *req.PosOrderID != "" {
		linkedOrderTotalAmount = u.linkCustomerToPOSOrderAndSalesOrder(ctx, *req.PosOrderID, customerRecord)
	}

	// If a POS order was linked, credit points for it immediately.
	if req.PosOrderID != nil && *req.PosOrderID != "" {
		_, _ = u.EarnPoints(ctx, &dto.EarnPointsRequest{
			MemberID:        member.ID,
			TransactionID:   *req.PosOrderID,
			TransactionType: "pos_order",
			TotalAmount:     linkedOrderTotalAmount,
			ProcessedBy:     &enrolledBy,
		})
	}

	// If an existing customer is being enrolled for the first time, backfill historical
	// purchase points from previously completed POS orders.
	if req.CustomerID != nil {
		u.backfillHistoricalPoints(ctx, member, cfg, customerID, enrolledBy, req.PosOrderID)
	}

	// Re-fetch member to reflect any points earned or backfilled above.
	if req.PosOrderID != nil || req.CustomerID != nil {
		member, _ = u.memberRepo.GetByID(ctx, member.ID)
	}

	return u.buildMemberResponse(ctx, member, customerRecord)
}

func (u *loyaltyUsecase) ChangeProgram(ctx context.Context, id, updatedBy string, req *dto.ChangeProgramRequest) (*dto.LoyaltyMemberResponse, error) {
	member, err := u.memberRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrLoyaltyMemberNotFound
	}

	if member.ProgramID == req.ProgramID {
		return u.GetMember(ctx, id)
	}

	program, err := u.programRepo.GetByID(ctx, req.ProgramID)
	if err != nil || program == nil {
		return nil, ErrLoyaltyProgramNotFound
	}

	cfg, err := parseConfig(program)
	if err != nil {
		return nil, err
	}

	tier := computeTier(member.LifetimePoints, cfg.Tiers)

	err = u.db.WithContext(ctx).Model(&models.LoyaltyMember{}).Where("id = ?", id).Updates(map[string]interface{}{
		"program_id":       req.ProgramID,
		"current_tier":     tier.Name,
		"tier_badge_color": tier.BadgeColor,
		"updated_by":       &updatedBy,
	}).Error
	if err != nil {
		return nil, err
	}

	return u.GetMember(ctx, id)
}

func (u *loyaltyUsecase) GetMember(ctx context.Context, id string) (*dto.LoyaltyMemberResponse, error) {
	member, err := u.memberRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrLoyaltyMemberNotFound
	}

	customer, _ := u.customerRepo.FindByID(ctx, member.CustomerID)
	return u.buildMemberResponse(ctx, member, customer)
}

func (u *loyaltyUsecase) GetMemberByCustomerID(ctx context.Context, customerID string) (*dto.LoyaltyMemberResponse, error) {
	member, err := u.memberRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrLoyaltyMemberNotFound
	}
	customer, _ := u.customerRepo.FindByID(ctx, customerID)
	return u.buildMemberResponse(ctx, member, customer)
}

func (u *loyaltyUsecase) LookupMember(ctx context.Context, name, outletID string) (*dto.LookupMemberResponse, error) {
	member, err := u.memberRepo.LookupByNameAndOutlet(ctx, name, outletID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return &dto.LookupMemberResponse{Found: false}, nil
	}

	// Load program config to build available rewards list.
	var availableRewards []dto.RewardConfig
	program, _ := u.programRepo.GetByID(ctx, member.ProgramID)
	if program != nil {
		cfg, parseErr := parseConfig(program)
		if parseErr == nil {
			for _, r := range cfg.Rewards {
				if r.IsActive && int64(r.PointsRequired) <= member.PointBalance {
					availableRewards = append(availableRewards, r)
				}
			}
		}
	}

	return &dto.LookupMemberResponse{
		Found:            true,
		MemberID:         &member.ID,
		MemberCode:       &member.MemberCode,
		CustomerID:       &member.CustomerID,
		CurrentTier:      &member.CurrentTier,
		TierBadgeColor:   &member.TierBadgeColor,
		PointBalance:     &member.PointBalance,
		AvailableRewards: availableRewards,
	}, nil
}

func (u *loyaltyUsecase) ListMembers(ctx context.Context, params repositories.MemberListParams) ([]dto.LoyaltyMemberResponse, *utils.PaginationResult, error) {
	// Outlet-scoped users see only members whose program belongs to their outlets.
	isScoped := isOutletScopedScope(ctx)
	outletIDs, err := u.resolveScopedOutletIDs(ctx)
	if err != nil {
		return nil, nil, err
	}
	if isScoped {
		params.ProgramOutletIDs = outletIDs
	}

	members, total, err := u.memberRepo.List(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	resp := make([]dto.LoyaltyMemberResponse, 0, len(members))
	for i := range members {
		customer, _ := u.customerRepo.FindByID(ctx, members[i].CustomerID)
		r, _ := u.buildMemberResponse(ctx, &members[i], customer)
		if r != nil {
			resp = append(resp, *r)
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.PerPage)))
	pagination := &utils.PaginationResult{
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      int(total),
		TotalPages: totalPages,
	}
	return resp, pagination, nil
}

// ─── Points Operations ────────────────────────────────────────────────────────

// EarnPoints awards points for a completed transaction atomically using a DB transaction.
func (u *loyaltyUsecase) EarnPoints(ctx context.Context, req *dto.EarnPointsRequest) (int64, error) {
	member, err := u.memberRepo.GetByID(ctx, req.MemberID)
	if err != nil || member == nil {
		return 0, ErrLoyaltyMemberNotFound
	}

	// Guard against double-awarding points for the same transaction.
	existing, err := u.ledgerRepo.FindByTransaction(ctx, req.MemberID, req.TransactionID, string(models.LedgerEntryTypeEarned))
	if err != nil {
		return 0, err
	}
	if existing != nil {
		return 0, ErrPointsAlreadyAwarded
	}

	program, err := u.programRepo.GetByID(ctx, member.ProgramID)
	if err != nil || program == nil {
		return 0, ErrLoyaltyProgramNotFound
	}

	cfg, err := parseConfig(program)
	if err != nil {
		return 0, err
	}

	// Skip if transaction is below minimum amount.
	effectiveTotalAmount := req.TotalAmount
	if effectiveTotalAmount <= 0 {
		effectiveTotalAmount = u.resolveTransactionAmount(ctx, req.TransactionType, req.TransactionID)
	}

	if effectiveTotalAmount < cfg.PointRules.MinTransactionAmount {
		return 0, nil
	}

	tier := computeTier(member.LifetimePoints, cfg.Tiers)
	multiplier := tier.Multiplier
	if multiplier <= 0 {
		multiplier = 1.0
	}

	basePoints := int64(0)
	if cfg.PointRules.AmountPerPoint > 0 {
		basePoints = int64(math.Floor(effectiveTotalAmount / cfg.PointRules.AmountPerPoint * cfg.PointRules.PointsPerAmount))
	}
	earnedPoints := int64(math.Floor(float64(basePoints) * multiplier))
	if earnedPoints <= 0 {
		return 0, nil
	}

	var expiresAt *time.Time
	if cfg.PointExpiryDays > 0 {
		expiry := apptime.Now().AddDate(0, 0, cfg.PointExpiryDays)
		expiresAt = &expiry
	}

	return earnedPoints, u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		entry := &models.LoyaltyPointLedger{
			MemberID:        req.MemberID,
			EntryType:       models.LedgerEntryTypeEarned,
			TransactionType: &req.TransactionType,
			TransactionID:   &req.TransactionID,
			Points:          earnedPoints,
			BalanceAfter:    member.PointBalance + earnedPoints,
			LifetimeAfter:   member.LifetimePoints + earnedPoints,
			Multiplier:      &multiplier,
			ExpiresAt:       expiresAt,
			ProcessedBy:     req.ProcessedBy,
		}

		if err := tx.Create(entry).Error; err != nil {
			return err
		}

		newLifetime := member.LifetimePoints + earnedPoints
		newTier := computeTier(newLifetime, cfg.Tiers)
		now := apptime.Now()

		return tx.Model(&models.LoyaltyMember{}).Where("id = ?", req.MemberID).Updates(map[string]interface{}{
			"point_balance":      gorm.Expr("point_balance + ?", earnedPoints),
			"lifetime_points":    gorm.Expr("lifetime_points + ?", earnedPoints),
			"current_tier":       newTier.Name,
			"tier_badge_color":   newTier.BadgeColor,
			"last_transaction_at": &now,
			"total_transactions": gorm.Expr("total_transactions + 1"),
		}).Error
	})
}

// RedeemPoints deducts points for a reward atomically using a DB transaction.
func (u *loyaltyUsecase) RedeemPoints(ctx context.Context, req *dto.RedeemPointsRequest) (*dto.RedeemPointsResponse, error) {
	member, err := u.memberRepo.GetByID(ctx, req.MemberID)
	if err != nil || member == nil {
		return nil, ErrLoyaltyMemberNotFound
	}

	program, err := u.programRepo.GetByID(ctx, member.ProgramID)
	if err != nil || program == nil {
		return nil, ErrLoyaltyProgramNotFound
	}

	cfg, err := parseConfig(program)
	if err != nil {
		return nil, err
	}

	// Find the requested reward.
	var reward *dto.RewardConfig
	for i := range cfg.Rewards {
		if cfg.Rewards[i].ID == req.RewardID {
			reward = &cfg.Rewards[i]
			break
		}
	}
	if reward == nil {
		return nil, ErrRewardNotFound
	}

	if member.PointBalance < reward.PointsRequired {
		return nil, ErrInsufficientPoints
	}

	rewardName := reward.Name
	pointsToDeduct := reward.PointsRequired

	txErr := u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		entry := &models.LoyaltyPointLedger{
			MemberID:        req.MemberID,
			EntryType:       models.LedgerEntryTypeRedeemed,
			TransactionType: &req.TransactionType,
			TransactionID:   &req.TransactionID,
			Points:          -int64(pointsToDeduct),
			BalanceAfter:    member.PointBalance - int64(pointsToDeduct),
			LifetimeAfter:   member.LifetimePoints,
			RewardID:        &req.RewardID,
			RewardName:      &rewardName,
			ProcessedBy:     req.ProcessedBy,
		}
		if err := tx.Create(entry).Error; err != nil {
			return err
		}

		return tx.Model(&models.LoyaltyMember{}).Where("id = ?", req.MemberID).Updates(map[string]interface{}{
			"point_balance": gorm.Expr("point_balance - ?", pointsToDeduct),
		}).Error
	})

	if txErr != nil {
		return nil, txErr
	}

	// Apply discount to the linked POS order so total_amount reflects the loyalty reward.
	// Merch-type rewards are physical items and do not reduce the monetary total.
	if req.TransactionType == "pos_order" && req.TransactionID != "" && reward.Type != "merch" {
		discountDelta := float64(reward.Value)

		if reward.Type == "discount_percent" && discountDelta > 0 {
			// Backward compatibility: legacy configs may store 5% as 5000.
			// Normalize to a 0..100 percentage before calculating discount.
			if discountDelta > 100 {
				discountDelta = discountDelta / 1000
			}
			discountDelta = math.Min(math.Max(discountDelta, 0), 100)

			var orderSubtotal float64
			_ = u.db.WithContext(ctx).Table("pos_orders").
				Select("subtotal").
				Where("id = ?", req.TransactionID).
				Scan(&orderSubtotal).Error
			discountDelta = orderSubtotal * discountDelta / 100
		}

		if discountDelta > 0 {
			rewardID := req.RewardID
			memberID := req.MemberID
			_ = u.db.WithContext(ctx).Table("pos_orders").
				Where("id = ? AND status = 'PENDING'", req.TransactionID).
				Updates(map[string]interface{}{
					"discount_amount":   gorm.Expr("discount_amount + ?", discountDelta),
					"total_amount":      gorm.Expr("GREATEST(total_amount - ?, 0)", discountDelta),
					"loyalty_member_id": memberID,
					"loyalty_reward_id": rewardID,
				}).Error
		}
	}

	return &dto.RedeemPointsResponse{
		PointsDeducted:   int64(pointsToDeduct),
		NewBalance:       member.PointBalance - int64(pointsToDeduct),
		DiscountAmount:   reward.Value,
		DiscountType:     reward.Type,
		RewardName:       reward.Name,
		MerchDescription: reward.MerchDescription,
	}, nil
}

func (u *loyaltyUsecase) AdjustPoints(ctx context.Context, req *dto.AdjustPointsRequest) error {
	member, err := u.memberRepo.GetByID(ctx, req.MemberID)
	if err != nil || member == nil {
		return ErrLoyaltyMemberNotFound
	}

	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		newBalance := member.PointBalance + req.Points
		newLifetime := member.LifetimePoints
		if req.Points > 0 {
			newLifetime += req.Points
		}

		entry := &models.LoyaltyPointLedger{
			MemberID:      req.MemberID,
			EntryType:     models.LedgerEntryTypeAdjusted,
			Points:        req.Points,
			BalanceAfter:  newBalance,
			LifetimeAfter: newLifetime,
			Notes:         req.Notes,
			ProcessedBy:   req.ProcessedBy,
		}
		if err := tx.Create(entry).Error; err != nil {
			return err
		}

		updates := map[string]interface{}{
			"point_balance": gorm.Expr("point_balance + ?", req.Points),
		}
		if req.Points > 0 {
			updates["lifetime_points"] = gorm.Expr("lifetime_points + ?", req.Points)
		}
		return tx.Model(&models.LoyaltyMember{}).Where("id = ?", req.MemberID).Updates(updates).Error
	})
}

// backfillHistoricalPoints awards adjusted points for a newly enrolled customer based on
// the total revenue of their previously completed POS orders. This is best-effort:
// errors are silently ignored so that enrollment itself always succeeds.
func (u *loyaltyUsecase) backfillHistoricalPoints(ctx context.Context, member *models.LoyaltyMember, cfg *dto.LoyaltyConfigJSON, customerID, enrolledBy string, excludePosOrderID *string) {
	if cfg.PointRules.AmountPerPoint <= 0 || cfg.PointRules.PointsPerAmount <= 0 {
		return
	}

	var totalRevenue float64
	query := u.db.WithContext(ctx).
		Table("pos_orders").
		Where("customer_id = ? AND status IN ?", customerID, []string{"PAID", "COMPLETED"})

	if excludePosOrderID != nil && *excludePosOrderID != "" {
		query = query.Where("id <> ?", *excludePosOrderID)
	}

	if err := query.
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&totalRevenue).Error; err != nil || totalRevenue <= 0 {
		return
	}

	backfillPoints := int64(math.Floor(totalRevenue / cfg.PointRules.AmountPerPoint * cfg.PointRules.PointsPerAmount))
	if backfillPoints <= 0 {
		return
	}

	note := "Historical purchase backfill"
	_ = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		entry := &models.LoyaltyPointLedger{
			MemberID:      member.ID,
			EntryType:     models.LedgerEntryTypeAdjusted,
			Points:        backfillPoints,
			BalanceAfter:  member.PointBalance + backfillPoints,
			LifetimeAfter: member.LifetimePoints + backfillPoints,
			Notes:         &note,
			ProcessedBy:   &enrolledBy,
		}
		if err := tx.Create(entry).Error; err != nil {
			return err
		}

		newLifetime := member.LifetimePoints + backfillPoints
		newTier := computeTier(newLifetime, cfg.Tiers)

		return tx.Model(&models.LoyaltyMember{}).Where("id = ?", member.ID).Updates(map[string]interface{}{
			"point_balance":   gorm.Expr("point_balance + ?", backfillPoints),
			"lifetime_points": gorm.Expr("lifetime_points + ?", backfillPoints),
			"current_tier":    newTier.Name,
			"tier_badge_color": newTier.BadgeColor,
		}).Error
	})
}

func (u *loyaltyUsecase) linkCustomerToPOSOrderAndSalesOrder(ctx context.Context, posOrderID string, customer *customerModels.Customer) float64 {
	if posOrderID == "" || customer == nil {
		return 0
	}

	now := apptime.Now()
	orderUpdates := map[string]interface{}{
		"customer_id":    customer.ID,
		"customer_name":  customer.Name,
		"customer_phone": customer.ContactPerson,
		"customer_email": customer.Email,
		"updated_at":     now,
	}
	_ = u.db.WithContext(ctx).Table("pos_orders").Where("id = ?", posOrderID).Updates(orderUpdates).Error

	salesOrderUpdates := map[string]interface{}{
		"customer_id":    customer.ID,
		"customer_name":  customer.Name,
		"customer_phone": customer.ContactPerson,
		"customer_email": customer.Email,
		"updated_at":     now,
	}
	_ = u.db.WithContext(ctx).Table("sales_orders").Where("source_pos_order_id = ?", posOrderID).Updates(salesOrderUpdates).Error

	var totalAmount float64
	if err := u.db.WithContext(ctx).
		Table("pos_orders").
		Where("id = ? AND status IN ?", posOrderID, []string{"PAID", "COMPLETED"}).
		Select("COALESCE(total_amount, 0)").
		Scan(&totalAmount).Error; err != nil {
		return 0
	}

	return totalAmount
}

func (u *loyaltyUsecase) resolveTransactionAmount(ctx context.Context, transactionType, transactionID string) float64 {
	if transactionID == "" {
		return 0
	}

	var totalAmount float64
	var tableName string

	switch transactionType {
	case "pos_order":
		tableName = "pos_orders"
	case "sales_order":
		tableName = "sales_orders"
	default:
		return 0
	}

	if err := u.db.WithContext(ctx).
		Table(tableName).
		Where("id = ?", transactionID).
		Select("COALESCE(total_amount, 0)").
		Scan(&totalAmount).Error; err != nil {
		return 0
	}

	return totalAmount
}

func (u *loyaltyUsecase) ListLedger(ctx context.Context, params repositories.LedgerListParams) ([]dto.PointLedgerResponse, *utils.PaginationResult, error) {
	entries, total, err := u.ledgerRepo.ListByMember(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	resp := make([]dto.PointLedgerResponse, 0, len(entries))
	for i := range entries {
		resp = append(resp, *mapper.ToPointLedgerResponse(&entries[i]))
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.PerPage)))
	pagination := &utils.PaginationResult{
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      int(total),
		TotalPages: totalPages,
	}
	return resp, pagination, nil
}

// ─── Public self-registration ─────────────────────────────────────────────────

// PublicSelfRegister allows a walk-in customer to self-register via a QR code on their receipt.
func (u *loyaltyUsecase) PublicSelfRegister(ctx context.Context, req *dto.PublicSelfRegisterRequest) (*dto.PublicSelfRegisterResponse, error) {
	phone := req.Phone
	enrollReq := &dto.EnrollMemberRequest{
		OutletID:   &req.OutletID,
		Name:       req.Name,
		Phone:      &phone,
		Email:      req.Email,
		PosOrderID: &req.OrderID,
	}

	resp, err := u.EnrollMember(ctx, "public", enrollReq)
	if err != nil {
		return nil, err
	}

	// Check if member was already enrolled (EnrollMember returns existing member gracefully).
	alreadyMember := resp.TotalTransactions > 1

	return &dto.PublicSelfRegisterResponse{
		MemberCode:     resp.MemberCode,
		CurrentTier:    resp.CurrentTier,
		TierBadgeColor: resp.TierBadgeColor,
		PointsEarned:   resp.PointBalance,
		PointBalance:   resp.PointBalance,
		AlreadyMember:  alreadyMember,
	}, nil
}

// ─── Private helpers ──────────────────────────────────────────────────────────

func (u *loyaltyUsecase) buildMemberResponse(ctx context.Context, m *models.LoyaltyMember, customer *customerModels.Customer) (*dto.LoyaltyMemberResponse, error) {
	name := ""
	var phone *string
	var enrolledOutletName *string
	if customer != nil {
		name = customer.Name
		if customer.ContactPerson != "" {
			cp := customer.ContactPerson
			phone = &cp
		}
	}

	if m.EnrolledOutletID != nil && *m.EnrolledOutletID != "" {
		outlet, err := u.outletRepo.GetByID(ctx, *m.EnrolledOutletID)
		if err == nil && outlet != nil {
			outletName := outlet.Name
			enrolledOutletName = &outletName
		}
	}

	return mapper.ToLoyaltyMemberResponse(m, name, phone, enrolledOutletName), nil
}

func parseConfig(program *models.LoyaltyProgram) (*dto.LoyaltyConfigJSON, error) {
	var cfg dto.LoyaltyConfigJSON
	if err := json.Unmarshal(program.ConfigJSON, &cfg); err != nil {
		return nil, ErrInvalidLoyaltyConfig
	}
	return &cfg, nil
}

func marshalAndValidateConfig(cfg *dto.LoyaltyConfigJSON) ([]byte, error) {
	if len(cfg.Tiers) == 0 {
		return nil, fmt.Errorf("%w: at least one tier is required", ErrInvalidLoyaltyConfig)
	}
	if cfg.PointRules.AmountPerPoint <= 0 {
		return nil, fmt.Errorf("%w: amount_per_point must be positive", ErrInvalidLoyaltyConfig)
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		return nil, ErrInvalidLoyaltyConfig
	}
	return b, nil
}

// resolveScopedOutletIDs returns the outlet IDs the current user is permitted to manage.
// Returns nil, nil for admin/global users (no restriction).
// Returns a non-empty slice for OUTLET/WAREHOUSE-scoped users.
// Returns ErrLoyaltyForbidden when a scoped user has no warehouse IDs in context.
func (u *loyaltyUsecase) resolveScopedOutletIDs(ctx context.Context) ([]string, error) {
	if !isOutletScopedScope(ctx) {
		return nil, nil
	}

	warehouseIDs, _ := ctx.Value("scope_warehouse_ids").([]string)
	if len(warehouseIDs) == 0 {
		log.Printf("[loyalty] scoped user has no warehouse assignments")
		return []string{}, nil
	}

	outlets, err := u.outletRepo.FindByWarehouseIDs(ctx, warehouseIDs)
	if err != nil {
		return nil, err
	}
	if len(outlets) == 0 {
		log.Printf("[loyalty] OUTLET-scoped user has no outlets for warehouses %v", warehouseIDs)
		return []string{}, nil
	}

	ids := make([]string, 0, len(outlets))
	for _, o := range outlets {
		ids = append(ids, o.ID)
	}
	return ids, nil
}

func isOutletScopedScope(ctx context.Context) bool {
	permissionScope, _ := ctx.Value("permission_scope").(string)
	return permissionScope == "WAREHOUSE" || permissionScope == "OUTLET"
}

// containsString reports whether s is present in the slice.
func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// computeTier returns the highest tier a member qualifies for based on lifetime points.
func computeTier(lifetimePoints int64, tiers []dto.TierConfig) dto.TierConfig {
	best := dto.TierConfig{Name: "Bronze", MinPoints: 0, Multiplier: 1.0, BadgeColor: "#CD7F32"}
	for _, t := range tiers {
		if lifetimePoints >= t.MinPoints && t.MinPoints >= best.MinPoints {
			best = t
		}
	}
	return best
}
