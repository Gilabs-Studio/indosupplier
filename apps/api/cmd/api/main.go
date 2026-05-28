package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	authUsecase "github.com/gilabs/indosupplier/api/internal/auth/domain/usecase"
	authHandler "github.com/gilabs/indosupplier/api/internal/auth/presentation/handler"
	authRouter "github.com/gilabs/indosupplier/api/internal/auth/presentation/router"
	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/audit"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/config"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/events"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/redis"
	coreRouter "github.com/gilabs/indosupplier/api/internal/core/infrastructure/router"
	"github.com/gilabs/indosupplier/api/internal/core/logger"
	"github.com/gilabs/indosupplier/api/internal/core/middleware"
	"github.com/gilabs/indosupplier/api/internal/core/response"
	"github.com/gilabs/indosupplier/api/internal/core/storage"
	refreshTokenRepo "github.com/gilabs/indosupplier/api/internal/refresh_token/data/repositories"
	refreshTokenWorker "github.com/gilabs/indosupplier/api/internal/refresh_token/worker"
	"github.com/gilabs/indosupplier/api/internal/user/data/repositories"
	userUsecase "github.com/gilabs/indosupplier/api/internal/user/domain/usecase"
	userHandler "github.com/gilabs/indosupplier/api/internal/user/presentation/handler"
	userRouter "github.com/gilabs/indosupplier/api/internal/user/presentation/router"
	"github.com/gilabs/indosupplier/api/seeders"
)

func initInfrastructure() {
	logger.Init()

	if err := config.Load(); err != nil {
		log.Fatal("failed to load config:", err)
	}

	apptime.Init(config.AppConfig.Server.Timezone)

	if err := storage.Init(
		config.AppConfig.Storage.R2AccountID,
		config.AppConfig.Storage.R2AccessKeyID,
		config.AppConfig.Storage.R2SecretAccessKey,
		config.AppConfig.Storage.R2BucketName,
		config.AppConfig.Storage.R2PublicURL,
	); err != nil {
		log.Printf("warning: failed to initialize storage: %v", err)
	}

	if err := database.Connect(); err != nil {
		log.Fatal("failed to connect database:", err)
	}

	if err := redis.InitRedis(config.AppConfig); err != nil {
		log.Printf("warning: redis init failed: %v", err)
	}

	if config.AppConfig.Startup.RunMigrations {
		if err := database.AutoMigrate(); err != nil {
			log.Fatal("failed to run migrations:", err)
		}
	}

	if config.AppConfig.Startup.RunSeeders {
		if err := seeders.SeedAll(); err != nil {
			log.Fatal("failed to run seeders:", err)
		}
	}
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

	return jwt.NewJWTManager(jwt.Options{
		AccessSecretKey:  accessSecret,
		RefreshSecretKey: refreshSecret,
		Issuer:           config.AppConfig.JWT.Issuer,
		AccessTokenTTL:   time.Duration(config.AppConfig.JWT.AccessTokenTTL) * time.Hour,
		RefreshTokenTTL:  time.Duration(config.AppConfig.JWT.RefreshTokenTTL) * 24 * time.Hour,
	})
}

func main() {
	initInfrastructure()
	defer database.Close()
	defer redis.Close()

	jwtManager := setupJWT()

	refreshTokenRepository := refreshTokenRepo.NewRefreshTokenRepository(database.DB)
	userRepository := repositories.NewUserRepository(database.DB)

	auditService := audit.NewAuditService(database.DB)
	eventPublisher := events.NewNoOpEventPublisher(true)

	authUC := authUsecase.NewAuthUsecase(
		database.DB,
		userRepository,
		refreshTokenRepository,
		jwtManager,
		eventPublisher,
	)

	userUC := userUsecase.NewUserUsecase(
		userRepository,
		auditService,
		eventPublisher,
		redis.GetClient(),
	)

	authH := authHandler.NewAuthHandler(authUC)
	userH := userHandler.NewUserHandler(userUC)

	rtWorker := refreshTokenWorker.NewRefreshTokenCleanupWorker(refreshTokenRepository, 24*time.Hour)
	rtWorker.Start()

	r := coreRouter.NewEngine(jwtManager)

	r.Use(middleware.MetricsMiddleware())
	if config.AppConfig.Observability.MetricsEnabled {
		r.GET("/metrics", middleware.MetricsHandler())
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "IndoSupplier API is running"})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/", func(c *gin.Context) {
			response.SuccessResponse(c, gin.H{"message": "IndoSupplier API v1", "version": "1.0.0"}, nil)
		})

		authRouter.RegisterAuthRoutes(v1, authH, jwtManager)
		userRouter.RegisterUserRoutes(v1, userH, jwtManager)
		coreRouter.RegisterUploadRoutes(v1, jwtManager)
	}

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
		log.Printf("server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("shutdown signal received: %v", sig)
	case err := <-serverErr:
		log.Printf("server error: %v", err)
	}

	rtWorker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.AppConfig.Server.ShutdownTimeoutSec)*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		_ = srv.Close()
	}
}
