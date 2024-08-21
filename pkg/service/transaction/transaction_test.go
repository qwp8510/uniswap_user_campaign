package transaction

import (
	"context"
	"testing"
	"tradingAce/internal/testutils"
	"tradingAce/pkg/model"
	"tradingAce/pkg/model/option"

	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestManager_Upsert(t *testing.T) {
	godotenv.Load("../../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	type args struct {
		ctx context.Context
		opt option.TransactionUpsertOptions
	}
	tests := []struct {
		name string
		args args
		want model.Transaction
	}{
		{
			name: "upsert transaction not exist",
			args: args{
				ctx: context.TODO(),
				opt: option.TransactionUpsertOptions{
					BlockNum:        1,
					PairAddress:     "0x0000000000000000000000000000000000000000",
					SenderAddress:   "0x0000000000000000000000000000000000000000",
					Amount0In:       decimal.NewFromInt(10),
					Amount1In:       decimal.NewFromInt(20),
					Amount0Out:      decimal.NewFromInt(30),
					Amount1Out:      decimal.NewFromInt(40),
					ReceiverAddress: "0x0000000000000000000000000000000000000000",
				},
			},
			want: model.Transaction{
				BlockNum:        1,
				PairAddress:     "0x0000000000000000000000000000000000000000",
				SenderAddress:   "0x0000000000000000000000000000000000000000",
				Amount0In:       decimal.NewFromInt(10),
				Amount1In:       decimal.NewFromInt(20),
				Amount0Out:      decimal.NewFromInt(30),
				Amount1Out:      decimal.NewFromInt(40),
				ReceiverAddress: "0x0000000000000000000000000000000000000000",
			},
		},
		{
			name: "upsert transaction if exist",
			args: args{
				ctx: context.TODO(),
				opt: option.TransactionUpsertOptions{
					BlockNum:        1,
					PairAddress:     "0x0000000000000000000000000000000000000000",
					SenderAddress:   "0x0000000000000000000000000000000000000000",
					Amount0In:       decimal.NewFromInt(44),
					Amount1In:       decimal.NewFromInt(55),
					Amount0Out:      decimal.NewFromInt(66),
					Amount1Out:      decimal.NewFromInt(77),
					ReceiverAddress: "0x0000000000000000000000000000000000000000",
				},
			},
			want: model.Transaction{
				BlockNum:        1,
				PairAddress:     "0x0000000000000000000000000000000000000000",
				SenderAddress:   "0x0000000000000000000000000000000000000000",
				Amount0In:       decimal.NewFromInt(44),
				Amount1In:       decimal.NewFromInt(55),
				Amount0Out:      decimal.NewFromInt(66),
				Amount1Out:      decimal.NewFromInt(77),
				ReceiverAddress: "0x0000000000000000000000000000000000000000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := Manager{db: d}

			if err := mgr.Upsert(tt.args.ctx, tt.args.opt); err != nil {
				t.Errorf("Upsert() error = %v", err)
			}

			var result model.Transaction
			if err := d.QueryRow(`
				SELECT "blockNum", "pairAddress", "senderAddress", "amount0In", "amount1In", "amount0Out", "amount1Out", "receiverAddress"
				FROM transaction 
				WHERE "pairAddress" = $1 and "blockNum" = $2`,
				tt.args.opt.PairAddress,
				tt.args.opt.BlockNum,
			).Scan(
				&result.BlockNum,
				&result.PairAddress,
				&result.SenderAddress,
				&result.Amount0In,
				&result.Amount1In,
				&result.Amount0Out,
				&result.Amount1Out,
				&result.ReceiverAddress,
			); err != nil {
				t.Errorf("Upsert() query error = %v", err)
			}

			assert.Equal(t, tt.want.BlockNum, result.BlockNum)
			assert.Equal(t, tt.want.PairAddress, result.PairAddress)
			assert.Equal(t, tt.want.SenderAddress, result.SenderAddress)
			assert.True(t, tt.want.Amount0In.Equal(result.Amount0In))
			assert.True(t, tt.want.Amount1In.Equal(result.Amount1In))
			assert.True(t, tt.want.Amount0Out.Equal(result.Amount0Out))
			assert.True(t, tt.want.Amount1Out.Equal(result.Amount1Out))
			assert.Equal(t, tt.want.ReceiverAddress, result.ReceiverAddress)
		})
	}
}

func TestManager_GetUserUSDC(t *testing.T) {
	godotenv.Load("../../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	mgr := Manager{db: d}

	// init data
	opt1 := option.TransactionUpsertOptions{
		BlockNum:        1,
		PairAddress:     "0x0000000000000000000000000000000000000000",
		SenderAddress:   "0x0000000000000000000000000000000000000111",
		Amount0In:       decimal.NewFromInt(100),
		Amount1In:       decimal.NewFromInt(55),
		Amount0Out:      decimal.NewFromInt(66),
		Amount1Out:      decimal.NewFromInt(77),
		ReceiverAddress: "0x0000000000000000000000000000000000000000",
	}
	if err := mgr.Upsert(context.TODO(), opt1); err != nil {
		t.Errorf("GetUserUSDC() error = %v", err)
	}
	opt2 := option.TransactionUpsertOptions{
		BlockNum:        2,
		PairAddress:     "0x0000000000000000000000000000000000000000",
		SenderAddress:   "0x0000000000000000000000000000000000000111",
		Amount0In:       decimal.NewFromInt(200),
		Amount1In:       decimal.NewFromInt(55),
		Amount0Out:      decimal.NewFromInt(66),
		Amount1Out:      decimal.NewFromInt(77),
		ReceiverAddress: "0x0000000000000000000000000000000000000000",
	}
	if err := mgr.Upsert(context.TODO(), opt2); err != nil {
		t.Errorf("GetUserUSDC() error = %v", err)
	}

	result, err := mgr.GetUserUSDC(context.TODO(), "0x0000000000000000000000000000000000000111")
	if err != nil {
		t.Errorf("GetUserUSDC() error = %v", err)
	}

	assert.True(t, decimal.NewFromInt(300).Equal(result))
}
