package task

import (
	"context"
	"testing"
	"time"
	"tradingAce/internal/testutils"
	"tradingAce/pkg/model"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

func TestManager_GetOnboardingTask(t *testing.T) {
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

func TestManager_GetSharePoolTask(t *testing.T) {
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

func TestManager_CreateSharePoolTask(t *testing.T) {
	godotenv.Load("../../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	startAt, parseErr := time.Parse("2006-01-02", "2024-07-02")
	if parseErr != nil {
		t.Errorf("parse time err: %v", parseErr)
		return
	}

	mgr := Manager{db: d}
	err = mgr.CreateSharePoolTask(context.Background(), "0xabc", startAt)
	if err != nil {
		t.Errorf("CreateSharePoolTask fail: %s", err)
		return
	}

	query := `
		SELECT "id", "createdAt", "name", "pairAddress", "startAt"
		FROM "task"
		WHERE "name" = $1 AND "pairAddress" = $2;
	`

	var task model.Task
	qErr := mgr.db.QueryRowContext(context.Background(), query, "share_pool", "0xabc").Scan(
		&task.ID,
		&task.CreatedAt,
		&task.Name,
		&task.PairAddress,
		&task.StartAt,
	)
	if qErr != nil {
		t.Error(qErr)
		return
	}
	assert.Equal(t, "share_pool", task.Name.String)
	assert.Equal(t, "0xabc", task.PairAddress.String)
	assert.True(t, startAt.Equal(task.StartAt))
}

func TestManager_CreateSharePoolTaskIfExist(t *testing.T) {
	godotenv.Load("../../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	startAt, parseErr := time.Parse("2006-01-02", "2024-07-02")
	if parseErr != nil {
		t.Errorf("parse time err: %v", parseErr)
		return
	}

	mgr := Manager{db: d}
	err = mgr.CreateSharePoolTask(context.Background(), "0xabc", startAt)
	if err != nil {
		t.Errorf("CreateSharePoolTask fail: %s", err)
		return
	}

	resultErr := mgr.CreateSharePoolTask(context.Background(), "0xabc", startAt)

	assert.True(t, resultErr != nil)
}
