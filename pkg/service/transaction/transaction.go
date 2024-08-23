package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"tradingAce/pkg/model/option"
	"tradingAce/pkg/utils"

	"github.com/shopspring/decimal"
)

type Manager struct {
	db *sql.DB
}

func (m *Manager) Upsert(ctx context.Context, opt option.TransactionUpsertOptions) error {
	query := `
		INSERT INTO transaction ("id", "blockNum", "pairAddress", "senderAddress", "amount0In", "amount1In", 
        	"amount0Out", "amount1Out", "receiverAddress", "transactionAt")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT ("blockNum", "pairAddress") 
		DO UPDATE SET
			"senderAddress" = EXCLUDED."senderAddress",
			"amount0In" = EXCLUDED."amount0In",
			"amount1In" = EXCLUDED."amount1In",
			"amount0Out" = EXCLUDED."amount0Out",
			"amount1Out" = EXCLUDED."amount1Out",
			"receiverAddress" = EXCLUDED."receiverAddress",
			"transactionAt" = EXCLUDED."transactionAt"
	`

	fmt.Println("in upsert")
	_, err := m.db.ExecContext(
		ctx,
		query,
		utils.GenDBID(),
		opt.BlockNum,
		opt.PairAddress,
		opt.SenderAddress,
		opt.Amount0In,
		opt.Amount1In,
		opt.Amount0Out,
		opt.Amount1Out,
		opt.ReceiverAddress,
		opt.TransactionAt,
	)
	fmt.Println("out upsert")

	return err
}

func (m *Manager) GetUserUSDC(ctx context.Context, address string) (decimal.Decimal, error) {
	query := `
		SELECT SUM("amount0In") AS amount
		FROM transaction
		WHERE "senderAddress" = $1;
    `

	var totalAmount string
	err := m.db.QueryRow(query, address).Scan(&totalAmount)
	if err != nil {
		if err == sql.ErrNoRows {
			totalAmount = "0"
		}
		return decimal.Decimal{}, fmt.Errorf("failed to query total amount0In: %v", err)
	}

	decAmount, err := decimal.NewFromString(totalAmount)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("failed to convert amount0In to decimal: %v", err)
	}

	return decAmount, nil
}
