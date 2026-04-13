package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/orimono/ito"
	"github.com/orimono/shutter/internal/capability/executor"
	_ "modernc.org/sqlite"
)

type ExecutorStore struct {
	db *sql.DB
}

func NewExecutorStore(path string) (*ExecutorStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS executors (
			kind                TEXT PRIMARY KEY,
			version             TEXT NOT NULL,
			runtime             TEXT NOT NULL,
			script              TEXT NOT NULL,
			platforms           TEXT NOT NULL DEFAULT '[]',
			risk                INTEGER NOT NULL DEFAULT 0,
			requires_elevation  INTEGER NOT NULL DEFAULT 0,
			warning             TEXT NOT NULL DEFAULT ''
		)
	`); err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}
	return &ExecutorStore{db: db}, nil
}

func (s *ExecutorStore) Save(ctx context.Context, reg ito.ExecutorRegistration) error {
	platforms, err := json.Marshal(reg.Platforms)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO executors (kind, version, runtime, script, platforms, risk, requires_elevation, warning)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(kind) DO UPDATE SET
			version            = excluded.version,
			runtime            = excluded.runtime,
			script             = excluded.script,
			platforms          = excluded.platforms,
			risk               = excluded.risk,
			requires_elevation = excluded.requires_elevation,
			warning            = excluded.warning
	`, reg.Kind, reg.Version, reg.Runtime, reg.Script, string(platforms),
		int(reg.Risk), reg.RequiresElevation, reg.Warning)
	return err
}

func (s *ExecutorStore) LoadAll() ([]*executor.ScriptExecutor, error) {
	rows, err := s.db.Query(`
		SELECT kind, version, runtime, script, platforms, risk, requires_elevation, warning
		FROM executors
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*executor.ScriptExecutor
	for rows.Next() {
		var (
			reg           ito.ExecutorRegistration
			platformsJSON string
			risk          int
			requiresElev  bool
		)
		if err := rows.Scan(
			&reg.Kind, &reg.Version, &reg.Runtime, &reg.Script,
			&platformsJSON, &risk, &requiresElev, &reg.Warning,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(platformsJSON), &reg.Platforms); err != nil {
			return nil, err
		}
		reg.Risk = ito.RiskLevel(risk)
		reg.RequiresElevation = requiresElev
		result = append(result, executor.NewScriptExecutor(reg))
	}
	return result, rows.Err()
}

// NewScriptExecutor is a convenience re-export for dispatcher use.
func NewScriptExecutor(reg ito.ExecutorRegistration) *executor.ScriptExecutor {
	return executor.NewScriptExecutor(reg)
}
