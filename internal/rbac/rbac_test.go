package rbac

import (
	"reflect"
	"testing"
)

// The policy matrix is the whole authorization story, so it is worth asserting
// explicitly rather than trusting the map to stay correct through edits. The
// negative cases are the important ones: a salesman must never reach user
// management, and only a super admin may create users.
func TestRolePermissionMatrix(t *testing.T) {
	cases := []struct {
		role Role
		perm Permission
		want bool
	}{
		{RoleSuperAdmin, PermUserManage, true},
		{RoleSuperAdmin, PermAccountManage, true},

		{RoleAdmin, PermProductManage, true},
		{RoleAdmin, PermAccountManage, true},
		{RoleAdmin, PermUserManage, false}, // deliberate: only super admin manages users

		{RoleManager, PermSalesManage, true},
		{RoleManager, PermPurchaseManage, true},
		{RoleManager, PermReportAccess, false},
		{RoleManager, PermAccountManage, false},

		{RoleSalesman, PermSalesManage, true},
		{RoleSalesman, PermProductManage, false},
		{RoleSalesman, PermUserManage, false},
		{RoleSalesman, PermAccountManage, false},

		{Role("auditor"), PermReportAccess, false}, // unknown role gets nothing
		{Role(""), PermSalesManage, false},
	}

	for _, tc := range cases {
		if got := HasPermission(tc.role, tc.perm); got != tc.want {
			t.Errorf("HasPermission(%q, %q) = %v, want %v", tc.role, tc.perm, got, tc.want)
		}
	}
}

// An explicit per-user permission list overrides the role defaults; an empty
// list falls back to them. Getting this backwards would silently grant a user
// everything their role allows after an admin had revoked it.
func TestEffectivePermissions(t *testing.T) {
	t.Run("custom list wins over the role defaults", func(t *testing.T) {
		custom := []string{string(PermReportAccess)}
		got := EffectivePermissions(string(RoleSalesman), custom)
		if !reflect.DeepEqual(got, custom) {
			t.Errorf("got %v, want %v", got, custom)
		}
	})

	t.Run("empty custom list falls back to the role", func(t *testing.T) {
		got := EffectivePermissions(string(RoleManager), nil)
		want := []string{
			string(PermProductManage),
			string(PermPurchaseManage),
			string(PermSalesManage),
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("unknown role grants nothing", func(t *testing.T) {
		if got := EffectivePermissions("auditor", nil); len(got) != 0 {
			t.Errorf("unknown role got %v, want no permissions", got)
		}
	})
}

// Permissions returns a copy: a caller mutating the returned slice must not be
// able to rewrite the policy for every other request in the process.
func TestPermissionsReturnsACopy(t *testing.T) {
	perms := Permissions(RoleSalesman)
	if len(perms) == 0 {
		t.Fatal("salesman should have at least one permission")
	}
	perms[0] = PermUserManage

	if HasPermission(RoleSalesman, PermUserManage) {
		t.Error("mutating the returned slice escalated the salesman role")
	}
}

func TestValidation(t *testing.T) {
	for _, r := range []Role{RoleSuperAdmin, RoleAdmin, RoleManager, RoleSalesman} {
		if !IsValidRole(r) {
			t.Errorf("IsValidRole(%q) = false, want true", r)
		}
	}
	for _, r := range []Role{"", "root", "Admin"} {
		if IsValidRole(r) {
			t.Errorf("IsValidRole(%q) = true, want false", r)
		}
	}

	for _, p := range AllPermissions() {
		if !IsValidPermission(string(p)) {
			t.Errorf("IsValidPermission(%q) = false, want true", p)
		}
	}
	for _, p := range []string{"", "product.delete", "Product.Manage"} {
		if IsValidPermission(p) {
			t.Errorf("IsValidPermission(%q) = true, want false", p)
		}
	}
}

// Every permission a role is granted must be a known permission, or a route
// guard could require something no role can ever satisfy.
func TestMatrixOnlyReferencesKnownPermissions(t *testing.T) {
	for role := range rolePermissions {
		for _, p := range rolePermissions[role] {
			if !IsValidPermission(string(p)) {
				t.Errorf("role %q references unknown permission %q", role, p)
			}
		}
	}
}
