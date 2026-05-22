package presentation

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/gilabs/gims/api/internal/ai/data/repositories"
	"github.com/gilabs/gims/api/internal/ai/domain/mapper"
	"github.com/gilabs/gims/api/internal/ai/domain/usecase"
	aiContext "github.com/gilabs/gims/api/internal/ai/domain/usecase/context"
	"github.com/gilabs/gims/api/internal/ai/domain/usecase/tools"
	"github.com/gilabs/gims/api/internal/ai/presentation/handler"
	"github.com/gilabs/gims/api/internal/ai/presentation/router"
	"github.com/gilabs/gims/api/internal/core/apptime"
	coreUsecase "github.com/gilabs/gims/api/internal/core/domain/usecase"
	"github.com/gilabs/gims/api/internal/core/infrastructure/cerebras"
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	financeUsecase "github.com/gilabs/gims/api/internal/finance/domain/usecase"

	hrdUsecase "github.com/gilabs/gims/api/internal/hrd/domain/usecase"
	inventoryUsecase "github.com/gilabs/gims/api/internal/inventory/domain/usecase"
	purchaseUsecase "github.com/gilabs/gims/api/internal/purchase/domain/usecase"
	salesUsecase "github.com/gilabs/gims/api/internal/sales/domain/usecase"
)

// AIDeps holds the resolved domain usecase dependencies for AI actions
type AIDeps struct {
	HolidayUC         hrdUsecase.HolidayUsecase
	LeaveRequestUC    hrdUsecase.LeaveRequestUsecase
	AttendanceUC      hrdUsecase.AttendanceRecordUsecase
	SalesQuotationUC  salesUsecase.SalesQuotationUsecase
	SalesOrderUC      salesUsecase.SalesOrderUsecase
	DeliveryOrderUC   salesUsecase.DeliveryOrderUsecase
	CustomerInvoiceUC salesUsecase.CustomerInvoiceUsecase
	InventoryUC       inventoryUsecase.InventoryUsecase
	PurchaseOrderUC   purchaseUsecase.PurchaseOrderUsecase
	PurchaseReqUC     purchaseUsecase.PurchaseRequisitionUsecase
	GoodsReceiptUC    purchaseUsecase.GoodsReceiptUsecase
	SupplierInvoiceUC purchaseUsecase.SupplierInvoiceUsecase
	CoaUC             financeUsecase.ChartOfAccountUsecase
	JournalUC         financeUsecase.JournalEntryUsecase
	FinancePaymentUC  financeUsecase.PaymentUsecase
	BudgetUC          financeUsecase.BudgetUsecase
	CashBankUC        financeUsecase.CashBankJournalUsecase
	TaxInvoiceUC      financeUsecase.TaxInvoiceUsecase
	AssetUC           financeUsecase.AssetUsecase
	SalaryUC          hrdUsecase.SalaryStructureUsecase
	BankAccountUC     coreUsecase.BankAccountUsecase
}

// RegisterRoutes registers all AI assistant routes
func RegisterRoutes(
	_ *gin.Engine,
	api *gin.RouterGroup,
	db *gorm.DB,
	jwtManager *jwt.JWTManager,
	permService interface {
		GetPermissions(roleCode string) ([]string, error)
		GetPermissionsWithScope(roleCode string) (map[string]string, error)
	},
	cerebrasClient *cerebras.Client,
	deps *AIDeps,
) {
	// Initialize AI repositories
	sessionRepo := repositories.NewChatSessionRepository(db)
	messageRepo := repositories.NewChatMessageRepository(db)
	actionRepo := repositories.NewActionLogRepository(db)
	intentRepo := repositories.NewIntentRegistryRepository(db)

	// Initialize AI domain components
	chatMapper := mapper.NewChatMapper()
	intentResolver := usecase.NewIntentResolver(cerebrasClient, intentRepo)
	paramExtractor := usecase.NewParameterExtractor(cerebrasClient, intentRepo)
	permValidator := usecase.NewPermissionValidator(intentRepo)
	entityResolver := usecase.NewEntityResolver(db)
	requestValidator := usecase.NewRequestValidator(db, entityResolver)

	executorDeps := &usecase.ActionExecutorDeps{}
	if deps != nil {
		executorDeps.HolidayUsecase = deps.HolidayUC
		executorDeps.LeaveRequestUsecase = deps.LeaveRequestUC
		executorDeps.AttendanceUsecase = deps.AttendanceUC
		executorDeps.SalesQuotationUsecase = deps.SalesQuotationUC
		executorDeps.SalesOrderUsecase = deps.SalesOrderUC
		executorDeps.DeliveryOrderUsecase = deps.DeliveryOrderUC
		executorDeps.CustomerInvoiceUsecase = deps.CustomerInvoiceUC
		executorDeps.InventoryUsecase = deps.InventoryUC
		executorDeps.PurchaseOrderUsecase = deps.PurchaseOrderUC
		executorDeps.PurchaseRequisitionUsecase = deps.PurchaseReqUC
		executorDeps.GoodsReceiptUsecase = deps.GoodsReceiptUC
		executorDeps.SupplierInvoiceUsecase = deps.SupplierInvoiceUC
		executorDeps.CoaUsecase = deps.CoaUC
		executorDeps.JournalUsecase = deps.JournalUC
		executorDeps.FinancePaymentUsecase = deps.FinancePaymentUC
		executorDeps.BudgetUsecase = deps.BudgetUC
		executorDeps.CashBankUsecase = deps.CashBankUC
		executorDeps.TaxInvoiceUsecase = deps.TaxInvoiceUC
		executorDeps.AssetUsecase = deps.AssetUC
		executorDeps.SalaryUsecase = deps.SalaryUC
		executorDeps.BankAccountUsecase = deps.BankAccountUC
	}
	actionExecutor := usecase.NewActionExecutor(executorDeps, entityResolver)

	// Initialize usecase
	aiChatUC := usecase.NewAIChatUsecase(
		sessionRepo,
		messageRepo,
		actionRepo,
		intentRepo,
		cerebrasClient,
		chatMapper,
		intentResolver,
		paramExtractor,
		requestValidator,
		permValidator,
		entityResolver,
		actionExecutor,
	)

	// Initialize handlers (legacy pipeline — kept for backward compatibility)
	chatHandler := handler.NewChatHandler(aiChatUC, cerebrasClient)
	sessionHandler := handler.NewSessionHandler(aiChatUC)
	adminHandler := handler.NewAdminHandler(aiChatUC)

	// ── V2: Engine-based pipeline with tool registry & streaming ──
	// Build the tool registry by bridging AI intent registry entries
	toolRegistry := tools.NewRegistry()
	if intentRepo != nil {
		// Bridge: wrap ActionExecutor.Execute into the ActionExecutorFunc signature
		executorFunc := tools.ActionExecutorFunc(func(ctx context.Context, intentCode string, params map[string]interface{}, userID string, resolvedEntities map[string]*tools.ResolvedEntity) *tools.ToolResult {
			start := apptime.Now()

			// Convert to legacy IntentResult
			legacyIntent := &usecase.IntentResult{
				IntentCode: intentCode,
				Parameters: params,
			}

			// Convert resolved entities to legacy format
			var legacyResolved map[string]*usecase.ResolvedEntity
			if len(resolvedEntities) > 0 {
				legacyResolved = make(map[string]*usecase.ResolvedEntity, len(resolvedEntities))
				for k, v := range resolvedEntities {
					legacyResolved[k] = &usecase.ResolvedEntity{
						ID:          v.ID,
						DisplayName: v.Name,
						EntityType:  v.Type,
					}
				}
			}

			result := actionExecutor.Execute(ctx, legacyIntent, legacyResolved, userID)
			if result == nil {
				return &tools.ToolResult{
					Success:      false,
					ErrorMessage: "action executor returned nil",
					Action:       "execute",
					DurationMs:   apptime.Now().Sub(start).Milliseconds(),
				}
			}

			return &tools.ToolResult{
				Success:      result.Success,
				Data:         result.Data,
				Message:      result.Message,
				EntityType:   result.EntityType,
				EntityID:     result.EntityID,
				Action:       result.Action,
				DurationMs:   result.DurationMs,
				ErrorCode:    result.ErrorCode,
				ErrorMessage: result.ErrorMessage,
			}
		})

		// Bridge: wrap EntityResolver for parameter entity resolution
		resolverFunc := tools.EntityResolverFunc(func(ctx context.Context, params map[string]interface{}) (map[string]*tools.ResolvedEntity, error) {
			legacyResolved, err := entityResolver.ResolveEntitiesFromParameters(ctx, params)
			if err != nil {
				return nil, err
			}
			resolved := make(map[string]*tools.ResolvedEntity, len(legacyResolved))
			for k, v := range legacyResolved {
				resolved[k] = &tools.ResolvedEntity{
					ID:   v.ID,
					Name: v.DisplayName,
					Type: v.EntityType,
				}
			}
			return resolved, nil
		})

		bgCtx := context.Background()
		if regErr := tools.RegisterFromIntentRegistry(bgCtx, toolRegistry, intentRepo, executorFunc, resolverFunc); regErr != nil {
			log.Printf("[AI] Warning: failed to register tools from intent registry: %v", regErr)
		} else {
			log.Printf("[AI] Registered %d tools from intent registry", toolRegistry.Count())
		}
	}

	// Build context builder for system prompt assembly
	contextBuilder := aiContext.NewBuilder(toolRegistry)

	// Initialize engine-based usecase
	chatEngineUC := usecase.NewChatEngineUsecase(
		sessionRepo,
		messageRepo,
		actionRepo,
		cerebrasClient,
		toolRegistry,
		contextBuilder,
		entityResolver,
	)

	// Initialize streaming handler
	streamHandler := handler.NewStreamHandler(chatEngineUC, cerebrasClient)

	// Create AI group under API with auth middleware
	aiGroup := api.Group("/ai")
	aiGroup.Use(middleware.AuthMiddleware(jwtManager, permService))
	aiGroup.Use(middleware.ScopeMiddleware(db))

	// Register routes — legacy and v2 coexist
	router.RegisterChatRoutes(aiGroup, chatHandler)
	router.RegisterV2ChatRoutes(aiGroup, streamHandler)
	router.RegisterSessionRoutes(aiGroup, sessionHandler)
	router.RegisterAdminRoutes(aiGroup, adminHandler)
}
