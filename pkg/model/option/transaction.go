package option

import (
	"time"

	"github.com/shopspring/decimal"
)

type TransactionUpsertOptions struct {
	BlockNum        uint64
	PairAddress     string
	SenderAddress   string
	Amount0In       decimal.Decimal
	Amount1In       decimal.Decimal
	Amount0Out      decimal.Decimal
	Amount1Out      decimal.Decimal
	ReceiverAddress string
	TransactionAt   time.Time
}
