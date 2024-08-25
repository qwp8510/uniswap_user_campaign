package usertask

import (
	"context"
	"database/sql"
	"testing"
	"time"
	"tradingAce/internal/testutils"
	"tradingAce/pkg/constants"
	"tradingAce/pkg/model"
	"tradingAce/pkg/model/option"
	"tradingAce/pkg/service/task"
	"tradingAce/pkg/service/transaction"
	"tradingAce/pkg/service/userpoint"

	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func setOnbardingTask() *model.Task {
	onboardingTask = &model.Task{
		ID:          "onboardingtask",
		CreatedAt:   time.Now(),
		Name:        sql.NullString{String: "onboarding", Valid: true},
		PairAddress: sql.NullString{},
		StartAt:     time.Now(),
	}

	return onboardingTask
}

func TestManager_upsert(t *testing.T) {
	godotenv.Load("../../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	type args struct {
		ctx     context.Context
		address string
		taskId  string
		state   string
		amount  decimal.Decimal
	}
	tests := []struct {
		name string
		args args
		want model.UserTask
	}{
		{
			name: "upsert not exist usertask",
			args: args{
				ctx:     context.TODO(),
				address: "0x0000000000000000000000000000000000000000",
				taskId:  "task1",
				state:   "pending",
				amount:  decimal.NewFromInt(10),
			},
			want: model.UserTask{
				UserAddress: "0x0000000000000000000000000000000000000000",
				TaskID:      "task1",
				State:       "pending",
				Amount:      decimal.NewFromInt(10),
			},
		},
		{
			name: "upsert not exist usertask",
			args: args{
				ctx:     context.TODO(),
				address: "0x0000000000000000000000000000000000000000",
				taskId:  "task1",
				state:   "completed",
				amount:  decimal.NewFromInt(20),
			},
			want: model.UserTask{
				UserAddress: "0x0000000000000000000000000000000000000000",
				TaskID:      "task1",
				State:       "completed",
				Amount:      decimal.NewFromInt(20),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := Manager{db: d}
			if err := mgr.Upsert(context.TODO(), tt.args.address, tt.args.taskId, tt.args.state, tt.args.amount); err != nil {
				t.Errorf("Upsert err: %v", err)
				return
			}

			var result model.UserTask
			if err := d.QueryRow(
				`SELECT "userAddress", "taskId", "state", "amount" FROM "userTask" WHERE "userAddress" = $1 AND "taskId" = $2;`,
				tt.args.address, tt.args.taskId,
			).Scan(
				&result.UserAddress,
				&result.TaskID,
				&result.State,
				&result.Amount,
			); err != nil {
				t.Errorf("scan err: %v", err)
				return
			}
			assert.True(
				t, tt.want.Amount.Equal(result.Amount), "amount not equal want: %v, got: %v", tt.want.Amount, result.Amount,
			)
			assert.Equal(t, tt.want.UserAddress, result.UserAddress)
			assert.Equal(t, tt.want.TaskID, result.TaskID)
			assert.Equal(t, tt.want.State, result.State)
		})
	}
}

func TestManager_getUserTask(t *testing.T) {
	godotenv.Load("../../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	mgr := Manager{db: d}
	if err := mgr.Upsert(context.TODO(), "0x0000000000000000000000000000000000000000", "task1", "pending", decimal.NewFromInt(10)); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}

	model, err := mgr.getUserTask(context.TODO(), "0x0000000000000000000000000000000000000000", "task1")
	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, "task1", model.TaskID)
	assert.Equal(t, "pending", model.State)
	assert.True(t, decimal.NewFromInt(10).Equal(model.Amount))
}

func TestManager_checkSharePoolTask(t *testing.T) {
	godotenv.Load("../../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	ctx := context.TODO()

	trMgr := transaction.NewManager(d)
	mgr := Manager{
		db:             d,
		taskMgr:        task.NewManager(d),
		transactionMgr: trMgr,
		userPointMgr:   userpoint.NewManager(d),
	}

	sender1 := "0x0000000000000000000000000000000000000000"
	sender2 := "0x0000000000000000000000000000000000000001"
	senderNoOnboarding := "0x0000000000000000000000000000000000000002"

	// init test transaction data
	transactionAt1, parseErr := time.Parse("2006-01-02", "2024-07-02")
	if parseErr != nil {
		t.Errorf("parse time err: %v", parseErr)
		return
	}
	transactionAt2, parseErr := time.Parse("2006-01-02", "2024-07-12")
	if parseErr != nil {
		t.Errorf("parse time err: %v", parseErr)
		return
	}
	transactionAtOutOfRange, parseErr := time.Parse("2006-01-02", "2024-06-12")
	if parseErr != nil {
		t.Errorf("parse time err: %v", parseErr)
		return
	}

	if err := trMgr.Upsert(ctx, option.TransactionUpsertOptions{
		BlockNum:        1,
		PairAddress:     "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc",
		SenderAddress:   sender1,
		Amount0In:       constants.UsdcPrecision.Mul(decimal.NewFromInt(1000)),
		Amount1In:       constants.EthPrecision.Mul(decimal.NewFromInt(50)),
		Amount0Out:      decimal.NewFromInt(30),
		Amount1Out:      decimal.NewFromInt(40),
		ReceiverAddress: "0x0000000000000000000000000000000000000000",
		TransactionAt:   transactionAt1,
	}); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}
	if err := trMgr.Upsert(ctx, option.TransactionUpsertOptions{
		BlockNum:        2,
		PairAddress:     "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc",
		SenderAddress:   sender1,
		Amount0In:       constants.UsdcPrecision.Mul(decimal.NewFromInt(1000)),
		Amount1In:       constants.EthPrecision.Mul(decimal.NewFromInt(0)),
		Amount0Out:      decimal.NewFromInt(30),
		Amount1Out:      decimal.NewFromInt(40),
		ReceiverAddress: "0x0000000000000000000000000000000000000000",
		TransactionAt:   transactionAt2,
	}); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}
	if err := trMgr.Upsert(ctx, option.TransactionUpsertOptions{
		BlockNum:        3,
		PairAddress:     "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc",
		SenderAddress:   sender2,
		Amount0In:       constants.UsdcPrecision.Mul(decimal.NewFromInt(1000)),
		Amount1In:       constants.EthPrecision.Mul(decimal.NewFromInt(10)),
		Amount0Out:      decimal.NewFromInt(30),
		Amount1Out:      decimal.NewFromInt(40),
		ReceiverAddress: "0x0000000000000000000000000000000000000000",
		TransactionAt:   transactionAt1,
	}); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}
	if err := trMgr.Upsert(ctx, option.TransactionUpsertOptions{
		BlockNum:        3,
		PairAddress:     "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc",
		SenderAddress:   sender2,
		Amount0In:       constants.UsdcPrecision.Mul(decimal.NewFromInt(4000)),
		Amount1In:       constants.EthPrecision.Mul(decimal.NewFromInt(10)),
		Amount0Out:      decimal.NewFromInt(30),
		Amount1Out:      decimal.NewFromInt(40),
		ReceiverAddress: "0x0000000000000000000000000000000000000000",
		TransactionAt:   transactionAt1,
	}); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}
	// diffrent PairAddress
	if err := trMgr.Upsert(ctx, option.TransactionUpsertOptions{
		BlockNum:        4,
		PairAddress:     "0xnotReelAddress",
		SenderAddress:   sender2,
		Amount0In:       constants.UsdcPrecision.Mul(decimal.NewFromInt(80000)),
		Amount1In:       constants.EthPrecision.Mul(decimal.NewFromInt(0)),
		Amount0Out:      decimal.NewFromInt(30),
		Amount1Out:      decimal.NewFromInt(40),
		ReceiverAddress: "0x0000000000000000000000000000000000000000",
		TransactionAt:   transactionAt1,
	}); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}
	// out of transaction range
	if err := trMgr.Upsert(ctx, option.TransactionUpsertOptions{
		BlockNum:        40,
		PairAddress:     "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc",
		SenderAddress:   sender2,
		Amount0In:       constants.UsdcPrecision.Mul(decimal.NewFromInt(80000)),
		Amount1In:       constants.EthPrecision.Mul(decimal.NewFromInt(0)),
		Amount0Out:      decimal.NewFromInt(30),
		Amount1Out:      decimal.NewFromInt(40),
		ReceiverAddress: "0x0000000000000000000000000000000000000000",
		TransactionAt:   transactionAtOutOfRange,
	}); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}
	if err := trMgr.Upsert(ctx, option.TransactionUpsertOptions{
		BlockNum:        9999,
		PairAddress:     "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc",
		SenderAddress:   senderNoOnboarding,
		Amount0In:       constants.UsdcPrecision.Mul(decimal.NewFromInt(99999)),
		Amount1In:       constants.EthPrecision.Mul(decimal.NewFromInt(99999)),
		Amount0Out:      decimal.NewFromInt(30),
		Amount1Out:      decimal.NewFromInt(40),
		ReceiverAddress: "0x0000000000000000000000000000000000000000",
		TransactionAt:   transactionAt1,
	}); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}

	// init sender onboarding user task
	onboardingTask := setOnbardingTask()
	if err := mgr.Upsert(ctx, sender1, onboardingTask.ID, "completed", decimal.NewFromInt(1000)); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}
	if err := mgr.Upsert(ctx, sender2, onboardingTask.ID, "completed", decimal.NewFromInt(1030)); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}

	startAt, parseErr := time.Parse("2006-01-02", "2024-07-01")
	if parseErr != nil {
		t.Errorf("Parse err: %v", err)
		return
	}

	sharePoolTask := model.Task{
		ID:        "sharePoolTask",
		CreatedAt: time.Now(),
		Name:      sql.NullString{String: "share_pool", Valid: true},
		PairAddress: sql.NullString{
			String: "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc",
			Valid:  true,
		},
		StartAt: startAt,
	}
	if err := mgr.checkSharePoolTask(ctx, sharePoolTask); err != nil {
		t.Errorf("checkSharePoolTask err: %v", err)
	}

	ut1, ut1Err := mgr.getUserTask(ctx, sender1, sharePoolTask.ID)
	if ut1Err != nil {
		t.Errorf("getUserTask 1 err: %v", ut1Err)
		return
	}
	assert.Equal(t, sharePoolTask.ID, ut1.TaskID)
	assert.Equal(t, "completed", ut1.State)
	var result1 model.UserPoint
	if err := d.QueryRow(
		`SELECT "userAddress", "taskId", "point" FROM "userPoint" 
		WHERE "userAddress"=$1 AND "taskId"=$2`,
		sender1, sharePoolTask.ID,
	).Scan(
		&result1.UserAddress,
		&result1.TaskID,
		&result1.Point,
	); err != nil {
		t.Errorf("get user point query error = %v", err)
		return
	}
	assert.Equal(t, sender1, result1.UserAddress)
	assert.Equal(t, sharePoolTask.ID, result1.TaskID)
	assert.Equal(t, 18080, result1.Point)

	ut2, ut2Err := mgr.getUserTask(ctx, sender2, sharePoolTask.ID)
	if ut2Err != nil {
		t.Errorf("getUserTask 2 err: %v", ut2Err)
		return
	}
	assert.Equal(t, sharePoolTask.ID, ut2.TaskID)
	assert.Equal(t, "completed", ut2.State)
	var result2 model.UserPoint
	if err := d.QueryRow(
		`SELECT "userAddress", "taskId", "point" FROM "userPoint" 
		WHERE "userAddress"=$1 AND "taskId"=$2`,
		sender2, sharePoolTask.ID,
	).Scan(
		&result2.UserAddress,
		&result2.TaskID,
		&result2.Point,
	); err != nil {
		t.Errorf("get user point query error = %v", err)
		return
	}
	assert.Equal(t, sender2, result2.UserAddress)
	assert.Equal(t, sharePoolTask.ID, result2.TaskID)
	assert.Equal(t, 1920, result2.Point)

	_, ut3Err := mgr.getUserTask(ctx, senderNoOnboarding, sharePoolTask.ID)
	assert.EqualError(t, ut3Err, sql.ErrNoRows.Error())
}

func TestManager_checkSharePoolTaskNotFinished(t *testing.T) {
	godotenv.Load("../../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	ctx := context.TODO()

	trMgr := transaction.NewManager(d)
	mgr := Manager{
		db:             d,
		taskMgr:        task.NewManager(d),
		transactionMgr: trMgr,
		userPointMgr:   userpoint.NewManager(d),
	}

	sender1 := "0x0000000000000000000000000000000000000000"

	now := time.Now()
	transactionAt1 := now.AddDate(0, 0, -10)
	twoWeeksAgo := now.AddDate(0, 0, -14)

	if err := trMgr.Upsert(ctx, option.TransactionUpsertOptions{
		BlockNum:        1,
		PairAddress:     "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc",
		SenderAddress:   sender1,
		Amount0In:       constants.UsdcPrecision.Mul(decimal.NewFromInt(1000)),
		Amount1In:       constants.EthPrecision.Mul(decimal.NewFromInt(50)),
		Amount0Out:      decimal.NewFromInt(30),
		Amount1Out:      decimal.NewFromInt(40),
		ReceiverAddress: "0x0000000000000000000000000000000000000000",
		TransactionAt:   transactionAt1,
	}); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}

	// init sender onboarding user task
	onboardingTask := setOnbardingTask()
	if err := mgr.Upsert(ctx, sender1, onboardingTask.ID, "completed", decimal.NewFromInt(1000)); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}

	sharePoolTask := model.Task{
		ID:        "checkSharePoolTaskNotFinished",
		CreatedAt: time.Now(),
		Name:      sql.NullString{String: "share_pool", Valid: true},
		PairAddress: sql.NullString{
			String: "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc",
			Valid:  true,
		},
		StartAt: twoWeeksAgo,
	}

	if err := mgr.checkSharePoolTask(ctx, sharePoolTask); err != nil {
		t.Errorf("checkSharePoolTask err: %v", err)
		return
	}

	ut1, ut1Err := mgr.getUserTask(ctx, sender1, sharePoolTask.ID)
	if ut1Err != nil {
		t.Errorf("getUserTask 1 err: %v", ut1Err)
		return
	}
	assert.Equal(t, sharePoolTask.ID, ut1.TaskID)
	assert.Equal(t, "pending", ut1.State)
	var result1 model.UserPoint
	if err := d.QueryRow(
		`SELECT "userAddress", "taskId", "point" FROM "userPoint" 
		WHERE "userAddress"=$1 AND "taskId"=$2`,
		sender1, sharePoolTask.ID,
	).Scan(
		&result1.UserAddress,
		&result1.TaskID,
		&result1.Point,
	); err != nil {
		t.Errorf("get user point query error = %v", err)
		return
	}
	assert.Equal(t, sender1, result1.UserAddress)
	assert.Equal(t, sharePoolTask.ID, result1.TaskID)
	assert.Equal(t, 10000, result1.Point)
}

func TestManager_CheckOnboardingTask(t *testing.T) {
	godotenv.Load("../../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	ctx := context.TODO()

	trMgr := transaction.NewManager(d)
	mgr := Manager{
		db:             d,
		taskMgr:        task.NewManager(d),
		transactionMgr: trMgr,
		userPointMgr:   userpoint.NewManager(d),
	}

	sender1 := "0x0000000000000000000000000000000000000000"
	sender2 := "0x0000000000000000000000000000000000000001"
	senderNoOnboarding := "0xnononononon"

	// init transaction
	transactionAt1, parseErr := time.Parse("2006-01-02", "2024-07-02")
	if parseErr != nil {
		t.Errorf("parse time err: %v", parseErr)
		return
	}
	transactionAt2, parseErr := time.Parse("2006-01-02", "2024-07-12")
	if parseErr != nil {
		t.Errorf("parse time err: %v", parseErr)
		return
	}
	transactionAtOutOfRange, parseErr := time.Parse("2006-01-02", "2024-06-12")
	if parseErr != nil {
		t.Errorf("parse time err: %v", parseErr)
		return
	}

	if err := trMgr.Upsert(ctx, option.TransactionUpsertOptions{
		BlockNum:        1,
		PairAddress:     "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc",
		SenderAddress:   sender1,
		Amount0In:       constants.UsdcPrecision.Mul(decimal.NewFromInt(700)),
		Amount1In:       constants.EthPrecision.Mul(decimal.NewFromInt(50)),
		Amount0Out:      decimal.NewFromInt(30),
		Amount1Out:      decimal.NewFromInt(40),
		ReceiverAddress: "0x0000000000000000000000000000000000000000",
		TransactionAt:   transactionAt1,
	}); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}
	if err := trMgr.Upsert(ctx, option.TransactionUpsertOptions{
		BlockNum:        2,
		PairAddress:     "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc",
		SenderAddress:   sender1,
		Amount0In:       constants.UsdcPrecision.Mul(decimal.NewFromInt(400)),
		Amount1In:       constants.EthPrecision.Mul(decimal.NewFromInt(0)),
		Amount0Out:      decimal.NewFromInt(30),
		Amount1Out:      decimal.NewFromInt(40),
		ReceiverAddress: "0x0000000000000000000000000000000000000000",
		TransactionAt:   transactionAt2,
	}); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}
	if err := trMgr.Upsert(ctx, option.TransactionUpsertOptions{
		BlockNum:        3,
		PairAddress:     "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc",
		SenderAddress:   sender2,
		Amount0In:       constants.UsdcPrecision.Mul(decimal.NewFromInt(1000)),
		Amount1In:       constants.EthPrecision.Mul(decimal.NewFromInt(10)),
		Amount0Out:      decimal.NewFromInt(30),
		Amount1Out:      decimal.NewFromInt(40),
		ReceiverAddress: "0x0000000000000000000000000000000000000000",
		TransactionAt:   transactionAt1,
	}); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}
	if err := trMgr.Upsert(ctx, option.TransactionUpsertOptions{
		BlockNum:        999,
		PairAddress:     "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc",
		SenderAddress:   senderNoOnboarding,
		Amount0In:       constants.UsdcPrecision.Mul(decimal.NewFromInt(500)),
		Amount1In:       constants.EthPrecision.Mul(decimal.NewFromInt(1000)),
		Amount0Out:      decimal.NewFromInt(30),
		Amount1Out:      decimal.NewFromInt(40),
		ReceiverAddress: "0x0000000000000000000000000000000000000000",
		TransactionAt:   transactionAtOutOfRange,
	}); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}

	onboardingTask := setOnbardingTask()

	if err := mgr.CheckOnboardingTask(ctx, sender1); err != nil {
		t.Errorf("CheckOnboardingTask err: %v", err)
		return
	}
	ut1, ut1Err := mgr.getUserTask(ctx, sender1, onboardingTask.ID)
	if ut1Err != nil {
		t.Errorf("getUserTask 1 err: %v", ut1Err)
		return
	}
	assert.Equal(t, onboardingTask.ID, ut1.TaskID)
	assert.Equal(t, "completed", ut1.State)
	assert.True(t, decimal.NewFromInt(1100).Equal(ut1.Amount), "amount should be 1100")
	var result1 model.UserPoint
	if err := d.QueryRow(
		`SELECT "userAddress", "taskId", "point" FROM "userPoint" 
		WHERE "userAddress"=$1 AND "taskId"=$2`,
		sender1, onboardingTask.ID,
	).Scan(
		&result1.UserAddress,
		&result1.TaskID,
		&result1.Point,
	); err != nil {
		t.Errorf("get user point query error = %v", err)
		return
	}
	assert.Equal(t, sender1, result1.UserAddress)
	assert.Equal(t, onboardingTask.ID, result1.TaskID)
	assert.Equal(t, constants.OnboardingPoint, result1.Point)

	if err := mgr.CheckOnboardingTask(ctx, sender2); err != nil {
		t.Errorf("CheckOnboardingTask err: %v", err)
		return
	}
	ut2, ut2Err := mgr.getUserTask(ctx, sender2, onboardingTask.ID)
	if ut2Err != nil {
		t.Errorf("getUserTask 2 err: %v", ut2Err)
		return
	}
	assert.Equal(t, onboardingTask.ID, ut2.TaskID)
	assert.Equal(t, "completed", ut2.State)
	assert.True(t, decimal.NewFromInt(1000).Equal(ut2.Amount), "amount should be 1000")
	var result2 model.UserPoint
	if err := d.QueryRow(
		`SELECT "userAddress", "taskId", "point" FROM "userPoint" 
		WHERE "userAddress"=$1 AND "taskId"=$2`,
		sender2, onboardingTask.ID,
	).Scan(
		&result2.UserAddress,
		&result2.TaskID,
		&result2.Point,
	); err != nil {
		t.Errorf("get user point query error = %v", err)
		return
	}
	assert.Equal(t, sender2, result2.UserAddress)
	assert.Equal(t, onboardingTask.ID, result2.TaskID)
	assert.Equal(t, constants.OnboardingPoint, result2.Point)

	if err := mgr.CheckOnboardingTask(ctx, senderNoOnboarding); err != nil {
		t.Errorf("CheckOnboardingTask err: %v", err)
		return
	}
	ut3, ut3Err := mgr.getUserTask(ctx, senderNoOnboarding, onboardingTask.ID)
	if ut3Err != nil {
		t.Errorf("getUserTask 3 err: %v", ut3Err)
		return
	}
	assert.Equal(t, onboardingTask.ID, ut3.TaskID)
	assert.Equal(t, "pending", ut3.State)
	assert.True(t, decimal.NewFromInt(500).Equal(ut3.Amount), "amount should be 5000")
	var result3 model.UserPoint
	result3Err := d.QueryRow(
		`SELECT "userAddress", "taskId", "point" FROM "userPoint" 
		WHERE "userAddress"=$1 AND "taskId"=$2`,
		senderNoOnboarding, onboardingTask.ID,
	).Scan(
		&result3.UserAddress,
		&result3.TaskID,
		&result3.Point,
	)
	assert.EqualError(t, result3Err, sql.ErrNoRows.Error())
}

func TestManager_GetUserTasks(t *testing.T) {
	godotenv.Load("../../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	ctx := context.TODO()

	now := time.Now()
	data1 := option.GetUserTaskPoint{
		TaskID:      "onboardingId",
		State:       "completed",
		Amount:      decimal.NewFromInt(1200),
		CreatedAt:   now,
		UserAddress: "0x123",
		Point:       constants.OnboardingPoint,
		TaskName:    "onboarding",
	}

	mgr := Manager{db: d}
	userPointMgr := userpoint.NewManager(d)

	if _, err := d.Exec(
		`INSERT INTO task("id", "createdAt", "name", "pairAddress", "startAt") VALUES ($1, $2, $3, $4, $5) ON CONFLICT ("pairAddress") DO NOTHING;`,
		data1.TaskID, time.Now(), data1.TaskName, nil, now,
	); err != nil {
		t.Errorf("insert task err: %v", err)
		return
	}
	if err := mgr.Upsert(ctx, data1.UserAddress, data1.TaskID, data1.State, data1.Amount); err != nil {
		t.Errorf("Upsert() error = %v", err)
		return
	}
	if err := userPointMgr.UpsertForUserTask(ctx, data1.UserAddress, data1.TaskID, data1.Point); err != nil {
		t.Errorf("UpsertForUserTask() error = %v", err)
		return
	}

	type args struct {
		ctx     context.Context
		address string
	}
	tests := []struct {
		name string
		args args
		want option.GetUserTaskPoint
	}{
		{
			name: "query",
			args: args{
				ctx:     ctx,
				address: data1.UserAddress,
			},
			want: data1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mgr.GetUserTasks(tt.args.ctx, tt.args.address)
			if err != nil {
				t.Errorf("GetUserTasks() error = %v", err)
				return
			}

			assert.Equal(t, tt.want.TaskID, result[0].TaskID)
			assert.Equal(t, tt.want.State, result[0].State)
			assert.True(t, tt.want.Amount.Equal(result[0].Amount))
			assert.Equal(t, tt.want.UserAddress, result[0].UserAddress)
			assert.Equal(t, tt.want.Point, result[0].Point)
			assert.Equal(t, tt.want.TaskName, result[0].TaskName)
			assert.Equal(t, tt.want.PairAddress, result[0].PairAddress)
		})
	}
}

func TestManager_CheckOnboardingTaskNonExistTask(t *testing.T) {
	godotenv.Load("../../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	ctx := context.TODO()

	trMgr := transaction.NewManager(d)
	mgr := Manager{
		db:             d,
		taskMgr:        task.NewManager(d),
		transactionMgr: trMgr,
		userPointMgr:   userpoint.NewManager(d),
	}
	onboardingTask = nil
	err = mgr.CheckOnboardingTask(ctx, "0x123")
	assert.EqualError(t, err, sql.ErrNoRows.Error())
}

func TestManager_CheckOnboardingTaskUpdateExistUserTask(t *testing.T) {
	godotenv.Load("../../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	ctx := context.TODO()

	trMgr := transaction.NewManager(d)
	mgr := Manager{
		db:             d,
		taskMgr:        task.NewManager(d),
		transactionMgr: trMgr,
		userPointMgr:   userpoint.NewManager(d),
	}

	sender1 := "0x0000000000000000000000000000000000000000"

	// init transaction
	transactionAt1, parseErr := time.Parse("2006-01-02", "2024-07-02")
	if parseErr != nil {
		t.Errorf("parse time err: %v", parseErr)
		return
	}

	if err := trMgr.Upsert(ctx, option.TransactionUpsertOptions{
		BlockNum:        1,
		PairAddress:     "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc",
		SenderAddress:   sender1,
		Amount0In:       constants.UsdcPrecision.Mul(decimal.NewFromInt(700)),
		Amount1In:       constants.EthPrecision.Mul(decimal.NewFromInt(50)),
		Amount0Out:      decimal.NewFromInt(30),
		Amount1Out:      decimal.NewFromInt(40),
		ReceiverAddress: "0x0000000000000000000000000000000000000000",
		TransactionAt:   transactionAt1,
	}); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}
	if err := trMgr.Upsert(ctx, option.TransactionUpsertOptions{
		BlockNum:        2,
		PairAddress:     "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc",
		SenderAddress:   sender1,
		Amount0In:       constants.UsdcPrecision.Mul(decimal.NewFromInt(400)),
		Amount1In:       constants.EthPrecision.Mul(decimal.NewFromInt(0)),
		Amount0Out:      decimal.NewFromInt(30),
		Amount1Out:      decimal.NewFromInt(40),
		ReceiverAddress: "0x0000000000000000000000000000000000000000",
		TransactionAt:   transactionAt1,
	}); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}

	onboardingTask := setOnbardingTask()

	if err := mgr.Upsert(ctx, sender1, onboardingTask.ID, "pending", decimal.NewFromInt(800)); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}

	if err := mgr.CheckOnboardingTask(ctx, sender1); err != nil {
		t.Errorf("CheckOnboardingTask err: %v", err)
		return
	}
	ut1, ut1Err := mgr.getUserTask(ctx, sender1, onboardingTask.ID)
	if ut1Err != nil {
		t.Errorf("getUserTask 1 err: %v", ut1Err)
		return
	}
	assert.Equal(t, onboardingTask.ID, ut1.TaskID)
	assert.Equal(t, "completed", ut1.State)
	assert.True(t, decimal.NewFromInt(1100).Equal(ut1.Amount), "amount should be 1100")
	var result1 model.UserPoint
	if err := d.QueryRow(
		`SELECT "userAddress", "taskId", "point" FROM "userPoint" 
		WHERE "userAddress"=$1 AND "taskId"=$2`,
		sender1, onboardingTask.ID,
	).Scan(
		&result1.UserAddress,
		&result1.TaskID,
		&result1.Point,
	); err != nil {
		t.Errorf("get user point query error = %v", err)
		return
	}
	assert.Equal(t, sender1, result1.UserAddress)
	assert.Equal(t, onboardingTask.ID, result1.TaskID)
	assert.Equal(t, constants.OnboardingPoint, result1.Point)
}

func TestManager_CheckFinishedOnboardingTask(t *testing.T) {
	godotenv.Load("../../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	ctx := context.TODO()

	trMgr := transaction.NewManager(d)
	mgr := Manager{
		db:             d,
		taskMgr:        task.NewManager(d),
		transactionMgr: trMgr,
		userPointMgr:   userpoint.NewManager(d),
	}

	sender1 := "0x0000000000000000000000000000000000000000"

	onboardingTask := setOnbardingTask()

	if err := mgr.Upsert(ctx, sender1, onboardingTask.ID, "completed", decimal.NewFromInt(1000)); err != nil {
		t.Errorf("Upsert err: %v", err)
		return
	}

	if err := mgr.CheckOnboardingTask(ctx, sender1); err != nil {
		t.Errorf("CheckOnboardingTask err: %v", err)
		return
	}
	ut1, ut1Err := mgr.getUserTask(ctx, sender1, onboardingTask.ID)
	if ut1Err != nil {
		t.Errorf("getUserTask 1 err: %v", ut1Err)
		return
	}
	assert.Equal(t, onboardingTask.ID, ut1.TaskID)
	assert.Equal(t, "completed", ut1.State)
	assert.True(t, decimal.NewFromInt(1000).Equal(ut1.Amount), "amount should be 1000")
}
