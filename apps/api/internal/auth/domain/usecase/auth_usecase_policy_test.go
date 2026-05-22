package usecase

import "testing"

func TestIsAssignmentAllowedByPolicy_AllowsExactAndPrefixMenuMatches(t *testing.T) {
	policy := planPermissionPolicy{
		menuPrefixes: []string{"/sales/orders", masterDataMenuPrefix},
		hasRules:     true,
	}

	tests := []struct {
		name string
		row  rolePermissionAssignment
		want bool
	}{
		{
			name: "allows exact menu entitlement",
			row:  rolePermissionAssignment{MenuURL: "/sales/orders"},
			want: true,
		},
		{
			name: "allows master data subtree",
			row:  rolePermissionAssignment{MenuURL: "/master-data/users"},
			want: true,
		},
		{
			name: "blocks unrelated menu",
			row:  rolePermissionAssignment{MenuURL: "/purchase/purchase-orders"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAssignmentAllowedByPolicy(tt.row, policy)
			if got != tt.want {
				t.Fatalf("isAssignmentAllowedByPolicy(%q) = %v, want %v", tt.row.MenuURL, got, tt.want)
			}
		})
	}
}