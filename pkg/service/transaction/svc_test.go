package transaction

import (
	"testing"
	"tradingAce/internal/testutils"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func Test_NewManager(t *testing.T) {
	godotenv.Load("../../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	manager := NewManager(d)
	mgr := manager.(*Manager)

	assert.Equal(t, d, mgr.db)
}
