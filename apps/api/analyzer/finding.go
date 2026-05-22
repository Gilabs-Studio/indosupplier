package analyzer

import "fmt"

// Severity levels for findings
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityError    Severity = "ERROR"
	SeverityWarning  Severity = "WARNING"
	SeverityInfo     Severity = "INFO"
	SeverityPass     Severity = "PASS"
	SeveritySkipped  Severity = "SKIPPED"
)

// Finding represents a single audit result
type Finding struct {
	Code           string   `json:"code"`
	Severity       Severity `json:"severity"`
	Module         string   `json:"module"`
	Entity         string   `json:"entity,omitempty"`
	Flow           string   `json:"flow,omitempty"`
	ScenarioID     string   `json:"scenario_id,omitempty"`
	Message        string   `json:"message"`
	Evidence       string   `json:"evidence,omitempty"`
	Recommendation string   `json:"recommendation,omitempty"`
	RequestID      string   `json:"request_id,omitempty"`
}

// String formats a finding for console display
func (f Finding) String() string {
	icon := f.Icon()
	prefix := fmt.Sprintf("[%s]", f.Code)
	if f.ScenarioID != "" {
		prefix = fmt.Sprintf("[%s] %s —", f.Code, f.ScenarioID)
	}
	s := fmt.Sprintf("%s %s %s: %s", icon, prefix, f.Module, f.Message)
	if f.Evidence != "" {
		s += fmt.Sprintf("\n     Evidence: %s", f.Evidence)
	}
	if f.RequestID != "" {
		s += fmt.Sprintf("\n     Request ID: %s", f.RequestID)
	}
	if f.Recommendation != "" {
		s += fmt.Sprintf("\n     Fix: %s", f.Recommendation)
	}
	return s
}

// Icon returns the severity icon
func (f Finding) Icon() string {
	switch f.Severity {
	case SeverityCritical:
		return "🔴"
	case SeverityError:
		return "❌"
	case SeverityWarning:
		return "⚠️"
	case SeverityInfo:
		return "ℹ️"
	case SeverityPass:
		return "✅"
	case SeveritySkipped:
		return "⏭️"
	default:
		return "  "
	}
}

// IsFailure returns true if this finding represents a problem
func (f Finding) IsFailure() bool {
	return f.Severity == SeverityCritical || f.Severity == SeverityError
}

// Checker is the interface all validators/scanners must implement
type Checker interface {
	Name() string
	Run(cfg *Config) []Finding
}
