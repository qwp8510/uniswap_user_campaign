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
	"tradingAce/pkg/model/option"
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

	threshold := decimal.NewFromFloat(1000.00).Mul(constants.UsdcPrecision) // 1000 USDC
	if amount.GreaterThanOrEqual(threshold) {
		userTask.State = "completed"
	}

	if notExist {
		query := `
			INSERT INTO "userTask" ("id", "userAddress", "taskId", "state")
			VALUES ($1, $2, $3, $4);
		`

		_, err := m.db.ExecContext(ctx, query, userTask.ID, userTask.UserAddress, userTask.TaskID, userTask.State)
		if err != nil {
			return fmt.Errorf("failed to create user task: %v", err)
		}
		if userTask.State == "completed" {
			if err := m.userPointMgr.UpsertForUserTask(ctx, userTask.UserAddress, onboardingTask.ID, constants.OnboardingPoint); err != nil {
				log.Printf("checkSharePoolTask upsert point fail: %v", err)
			}
		}
	} else {
		query := `
			UPDATE "userTask" SET "state" = $1 WHERE "id" = $2;
		`

		_, err := m.db.ExecContext(ctx, query, userTask.State, userTask.ID)
		if err != nil {
			return fmt.Errorf("failed to update user task: %v", err)
		}
		if userTask.State == "completed" {
			if err := m.userPointMgr.UpsertForUserTask(ctx, userTask.UserAddress, onboardingTask.ID, constants.OnboardingPoint); err != nil {
				log.Printf("checkSharePoolTask upsert point fail: %v", err)
			}
		}
	}

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

func (m *Manager) GetUserTasks(ctx context.Context, address string) ([]option.GetUserTaskPoint, error) {
	query := `
SELECT 
    ut."id" AS "id",
    ut."createdAt" AS "createdAt",
    ut."userAddress" AS "userAddress",
    ut."taskId" AS "taskID",
    ut."state" AS "state",
    ut."amount" AS "amount",
    up."point" AS "point",
    t."name" AS "taskName",
    t."pairAddress" AS "pairAddress"
FROM 
    "userTask" ut
JOIN 
    "userPoint" up 
    ON ut."taskId" = up."taskId" 
    AND ut."userAddress" = up."userAddress"
JOIN 
    "task" t
    ON ut."taskId" = t."id"
WHERE 
    ut."userAddress" = $1;
	`

	result := make([]option.GetUserTaskPoint, 0)

	stmt, err := m.db.Prepare(query)
	if err != nil {
		return result, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(address)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var data option.GetUserTaskPoint
		rows.Scan(
			&data.ID,
			&data.CreatedAt,
			&data.UserAddress,
			&data.TaskID,
			&data.State,
			&data.Amount,
			&data.Point,
			&data.TaskName,
			&data.PairAddress,
		)

		result = append(result, data)
	}

	return result, nil
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
			SELECT t."senderAddress" AS "senderAddress", 
				SUM(t."amount0In") AS "totalAmount0In", 
				SUM(t."amount1In") AS "totalAmount1In"
			FROM transaction t
			JOIN "userTask" ut 
				ON t."senderAddress" = ut."userAddress"
			WHERE t."transactionAt" >= $1 
				AND t."transactionAt" < $2
				AND t."pairAddress" = $3
				AND ut."taskId" = $4
				AND ut.state = 'completed'
			GROUP BY t."senderAddress";
		`, startTime, endTime, task.PairAddress, onboardingTask.ID)
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
				fmt.Println(sender, volume, totalVolumeUSD)
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
		if err := m.Upsert(ctx, sender, task.ID, state, senderAmounts[sender]); err != nil {
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

func (m *Manager) Upsert(ctx context.Context, address string, taskId string, state string, amount decimal.Decimal) error {
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
		SELECT "id", "createdAt", "userAddress", "taskId", "state", "amount"
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
		&userTask.Amount,
	)

	return userTask, err
}
