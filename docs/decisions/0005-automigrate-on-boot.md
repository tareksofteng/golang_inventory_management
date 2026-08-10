# 0005 — Schema is applied by AutoMigrate on boot

**Status:** Accepted for now, will not survive production

## Context

`cmd/api/main.go` calls `models.AutoMigrate(db)` at startup, then seeds the first super
admin and the default chart of accounts. Clone, set `.env`, run — the database is ready.
That property is worth a lot while the model layer is still changing weekly.

## Decision

GORM `AutoMigrate` applies the schema on every boot. Seeding is idempotent: the super admin
and the chart of accounts are created only when absent.

## Alternatives considered

**Versioned migration files (golang-migrate, goose).** The production answer. Rejected for
now because the models are still moving; a migration file per model change would be pure
overhead at this stage, and the ordering constraints would slow down exactly the
experimentation this phase needs.

**Schema dump checked into the repository.** Reproducible, but no upgrade path for an
existing database, which is the whole problem migrations solve.

## Consequences

- Setup is one command, and CI or a fresh clone never needs a migration step.
- `AutoMigrate` will add columns and indexes but will not drop or rename them, and it will
  not carry data through a type change. Any destructive change has to be done by hand,
  which is a trap once real data exists.
- There is no record of *when* a schema change happened, so two deployments can silently
  differ.
- The trigger for switching: the first deployment holding data someone would miss. At that
  point the current schema becomes migration `0001` and AutoMigrate comes out of the boot
  path. Doing it earlier costs more than it returns.
