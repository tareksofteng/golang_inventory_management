# 0002 — Accounting posts are best-effort, outside the business transaction

**Status:** Accepted, with a known cost

## Context

Creating a sale does several things: insert the invoice and its lines, decrement stock,
increase the customer's due, and post a journal entry (Dr Cash / Dr A/R, Cr Sales, plus the
COGS pair). The first three are the transaction the user asked for. The fourth is
bookkeeping that follows from it.

The question is what happens when the fourth one fails — a missing account in the chart, a
seeding gap, a validation error in the journal service.

## Decision

The business transaction commits first. Posting runs afterwards and is best-effort: on
failure it logs `accounting: auto-post failed for <ref>` and returns. The sale stands.

```go
if _, err := p.journal.Create(...); err != nil {
    log.Printf("accounting: auto-post failed for %s: %v", ref, err)
}
```

## Alternatives considered

**Post inside the same database transaction and roll the sale back on failure.** Books can
never drift from operations. Rejected because it makes the accounting subsystem a hard
dependency of the till: a missing account code in the seeded chart would stop a shop from
selling. For a retail system, refusing a sale because bookkeeping is misconfigured is the
worse failure.

**Queue the posting for retry.** The correct answer, and where this should end up. Rejected
for now because it needs a job runner and an outbox table, which is more infrastructure
than this codebase currently carries.

## Consequences

- A posting failure is silent to the user and visible only in the log. The ledger can drift
  from operations without anyone noticing that day.
- That risk is what `accounting_poster_test.go` exists for: every posting type is asserted
  to balance, to have at least two lines, and to resolve every account, so the common
  causes of failure are caught before deployment rather than in the log.
- Two follow-ups are open: a reconciliation report that compares sales totals against the
  Sales Income account for a period, and an outbox with retry to replace the log line.
