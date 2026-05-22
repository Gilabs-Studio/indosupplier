package repositories

import (
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// companyTimezoneProvider implements apptime.CompanyTimezoneProvider using
// an in-memory cache backed by direct DB queries. The cache uses a short
// TTL because timezone changes are rare but must be picked up eventually.
//
// WHY: apptime (core/) cannot import organization models. This concrete
// implementation lives in organization/data/repositories and is wired
// at startup via apptime.RegisterProvider().
type companyTimezoneProvider struct {
	db *gorm.DB

	mu    sync.RWMutex
	cache map[string]cacheEntry // key = company_id or "emp:"+employee_id
	ttl   time.Duration
}

type cacheEntry struct {
	timezone  string
	expiresAt time.Time
}

// NewCompanyTimezoneProvider creates a provider backed by the given DB.
func NewCompanyTimezoneProvider(db *gorm.DB) *companyTimezoneProvider {
	return &companyTimezoneProvider{
		db:    db,
		cache: make(map[string]cacheEntry),
		ttl:   5 * time.Minute,
	}
}

// GetCompanyTimezone returns the IANA timezone for the given company ID.
// Returns "" if not found (caller falls back to instance default).
func (p *companyTimezoneProvider) GetCompanyTimezone(companyID string) string {
	if companyID == "" {
		return ""
	}

	// Check cache
	p.mu.RLock()
	if entry, ok := p.cache[companyID]; ok && time.Now().Before(entry.expiresAt) {
		p.mu.RUnlock()
		return entry.timezone
	}
	p.mu.RUnlock()

	// Query DB — lightweight single-column select
	var timezone string
	err := p.db.Table("companies").
		Select("timezone").
		Where("id = ? AND deleted_at IS NULL", companyID).
		Limit(1).
		Scan(&timezone).Error
	if err != nil {
		log.Printf("companyTimezoneProvider: failed to get timezone for company %s: %v", companyID, err)
		return ""
	}

	// Store in cache
	p.mu.Lock()
	p.cache[companyID] = cacheEntry{
		timezone:  timezone,
		expiresAt: time.Now().Add(p.ttl),
	}
	p.mu.Unlock()

	return timezone
}

// GetEmployeeTimezone returns the IANA timezone for the given employee ID
// by resolving employee → company_id → timezone.
// Returns "" if no company is assigned or not found.
func (p *companyTimezoneProvider) GetEmployeeTimezone(employeeID string) string {
	if employeeID == "" {
		return ""
	}

	cacheKey := "emp:" + employeeID

	// Check cache
	p.mu.RLock()
	if entry, ok := p.cache[cacheKey]; ok && time.Now().Before(entry.expiresAt) {
		p.mu.RUnlock()
		return entry.timezone
	}
	p.mu.RUnlock()

	// Query: join employees → companies to get timezone in one query
	var timezone string
	err := p.db.Table("employees e").
		Select("c.timezone").
		Joins("JOIN companies c ON c.id = e.company_id AND c.deleted_at IS NULL").
		Where("e.id = ? AND e.deleted_at IS NULL AND e.company_id IS NOT NULL", employeeID).
		Limit(1).
		Scan(&timezone).Error
	if err != nil {
		log.Printf("companyTimezoneProvider: failed to get timezone for employee %s: %v", employeeID, err)
		return ""
	}

	// Store in cache
	p.mu.Lock()
	p.cache[cacheKey] = cacheEntry{
		timezone:  timezone,
		expiresAt: time.Now().Add(p.ttl),
	}
	p.mu.Unlock()

	return timezone
}

// InvalidateCompany removes the cached timezone for a company so the
// next lookup fetches fresh data. Call this when a company's timezone changes.
func (p *companyTimezoneProvider) InvalidateCompany(companyID string) {
	p.mu.Lock()
	delete(p.cache, companyID)
	// Also invalidate any employee entries for this company — we don't track
	// the reverse mapping, so we clear all employee entries. This is fine
	// because timezone changes are extremely rare.
	for k := range p.cache {
		if len(k) > 4 && k[:4] == "emp:" {
			delete(p.cache, k)
		}
	}
	p.mu.Unlock()
}

// GetDistinctCompanyTimezones returns all distinct (companyID, timezone) pairs
// for active companies. Used by the auto-absent worker to iterate per-timezone.
func (p *companyTimezoneProvider) GetDistinctCompanyTimezones() (map[string]string, error) {
	type result struct {
		ID       string `gorm:"column:id"`
		Timezone string `gorm:"column:timezone"`
	}
	var results []result
	err := p.db.Table("companies").
		Select("id, timezone").
		Where("deleted_at IS NULL AND is_active = ?", true).
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	m := make(map[string]string, len(results))
	for _, r := range results {
		tz := r.Timezone
		if tz == "" {
			tz = "Asia/Jakarta"
		}
		m[r.ID] = tz
	}
	return m, nil
}
