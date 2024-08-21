package userpoint

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"tradingAce/pkg/utils"
)

type Manager struct {
	db *sql.DB
}

func (m *Manager) UpsertForUserTask(ctx context.Context, address string, taskId string, point int) error {
	query := `
		INSERT INTO "userPoint" ("id", "userAddress", "taskId", "point", "createdAt")
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT ("userAddress", "taskId")
		DO UPDATE SET "point" = EXCLUDED."point"
	`

	_, err := m.db.ExecContext(ctx, query, utils.GenDBID(), address, taskId, point, time.Now())
	if err != nil {
		return fmt.Errorf("UpsertForUserTask failed to upsert user task: %v", err)
	}

	return nil
}
