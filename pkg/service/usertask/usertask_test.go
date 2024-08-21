package usertask

import (
	"context"
	"testing"
	"tradingAce/internal/testutils"
	"tradingAce/pkg/model"

	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

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
			if err := mgr.upsert(context.TODO(), tt.args.address, tt.args.taskId, tt.args.state, tt.args.amount); err != nil {
				t.Errorf("upsert err: %v", err)
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
	if err := mgr.upsert(context.TODO(), "0x0000000000000000000000000000000000000000", "task1", "pending", decimal.NewFromInt(10)); err != nil {
		t.Errorf("upsert err: %v", err)
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
