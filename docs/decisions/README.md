# Architecture decisions

Short records of the decisions that shaped this codebase, including the options that were
rejected and what each choice costs. Written so that someone reading the code six months
from now (including me) does not have to reverse-engineer the reasoning.

Format: context, decision, alternatives considered, consequences. One file per decision,
numbered in the order they were made. A decision that gets reversed later gets a new
record rather than an edit to the old one.

| # | Decision | Status |
|---|---|---|
| [0001](0001-layered-architecture.md) | Controller / service / repository layering | Accepted |
| [0002](0002-best-effort-accounting-posts.md) | Accounting posts are best-effort, outside the business transaction | Accepted, with a known cost |
| [0003](0003-money-as-float64.md) | Money is `float64` with a half-cent tolerance | Accepted reluctantly, migration planned |
| [0004](0004-rbac-as-a-code-matrix.md) | Authorization policy lives in code, not in tables | Accepted |
| [0005](0005-automigrate-on-boot.md) | Schema is applied by AutoMigrate on boot | Accepted for now, will not survive production |
