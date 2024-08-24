package userpoint

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"tradingAce/pkg/model"
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

func (m *Manager) GetUserPointsForTask(ctx context.Context, taskID string) ([]model.UserPoint, error) {
	query := `SELECT "id", "userAddress", "createdAt", "taskId", "point" FROM "userPoint"`

	var args []interface{}

	if len(taskID) != 0 {
		query += ` WHERE "taskId" = $1`
		args = append(args, taskID)
	}

	stmt, err := m.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	var userPoints []model.UserPoint
	for rows.Next() {
		var userPoint model.UserPoint
		if err := rows.Scan(&userPoint.ID, &userPoint.UserAddress, &userPoint.CreatedAt, &userPoint.TaskID, &userPoint.Point); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		userPoints = append(userPoints, userPoint)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %v", err)
	}

	return userPoints, nil
}
