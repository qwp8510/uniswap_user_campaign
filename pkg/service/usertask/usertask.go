package usertask

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
	"tradingAce/pkg/constants"
	iface "tradingAce/pkg/interface"
	"tradingAce/pkg/model"
	"tradingAce/pkg/utils"

	"github.com/shopspring/decimal"
)

type Manager struct {
	db             *sql.DB
	taskMgr        iface.TaskManager
	transactionMgr iface.TransactionManager
	userPointMgr   iface.UserPointManager
}

// cache onboarding task
var onboardingTask *model.Task

func (m *Manager) CheckOnboardingTask(ctx context.Context, address string) error {
	if onboardingTask == nil {
		t, err := m.taskMgr.GetOnboardingTask(ctx)
		if err != nil {
			return err
		}

		onboardingTask = &t
	}

	fmt.Println("in check")
	notExist := false
	userTask, err := m.getUserTask(ctx, address, onboardingTask.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			notExist = true
			userTask = model.UserTask{
				ID:          utils.GenDBID(),
				UserAddress: address,
				TaskID:      onboardingTask.ID,
				State:       "pending",
			}
		} else {
			return fmt.Errorf("failed to query onboarding usertask: %v", err)
		}
	}

	if userTask.State == "completed" {
		return nil
	}

	amount, err := m.transactionMgr.GetUserUSDC(ctx, address)
	if err != nil {
		return fmt.Errorf("failed to GetUserUSDC: %v", err)
	}

	threshold := decimal.NewFromFloat(1000.00) // 1000 USDC
	if amount.GreaterThanOrEqual(threshold) {
		userTask.State = "completed"
	}

	if notExist {
		query := `
			INSERT INTO "userTask" ("id", "userAddress", "taskId", "state")
			VALUES ($1, $2, $3, $4);
		`

		row := m.db.QueryRowContext(ctx, query, userTask.ID, userTask.UserAddress, userTask.TaskID, userTask.State)
		if row.Err() != nil {
			return fmt.Errorf("failed to create user task: %v", err)
		}
	} else {
		query := `
			UPDATE "userTask" SET "state" = $1 WHERE "id" = $2;
		`

		_, err := m.db.ExecContext(ctx, query, userTask.State, userTask.ID)
		if err != nil {
			return fmt.Errorf("failed to update user task: %v", err)
		}
	}

	fmt.Println("out check")

	return nil
}

func (m *Manager) CheckSharePoolTasks(ctx context.Context) error {
	tasks, getSharePoolErr := m.taskMgr.GetSharePoolTask(ctx)
	if getSharePoolErr != nil {
		return getSharePoolErr
	}

	for _, task := range tasks {
		err := m.checkSharePoolTask(ctx, task)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) checkSharePoolTask(ctx context.Context, task model.Task) error {
	startTime := task.StartAt
	senderPoints := make(map[string]decimal.Decimal)
	senderAmounts := make(map[string]decimal.Decimal)
	state := "pending"

	for week := 1; week <= 4; week++ {
		endTime := utils.GetLastTimeOfWeek(startTime)
		if time.Now().Before(endTime) {
			continue
		}
		if week == 4 {
			state = "completed"
		}

		rows, err := m.db.Query(`
			 SELECT sender, SUM(amount0In) AS totalAmount0In, SUM(amount1In) AS totalAmount1In
			 FROM transaction
			 WHERE timestamp >= $1 AND timestamp < $2
			 GROUP BY sender
		`, startTime, endTime)
		if err != nil {
			return fmt.Errorf("checkSharePoolTask query sum transaction: %v", err)
		}
		defer rows.Close()

		totalVolumeUSD := decimal.NewFromFloat(0)
		senderVolumes := make(map[string]decimal.Decimal)

		for rows.Next() {
			var sender string
			var totalAmount0In, totalAmount1In decimal.Decimal

			err := rows.Scan(&sender, &totalAmount0In, &totalAmount1In)
			if err != nil {
				return fmt.Errorf("CheckSharePoolTask scan error: %v", err)
			}

			// To USD
			totalAmount0InUSD := totalAmount0In.Div(constants.UsdcPrecision).Mul(constants.UsdcPrice)
			totalAmount1InUSD := totalAmount1In.Div(constants.EthPrecision).Mul(constants.EthPrice)

			totalAmountUSD := totalAmount0InUSD.Add(totalAmount1InUSD)
			senderVolumes[sender] = totalAmountUSD

			// sum all amount
			totalVolumeUSD = totalVolumeUSD.Add(totalAmountUSD)
		}

		for sender, volume := range senderVolumes {
			if !totalVolumeUSD.IsZero() {
				proportion := volume.Div(totalVolumeUSD)
				points := proportion.Mul(constants.PointsPerWeek)
				senderPoints[sender] = senderPoints[sender].Add(points)
				senderAmounts[sender] = senderAmounts[sender].Add(volume)
			} else {
				senderPoints[sender] = decimal.NewFromInt(0)
				senderAmounts[sender] = decimal.NewFromInt(0)
			}
		}

		// to next week
		startTime = endTime.AddDate(0, 0, 1).Truncate(24 * time.Hour)
	}

	// save point to
	for sender, points := range senderPoints {
		// FIXME: replace decimal.NewFromInt(0)
		if err := m.upsert(ctx, sender, task.ID, state, decimal.NewFromInt(0)); err != nil {
			log.Printf("checkSharePoolTask upsert user task fail: %v", err)
			continue
		}

		if err := m.userPointMgr.UpsertForUserTask(ctx, sender, task.ID, int(points.IntPart())); err != nil {
			log.Printf("checkSharePoolTask upsert point fail: %v", err)
			continue
		}
	}

	return nil
}

func (m *Manager) upsert(ctx context.Context, address string, taskId string, state string, amount decimal.Decimal) error {
	query := `
		INSERT INTO "userTask" ("id", "userAddress", "taskId", "state", "createdAt", "amount")
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT ("userAddress", "taskId")
		DO UPDATE SET "state" = EXCLUDED."state", "amount" = EXCLUDED."amount"
	`

	_, err := m.db.ExecContext(ctx, query, utils.GenDBID(), address, taskId, state, time.Now(), amount)
	if err != nil {
		return fmt.Errorf("failed to upsert user task: %w", err)
	}

	return nil
}

func (m *Manager) getUserTask(ctx context.Context, address string, taskId string) (model.UserTask, error) {
	query := `
		SELECT "id", "createdAt", "userAddress", "taskId", "state"
		FROM "userTask"
		WHERE "userAddress" = $1 AND "taskId" = $2;
	`

	var userTask model.UserTask
	err := m.db.QueryRowContext(ctx, query, address, taskId).Scan(
		&userTask.ID,
		&userTask.CreatedAt,
		&userTask.UserAddress,
		&userTask.TaskID,
		&userTask.State,
	)

	return userTask, err
}
