package model

import (
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

type UserTask struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UserAddress string    `json:"userAddress"`
	TaskID      string    `json:"taskId"`
	State       string    `json:"state"`
}

type UserPoint struct {
	UserAddress string    `json:"userAddress"`
	CreatedAt   time.Time `json:"createdAt"`
	TaskID      string    `json:"taskId"`
	Point       int       `json:"point"`
}

type Task struct {
	ID          string         `json:"id"`
	CreatedAt   time.Time      `json:"createdAt"`
	Name        sql.NullString `json:"name"`
	PairAddress sql.NullString `json:"pairAddress"`
	StartAt     time.Time      `json:"startAt"`
}

type Transaction struct {
	ID              string          `json:"id"`
	BlockNum        uint64          `json:"blockNum"`
	PairAddress     string          `json:"pairAddress"`
	CreatedAt       time.Time       `json:"createdAt"`
	SenderAddress   string          `json:"senderAddress"`
	Amount0In       decimal.Decimal `json:"amount0In"`
	Amount1In       decimal.Decimal `json:"amount1In"`
	Amount0Out      decimal.Decimal `json:"amount0Out"`
	Amount1Out      decimal.Decimal `json:"amount1Out"`
	ReceiverAddress string          `json:"receiverAddress"`
}
