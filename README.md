# Inventory & Accounting API (Go)

Inventory, sales, purchasing and double-entry accounting behind a REST API in Go, with a
Vue 3 dashboard on top. Built to learn Go properly by writing something with real
constraints: money that has to balance, stock that has to reconcile, and reports an
accountant would recognise.

[![CI](https://github.com/tareksofteng/golang_inventory_management/actions/workflows/ci.yml/badge.svg)](https://github.com/tareksofteng/golang_inventory_management/actions/workflows/ci.yml)
![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-1.12-008ECF)
![GORM](https://img.shields.io/badge/GORM-1.31-red)
![Vue](https://img.shields.io/badge/Vue-3-4FC08D?logo=vue.js&logoColor=white)

## What it does

- **Inventory** — products, categories, stock movement history per product
- **Purchasing** — purchase orders, purchase returns, suppliers, supplier payments
- **Sales** — sales orders, sales returns, customers, customer payments
- **Accounting** — double-entry chart of accounts, journal entries, ledger, trial balance,
  profit & loss over a date range, balance sheet as of a date, COGS and expenses
- **Reports** — dashboard analytics plus CSV export on the trial balance and account ledger
- **Access control** — JWT auth with role-based permissions, per-user permission overrides

Business events post their own journal entries. Recording a sale writes the inventory
movement and the accounting entries in the same flow, so the books never drift from the
stock ledger.

## Layout

```
cmd/api/main.go              entry point
config/                      env loading and app config
internal/
├── controllers/             HTTP handlers, one per resource (18)
├── services/                business logic, transactions, posting rules (20)
├── repositories/            data access, one per aggregate (17)
├── models/                  domain models (15)
├── middleware/              auth, CORS
├── rbac/                    role and permission resolution
├── routes/                  route registration
└── seeder/                  chart of accounts + demo data
pkg/
├── auth/                    JWT issue and verify
├── database/                connection and migration bootstrap
├── pagination/              shared list pagination
└── response/                uniform JSON envelope + validation errors
docs/                        generated Swagger (swagger.json, swagger.yaml)
frontend/                    Vue 3 + Vite + Tailwind dashboard
```

Controller → service → repository, with models shared across layers. Handlers do no
business logic; services own transaction boundaries; repositories are the only code that
touches GORM. That separation is the point of the project — it makes the accounting rules
testable without a HTTP request.

## Why it looks the way it does

The longer versions, with the options that were rejected and what each choice costs, are in
[`docs/decisions/`](docs/decisions/).

**Double-entry from the first commit, not bolted on.** Every posting writes balanced debit
and credit lines. Reports are derived from the journal rather than stored as running
totals, so a corrected entry corrects every downstream report automatically.

**System accounts are protected from deletion.** Cash, inventory, COGS and similar accounts
carry a flag the delete handler refuses. A deleted system account would silently break
every future posting.

**Soft ordering on reports.** Trial balance, P&L and balance sheet take date filters that
default to today, so the first screen a user sees is never an accidental full-history scan.

**Uniform response envelope.** `pkg/response` gives every endpoint the same success and
error shape, including validation errors from `validator/v10`. The Vue client has one error
path, not one per endpoint.

## Running it

Requires Go 1.26+, MySQL 8, Node 20+.

```bash
cp .env.example .env         # set DB credentials and JWT secret
go mod download
go run ./cmd/api             # or: make run
```

The API listens on `APP_PORT` (9000 by default). `AutoMigrate` and the seeder run on boot,
creating the schema, the default chart of accounts and the first super-admin.

Frontend:

```bash
cd frontend
npm install
npm run dev
```

Swagger UI is mounted at `/swagger/index.html`. Regenerate after changing annotations:

```bash
swag init -g cmd/api/main.go
```

## Tests

```bash
go test ./...
```

The suite covers the parts where a bug is expensive and silent:

- **Auto-posting balances.** Every transaction type (sale, purchase, both payment
  directions, both return types) is posted through a recording fake and asserted to have
  equal debits and credits, at least two lines, no line carrying both a debit and a credit,
  and no line pointing at an unresolved account. An unbalanced entry would corrupt the
  trial balance and both statements without raising an error anywhere.
- **Invoice totals.** Tax is charged on the net amount, the discount is clamped to
  `[0, subtotal]`, and the grand total is always `taxable + tax` across a matrix of
  subtotals, discounts and rate bands. The clamp is the one that matters: an over-discount
  would otherwise flip the sign of the VAT.
- **Authorization matrix.** Positive and negative cases per role, including the deliberate
  rule that an admin cannot manage users, and that `Permissions()` hands back a copy so a
  caller cannot mutate the policy for the whole process.
- **Pagination.** Bad input falls back to defaults, `per_page` is clamped to the ceiling,
  and the OFFSET and page-count arithmetic is checked at the boundaries.

CI runs `gofmt`, `go build`, `go vet` and the tests on every push and pull request.

## Stack

Go 1.26 · Gin · GORM (MySQL driver) · golang-jwt v5 · validator/v10 · swaggo ·
Vue 3 · Vite · Tailwind CSS

## Status

Active. Working next on stock reconciliation coverage and Docker packaging for a
one-command run.
