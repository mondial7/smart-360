# ADR 0001: Record architecture decisions

- **Status**: accepted
- **Date**: 2026-07-08
- **Deciders**: Marco Mondini

## Context

Smart 360 is undergoing a significant architectural change (a full rewrite of its
stack). Decisions of this weight — framework choices, data model, auth model —
have long-lived consequences, and the reasoning behind them is easily lost once
the code lands. Until now the repo had no durable record of *why* the
architecture is the way it is; context lived in PR descriptions and memory.

## Decision

We will keep Architecture Decision Records (ADRs) in `docs/adr/`, one Markdown
file per decision, numbered sequentially (`NNNN-kebab-title.md`). Each record
follows the lightweight MADR-style template in `0000-template.md`: Context,
Decision, Consequences, Alternatives considered.

ADRs are immutable once accepted. A decision that reverses an earlier one gets a
new ADR and marks the old one *superseded by ADR-XXXX* rather than editing it.

## Consequences

- New contributors (human or AI) can read the ADR log to understand the "why"
  without archaeology through git history.
- Small overhead per significant decision; trivial changes do not need an ADR.
- The log becomes the canonical place to look before revisiting a settled
  question.

## Alternatives considered

- **A single ARCHITECTURE.md** — harder to keep decisions atomic and immutable;
  tends to be rewritten in place, losing the history we want to preserve.
- **PR descriptions only** — not discoverable; scattered across a provider that
  may change.
