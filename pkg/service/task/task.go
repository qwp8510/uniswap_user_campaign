package task

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

func (m *Manager) GetOnboardingTask(ctx context.Context) (model.Task, error) {
	query := `
		SELECT "id", "createdAt", "name", "pairAddress", "startAt"
		FROM "task"
		WHERE "name" = $1;
    `

	var task model.Task
	err := m.db.QueryRowContext(ctx, query, "onboarding").Scan(
		&task.ID,
		&task.CreatedAt,
		&task.Name,
		&task.PairAddress,
		&task.StartAt,
	)

	return task, err
}

func (m *Manager) GetSharePoolTask(ctx context.Context) ([]model.Task, error) {
	query := `
		SELECT "id", "createdAt", "name", "pairAddress", "startAt"
		FROM "task"
		WHERE "name" = $1;
    `

	tasks := make([]model.Task, 0)
	rows, err := m.db.Query(query, "share_pool")
	if err != nil {
		return tasks, fmt.Errorf("GetSharePoolTask query fail: %v", err)
	}
	for rows.Next() {
		var task model.Task
		err := rows.Scan(
			&task.ID,
			&task.CreatedAt,
			&task.Name,
			&task.PairAddress,
			&task.StartAt,
		)
		if err != nil {
			return tasks, fmt.Errorf("GetSharePoolTask scan fail: %v", err)
		}

		tasks = append(tasks, task)
	}

	return tasks, err
}

func (m *Manager) CreateSharePoolTask(ctx context.Context, pairAddress string, startAt time.Time) error {
	query := `
		SELECT "id", "createdAt", "name", "pairAddress", "startAt"
		FROM "task"
		WHERE "name" = $1 AND "pairAddress" = $2;
	`

	var task model.Task
	qErr := m.db.QueryRowContext(ctx, query, "share_pool", pairAddress).Scan(
		&task.ID,
		&task.CreatedAt,
		&task.Name,
		&task.PairAddress,
		&task.StartAt,
	)
	if qErr != sql.ErrNoRows {
		return fmt.Errorf("task pairAddress exist: %s", pairAddress)
	} else if qErr != nil && qErr != sql.ErrNoRows {
		return qErr
	}

	insertQuery := `
		INSERT INTO task ("id", "createdAt", "name", "pairAddress", "startAt")
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := m.db.ExecContext(ctx, insertQuery, utils.GenDBID(), time.Now(), "share_pool", pairAddress, startAt)
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	return nil
}
