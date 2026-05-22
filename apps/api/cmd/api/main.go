package main

import (
	"context"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gilabs/gims/api/internal/core/apptime"
	coreRepos "github.com/gilabs/gims/api/internal/core/data/repositories"
	coreUsecase "github.com/gilabs/gims/api/internal/core/domain/usecase"
	"github.com/gilabs/gims/api/internal/core/infrastructure/audit"
	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/events"
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/infrastructure/redis"
	coreRouter "github.com/gilabs/gims/api/internal/core/infrastructure/router"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/core/infrastructure/ws"
	"github.com/gilabs/gims/api/internal/core/logger"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/core/storage"
	notificationWs "github.com/gilabs/gims/api/internal/notification/infrastructure/ws"
	posWs "github.com/gilabs/gims/api/internal/pos/infrastructure/ws"
	"github.com/gilabs/gims/api/seeders"

	authUsecase "github.com/gilabs/gims/api/internal/auth/domain/usecase"
	authHandler "github.com/gilabs/gims/api/internal/auth/presentation/handler"
	authRouter "github.com/gilabs/gims/api/internal/auth/presentation/router"

	permissionRepo "github.com/gilabs/gims/api/internal/permission/data/repositories"
	permissionUsecase "github.com/gilabs/gims/api/internal/permission/domain/usecase"
	permissionHandler "github.com/gilabs/gims/api/internal/permission/presentation/handler"
	permissionRouter "github.com/gilabs/gims/api/internal/permission/presentation/router"

	refreshTokenRepo "github.com/gilabs/gims/api/internal/refresh_token/data/repositories"
	refreshTokenWorker "github.com/gilabs/gims/api/internal/refresh_token/worker"

	roleRepo "github.com/gilabs/gims/api/internal/role/data/repositories"
	roleUsecase "github.com/gilabs/gims/api/internal/role/domain/usecase"
	roleHandler "github.com/gilabs/gims/api/internal/role/presentation/handler"
	roleRouter "github.com/gilabs/gims/api/internal/role/presentation/router"

	userRepo "github.com/gilabs/gims/api/internal/user/data/repositories"
	userUsecase "github.com/gilabs/gims/api/internal/user/domain/usecase"
	userHandler "github.com/gilabs/gims/api/internal/user/presentation/handler"
	userRouter "github.com/gilabs/gims/api/internal/user/presentation/router"

	passwordResetRepo "github.com/gilabs/gims/api/internal/password_reset/data/repositories"
	passwordResetUsecase "github.com/gilabs/gims/api/internal/password_reset/domain/usecase"
	passwordResetHandler "github.com/gilabs/gims/api/internal/password_reset/presentation/handler"
	passwordResetRouter "github.com/gilabs/gims/api/internal/password_reset/presentation/router"

	notificationRepo "github.com/gilabs/gims/api/internal/notification/data/repositories"

	corePresentation "github.com/gilabs/gims/api/internal/core/presentation"
	customerRepos "github.com/gilabs/gims/api/internal/customer/data/repositories"
	customerPresentation "github.com/gilabs/gims/api/internal/customer/presentation"
	financePresentation "github.com/gilabs/gims/api/internal/finance/presentation"
	geographicPresentation "github.com/gilabs/gims/api/internal/geographic/presentation"
	hrdPresentation "github.com/gilabs/gims/api/internal/hrd/presentation"
	hrdWorker "github.com/gilabs/gims/api/internal/hrd/worker"
	inventoryRepo "github.com/gilabs/gims/api/internal/inventory/data/repositories" // Import repo
	inventoryUsecase "github.com/gilabs/gims/api/internal/inventory/domain/usecase" // Import usecase
	inventoryPresentation "github.com/gilabs/gims/api/internal/inventory/presentation"
	notificationPresentation "github.com/gilabs/gims/api/internal/notification/presentation"
	orgRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	organizationPresentation "github.com/gilabs/gims/api/internal/organization/presentation"
	productPresentation "github.com/gilabs/gims/api/internal/product/presentation"
	purchasePresentation "github.com/gilabs/gims/api/internal/purchase/presentation"
	salesPresentation "github.com/gilabs/gims/api/internal/sales/presentation"
	stockOpnamePresentation "github.com/gilabs/gims/api/internal/stock_opname/presentation"
	supplierPresentation "github.com/gilabs/gims/api/internal/supplier/presentation"
	travelPlannerPresentation "github.com/gilabs/gims/api/internal/travel_planner/presentation"
	warehousePresentation "github.com/gilabs/gims/api/internal/warehouse/presentation"

	crmPresentation "github.com/gilabs/gims/api/internal/crm/presentation"
	feedbackPresentation "github.com/gilabs/gims/api/internal/feedback/presentation"
	generalPresentation "github.com/gilabs/gims/api/internal/general/presentation"
	loyaltyPresentation "github.com/gilabs/gims/api/internal/loyalty/presentation"
	posPresentation "github.com/gilabs/gims/api/internal/pos/presentation"
	reportPresentation "github.com/gilabs/gims/api/internal/report/presentation"

	aiPresentation "github.com/gilabs/gims/api/internal/ai/presentation"
	"github.com/gilabs/gims/api/internal/core/infrastructure/cerebras"

	tenantRepos "github.com/gilabs/gims/api/internal/tenant/data/repositories"
	tenantUsecase "github.com/gilabs/gims/api/internal/tenant/domain/usecase"
	tenantHandler "github.com/gilabs/gims/api/internal/tenant/presentation/handler"
	tenantRouter "github.com/gilabs/gims/api/internal/tenant/presentation/router"
	tenantWorker "github.com/gilabs/gims/api/internal/tenant/worker"

	tenantScope "github.com/gilabs/gims/api/internal/core/infrastructure/tenant"
	xenditClient "github.com/gilabs/gims/api/internal/core/infrastructure/xendit"
)

func initInfrastructure() {
	// Initialize logger
	logger.Init()

	// Load configuration
	if err := config.Load(); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize application timezone (must be before any time-dependent logic)
	apptime.Init(config.AppConfig.Server.Timezone)

	// Initialize Cloudflare R2 storage
	if err := storage.Init(
		config.AppConfig.Storage.R2AccountID,
		config.AppConfig.Storage.R2AccessKeyID,
		config.AppConfig.Storage.R2SecretAccessKey,
		config.AppConfig.Storage.R2BucketName,
		config.AppConfig.Storage.R2PublicURL,
	); err != nil {
		log.Fatal("Failed to initialize R2 storage:", err)
	}

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	// Note: Defer database.Close() must be handled in main if we returned db,
	// but here we are using global or singleton access patterns in this codebase
	// seeing later usage of database.DB.
	// However, original main had defer database.Close().
	// We will keep defer in main and just call connect here.

	// Connect to Redis
	if err := redis.InitRedis(config.AppConfig); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	}
	// Start the Redis Pub/Sub bridge for cross-instance location broadcasting.
	// This enables horizontal scaling — each pod fans out events to its local WS clients.
	ws.InitLocationPubSub()
	posWs.InitPosPubSub()
	notificationWs.InitNotificationPubSub()
	// Defer redis.Close() also needs to be in main

	// Run migrations (optionally only once)
	migrateOnce := os.Getenv("MIGRATE_ONCE") == "true"
	migrateFlagFile := os.Getenv("MIGRATE_FLAG_FILE")
	if migrateFlagFile == "" {
		migrateFlagFile = "/app/.migrated"
	}

	if config.AppConfig.Startup.RunMigrations {
		if migrateOnce {
			if _, err := os.Stat(migrateFlagFile); err == nil {
				log.Println("Skipping migrations (already migrated)")
			} else {
				if err := database.AutoMigrate(); err != nil {
					log.Fatal("Failed to run migrations:", err)
				}
				_ = os.WriteFile(migrateFlagFile, []byte(time.Now().Format(time.RFC3339)), 0644)
			}
		} else {
			if err := database.AutoMigrate(); err != nil {
				log.Fatal("Failed to run migrations:", err)
			}
		}
	} else {
		log.Println("Skipping migrations (RUN_MIGRATIONS=false)")
	}

	// Seed data (optionally only once)
	seedOnce := os.Getenv("SEED_ONCE") == "true"
	seedFlagFile := os.Getenv("SEED_FLAG_FILE")
	if seedFlagFile == "" {
		seedFlagFile = "/app/.seeded"
	}

	if config.AppConfig.Startup.RunSeeders {
		if seedOnce {
			if _, err := os.Stat(seedFlagFile); err == nil {
				log.Println("Skipping seeders (already seeded)")
			} else {
				if err := seeders.SeedAll(); err != nil {
					log.Fatal("Failed to seed data:", err)
				}
				_ = os.WriteFile(seedFlagFile, []byte(time.Now().Format(time.RFC3339)), 0644)
			}
		} else {
			if err := seeders.SeedAll(); err != nil {
				log.Fatal("Failed to seed data:", err)
			}
		}
	} else {
		log.Println("Skipping seeders (RUN_SEEDERS=false)")
	}

	// Register per-company timezone provider so apptime can resolve
	// employee/company-specific timezones from the database.
	// WHY: Must be after migrations+seeders so the companies table exists.
	tzProvider := orgRepos.NewCompanyTimezoneProvider(database.DB)
	apptime.RegisterProvider(tzProvider)
}

func setupJWT() *jwt.JWTManager {
	accessSecret := config.AppConfig.JWT.AccessSecretKey
	refreshSecret := config.AppConfig.JWT.RefreshSecretKey
	if accessSecret == "" {
		accessSecret = config.AppConfig.JWT.SecretKey
	}
	if refreshSecret == "" {
		refreshSecret = config.AppConfig.JWT.SecretKey
	}

	accessKeys := jwt.ParseKeyRing(config.AppConfig.JWT.AccessKeysRaw)
	refreshKeys := jwt.ParseKeyRing(config.AppConfig.JWT.RefreshKeysRaw)
	accessKID := config.AppConfig.JWT.AccessKID
	refreshKID := config.AppConfig.JWT.RefreshKID

	// If an active kid is set and exists in the ring, prefer it for signing.
	if accessKID != "" {
		if s, ok := accessKeys[accessKID]; ok && s != "" {
			accessSecret = s
		}
	}
	if refreshKID != "" {
		if s, ok := refreshKeys[refreshKID]; ok && s != "" {
			refreshSecret = s
		}
	}

	return jwt.NewJWTManager(jwt.Options{
		AccessSecretKey:  accessSecret,
		RefreshSecretKey: refreshSecret,
		AccessKeys:       accessKeys,
		RefreshKeys:      refreshKeys,
		AccessKID:        accessKID,
		RefreshKID:       refreshKID,
		Issuer:           config.AppConfig.JWT.Issuer,
		AccessTokenTTL:   time.Duration(config.AppConfig.JWT.AccessTokenTTL) * time.Hour,
		RefreshTokenTTL:  time.Duration(config.AppConfig.JWT.RefreshTokenTTL) * 24 * time.Hour,
	})
}

func main() {
	// 1. Initialize Infrastructure (Config, Logger, DB, Migrations)
	initInfrastructure()

	// Ensure cleanup
	defer database.Close()
	defer redis.Close()
	defer ws.StopLocationPubSub()
	defer posWs.StopPosPubSub()
	defer notificationWs.StopNotificationPubSub()

	// 2. Setup JWT
	jwtManager := setupJWT()

	// Setup repositories
	refreshTokenRepository := refreshTokenRepo.NewRefreshTokenRepository(database.DB)
	userRepository := userRepo.NewUserRepository(database.DB)
	roleRepository := roleRepo.NewRoleRepository(database.DB)
	permissionRepository := permissionRepo.NewPermissionRepository(database.DB)
	menuRepository := permissionRepo.NewMenuRepository(database.DB)
	passwordResetRepository := passwordResetRepo.NewPasswordResetRequestRepository(database.DB)
	notificationRepository := notificationRepo.NewNotificationRepository(database.DB)
	_ = menuRepository // potentially unused in main, but good to init if needed later

	// SaaS: System admin & tenant repositories
	sysAdminRepository := tenantRepos.NewSystemAdminRepository(database.DB)
	tenantRepository := tenantRepos.NewTenantRepository(database.DB)
	couponRepository := tenantRepos.NewCouponRepository(database.DB)
	subscriptionRepository := tenantRepos.NewSubscriptionRepository(database.DB)
	subscriptionPlanRepository := tenantRepos.NewSubscriptionPlanRepository(database.DB)

	// Setup Services
	permissionService := security.NewPermissionService(database.DB)
	auditService := audit.NewAuditService(database.DB)

	// Setup Event Publisher (NoOp for now - logs events to stdout)
	// Replace with Kafka publisher when Kafka is configured
	eventPublisher := events.NewNoOpEventPublisher(true) // Enable logging for development

	// Setup Usecases
	couponUC := tenantUsecase.NewCouponUsecase(couponRepository, subscriptionRepository, database.DB)
	subUC := tenantUsecase.NewSubscriptionUsecase(subscriptionRepository, tenantRepository)
	planUC := tenantUsecase.NewSubscriptionPlanUsecase(subscriptionPlanRepository)
	authUC := authUsecase.NewAuthUsecase(database.DB, userRepository, refreshTokenRepository, jwtManager, eventPublisher, couponUC, redis.GetClient(), subscriptionPlanRepository)
	userUC := userUsecase.NewUserUsecase(userRepository, roleRepository, tenantRepository, subscriptionRepository, auditService, eventPublisher, redis.GetClient())
	roleUC := roleUsecase.NewRoleUsecase(roleRepository, eventPublisher, redis.GetClient(), permissionService, database.DB, subscriptionPlanRepository)
	permissionUC := permissionUsecase.NewPermissionUsecase(permissionRepository, userRepository)
	passwordResetUC := passwordResetUsecase.NewPasswordResetUsecase(passwordResetRepository, userRepository, roleRepository, notificationRepository, auditService, eventPublisher, redis.GetClient())

	// Setup Handlers
	authH := authHandler.NewAuthHandler(authUC, couponUC, planUC)
	xenditH := authHandler.NewXenditHandler(authUC)
	userH := userHandler.NewUserHandler(userUC)
	roleH := roleHandler.NewRoleHandler(roleUC)
	permissionH := permissionHandler.NewPermissionHandler(permissionUC)
	passwordResetH := passwordResetHandler.NewPasswordResetHandler(passwordResetUC)

	// SaaS: System admin usecases & handler
	sysAdminUC := tenantUsecase.NewSystemAdminUsecase(sysAdminRepository, jwtManager)
	sysAdminTenantUC := tenantUsecase.NewTenantUsecase(tenantRepository)
	sysAdminH := tenantHandler.NewSystemAdminHandler(sysAdminUC, sysAdminTenantUC, couponUC, subUC, planUC, permissionRepository)

	// Payment transaction usecase
	paymentTxnRepository := tenantRepos.NewPaymentTransactionRepository(database.DB)
	paymentTxnUC := tenantUsecase.NewPaymentTransactionUsecase(paymentTxnRepository, subscriptionRepository)

	billingH := tenantHandler.NewBillingHandler(subUC, paymentTxnUC, subscriptionPlanRepository, authUC)

	// Payment method: saved card token management for auto-renewal.
	paymentMethodRepo := tenantRepos.NewPaymentMethodRepository(database.DB)
	paymentMethodUC := tenantUsecase.NewPaymentMethodUsecase(paymentMethodRepo, xenditClient.NewClient(
		func() string {
			if config.AppConfig != nil {
				return config.AppConfig.Xendit.SecretKey
			}
			return ""
		}(),
		func() string {
			if config.AppConfig != nil {
				return config.AppConfig.Xendit.BaseURL
			}
			return ""
		}(),
	))
	paymentMethodH := tenantHandler.NewPaymentMethodHandler(paymentMethodUC)

	// Setup refresh token cleanup worker
	// Run every 24 hours to clean up expired refresh tokens
	rtWorker := refreshTokenWorker.NewRefreshTokenCleanupWorker(
		refreshTokenRepository,
		24*time.Hour,
	)
	rtWorker.Start()

	// Cleanup worker for tenants that passed 30-day deletion grace period.
	delWorker := tenantWorker.NewTenantDeletionWorker(database.DB, 12*time.Hour)
	delWorker.Start()

	// Setup Router
	r := coreRouter.NewEngine(jwtManager)

	// Metrics (counts/avg latency). Endpoint is token-gated and only enabled via env.
	r.Use(middleware.MetricsMiddleware())
	if config.AppConfig.Observability.MetricsEnabled {
		r.GET("/metrics", middleware.MetricsHandler())
		log.Println("Metrics enabled at /metrics (token required)")
	}

	// pprof (debugging). Token-gated and only enabled via env (non-prod enforced in config).
	if config.AppConfig.Observability.PprofEnabled {
		pp := r.Group("/debug/pprof")
		pp.Use(func(c *gin.Context) {
			if c.GetHeader("X-Internal-Token") != config.AppConfig.Observability.PprofToken {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			c.Next()
		})
		pp.GET("/", gin.WrapF(pprof.Index))
		pp.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		pp.GET("/profile", gin.WrapF(pprof.Profile))
		pp.POST("/symbol", gin.WrapF(pprof.Symbol))
		pp.GET("/symbol", gin.WrapF(pprof.Symbol))
		pp.GET("/trace", gin.WrapF(pprof.Trace))
		pp.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
		pp.GET("/block", gin.WrapH(pprof.Handler("block")))
		pp.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		pp.GET("/heap", gin.WrapH(pprof.Handler("heap")))
		pp.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
		pp.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
		log.Println("pprof enabled at /debug/pprof (token required)")
	}

	// Health check endpoints
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "API is running",
		})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// SaaS: Register system admin routes (completely separate from tenant API)
	tenantRouter.RegisterSystemAdminRoutes(r, sysAdminH, jwtManager, sysAdminRepository)

	// SaaS: Register tenant auto-inject callback on Create
	tenantScope.SetTenantCallback(database.DB)

	// API v1 routes
	var autoAbsentWorker *hrdWorker.AutoAbsentWorker
	v1 := r.Group("/api/v1")
	// TenantResolverMiddleware is intentionally NOT registered here.
	// Tenant isolation is now handled by AuthMiddleware, which reads tenant_id
	// directly from the JWT claim and sets it in context before any handler runs.
	// This eliminates the middleware ordering bug where TenantResolver ran before Auth.
	{
		v1.GET("/", func(c *gin.Context) {
			response.SuccessResponse(c, gin.H{
				"message": "API v1",
				"version": "1.0.0",
			}, nil)
		})

		authRouter.RegisterAuthRoutes(v1, authH, jwtManager, permissionService)
		authRouter.RegisterWebhookRoutes(v1, xenditH)

		// Tenant billing: authenticated users can query their subscription + entitlements.
		billing := v1.Group("/billing")
		billing.Use(middleware.AuthMiddleware(jwtManager, permissionService))
		{
			billing.GET("/subscription", billingH.GetMySubscription)
			billing.GET("/entitlements", billingH.GetMyEntitlements)
			billing.GET("/payment-history", billingH.GetMyPaymentHistory)
			billing.POST("/change", billingH.ChangeSubscription)
			billing.POST("/change/confirm", billingH.ConfirmSubscriptionChange)
			billing.DELETE("/subscription", billingH.CancelSubscription)
			// Payment method (saved card tokens for auto-renewal)
			billing.GET("/payment-methods", paymentMethodH.ListPaymentMethods)
			billing.POST("/payment-methods", paymentMethodH.AddPaymentMethod)
			billing.PATCH("/payment-methods/:id/default", paymentMethodH.SetDefaultPaymentMethod)
			billing.DELETE("/payment-methods/:id", paymentMethodH.RemovePaymentMethod)
		}
		passwordResetRouter.RegisterPasswordResetRoutes(v1, passwordResetH, jwtManager, permissionService)
		userRouter.RegisterUserRoutes(v1, userH, permissionH, jwtManager, permissionService)
		roleRouter.RegisterRoleRoutes(v1, roleH, jwtManager, permissionService)
		permissionRouter.RegisterPermissionRoutes(v1, permissionH, jwtManager, permissionService)
		coreRouter.RegisterUploadRoutes(v1, jwtManager, permissionService)
		notificationPresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService)

		// Geographic module (Sprint 1)
		geographicPresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService)

		// Organization module (Sprint 2)
		organizationPresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService)

		// Supplier module (Sprint 4)
		supplierPresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService)

		// Customer module (Master Data)
		customerPresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService)

		// Product module (Sprint 4)
		productPresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService)

		// Warehouse module (Sprint 4)
		warehousePresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService)

		// Core Master Data (Sprint 4 - PaymentTerms, CourierAgency, SOSource, LeaveType)
		corePresentation.RegisterMasterDataRoutes(r, v1, database.DB, jwtManager, permissionService)

		// Finance module (Sprint 10 - COA & Journals)
		financeDeps := financePresentation.RegisterRoutes(r, v1, database.DB, jwtManager, auditService, permissionService)

		// Inventory Setup (Shared Dependency)
		invRepo := inventoryRepo.NewInventoryRepository(database.DB)
		invUC := inventoryUsecase.NewInventoryUsecase(database.DB, invRepo, financeDeps.JournalUC, financeDeps.Engine)

		// Sales module (Sprint 5 - Sales Quotation)
		salesDeps := salesPresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService, salesPresentation.SalesRouteDeps{
			InventoryUC: invUC,
			JournalUC:   financeDeps.JournalUC,
			CoaUC:       financeDeps.CoaUC,
			Engine:      financeDeps.Engine,
			SettingsUC:  financeDeps.SettingsUC,
		})

		// HRD module (Sprint 13 - Attendance)
		hrdDeps := hrdPresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService, &hrdPresentation.HRDFinanceDeps{
			JournalUC:  financeDeps.JournalUC,
			CoaUC:      financeDeps.CoaUC,
			SettingsUC: financeDeps.SettingsUC,
		})

		// Start auto-absent worker (runs daily at startup + every 24 hours)
		// Create a CompanyTimezoneProvider so the worker can iterate companies per-timezone.
		workerTzProvider := orgRepos.NewCompanyTimezoneProvider(database.DB)
		autoAbsentWorker = hrdWorker.NewAutoAbsentWorker(hrdDeps.AttendanceUC, 24*time.Hour, workerTzProvider)
		autoAbsentWorker.Start()
		// Inventory module (Sprint 9)
		inventoryPresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService, invUC)

		// Stock Opname module (Sprint 9) — depends on inventory for stock movement on Post
		stockOpnamePresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService, invUC, financeDeps.JournalUC, financeDeps.CoaUC, financeDeps.SettingsUC)

		// CRM module (Sprint 17 - Foundation & Settings)
		// CRM module
		crmPresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService)

		// Feedback module (POS QR code feedback)
		feedbackPresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService)

		// Loyalty module
		customerRepo := customerRepos.NewCustomerRepository(database.DB)
		loyaltyDeps := loyaltyPresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService, customerRepo)

		// Reports module (Sales Overview)
		reportPresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService)

		// Travel Planner module
		travelPlannerPresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService)

		// General module (Dashboard)
		generalPresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService)

		// POS module (Floor Layout Designer)
		posPresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService, loyaltyDeps.Usecase, salesDeps.CustomerInvoiceUC, salesDeps.SalesPaymentUC)

		// Purchase module (Sprint 8 - Purchase Requisitions)
		purchaseDeps := purchasePresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService, invUC, financeDeps.JournalUC, financeDeps.CoaUC, financeDeps.AssetUC, financeDeps.Engine, financeDeps.SettingsUC, financeDeps.CashBankTransactionUC)

		// AI Assistant module
		currencyRepo := coreRepos.NewCurrencyRepository(database.DB)
		bankAccountRepo := coreRepos.NewBankAccountRepository(database.DB)
		bankAccountUC := coreUsecase.NewBankAccountUsecaseWithCurrency(database.DB, bankAccountRepo, currencyRepo)

		cerebrasClient := cerebras.NewClient(
			config.AppConfig.Cerebras.BaseURL,
			config.AppConfig.Cerebras.APIKey,
			config.AppConfig.Cerebras.Model,
		)
		aiPresentation.RegisterRoutes(r, v1, database.DB, jwtManager, permissionService, cerebrasClient, &aiPresentation.AIDeps{
			HolidayUC:         hrdDeps.HolidayUC,
			LeaveRequestUC:    hrdDeps.LeaveRequestUC,
			AttendanceUC:      hrdDeps.AttendanceUC,
			SalesQuotationUC:  salesDeps.QuotationUC,
			SalesOrderUC:      salesDeps.OrderUC,
			DeliveryOrderUC:   salesDeps.DeliveryOrderUC,
			CustomerInvoiceUC: salesDeps.CustomerInvoiceUC,

			InventoryUC:       invUC,
			PurchaseOrderUC:   purchaseDeps.OrderUC,
			PurchaseReqUC:     purchaseDeps.RequisitionUC,
			GoodsReceiptUC:    purchaseDeps.GoodsReceiptUC,
			SupplierInvoiceUC: purchaseDeps.SupplierInvoiceUC,
			CoaUC:             financeDeps.CoaUC,
			JournalUC:         financeDeps.JournalUC,
			FinancePaymentUC:  financeDeps.PaymentUC,
			BudgetUC:          financeDeps.BudgetUC,
			CashBankUC:        financeDeps.CashBankUC,
			TaxInvoiceUC:      financeDeps.TaxInvoiceUC,
			AssetUC:           financeDeps.AssetUC,
			SalaryUC:          hrdDeps.SalaryUC,
			BankAccountUC:     bankAccountUC,
		})
	}

	// Run server with explicit timeouts and graceful shutdown
	port := config.AppConfig.Server.Port
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: time.Duration(config.AppConfig.Server.ReadHeaderTimeoutSec) * time.Second,
		ReadTimeout:       time.Duration(config.AppConfig.Server.ReadTimeoutSec) * time.Second,
		WriteTimeout:      time.Duration(config.AppConfig.Server.WriteTimeoutSec) * time.Second,
		IdleTimeout:       time.Duration(config.AppConfig.Server.IdleTimeoutSec) * time.Second,
		MaxHeaderBytes:    config.AppConfig.Server.MaxHeaderBytes,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for shutdown signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	select {
	case sig := <-quit:
		log.Printf("Shutdown signal received: %v", sig)
	case err := <-serverErr:
		log.Printf("Server error: %v", err)
	}

	// Stop background worker before shutting down server/resources
	rtWorker.Stop()
	delWorker.Stop()
	if autoAbsentWorker != nil {
		autoAbsentWorker.Stop()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.AppConfig.Server.ShutdownTimeoutSec)*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Graceful shutdown failed: %v", err)
		_ = srv.Close()
	}
	log.Println("Server stopped")
}
