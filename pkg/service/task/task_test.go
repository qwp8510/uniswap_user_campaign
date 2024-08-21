package task

import (
	"context"
	"testing"
	"time"
	"tradingAce/internal/testutils"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

func Test_GetOnboardingTask(t *testing.T) {
	godotenv.Load("../../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	if _, err := d.Exec(
		`INSERT INTO task("id", "createdAt", "name", "pairAddress", "startAt")
		SELECT $1, $2, $3, $4, $5
		WHERE NOT EXISTS (SELECT 1 FROM task WHERE name = 'onboarding');`,
		"aaa", time.Now(), "onboarding", nil, "2006-01-02",
	); err != nil {
		t.Error(err)
		return
	}

	mgr := Manager{db: d}
	model, err := mgr.GetOnboardingTask(context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, "onboarding", model.Name.String)
}

func Test_GetSharePoolTask(t *testing.T) {
	godotenv.Load("../../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	if _, err := d.Exec(
		`INSERT INTO task("id", "createdAt", "name", "pairAddress", "startAt")
		SELECT $1, $2, $3, $4, $5
		WHERE NOT EXISTS (SELECT 1 FROM task WHERE name = 'onboarding');`,
		"aaa", time.Now(), "onboarding", nil, "2006-01-02",
	); err != nil {
		t.Error(err)
		return
	}

	if _, err := d.Exec(
		`INSERT INTO task("id", "createdAt", "name", "pairAddress", "startAt") VALUES ($1, $2, $3, $4, $5) ON CONFLICT ("pairAddress") DO NOTHING;`,
		"bbb", time.Now(), "share_pool", "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc", "2024-07-02",
	); err != nil {
		t.Error(err)
		return
	}
	if _, err := d.Exec(
		`INSERT INTO task("id", "createdAt", "name", "pairAddress", "startAt") VALUES ($1, $2, $3, $4, $5) ON CONFLICT ("pairAddress") DO NOTHING;`,
		"ccc", time.Now(), "share_pool", "0xhihihihhihihihi", "2024-08-02",
	); err != nil {
		t.Error(err)
		return
	}

	mgr := Manager{db: d}
	ids := []string{"bbb", "ccc"}
	models, err := mgr.GetSharePoolTask(context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	for _, task := range models {
		assert.True(t, slices.Contains(ids, task.ID))
	}
}
