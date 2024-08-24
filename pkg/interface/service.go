package iface

import (
	"context"
	"tradingAce/pkg/model"
	"tradingAce/pkg/model/option"

	"github.com/shopspring/decimal"
)

type TaskManager interface {
	GetOnboardingTask(ctx context.Context) (model.Task, error)
	GetSharePoolTask(ctx context.Context) ([]model.Task, error)
}

type UserTaskManager interface {
	CheckOnboardingTask(ctx context.Context, address string) error
	CheckSharePoolTasks(ctx context.Context) error
	Upsert(ctx context.Context, address string, taskId string, state string, amount decimal.Decimal) error
	GetUserTasks(ctx context.Context, address string) ([]option.GetUserTaskPoint, error)
}

type TransactionManager interface {
	Upsert(ctx context.Context, opt option.TransactionUpsertOptions) error
	GetUserUSDC(ctx context.Context, address string) (decimal.Decimal, error)
}

type UserPointManager interface {
	UpsertForUserTask(ctx context.Context, address string, taskId string, point int) error
	GetUserPointsForTask(ctx context.Context, taskID string) ([]model.UserPoint, error)
}
