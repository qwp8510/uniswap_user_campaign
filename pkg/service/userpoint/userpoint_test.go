package userpoint

import (
	"context"
	"testing"
	"tradingAce/internal/testutils"
	"tradingAce/pkg/model"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestManager_UpsertForUserTask(t *testing.T) {
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
		point   int
	}
	tests := []struct {
		name string
		args args
		want model.UserPoint
	}{
		{
			name: "upsert not exist userpoint",
			args: args{
				ctx:     context.TODO(),
				address: "0x0000000000000000000000000000000000000000",
				taskId:  "task1",
				point:   10,
			},
			want: model.UserPoint{
				UserAddress: "0x0000000000000000000000000000000000000000",
				TaskID:      "task1",
				Point:       10,
			},
		},
		{
			name: "upsert userpoint if exist",
			args: args{
				ctx:     context.TODO(),
				address: "0x0000000000000000000000000000000000000000",
				taskId:  "task1",
				point:   88,
			},
			want: model.UserPoint{
				UserAddress: "0x0000000000000000000000000000000000000000",
				TaskID:      "task1",
				Point:       88,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := Manager{db: d}

			if err := mgr.UpsertForUserTask(tt.args.ctx, tt.args.address, tt.args.taskId, tt.args.point); err != nil {
				t.Errorf("UpsertForUserTask() error = %v", err)
				return
			}

			var result model.UserPoint
			if err := d.QueryRow(
				`SELECT "userAddress", "taskId", "point" FROM "userPoint" 
				WHERE "userAddress"=$1 AND "taskId"=$2`,
				tt.args.address, tt.args.taskId,
			).Scan(
				&result.UserAddress,
				&result.TaskID,
				&result.Point,
			); err != nil {
				t.Errorf("UpsertForUserTask() query error = %v", err)
				return
			}

			assert.Equal(t, tt.want.UserAddress, result.UserAddress)
			assert.Equal(t, tt.want.TaskID, result.TaskID)
			assert.Equal(t, tt.want.Point, result.Point)
		})
	}
}
