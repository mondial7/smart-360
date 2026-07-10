# ADR 0007: Normalize the relational schema (join tables + typed jsonb)

- **Status**: accepted
- **Date**: 2026-07-08
- **Deciders**: Marco Mondini

## Context

The MongoDB model leaned on document idioms that don't translate cleanly to a
relational store: embedded arrays (a round's `reviewers[]`, a team's
`member_ids[]`), and several consolidation fields persisted as **JSON strings**
inside the document (`strengths`, `areas_for_improvement`, `actionable_insights`,
`question_summaries`) because the old code found that expedient. The rewrite is
the moment to model this data the way Postgres wants it.

## Decision

We will use a **normalized relational schema** with `uuid` primary keys:

- Embedded arrays become **join tables**: `round_reviewers` (round_id,
  reviewer_id, unique together) and `team_members` (team_id, user_id).
- The `submissions` uniqueness rule ("one submission per reviewer per round")
  becomes a real `UNIQUE (round_id, reviewer_id)` constraint.
- Config/snapshot data with no query needs stays as **`jsonb`** columns:
  template `questions`/`competencies`, submission `responses`/`ratings`, and the
  consolidation sub-documents. The legacy JSON-*string* fields are promoted to
  proper typed values in Go and stored as real `jsonb`.
- `audit_logs` deliberately **caches display fields and is not foreign-keyed**,
  so the trail survives deletion of the entities it references.

`users.team_id` is kept as a denormalized single-team pointer (maintained
alongside `team_members`) because the role checks key on it.

## Consequences

- Referential integrity and uniqueness are enforced by the database, not by
  application code that "mostly" gets it right.
- jsonb keeps schema-flexible, non-queried blobs simple while letting us index
  and constrain the relational parts.
- The denormalized `users.team_id` must be maintained in tandem with
  `team_members`; the team repository owns that dual-write.
- Reviewer/member membership needs a follow-up query (or join) on read; at this
  scale that's cheap and explicit.

## Alternatives considered

- **Mirror the document model with array/jsonb columns for reviewers/members** —
  keeps writes simple but loses FK integrity, uniqueness, and easy reverse
  lookups ("which rounds is this user reviewing?").
- **Fully normalize everything, including template questions** — needless: that
  data is read as a whole snapshot and never queried by inner field.
