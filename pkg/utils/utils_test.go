package utils

import (
	"strings"
	"testing"
	"time"

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
