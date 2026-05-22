// Package apptime provides a centralized, timezone-aware time source
// for the entire application.
//
// WHY: Docker containers default to UTC. Business logic (attendance,
// invoicing, code generation) must use the configured business timezone
// (default: Asia/Jakarta / WIB UTC+7) to correctly determine the current
// day, working hours, and calendar dates.
//
// Per-company timezone support:
// Companies may operate in different timezones (e.g. WIB vs WITA).
// Use NowForEmployee / NowForCompany to get the correct local time.
// The global Now() / Location() remain as the system-wide default fallback.
//
// USAGE:
//
//	apptime.Init("Asia/Jakarta")            // called once at startup
//	now := apptime.Now()                     // global default "now"
//	loc := apptime.Location()                // global *time.Location
//	now := apptime.NowForEmployee(empID)     // per-employee timezone "now"
//	loc := apptime.LocationForEmployee(empID)// per-employee *time.Location
package apptime

import (
	"log"
	"sync"
	"time"
)

var (
	loc  *time.Location
	once sync.Once

	// locationCache caches parsed *time.Location by IANA name to avoid
	// repeated syscalls to time.LoadLocation. Protected by locationCacheMu.
	locationCache   = make(map[string]*time.Location)
	locationCacheMu sync.RWMutex
)

// Init sets the application timezone from an IANA timezone name
// (e.g. "Asia/Jakarta", "Asia/Tokyo", "America/New_York").
// Falls back to a fixed UTC+7 zone if the name cannot be loaded
// (e.g. missing tzdata on the host).
// Init is safe to call multiple times; only the first call takes effect.
func Init(timezone string) {
	once.Do(func() {
		if timezone == "" {
			timezone = "Asia/Jakarta"
		}
		var err error
		loc, err = time.LoadLocation(timezone)
		if err != nil {
			log.Printf("apptime: failed to load timezone %q: %v — falling back to fixed UTC+7 (WIB)", timezone, err)
			loc = time.FixedZone("WIB", 7*60*60)
		}
		log.Printf("apptime: application timezone set to %s", loc)

		// Seed cache with the default location
		locationCacheMu.Lock()
		locationCache[timezone] = loc
		locationCacheMu.Unlock()
	})
}

// ensureInit guarantees the package has been initialized.
// If Init was never called, it defaults to Asia/Jakarta.
func ensureInit() {
	if loc == nil {
		Init("Asia/Jakarta")
	}
}

// ResolveLocation returns a *time.Location for the given IANA timezone name.
// Results are cached in-memory for fast repeated lookups.
// Returns the global default location if the name is empty or cannot be loaded.
func ResolveLocation(timezone string) *time.Location {
	ensureInit()
	if timezone == "" {
		return loc
	}

	// Fast path: read from cache
	locationCacheMu.RLock()
	cached, ok := locationCache[timezone]
	locationCacheMu.RUnlock()
	if ok {
		return cached
	}

	// Slow path: load and cache
	loaded, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("apptime: failed to resolve timezone %q: %v — using default %s", timezone, err, loc)
		return loc
	}

	locationCacheMu.Lock()
	locationCache[timezone] = loaded
	locationCacheMu.Unlock()
	return loaded
}

// ---------------------------------------------------------------------------
// Global (system-wide default) helpers — unchanged API
// ---------------------------------------------------------------------------

// Now returns the current time in the application timezone.
func Now() time.Time {
	ensureInit()
	return time.Now().In(loc)
}

// Location returns the configured application *time.Location.
// Use this wherever you need to construct a time.Date(..., loc).
func Location() *time.Location {
	ensureInit()
	return loc
}

// Today returns the start of today (00:00:00) in the application timezone.
func Today() time.Time {
	n := Now()
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, loc)
}

// StartOfMonth returns the first day of a given month/year at 00:00:00
// in the application timezone.
func StartOfMonth(year int, month time.Month) time.Time {
	ensureInit()
	return time.Date(year, month, 1, 0, 0, 0, 0, loc)
}

// ---------------------------------------------------------------------------
// Per-timezone helpers — for explicit timezone-aware operations
// ---------------------------------------------------------------------------

// NowIn returns the current time in the given timezone.
// Falls back to the global default if timezone is empty or invalid.
func NowIn(timezone string) time.Time {
	return time.Now().In(ResolveLocation(timezone))
}

// TodayIn returns the start of today (00:00:00) in the given timezone.
func TodayIn(timezone string) time.Time {
	l := ResolveLocation(timezone)
	n := time.Now().In(l)
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, l)
}

// StartOfMonthIn returns the first day of a given month/year at 00:00:00
// in the given timezone.
func StartOfMonthIn(year int, month time.Month, timezone string) time.Time {
	return time.Date(year, month, 1, 0, 0, 0, 0, ResolveLocation(timezone))
}
