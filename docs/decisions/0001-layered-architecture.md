# 0001 — Controller / service / repository layering

**Status:** Accepted

## Context

This is an inventory system with accounting attached, so the same operation is reached from
more than one direction: a sale is created by the sales handler, but its stock effect and
its journal entry also need to happen when a return is processed or when data is seeded.
If the rules live in the HTTP handler, every new entry point copies them, and the copies
drift.

Go has no framework opinion to inherit here. GORM makes it very easy to write a handler
that queries the database directly, which is exactly the shape that becomes untestable.

## Decision

Three layers, with a strict direction of dependency:

- **Controllers** parse and validate the request, call one service method, and shape the
  response. No business rules, no GORM.
- **Services** own the business rules and the transaction boundaries. This is where "a sale
  decrements stock, adds to the customer's due, and posts a journal entry" lives.
- **Repositories** are the only code that touches GORM. They expose intent-named methods
  (`FindByCode`, `CountAll`) rather than a generic query builder.

Services depend on repository *interfaces*, not on concrete types.

## Alternatives considered

**Handlers calling GORM directly.** Fastest to write and normal in small Go services.
Rejected because the accounting rules are the interesting part of this project, and they
would have ended up spread across a dozen handlers with no way to test them without a
database.

**A single service package with no repository layer.** Less indirection, but every test
would then need a live MySQL instance, and the posting rules are exactly what needs to be
tested cheaply and often.

## Consequences

- The posting rules are unit-testable with fakes: `accounting_poster_test.go` asserts that
  every transaction type balances without a database anywhere in the test.
- More files and more wiring in `main.go` than a flat design needs. Accepted cost.
- The repository interfaces are wide, since each aggregate exposes its own methods. Tests
  embed the interface in a stub and override only the method under test, so the width does
  not turn into test boilerplate.
