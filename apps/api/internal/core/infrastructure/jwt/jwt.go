package jwt

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	accessSecretKey  string
	refreshSecretKey string
	accessKeys       map[string]string
	refreshKeys      map[string]string
	accessKID        string
	refreshKID       string
	issuer           string
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
}

type Options struct {
	AccessSecretKey  string
	RefreshSecretKey string
	AccessKeys       map[string]string // kid -> secret
	RefreshKeys      map[string]string // kid -> secret
	AccessKID        string            // kid used for signing access tokens (optional)
	RefreshKID       string            // kid used for signing refresh tokens (optional)
	Issuer           string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
}

// ParseKeyRing parses a comma-separated key ring with format: "kid1:secret1,kid2:secret2".
// Whitespace around entries is ignored.
func ParseKeyRing(raw string) map[string]string {
	out := make(map[string]string)
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return out
	}
	entries := strings.Split(raw, ",")
	for _, e := range entries {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		parts := strings.SplitN(e, ":", 2)
		if len(parts) != 2 {
			continue
		}
		kid := strings.TrimSpace(parts[0])
		secret := strings.TrimSpace(parts[1])
		if kid == "" || secret == "" {
			continue
		}
		out[kid] = secret
	}
	return out
}

func mergeKeyRings(primary map[string]string, secondary map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range secondary {
		out[k] = v
	}
	for k, v := range primary {
		out[k] = v
	}
	return out
}

// AccessTokenTTL returns the access token TTL
func (m *JWTManager) AccessTokenTTL() time.Duration {
	return m.accessTokenTTL
}

// RefreshTokenTTL returns the refresh token TTL
func (m *JWTManager) RefreshTokenTTL() time.Duration {
	return m.refreshTokenTTL
}

func NewJWTManager(opts Options) *JWTManager {
	return &JWTManager{
		accessSecretKey:  opts.AccessSecretKey,
		refreshSecretKey: opts.RefreshSecretKey,
		accessKeys:       mergeKeyRings(opts.AccessKeys, nil),
		refreshKeys:      mergeKeyRings(opts.RefreshKeys, nil),
		accessKID:        opts.AccessKID,
		refreshKID:       opts.RefreshKID,
		issuer:           opts.Issuer,
		accessTokenTTL:   opts.AccessTokenTTL,
		refreshTokenTTL:  opts.RefreshTokenTTL,
	}
}

// GenerateAccessToken generates a new access token.
func (m *JWTManager) GenerateAccessToken(userID, email string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(apptime.Now().Add(m.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(apptime.Now()),
			NotBefore: jwt.NewNumericDate(apptime.Now()),
			ID:        uuid.New().String(),
			Issuer:    m.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	if m.accessKID != "" {
		token.Header["kid"] = m.accessKID
	}
	return token.SignedString([]byte(m.accessSecretKey))
}

// GenerateRefreshToken generates a new refresh token
func (m *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(apptime.Now().Add(m.refreshTokenTTL)),
		IssuedAt:  jwt.NewNumericDate(apptime.Now()),
		NotBefore: jwt.NewNumericDate(apptime.Now()),
		ID:        uuid.New().String(),
		Subject:   userID,
		Issuer:    m.issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	if m.refreshKID != "" {
		token.Header["kid"] = m.refreshKID
	}
	return token.SignedString([]byte(m.refreshSecretKey))
}

func (m *JWTManager) accessSecretsToTry(kid string) []string {
	var secrets []string
	if kid != "" {
		if s, ok := m.accessKeys[kid]; ok {
			secrets = append(secrets, s)
		}
	}
	if m.accessSecretKey != "" {
		secrets = append(secrets, m.accessSecretKey)
	}
	for _, s := range m.accessKeys {
		secrets = append(secrets, s)
	}
	return uniqueNonEmpty(secrets)
}

func (m *JWTManager) refreshSecretsToTry(kid string) []string {
	var secrets []string
	if kid != "" {
		if s, ok := m.refreshKeys[kid]; ok {
			secrets = append(secrets, s)
		}
	}
	if m.refreshSecretKey != "" {
		secrets = append(secrets, m.refreshSecretKey)
	}
	for _, s := range m.refreshKeys {
		secrets = append(secrets, s)
	}
	return uniqueNonEmpty(secrets)
}

func uniqueNonEmpty(in []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func tokenKIDFromString(tokenString string) string {
	parts := strings.Split(tokenString, ".")
	if len(parts) < 2 {
		return ""
	}
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return ""
	}
	var header map[string]interface{}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return ""
	}
	if v, ok := header["kid"]; ok {
		if s, ok2 := v.(string); ok2 {
			return s
		}
	}
	return ""
}

// ValidateToken validates and parses a token
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	// We may need to try multiple keys for rotation.
	// Strategy:
	// 1) If token has kid and we have a matching key, try it first.
	// 2) Fallback: try configured access secret + all ring keys.
	kid := tokenKIDFromString(tokenString)
	var lastErr error

	for _, secret := range m.accessSecretsToTry(kid) {
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return []byte(secret), nil
		})
		if err != nil {
			lastErr = err
			continue
		}
		parsedClaims, ok := token.Claims.(*Claims)
		if !ok || !token.Valid {
			lastErr = ErrInvalidToken
			continue
		}

		// Validate issuer (prevents token reuse across apps/environments)
		if m.issuer != "" && parsedClaims.Issuer != m.issuer {
			return nil, ErrInvalidToken
		}

		// Validate critical claims are not empty
		if parsedClaims.UserID == "" || parsedClaims.Email == "" {
			return nil, ErrInvalidToken
		}

		// Validate timestamps are valid
		if parsedClaims.NotBefore != nil && parsedClaims.NotBefore.Time.IsZero() {
			return nil, ErrInvalidToken
		}
		if parsedClaims.ExpiresAt != nil && parsedClaims.ExpiresAt.Time.IsZero() {
			return nil, ErrInvalidToken
		}

		return parsedClaims, nil
	}

	if lastErr != nil && errors.Is(lastErr, jwt.ErrTokenExpired) {
		return nil, ErrExpiredToken
	}
	return nil, ErrInvalidToken
}

// ValidateRefreshToken validates a refresh token and returns user ID
func (m *JWTManager) ValidateRefreshToken(tokenString string) (string, error) {
	kid := tokenKIDFromString(tokenString)

	var lastErr error
	for _, secret := range m.refreshSecretsToTry(kid) {
		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return []byte(secret), nil
		})
		if err != nil {
			lastErr = err
			continue
		}
		parsedClaims, ok := token.Claims.(*jwt.RegisteredClaims)
		if !ok || !token.Valid {
			lastErr = ErrInvalidToken
			continue
		}
		if m.issuer != "" && parsedClaims.Issuer != m.issuer {
			return "", ErrInvalidToken
		}
		return parsedClaims.Subject, nil
	}

	if lastErr != nil && errors.Is(lastErr, jwt.ErrTokenExpired) {
		return "", ErrExpiredToken
	}
	return "", ErrInvalidToken
}

// ExtractRefreshTokenID extracts the token ID (jti) from a refresh token
func (m *JWTManager) ExtractRefreshTokenID(tokenString string) (string, error) {
	kid := tokenKIDFromString(tokenString)

	var lastErr error
	for _, secret := range m.refreshSecretsToTry(kid) {
		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return []byte(secret), nil
		})
		if err != nil {
			lastErr = err
			continue
		}
		parsedClaims, ok := token.Claims.(*jwt.RegisteredClaims)
		if !ok || !token.Valid {
			lastErr = ErrInvalidToken
			continue
		}
		if m.issuer != "" && parsedClaims.Issuer != m.issuer {
			return "", ErrInvalidToken
		}
		if parsedClaims.ID == "" {
			return "", ErrInvalidToken
		}
		return parsedClaims.ID, nil
	}

	if lastErr != nil && errors.Is(lastErr, jwt.ErrTokenExpired) {
		return "", ErrExpiredToken
	}
	return "", ErrInvalidToken
}

// ValidateRefreshTokenWithID validates a refresh token and returns both user ID and token ID
func (m *JWTManager) ValidateRefreshTokenWithID(tokenString string) (userID string, tokenID string, err error) {
	kid := tokenKIDFromString(tokenString)

	var lastErr error
	for _, secret := range m.refreshSecretsToTry(kid) {
		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return []byte(secret), nil
		})
		if err != nil {
			lastErr = err
			continue
		}
		parsedClaims, ok := token.Claims.(*jwt.RegisteredClaims)
		if !ok || !token.Valid {
			lastErr = ErrInvalidToken
			continue
		}
		if m.issuer != "" && parsedClaims.Issuer != m.issuer {
			return "", "", ErrInvalidToken
		}
		if parsedClaims.ID == "" {
			return "", "", ErrInvalidToken
		}
		return parsedClaims.Subject, parsedClaims.ID, nil
	}

	if lastErr != nil && errors.Is(lastErr, jwt.ErrTokenExpired) {
		return "", "", ErrExpiredToken
	}
	return "", "", ErrInvalidToken
}
