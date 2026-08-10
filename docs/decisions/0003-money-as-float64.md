# 0003 — Money is `float64` with a half-cent tolerance

**Status:** Accepted reluctantly, migration planned

## Context

Every amount in this system is `float64` in Go and `decimal(14,2)` in MySQL. Binary
floating point cannot represent most decimal fractions exactly, so sums of prices and
percentages accumulate small errors. In an accounting system that matters twice over:
once on the invoice a customer reads, and once on whether a journal entry balances.

## Decision

Keep `float64` for now, and make the imprecision explicit everywhere it could cause a wrong
decision:

- The journal accepts an entry as balanced when `|debit − credit| ≤ 0.005` — half a cent —
  rather than comparing for exact equality, which would reject correct entries.
- The database column is `decimal(14,2)`, so what is stored is exact even when what was
  computed was not.
- Tests compare money with the same half-cent tolerance rather than `==`.

## Alternatives considered

**Integer minor units (store cents as `int64`).** The standard answer, exact by
construction. Rejected at this stage only because it touches every model, every request
payload, every report and the whole frontend at once. It is the right migration, not a
cheap one.

**A decimal library (`shopspring/decimal`).** Exact, and a much smaller diff than switching
to integers. Rejected for now because GORM scanning, JSON encoding and the Vue client all
need adapting, and the half-cent tolerance covers the failure mode that actually bites at
this scale.

## Consequences

- Totals are correct to the cent for realistic invoice sizes, and the tolerance keeps
  correct entries from being rejected.
- A long-running report that sums thousands of rows in Go can still drift. Reports should
  aggregate in SQL over the `decimal` columns rather than in Go — worth auditing.
- The migration to integer minor units stays on the roadmap, and this record is the reason
  it is there. Whoever does it should start with `computeTotals` and the journal balance
  check, since those already have tests to migrate against.
