package usecase

// activityTypeFollowUpID is the seeded ID for "Follow Up" activity types.
// This is used when generating audit trail entries for record edits and stage changes.
const activityTypeFollowUpID = "ce000001-0000-0000-0000-000000000005"

func strPtr(s string) *string {
	return &s
}
