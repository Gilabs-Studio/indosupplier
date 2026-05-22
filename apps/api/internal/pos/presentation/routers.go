package presentation

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	coreRepos "github.com/gilabs/gims/api/internal/core/data/repositories"
	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	securityInfra "github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/core/middleware"
	feedbackDataRepos "github.com/gilabs/gims/api/internal/feedback/data/repositories"
	feedbackUC "github.com/gilabs/gims/api/internal/feedback/domain/usecase"
	invDataRepos "github.com/gilabs/gims/api/internal/inventory/data/repositories"
	invUsecase "github.com/gilabs/gims/api/internal/inventory/domain/usecase"
	loyaltyUsecasePkg "github.com/gilabs/gims/api/internal/loyalty/domain/usecase"
	orgRepo "github.com/gilabs/gims/api/internal/organization/data/repositories"
	"github.com/gilabs/gims/api/internal/pos/data/repositories"
	"github.com/gilabs/gims/api/internal/pos/domain/provider"
	"github.com/gilabs/gims/api/internal/pos/domain/usecase"
	"github.com/gilabs/gims/api/internal/pos/infrastructure/ws"
	"github.com/gilabs/gims/api/internal/pos/presentation/handler"
	"github.com/gilabs/gims/api/internal/pos/presentation/router"
	salesRepos "github.com/gilabs/gims/api/internal/sales/data/repositories"
	salesUsecase "github.com/gilabs/gims/api/internal/sales/domain/usecase"
)

// RegisterRoutes registers all POS domain routes under /api/v1/pos.
// loyaltyUC is optional; pass nil to disable loyalty integration.
func RegisterRoutes(r *gin.Engine, api *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.JWTManager, permService interface {
	GetPermissions(roleCode string) ([]string, error)
	GetPermissionsWithScope(roleCode string) (map[string]string, error)
}, loyaltyUC loyaltyUsecasePkg.LoyaltyUsecase, customerInvoiceUC salesUsecase.CustomerInvoiceUsecase, salesPaymentUC salesUsecase.SalesPaymentUsecase) {
	credentialCipher, err := securityInfra.NewCredentialCipher(config.AppConfig.Security.XenditCredentialEncryptionKey)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize Xendit credential cipher: %v", err))
	}

	// ─── Repositories ────────────────────────────────────────────────────────

	floorPlanRepo := repositories.NewFloorPlanRepository(db)
	orderRepo := repositories.NewPosOrderRepository(db)
	productRepo := repositories.NewPOSProductRepository(db)
	paymentRepo := repositories.NewPOSPaymentRepository(db)
	configRepo := repositories.NewPOSConfigRepository(db)
	xenditRepo := repositories.NewXenditConfigRepository(db)
	bankAccountRepo := coreRepos.NewBankAccountRepository(db)
	sessionRepo := repositories.NewPosSessionRepository(db)
	qrTokenRepo := repositories.NewTableQRTokenRepository(db)
	deviceTokenRepo := repositories.NewPOSDeviceTokenRepository(db)

	outletRepo := orgRepo.NewOutletRepository(db)
	salesOrderRepo := salesRepos.NewSalesOrderRepository(db)
	invoiceRepo := salesRepos.NewCustomerInvoiceRepository(db)
	salesPaymentRepo := salesRepos.NewSalesPaymentRepository(db)

	invRepo := invDataRepos.NewInventoryRepository(db)
	recipeService := invUsecase.NewRecipeConsumptionService(db, invRepo)
	posHub := ws.DefaultPosHub()

	// ─── Usecases ────────────────────────────────────────────────────────────

	floorPlanUC := usecase.NewFloorPlanUsecase(floorPlanRepo, outletRepo, qrTokenRepo)
	orderUC := usecase.NewPOSOrderUsecase(db, orderRepo, outletRepo, productRepo, configRepo, recipeService)
	orderUC = orderUC.WithPOSHub(posHub)
	paymentUC := usecase.NewPOSPaymentUsecase(db, paymentRepo, orderRepo, configRepo, xenditRepo, orderUC, salesOrderRepo, invoiceRepo, customerInvoiceUC, salesPaymentUC, salesPaymentRepo, bankAccountRepo, credentialCipher)
	paymentUC = paymentUC.WithPOSHub(posHub)
	configUC := usecase.NewPOSConfigUsecase(configRepo)
	xenditUC := usecase.NewXenditConfigUsecase(xenditRepo, credentialCipher)
	sessionUC := usecase.NewPOSSessionUsecase(sessionRepo, outletRepo)
	deviceTokenUC := usecase.NewPOSDeviceTokenUsecase(deviceTokenRepo, outletRepo)

	// ─── Handlers ────────────────────────────────────────────────────────────

	floorPlanH := handler.NewFloorPlanHandler(floorPlanUC)
	orderH := handler.NewPOSOrderHandler(orderUC)

	// Build receipt handler and attach feedback usecase for QR code generation
	// (best-effort: silently disabled when no active form exists for an outlet).
	receiptH := handler.NewPOSReceiptHandler(orderUC, paymentRepo, configRepo, outletRepo)
	fbUC := feedbackUC.NewFeedbackUsecase(
		feedbackDataRepos.NewFeedbackFormRepository(db),
		feedbackDataRepos.NewFeedbackTokenRepository(db),
		feedbackDataRepos.NewFeedbackResponseRepository(db),
		outletRepo,
	)
	receiptH = receiptH.WithFeedbackUsecase(fbUC)
	if loyaltyUC != nil {
		receiptH = receiptH.WithLoyaltyUsecase(loyaltyUC)
		paymentUC = paymentUC.WithLoyaltyUsecase(loyaltyUC)
	}
	paymentH := handler.NewPOSPaymentHandler(paymentUC)
	configH := handler.NewPOSConfigHandler(configUC)
	xenditH := handler.NewXenditConfigHandler(xenditUC)
	sessionH := handler.NewPOSSessionHandler(sessionUC)
	deviceTokenH := handler.NewPOSDeviceTokenHandler(deviceTokenUC)

	// ─── Public self-order handler ────────────────────────────────────────────

	publicPOSUC := usecase.NewPublicPOSUsecase(qrTokenRepo, orderUC, paymentUC, configRepo, outletRepo, productRepo, posHub, deviceTokenRepo, provider.NewMultiPlatformPushNotifier(), db)
	publicPOSH := handler.NewPublicPOSHandler(publicPOSUC)
	posWSH := handler.NewPOSWSHandler(posHub, outletRepo)

	// ─── Route group ─────────────────────────────────────────────────────────

	group := api.Group("/pos")
	group.Use(middleware.AuthMiddleware(jwtManager, permService))
	group.Use(middleware.ScopeMiddleware(db))

	router.RegisterFloorPlanRoutes(group, floorPlanH)
	router.RegisterPOSOrderRoutes(group, orderH, receiptH)
	router.RegisterPOSPaymentRoutes(group, paymentH)
	router.RegisterPOSConfigRoutes(group, configH)
	router.RegisterXenditConfigRoutes(group, xenditH)
	router.RegisterPOSSessionRoutes(group, sessionH)
	group.POST("/device-token", middleware.RequirePermission("pos.order.read"), deviceTokenH.Register)

	// Staff WebSocket — requires auth (already applied by group.Use above)
	group.GET("/ws", middleware.RequirePermission("pos.order.read"), posWSH.Subscribe)

	// Public self-order routes — unauthenticated, rate-limited
	router.RegisterPublicPOSRoutes(r, publicPOSH)

	// Xendit webhook is unauthenticated (token verified inside the handler by Xendit signature)
	r.POST("/api/v1/pos/payments/xendit/webhook", paymentH.XenditWebhook)
}
