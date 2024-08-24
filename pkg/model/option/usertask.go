package option

import (
	"time"

	"github.com/shopspring/decimal"
)

type GetUserTaskPoint struct {
	ID          string          `json:"id"`
	CreatedAt   time.Time       `json:"createdAt"`
	UserAddress string          `json:"userAddress"`
	TaskID      string          `json:"taskId"`
	State       string          `json:"state"`
	Amount      decimal.Decimal `json:"amount"`
	Point       int             `json:"point"`
	TaskName    string          `json:"taskName,omitempty"`
	PairAddress string          `json:"pairAddress,omitempty"`
}
