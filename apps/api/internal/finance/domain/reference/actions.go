package reference

import "strings"

// Standard action names for finance module.
// All finance modules MUST use these to ensure consistent permission keys,
// audit logging, and frontend action mapping.
const (
	ActionView    = "read"
	ActionCreate  = "create"
	ActionUpdate  = "update"
	ActionDelete  = "delete"
	ActionPost    = "post"
	ActionReverse = "reverse"
	ActionApprove = "approve"
	ActionReject  = "reject"
	ActionSubmit  = "submit"
	ActionExport  = "export"
	ActionPay     = "pay"
)

// PermissionKey builds a consistent permission key in the format:
// "<resource>.<action>"
// Example: PermissionKey("journal_entries", ActionRead) → "journal_entries.read"
func PermissionKey(resource, action string) string {
	return strings.ToLower(resource) + "." + action
}
