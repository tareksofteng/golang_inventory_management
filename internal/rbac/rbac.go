// Package rbac holds the authorization policy: which role may do what.
// Keeping it in one small file (a code matrix) is the pragmatic choice for a
// fixed set of roles — no roles/permissions DB tables to manage.
package rbac

// Role is a user's single assigned role.
type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleManager    Role = "manager"
	RoleSalesman   Role = "salesman"
)

// Permission is a capability the API can require on a route.
type Permission string

const (
	PermProductManage  Permission = "product.manage"
	PermPurchaseManage Permission = "purchase.manage"
	PermSalesManage    Permission = "sales.manage"
	PermReportAccess   Permission = "report.access"
	PermUserManage     Permission = "user.manage"
)

// rolePermissions is the policy matrix. Super admin gets everything; each lower
// role gets a subset. Adding a permission to a role = one line here.
var rolePermissions = map[Role][]Permission{
	RoleSuperAdmin: {PermProductManage, PermPurchaseManage, PermSalesManage, PermReportAccess, PermUserManage},
	RoleAdmin:      {PermProductManage, PermPurchaseManage, PermSalesManage, PermReportAccess},
	RoleManager:    {PermProductManage, PermPurchaseManage, PermSalesManage},
	RoleSalesman:   {PermSalesManage},
}

// AllPermissions lists every permission the system knows about (for building
// the user permission checkboxes).
func AllPermissions() []Permission {
	return []Permission{PermProductManage, PermPurchaseManage, PermSalesManage, PermReportAccess, PermUserManage}
}

// IsValidPermission reports whether p is a known permission string.
func IsValidPermission(p string) bool {
	for _, perm := range AllPermissions() {
		if string(perm) == p {
			return true
		}
	}
	return false
}

// EffectivePermissions returns a user's actual permissions: the explicit custom
// list if set, otherwise the defaults for their role.
func EffectivePermissions(role string, custom []string) []string {
	if len(custom) > 0 {
		return custom
	}
	perms := rolePermissions[Role(role)]
	out := make([]string, len(perms))
	for i, p := range perms {
		out[i] = string(p)
	}
	return out
}

// IsValidRole reports whether r is one of the known roles (used in validation).
func IsValidRole(r Role) bool {
	_, ok := rolePermissions[r]
	return ok
}

// HasPermission reports whether the role is allowed the given permission.
func HasPermission(r Role, p Permission) bool {
	for _, perm := range rolePermissions[r] {
		if perm == p {
			return true
		}
	}
	return false
}

// Permissions returns all permissions granted to a role (handy for /me responses
// so the frontend can show/hide menus).
func Permissions(r Role) []Permission {
	perms := rolePermissions[r]
	out := make([]Permission, len(perms))
	copy(out, perms)
	return out
}
