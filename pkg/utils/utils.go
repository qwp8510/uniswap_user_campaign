package utils

import (
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func GenDBID() string {
	u := uuid.New()

	return strings.ReplaceAll(u.String(), "-", "")
}

func BigIntToDecimal(bi *big.Int) (decimal.Decimal, error) {
	biStr := bi.String()

	return decimal.NewFromString(biStr)
}

func GetLastTimeOfWeek(t time.Time) time.Time {
	offset := int(time.Sunday - t.Weekday())
	if offset < 0 {
		offset += 7
	}

	sunday := t.AddDate(0, 0, offset)

	// to 23:59:59
	lastMoment := time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 999999999, sunday.Location())

	return lastMoment
}
