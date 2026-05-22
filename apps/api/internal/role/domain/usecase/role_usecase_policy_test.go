package usecase

import "testing"

func TestIsPermissionAllowedByPolicy_AllowsExactAndPrefixMenuMatches(t *testing.T) {
	policy := planPermissionPolicy{
		menuPrefixes: []string{"/sales/orders", masterDataMenuPrefix},
		hasRules:     true,
	}

	tests := []struct {
		name string
		row  permissionMetaRow
		want bool
	}{
		{
			name: "allows exact menu entitlement",
			row:  permissionMetaRow{MenuURL: "/sales/orders"},
			want: true,
		},
		{
			name: "allows master data subtree",
			row:  permissionMetaRow{MenuURL: "/master-data/users"},
			want: true,
		},
		{
			name: "blocks unrelated menu",
			row:  permissionMetaRow{MenuURL: "/purchase/purchase-orders"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPermissionAllowedByPolicy(tt.row, policy)
			if got != tt.want {
				t.Fatalf("isPermissionAllowedByPolicy(%q) = %v, want %v", tt.row.MenuURL, got, tt.want)
			}
		})
	}
}