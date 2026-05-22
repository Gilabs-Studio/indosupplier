package analyzer

import "time"

// Config holds all analyzer configuration from CLI flags
type Config struct {
	// Modules to analyze (empty = all)
	Modules []string

	// Date range for financial checks
	FromDate time.Time
	ToDate   time.Time

	// Strict mode: warnings become errors
	Strict bool

	// Output formats: "console", "json", "txt"
	OutputFormats []string

	// Output directory for file reports
	OutputDir string

	// DryRun: print what would be checked without running
	DryRun bool

	// DSN override (empty = use app config)
	DSN string

	// BatchLimit for heavy queries
	BatchLimit int

	// Mode: "validate", "simulation", "full"
	Mode string

	// Scenario: filter simulation by scenario ID
	Scenario string

	// Cleanup: auto-cleanup test data
	Cleanup bool

	// FailOnSkippedMandatory: fail if mandatory scenarios are not covered
	FailOnSkippedMandatory bool
}

const (
	ModeValidate   = "validate"
	ModeSimulation = "simulation"
	ModeFull       = "full"
)

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	now := time.Now().UTC()
	return &Config{
		Modules:       []string{},
		FromDate:      time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
		ToDate:        now,
		Strict:        false,
		OutputFormats: []string{"console"},
		OutputDir:              "./reports",
		DryRun:                 false,
		BatchLimit:             1000,
		Mode:                   ModeFull,
		Cleanup:                true,
		FailOnSkippedMandatory: false,
	}
}

// ShouldRunModule returns true if the given module should be analyzed
func (c *Config) ShouldRunModule(module string) bool {
	if len(c.Modules) == 0 {
		return true
	}
	for _, m := range c.Modules {
		if m == module {
			return true
		}
	}
	return false
}

func (c *Config) FromDateStr() string {
	return c.FromDate.Format("2006-01-02") + " 00:00:00"
}

func (c *Config) ToDateStr() string {
	return c.ToDate.Format("2006-01-02") + " 23:59:59"
}

