package task

import (
	"context"
	"database/sql"
	"fmt"
	"tradingAce/pkg/model"
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
