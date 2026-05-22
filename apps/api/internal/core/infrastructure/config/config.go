package config

import (
	"fmt"
	"os"
	"strings"

	jwtinfra "github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/joho/godotenv"
)

type Config struct {
	Server        ServerConfig
	Subscription  SubscriptionLifecycleConfig
	Startup       StartupConfig
	Security      SecurityConfig
	Database      DatabaseConfig
	Observability ObservabilityConfig
	JWT           JWTConfig
	Cerebras      CerebrasConfig
	Xendit        XenditConfig
	Storage       StorageConfig
	RateLimit     RateLimitConfig
	HSTS          HSTSConfig
	Redis         RedisConfig
}

type StartupConfig struct {
	RunMigrations bool
	RunSeeders    bool
}

type SecurityConfig struct {
	CSRFEnabled                   bool
	ProxyHeadersEnabled           bool
	TrustedProxies                []string
	XenditCredentialEncryptionKey string
}

type RedisConfig struct {
	Host            string
	Port            string
	Password        string
	DB              int
	DialTimeoutSec  int
	ReadTimeoutSec  int
	WriteTimeoutSec int
	PoolSize        int
	MinIdleConns    int
}

type ServerConfig struct {
	Port                    string
	Env                     string
	Timezone                string // IANA timezone name (e.g. "Asia/Jakarta")
	// FrontendBaseURL is the public web app origin used to build QR code / feedback URLs.
	FrontendBaseURL         string
	// RootDomain is the root domain for setting cookies (e.g., ".salesview.id" or ".gilabs.id")
	// Used in production to share cookies across subdomains. Must include leading dot.
	RootDomain              string
	ReadHeaderTimeoutSec    int
	ReadTimeoutSec          int
	WriteTimeoutSec         int
	IdleTimeoutSec          int
	ShutdownTimeoutSec      int
	MaxHeaderBytes          int
	MaxBodyBytes            int64
	MaxMultipartBodyBytes   int64
	MaxMultipartMemoryBytes int64
}

type SubscriptionLifecycleConfig struct {
	GracePeriodDays int
}

type DatabaseConfig struct {
	Host                   string
	Port                   string
	User                   string
	Password               string
	DBName                 string
	SSLMode                string
	MaxOpenConns           int
	MaxIdleConns           int
	ConnMaxLifetimeMinutes int
	ConnMaxIdleTimeMinutes int
	PrepareStmt            bool
	SkipDefaultTransaction bool
}

type ObservabilityConfig struct {
	PprofEnabled   bool
	PprofToken     string
	MetricsEnabled bool
	MetricsToken   string
}

type JWTConfig struct {
	SecretKey        string
	AccessSecretKey  string
	RefreshSecretKey string
	AccessKeysRaw    string
	RefreshKeysRaw   string
	AccessKID        string
	RefreshKID       string
	Issuer           string
	AccessTokenTTL   int // in hours
	RefreshTokenTTL  int // in days
}

type CerebrasConfig struct {
	BaseURL string
	APIKey  string
	Model   string // Default model name
}

// XenditConfig holds Xendit payment gateway credentials.
type XenditConfig struct {
	SecretKey    string // Server-side Xendit secret key (sk_live_... / sk_test_...)
	WebhookToken string // Xendit callback verification token
	BaseURL      string // Xendit API base URL (overridable for sandbox)
}

type StorageConfig struct {
	MaxUploadSize     int64  // Maximum upload size in bytes (default: 10MB)
	R2AccountID       string // Cloudflare account ID
	R2AccessKeyID     string // R2 API token access key ID
	R2SecretAccessKey string // R2 API token secret access key
	R2BucketName      string // R2 bucket name
	R2PublicURL       string // Public URL prefix for serving R2 objects
}

// RateLimitRule defines rate limit configuration for a specific endpoint type
type RateLimitRule struct {
	Requests int // Number of requests allowed
	Window   int // Time window in seconds
}

// RateLimitConfig defines rate limit configuration for different endpoint types
type RateLimitConfig struct {
	Login                  RateLimitRule // Login endpoint: 5 requests per 15 minutes (Level 1 - IP)
	Refresh                RateLimitRule // Refresh token endpoint: 10 requests per hour
	Upload                 RateLimitRule // File upload endpoint: 20 requests per hour
	General                RateLimitRule // General API endpoints: 100 requests per minute
	Public                 RateLimitRule // Public endpoints: 200 requests per minute
	FailClosedOnRedisError bool          // If true, reject requests when Redis limiter is unavailable
	// Multi-level rate limiting for login
	LoginByEmail RateLimitRule // Level 2: 10 attempts per 15 minutes per email
	LoginGlobal  RateLimitRule // Level 3: 100 attempts per minute globally
}

// HSTSConfig defines HTTP Strict Transport Security configuration
type HSTSConfig struct {
	MaxAge            int  // Max age in seconds (default: 31536000 = 1 year)
	IncludeSubDomains bool // Include subdomains in HSTS policy
	Preload           bool // Enable HSTS preload
}

var AppConfig *Config

func Load() error {
	// Load .env file if exists (for local development only)
	// Skip .env loading in production to use Docker environment variables.
	// IMPORTANT: Read env value BEFORE and AFTER loading .env, because ENV/APP_ENV may
	// only exist inside .env (not in OS env). On first read it defaults to "development",
	// so we always attempt to load .env, then re-read to get the correct value.
	envValue := getEnv("APP_ENV", getEnv("ENV", "development")) // Support both APP_ENV and ENV
	if envValue != "production" {
		_ = godotenv.Load()
		// Re-read after loading .env so ENV=production inside the file is honoured
		envValue = getEnv("APP_ENV", getEnv("ENV", envValue))
	}

	AppConfig = &Config{
		Server: ServerConfig{
			Port:                    getEnv("PORT", "8080"),
			Env:                     envValue, // Use the env value we determined above
			Timezone:                getEnv("APP_TIMEZONE", "Asia/Jakarta"),
			FrontendBaseURL:         getEnv("FRONTEND_BASE_URL", "http://localhost:3000"),
			RootDomain:              getEnv("ROOT_DOMAIN", ""), // e.g., ".salesview.id" or ".gilabs.id" for cookie sharing
			ReadHeaderTimeoutSec:    getEnvAsInt("SERVER_READ_HEADER_TIMEOUT_SEC", 10),
			ReadTimeoutSec:          getEnvAsInt("SERVER_READ_TIMEOUT_SEC", 30),
			WriteTimeoutSec:         getEnvAsInt("SERVER_WRITE_TIMEOUT_SEC", 30),
			IdleTimeoutSec:          getEnvAsInt("SERVER_IDLE_TIMEOUT_SEC", 120),
			ShutdownTimeoutSec:      getEnvAsInt("SERVER_SHUTDOWN_TIMEOUT_SEC", 20),
			MaxHeaderBytes:          getEnvAsInt("SERVER_MAX_HEADER_BYTES", 1<<20),
			MaxBodyBytes:            getEnvAsInt64("SERVER_MAX_BODY_BYTES", 1<<20),             // 1MB
			MaxMultipartBodyBytes:   getEnvAsInt64("SERVER_MAX_MULTIPART_BODY_BYTES", 50<<20),  // 50MB
			MaxMultipartMemoryBytes: getEnvAsInt64("SERVER_MAX_MULTIPART_MEMORY_BYTES", 8<<20), // 8MB
		},
		Subscription: SubscriptionLifecycleConfig{
			GracePeriodDays: getEnvAsInt("SUBSCRIPTION_GRACE_PERIOD_DAYS", 7),
		},
		Startup: StartupConfig{
			RunMigrations: getEnvAsBool("RUN_MIGRATIONS", envValue != "production"),
			RunSeeders:    getEnvAsBool("RUN_SEEDERS", envValue != "production"),
		},
		Security: SecurityConfig{
			CSRFEnabled:                   getEnvAsBool("CSRF_ENABLED", true),
			ProxyHeadersEnabled:           getEnvAsBool("PROXY_HEADERS_ENABLED", false),
			TrustedProxies:                getEnvAsStringSlice("TRUSTED_PROXIES", nil),
			XenditCredentialEncryptionKey: getEnv("XENDIT_CREDENTIAL_ENCRYPTION_KEY", ""),
		},
		Database: DatabaseConfig{
			Host:                   getEnv("DB_HOST", "localhost"),
			Port:                   getEnv("DB_PORT", "5432"),
			User:                   getEnv("DB_USER", "postgres"),
			Password:               getEnv("DB_PASSWORD", "postgres"),
			DBName:                 getEnv("DB_NAME", "gims_erp"),
			SSLMode:                getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:           getEnvAsInt("DB_MAX_OPEN_CONNS", 50),
			MaxIdleConns:           getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetimeMinutes: getEnvAsInt("DB_CONN_MAX_LIFETIME_MINUTES", 30),
			ConnMaxIdleTimeMinutes: getEnvAsInt("DB_CONN_MAX_IDLE_TIME_MINUTES", 10),
			PrepareStmt:            getEnvAsBool("DB_PREPARE_STMT", false),
			SkipDefaultTransaction: getEnvAsBool("DB_SKIP_DEFAULT_TRANSACTION", false),
		},
		Observability: ObservabilityConfig{
			PprofEnabled:   getEnvAsBool("PPROF_ENABLED", false),
			PprofToken:     getEnv("PPROF_TOKEN", ""),
			MetricsEnabled: getEnvAsBool("METRICS_ENABLED", false),
			MetricsToken:   getEnv("METRICS_TOKEN", ""),
		},
		JWT: JWTConfig{
			SecretKey:        getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			AccessSecretKey:  getEnv("JWT_ACCESS_SECRET", ""),
			RefreshSecretKey: getEnv("JWT_REFRESH_SECRET", ""),
			AccessKeysRaw:    getEnv("JWT_ACCESS_KEYS", ""),
			RefreshKeysRaw:   getEnv("JWT_REFRESH_KEYS", ""),
			AccessKID:        getEnv("JWT_ACCESS_KID", ""),
			RefreshKID:       getEnv("JWT_REFRESH_KID", ""),
			Issuer:           getEnv("JWT_ISSUER", "template-1n"),
			AccessTokenTTL:   getEnvAsInt("JWT_ACCESS_TTL", 24), // 24 hours
			RefreshTokenTTL:  getEnvAsInt("JWT_REFRESH_TTL", 7), // 7 days
		},
		Cerebras: CerebrasConfig{
			BaseURL: getEnv("CEREBRAS_BASE_URL", "https://api.cerebras.ai"),
			APIKey:  getEnv("CEREBRAS_API_KEY", ""),
			Model:   getEnv("CEREBRAS_MODEL", "llama-3.1-8b"), // Default model
		},
		Xendit: XenditConfig{
			SecretKey:    getEnv("XENDIT_SECRET_KEY", ""),
			WebhookToken: getEnv("XENDIT_WEBHOOK_TOKEN", ""),
			BaseURL:      getEnv("XENDIT_BASE_URL", "https://api.xendit.co"),
		},
		Storage: StorageConfig{
			MaxUploadSize:     getEnvAsInt64("STORAGE_MAX_UPLOAD_SIZE", 10*1024*1024), // 10MB default
			R2AccountID:       getEnv("R2_ACCOUNT_ID", ""),
			R2AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
			R2SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
			R2BucketName:      getEnv("R2_BUCKET_NAME", ""),
			R2PublicURL:       getEnv("R2_PUBLIC_URL", ""),
		},
		RateLimit: RateLimitConfig{
			Login: RateLimitRule{
				Requests: getEnvAsInt("RATE_LIMIT_LOGIN_REQUESTS", 3), //3 requests per 15 minutes (Level 1 - IP)
				Window:   getEnvAsInt("RATE_LIMIT_LOGIN_WINDOW", 900), // 15 minutes (900 seconds)
			},
			Refresh: RateLimitRule{
				Requests: getEnvAsInt("RATE_LIMIT_REFRESH_REQUESTS", 10), // 10 requests
				Window:   getEnvAsInt("RATE_LIMIT_REFRESH_WINDOW", 3600), // 1 hour (3600 seconds)
			},
			Upload: RateLimitRule{
				Requests: getEnvAsInt("RATE_LIMIT_UPLOAD_REQUESTS", 20), // 20 requests
				Window:   getEnvAsInt("RATE_LIMIT_UPLOAD_WINDOW", 3600), // 1 hour (3600 seconds)
			},
			General: RateLimitRule{
				Requests: getEnvAsInt("RATE_LIMIT_GENERAL_REQUESTS", 100), // 100 requests
				Window:   getEnvAsInt("RATE_LIMIT_GENERAL_WINDOW", 60),    // 1 minute (60 seconds)
			},
			Public: RateLimitRule{
				Requests: getEnvAsInt("RATE_LIMIT_PUBLIC_REQUESTS", 200), // 200 requests
				Window:   getEnvAsInt("RATE_LIMIT_PUBLIC_WINDOW", 60),    // 1 minute (60 seconds)
			},
			FailClosedOnRedisError: getEnvAsBool("RATE_LIMIT_FAIL_CLOSED_ON_REDIS_ERROR", envValue == "production"),
			// Level 2: Rate limit by email/username (prevents brute force even if IP changes)
			LoginByEmail: RateLimitRule{
				Requests: getEnvAsInt("RATE_LIMIT_LOGIN_BY_EMAIL_REQUESTS", 5), // 5 requests per 15 minutes per email
				Window:   getEnvAsInt("RATE_LIMIT_LOGIN_BY_EMAIL_WINDOW", 900), // 15 minutes (900 seconds)
			},
			// Level 3: Global rate limit (prevents DOS on entire system)
			LoginGlobal: RateLimitRule{
				Requests: getEnvAsInt("RATE_LIMIT_LOGIN_GLOBAL_REQUESTS", 50), // 50 requests per minute globally
				Window:   getEnvAsInt("RATE_LIMIT_LOGIN_GLOBAL_WINDOW", 60),   // 1 minute (60 seconds)
			},
		},
		HSTS: HSTSConfig{
			MaxAge:            getEnvAsInt("HSTS_MAX_AGE", 31536000), // 1 year in seconds
			IncludeSubDomains: getEnv("HSTS_INCLUDE_SUBDOMAINS", "true") == "true",
			Preload:           getEnv("HSTS_PRELOAD", "true") == "true",
		},
		Redis: RedisConfig{
			Host:            getEnv("REDIS_HOST", "localhost"),
			Port:            getEnv("REDIS_PORT", "6379"),
			Password:        getEnv("REDIS_PASSWORD", ""),
			DB:              getEnvAsInt("REDIS_DB", 0),
			DialTimeoutSec:  getEnvAsInt("REDIS_DIAL_TIMEOUT_SEC", 10),
			ReadTimeoutSec:  getEnvAsInt("REDIS_READ_TIMEOUT_SEC", 30),
			WriteTimeoutSec: getEnvAsInt("REDIS_WRITE_TIMEOUT_SEC", 30),
			PoolSize:        getEnvAsInt("REDIS_POOL_SIZE", 10),
			MinIdleConns:    getEnvAsInt("REDIS_MIN_IDLE_CONNS", 5),
		},
	}

	if AppConfig.Subscription.GracePeriodDays < 1 {
		AppConfig.Subscription.GracePeriodDays = 7
	}

	// SECURITY: Enforce JWT_SECRET in production
	if AppConfig.Server.Env == "production" {
		// Backward compatible behavior:
		// - If JWT_ACCESS_SECRET/JWT_REFRESH_SECRET are set, use them.
		// - Else if JWT_*_KID is set and the corresponding JWT_*_KEYS contains it, use that.
		// - Otherwise fall back to JWT_SECRET.
		accessKeys := jwtinfra.ParseKeyRing(AppConfig.JWT.AccessKeysRaw)
		refreshKeys := jwtinfra.ParseKeyRing(AppConfig.JWT.RefreshKeysRaw)

		accessSecret := strings.TrimSpace(AppConfig.JWT.AccessSecretKey)
		refreshSecret := strings.TrimSpace(AppConfig.JWT.RefreshSecretKey)
		fallback := strings.TrimSpace(AppConfig.JWT.SecretKey)

		if strings.TrimSpace(AppConfig.JWT.AccessKID) != "" {
			kid := strings.TrimSpace(AppConfig.JWT.AccessKID)
			secret, ok := accessKeys[kid]
			if !ok {
				return fmt.Errorf("JWT_ACCESS_KID is set but not found in JWT_ACCESS_KEYS")
			}
			if accessSecret == "" {
				accessSecret = strings.TrimSpace(secret)
			}
		}
		if strings.TrimSpace(AppConfig.JWT.RefreshKID) != "" {
			kid := strings.TrimSpace(AppConfig.JWT.RefreshKID)
			secret, ok := refreshKeys[kid]
			if !ok {
				return fmt.Errorf("JWT_REFRESH_KID is set but not found in JWT_REFRESH_KEYS")
			}
			if refreshSecret == "" {
				refreshSecret = strings.TrimSpace(secret)
			}
		}

		if accessSecret == "" {
			accessSecret = fallback
		}
		if refreshSecret == "" {
			refreshSecret = fallback
		}

		if accessSecret == "" || accessSecret == "your-secret-key-change-in-production" {
			return fmt.Errorf("JWT_ACCESS_SECRET (or JWT_ACCESS_KEYS+JWT_ACCESS_KID, or JWT_SECRET fallback) must be set in production environment")
		}
		if refreshSecret == "" || refreshSecret == "your-secret-key-change-in-production" {
			return fmt.Errorf("JWT_REFRESH_SECRET (or JWT_REFRESH_KEYS+JWT_REFRESH_KID, or JWT_SECRET fallback) must be set in production environment")
		}

		// SECURITY: Enforce minimum secret length (256 bits = 32 bytes)
		if len(accessSecret) < 32 {
			return fmt.Errorf("JWT access secret must be at least 32 characters for production (current: %d)", len(accessSecret))
		}
		if len(refreshSecret) < 32 {
			return fmt.Errorf("JWT refresh secret must be at least 32 characters for production (current: %d)", len(refreshSecret))
		}
		if strings.TrimSpace(AppConfig.JWT.Issuer) == "" {
			return fmt.Errorf("JWT_ISSUER must be set in production environment")
		}
		if strings.TrimSpace(AppConfig.Security.XenditCredentialEncryptionKey) == "" {
			return fmt.Errorf("XENDIT_CREDENTIAL_ENCRYPTION_KEY must be set in production environment")
		}

		// SAFETY: Never expose debug/observability endpoints in production.
		if AppConfig.Observability.PprofEnabled || AppConfig.Observability.MetricsEnabled {
			return fmt.Errorf("PPROF_ENABLED and METRICS_ENABLED must be false in production")
		}

		// CRITICAL SAFETY: Never run seeders in production
		if AppConfig.Startup.RunSeeders {
			return fmt.Errorf("CRITICAL: RUN_SEEDERS must be false in production environment. Seeders are for development only and will corrupt production data")
		}

		// WARNING: Manual migration approval recommended for production
		if AppConfig.Startup.RunMigrations {
			fmt.Println("WARNING: Auto-migrations are enabled in production. Ensure proper testing and backup procedures are in place.")
		}

		// SECURITY: Suggest SSL for database connections in production
		if AppConfig.Database.SSLMode == "disable" {
			fmt.Println("WARNING: SECURITY: DB_SSLMODE is 'disable' in production. Ensure you are using a secure connection method such as Cloud SQL Auth Proxy or a private VPC network.")
		}
	}

	// If enabled in non-production, require a token so it's not accidentally exposed.
	if AppConfig.Observability.PprofEnabled {
		if strings.TrimSpace(AppConfig.Observability.PprofToken) == "" {
			return fmt.Errorf("PPROF_TOKEN must be set when PPROF_ENABLED=true")
		}
	}
	if AppConfig.Observability.MetricsEnabled {
		if strings.TrimSpace(AppConfig.Observability.MetricsToken) == "" {
			return fmt.Errorf("METRICS_TOKEN must be set when METRICS_ENABLED=true")
		}
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	var value int
	_, err := fmt.Sscanf(valueStr, "%d", &value)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	var value int64
	_, err := fmt.Sscanf(valueStr, "%d", &value)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	switch valueStr {
	case "true", "1", "yes", "y", "on", "TRUE", "Yes", "ON":
		return true
	case "false", "0", "no", "n", "off", "FALSE", "No", "OFF":
		return false
	default:
		return defaultValue
	}
}

func getEnvAsStringSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if strings.TrimSpace(valueStr) == "" {
		return defaultValue
	}
	parts := strings.Split(valueStr, ",")
	var out []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return defaultValue
	}
	return out
}

func GetDSN() string {
	db := AppConfig.Database
	// Ensure sslmode is set, default to disable if empty
	sslmode := db.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}
	// WHY: TimeZone=UTC ensures all timestamps are stored/retrieved in UTC,
	// application-level timezone conversion is handled by apptime package.
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		db.Host, db.Port, db.User, db.Password, db.DBName, sslmode,
	)
}
