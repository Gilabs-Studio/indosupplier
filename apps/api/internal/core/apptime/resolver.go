package apptime

import (
	"log"
	"sync"
	"time"
)

// CompanyTimezoneProvider resolves the IANA timezone string for a given
// company or employee. Implementations live outside core/ (e.g. in
// organization/data/repositories) to avoid circular imports.
//
// WHY: core/apptime cannot import organization models, so we define
// an interface here and let the application wire the concrete
// implementation at startup via RegisterProvider.
type CompanyTimezoneProvider interface {
	// GetCompanyTimezone returns the IANA timezone for the given company ID.
	// Returns "" if not found (caller will fall back to instance default).
	GetCompanyTimezone(companyID string) string

	// GetEmployeeTimezone returns the IANA timezone for the given employee ID
	// by resolving employee → company → timezone.
	// Returns "" if not found (caller will fall back to instance default).
	GetEmployeeTimezone(employeeID string) string
}

var (
	provider   CompanyTimezoneProvider
	providerMu sync.RWMutex
)

// RegisterProvider sets the global CompanyTimezoneProvider.
// Must be called once after database connection is established.
func RegisterProvider(p CompanyTimezoneProvider) {
	providerMu.Lock()
	defer providerMu.Unlock()
	provider = p
	log.Println("apptime: company timezone provider registered")
}

// getProvider returns the registered provider (may be nil).
func getProvider() CompanyTimezoneProvider {
	providerMu.RLock()
	defer providerMu.RUnlock()
	return provider
}

// ---------------------------------------------------------------------------
// Per-company / per-employee convenience functions
// ---------------------------------------------------------------------------

// LocationForCompany returns the *time.Location for the given company ID.
// Falls back to the global default if no provider is registered or the
// company has no timezone configured.
func LocationForCompany(companyID string) *time.Location {
	p := getProvider()
	if p == nil || companyID == "" {
		return Location()
	}
	tz := p.GetCompanyTimezone(companyID)
	return ResolveLocation(tz)
}

// LocationForEmployee returns the *time.Location for the given employee ID.
// Resolves employee → company → timezone. Falls back to the global default
// if no provider is registered or the employee/company has no timezone.
func LocationForEmployee(employeeID string) *time.Location {
	p := getProvider()
	if p == nil || employeeID == "" {
		return Location()
	}
	tz := p.GetEmployeeTimezone(employeeID)
	return ResolveLocation(tz)
}

// NowForCompany returns the current time in the company's timezone.
func NowForCompany(companyID string) time.Time {
	return time.Now().In(LocationForCompany(companyID))
}

// NowForEmployee returns the current time in the employee's timezone.
func NowForEmployee(employeeID string) time.Time {
	return time.Now().In(LocationForEmployee(employeeID))
}

// TodayForCompany returns the start of today (00:00:00) in the company's timezone.
func TodayForCompany(companyID string) time.Time {
	l := LocationForCompany(companyID)
	n := time.Now().In(l)
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, l)
}

// TodayForEmployee returns the start of today (00:00:00) in the employee's timezone.
func TodayForEmployee(employeeID string) time.Time {
	l := LocationForEmployee(employeeID)
	n := time.Now().In(l)
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, l)
}

// StartOfMonthForEmployee returns the first day of a given month/year at 00:00:00
// in the employee's timezone.
func StartOfMonthForEmployee(employeeID string, year int, month time.Month) time.Time {
	return time.Date(year, month, 1, 0, 0, 0, 0, LocationForEmployee(employeeID))
}
