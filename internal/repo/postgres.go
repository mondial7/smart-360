package repo

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned by repositories when a lookup matches no row. Callers
// distinguish "absent" from "failed" with errors.Is(err, ErrNotFound).
var ErrNotFound = errors.New("not found")

// rowScanner is satisfied by both pgx.Row and pgx.Rows, so a single scan helper
// serves QueryRow and iterated Query results.
type rowScanner interface {
	Scan(dest ...any) error
}

// querier is satisfied by both *pgxpool.Pool and pgx.Tx, so repository methods
// can run either directly on the pool or inside a transaction.
type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// NewPostgres builds the full set of pgx-backed repositories from a pool.
func NewPostgres(pool *pgxpool.Pool) Repositories {
	return Repositories{
		Users:          &pgUsers{pool},
		Teams:          &pgTeams{pool},
		Rounds:         &pgRounds{pool},
		Submissions:    &pgSubmissions{pool},
		Templates:      &pgTemplates{pool},
		Consolidations: &pgConsolidations{pool},
		Audit:          &pgAudit{pool},
		Moderation:     &pgModeration{pool},
		Sessions:       &pgSessions{pool},
	}
}

// mustJSON marshals v to JSON for a jsonb parameter, returning "null" on error
// (errors here indicate a programming bug, not runtime data, and are rare).
func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("null")
	}
	return b
}

// decodeJSON unmarshals a jsonb column value into dst, tolerating NULL/empty.
func decodeJSON(src []byte, dst any) error {
	if len(src) == 0 || string(src) == "null" {
		return nil
	}
	return json.Unmarshal(src, dst)
}

// normalizeErr maps pgx's no-rows sentinel to our ErrNotFound.
func normalizeErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// prefixed qualifies each comma-separated column in cols with a table alias,
// e.g. prefixed("fr", "id, name") -> "fr.id, fr.name". Used for join queries
// that reuse a table's column list.
func prefixed(alias, cols string) string {
	parts := strings.Split(cols, ", ")
	for i, p := range parts {
		parts[i] = alias + "." + p
	}
	return strings.Join(parts, ", ")
}
