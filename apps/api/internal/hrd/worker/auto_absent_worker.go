package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/hrd/domain/usecase"
)

// CompanyTimezoneProvider is a subset interface matching the provider registered
// in apptime. We define it here to avoid importing the organization package.
type CompanyTimezoneProvider interface {
	GetDistinctCompanyTimezones() (map[string]string, error) // companyID → timezone
}

// AutoAbsentWorker runs daily to create ABSENT/LEAVE records for employees
// who didn't clock in on the previous working day.
// When a CompanyTimezoneProvider is set, it processes each company group in
// its own timezone so "yesterday" is correct for all timezones.
type AutoAbsentWorker struct {
	attendanceUC usecase.AttendanceRecordUsecase
	tzProvider   CompanyTimezoneProvider
	ticker       *time.Ticker
	stopChan     chan struct{}
	stopOnce     sync.Once
}

// NewAutoAbsentWorker creates a new AutoAbsentWorker.
// tzProvider may be nil; in that case, the worker falls back to the global
// apptime timezone for all employees.
func NewAutoAbsentWorker(
	attendanceUC usecase.AttendanceRecordUsecase,
	interval time.Duration,
	tzProvider CompanyTimezoneProvider,
) *AutoAbsentWorker {
	return &AutoAbsentWorker{
		attendanceUC: attendanceUC,
		tzProvider:   tzProvider,
		ticker:       time.NewTicker(interval),
		stopChan:     make(chan struct{}),
	}
}

// Start starts the auto-absent worker
func (w *AutoAbsentWorker) Start() {
	log.Println("Auto-absent worker started (runs daily)")

	// Run immediately on start for yesterday's date
	go w.processAutoAbsent()

	go func() {
		for {
			select {
			case <-w.ticker.C:
				w.processAutoAbsent()
			case <-w.stopChan:
				w.ticker.Stop()
				log.Println("Auto-absent worker stopped")
				return
			}
		}
	}()
}

// Stop stops the auto-absent worker
func (w *AutoAbsentWorker) Stop() {
	w.stopOnce.Do(func() {
		close(w.stopChan)
	})
}

// processAutoAbsent processes auto-absent for each company in its local
// "yesterday". If no provider is set, falls back to the global timezone.
func (w *AutoAbsentWorker) processAutoAbsent() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// If no provider is configured, fall back to single-timezone processing
	if w.tzProvider == nil {
		w.processSingleTimezone(ctx, "", apptime.Now())
		return
	}

	companyTimezones, err := w.tzProvider.GetDistinctCompanyTimezones()
	if err != nil {
		log.Printf("Auto-absent: failed to get company timezones: %v — falling back to global", err)
		w.processSingleTimezone(ctx, "", apptime.Now())
		return
	}

	if len(companyTimezones) == 0 {
		log.Println("Auto-absent: no active companies found, using global timezone")
		w.processSingleTimezone(ctx, "", apptime.Now())
		return
	}

	// Process each company using its own timezone for "yesterday"
	for companyID, timezone := range companyTimezones {
		loc := apptime.ResolveLocation(timezone)
		yesterday := time.Now().In(loc).AddDate(0, 0, -1)

		log.Printf("Auto-absent: processing company %s (tz=%s) date %s", companyID, timezone, yesterday.Format("2006-01-02"))

		result, err := w.attendanceUC.ProcessAutoAbsent(ctx, yesterday, companyID)
		if err != nil {
			log.Printf("Auto-absent error (company=%s): %v", companyID, err)
			continue
		}

		if result.HolidaySkipped {
			log.Printf("Auto-absent: %s is a holiday for company %s, skipped", result.Date, companyID)
			continue
		}

		log.Printf("Auto-absent completed for company %s on %s: %d employees, %d absent, %d leave, %d skipped, %d errors",
			companyID, result.Date, result.TotalEmployees, result.AbsentCreated, result.LeaveCreated, result.Skipped, result.Errors)
	}
}

// processSingleTimezone is the fallback for when no timezone provider is set.
func (w *AutoAbsentWorker) processSingleTimezone(ctx context.Context, companyID string, now time.Time) {
	yesterday := now.AddDate(0, 0, -1)

	log.Printf("Auto-absent: processing date %s", yesterday.Format("2006-01-02"))

	result, err := w.attendanceUC.ProcessAutoAbsent(ctx, yesterday, companyID)
	if err != nil {
		log.Printf("Auto-absent error: %v", err)
		return
	}

	if result.HolidaySkipped {
		log.Printf("Auto-absent: %s is a holiday, skipped", result.Date)
		return
	}

	log.Printf("Auto-absent completed for %s: %d employees, %d absent, %d leave, %d skipped, %d errors",
		result.Date, result.TotalEmployees, result.AbsentCreated, result.LeaveCreated, result.Skipped, result.Errors)
}
