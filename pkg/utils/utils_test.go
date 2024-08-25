package utils

import (
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func Test_GenDBID(t *testing.T) {
	r1 := GenDBID()
	r2 := GenDBID()

	assert.True(t, !strings.Contains(r1, "-"))
	assert.True(t, r1 != r2)
}

func Test_GetLastTimeOfWeek(t *testing.T) {
	d, err := time.Parse("2006-01-02", "2024-07-02")
	if err != nil {
		t.Errorf("parse time error: %v", err)
		return
	}

	result := GetLastTimeOfWeek(d)

	assert.Equal(t, result.Year(), 2024)
	assert.Equal(t, result.Month(), time.July)
	assert.Equal(t, result.Day(), 7)
	assert.Equal(t, result.Hour(), 23)
	assert.Equal(t, result.Minute(), 59)
	assert.Equal(t, result.Second(), 59)
}

func Test_BigIntToDecimal(t *testing.T) {
	tests := []struct {
		name     string
		bigInt   *big.Int
		expected decimal.Decimal
		wantErr  bool
	}{
		{
			name:     "Positive integer",
			bigInt:   big.NewInt(123456789),
			expected: decimal.NewFromInt(123456789),
			wantErr:  false,
		},
		{
			name:     "Negative integer",
			bigInt:   big.NewInt(-987654321),
			expected: decimal.NewFromInt(-987654321),
			wantErr:  false,
		},
		{
			name:     "Zero",
			bigInt:   big.NewInt(0),
			expected: decimal.NewFromInt(0),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BigIntToDecimal(tt.bigInt)
			if (err != nil) != tt.wantErr {
				t.Errorf("BigIntToDecimal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !got.Equal(tt.expected) {
				t.Errorf("BigIntToDecimal() = %v, want %v", got, tt.expected)
			}
		})
	}
}
