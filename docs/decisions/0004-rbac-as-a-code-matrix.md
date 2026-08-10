# 0004 — Authorization policy lives in code, not in tables

**Status:** Accepted

## Context

The system has four roles — super admin, admin, manager, salesman — and six permissions.
Users can additionally be given an explicit permission list that overrides their role's
defaults, which is what the per-user permission screen edits.

The usual instinct is a `roles` table, a `permissions` table and a join table.

## Decision

The role-to-permission matrix is a `map[Role][]Permission` in `internal/rbac/rbac.go`.
Adding a permission to a role is one line in that map. Only the per-user override is
persisted, as a list on the user row.

Effective permissions resolve as: explicit user list if present, otherwise the role's
defaults.

## Alternatives considered

**Roles and permissions in the database.** Necessary when customers define their own roles
at runtime. Rejected because this system has a fixed, small set of roles that ship with the
product: a database-driven matrix would mean three more tables, seeders, an admin UI and a
migration for every policy change — to express something that is already a constant.

**Permission strings checked ad hoc in each handler.** Rejected: with no single place
listing the policy, "who can do what" becomes unanswerable without grepping, and typos in a
permission string fail open in the worst case.

## Consequences

- The whole authorization policy is readable in one 90-line file, and it is testable
  without a database. `rbac_test.go` asserts both the grants and the denials, including the
  deliberate rule that an admin cannot manage users.
- A policy change requires a deployment. Acceptable, and arguably correct — a permission
  change is a product decision, not runtime configuration.
- `Permissions()` returns a copy, so a caller cannot mutate the shared matrix and escalate
  a role for every subsequent request in the process. There is a test for exactly that.
- If a customer ever needs custom roles, this record is the thing to revisit; the
  `EffectivePermissions` seam is where a database-backed policy would plug in.
