package jwt

import (
	"encoding/base64"
	"testing"
	"time"
)

func TestParseKeyRing(t *testing.T) {
	got := ParseKeyRing(" kid1:secret1, kid2:secret2 ")
	if got["kid1"] != "secret1" {
		t.Fatalf("expected kid1=secret1, got %q", got["kid1"])
	}
	if got["kid2"] != "secret2" {
		t.Fatalf("expected kid2=secret2, got %q", got["kid2"])
	}

	got = ParseKeyRing(" , bad, :nope, kid3:, : , kid4:secret4 ")
	if _, ok := got["kid3"]; ok {
		t.Fatalf("expected kid3 to be ignored")
	}
	if got["kid4"] != "secret4" {
		t.Fatalf("expected kid4=secret4, got %q", got["kid4"])
	}

	got = ParseKeyRing("")
	if len(got) != 0 {
		t.Fatalf("expected empty map, got len=%d", len(got))
	}
}

func TestTokenKIDFromString(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","kid":"k1"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{}`))
	tok := header + "." + payload + ".sig"

	if got := tokenKIDFromString(tok); got != "k1" {
		t.Fatalf("expected kid k1, got %q", got)
	}

	if got := tokenKIDFromString("not-a-jwt"); got != "" {
		t.Fatalf("expected empty kid for invalid token, got %q", got)
	}

	badHeader := "###" + "." + payload + ".sig"
	if got := tokenKIDFromString(badHeader); got != "" {
		t.Fatalf("expected empty kid for bad header encoding, got %q", got)
	}
}

func newTestManager() *JWTManager {
	return NewJWTManager(Options{
		AccessSecretKey:  "test-secret-key-for-unit-tests-only",
		RefreshSecretKey: "test-refresh-key-for-unit-tests-only",
		Issuer:           "test",
		AccessTokenTTL:   15 * time.Minute,
		RefreshTokenTTL:  7 * 24 * time.Hour,
	})
}

// TestGenerateAccessToken_TenantIDEmbedded verifies that the tenant_id claim
// is correctly embedded in the JWT and survives the validate round-trip.
func TestGenerateAccessToken_TenantIDEmbedded(t *testing.T) {
	mgr := newTestManager()
	wantTenantID := "tenant-abc-123"

	token, err := mgr.GenerateAccessToken("user-1", "user@tenant.com", "admin", wantTenantID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	claims, err := mgr.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.TenantID != wantTenantID {
		t.Errorf("TenantID: want %q, got %q", wantTenantID, claims.TenantID)
	}
	if claims.UserID != "user-1" {
		t.Errorf("UserID: want %q, got %q", "user-1", claims.UserID)
	}
}

// TestGenerateAccessToken_SystemAdminNoTenant verifies that system-admin tokens
// are generated with an empty tenant_id (platform-level access).
func TestGenerateAccessToken_SystemAdminNoTenant(t *testing.T) {
	mgr := newTestManager()

	token, err := mgr.GenerateAccessToken("admin-1", "admin@system.com", "system_admin", "")
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	claims, err := mgr.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.TenantID != "" {
		t.Errorf("system_admin TenantID should be empty, got %q", claims.TenantID)
	}
	if claims.Role != "system_admin" {
		t.Errorf("Role: want %q, got %q", "system_admin", claims.Role)
	}
}

// TestGenerateAccessToken_CrossTenantIsolation verifies that tokens for two
// different tenants carry distinct tenant_id values (no bleed-over).
func TestGenerateAccessToken_CrossTenantIsolation(t *testing.T) {
	mgr := newTestManager()

	token1, _ := mgr.GenerateAccessToken("u1", "u1@t1.com", "admin", "tenant-111")
	token2, _ := mgr.GenerateAccessToken("u2", "u2@t2.com", "admin", "tenant-222")

	claims1, err := mgr.ValidateToken(token1)
	if err != nil {
		t.Fatalf("ValidateToken tenant1 failed: %v", err)
	}
	claims2, err := mgr.ValidateToken(token2)
	if err != nil {
		t.Fatalf("ValidateToken tenant2 failed: %v", err)
	}

	if claims1.TenantID == claims2.TenantID {
		t.Errorf("different tenants must have different TenantID in JWT, both got %q", claims1.TenantID)
	}
	if claims1.TenantID != "tenant-111" {
		t.Errorf("tenant1: want %q, got %q", "tenant-111", claims1.TenantID)
	}
	if claims2.TenantID != "tenant-222" {
		t.Errorf("tenant2: want %q, got %q", "tenant-222", claims2.TenantID)
	}
}

