package usertask

import (
	"testing"
	"tradingAce/internal/testutils"
	"tradingAce/pkg/service/task"
	"tradingAce/pkg/service/transaction"
	"tradingAce/pkg/service/userpoint"

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

	taskMgr := task.NewManager(d)
	transactionMgr := transaction.NewManager(d)
	userPointMgr := userpoint.NewManager(d)
	manager := NewManager(d, taskMgr, transactionMgr, userPointMgr)
	mgr := manager.(*Manager)

	assert.Equal(t, d, mgr.db)
	assert.Equal(t, taskMgr, mgr.taskMgr)
	assert.Equal(t, transactionMgr, mgr.transactionMgr)
	assert.Equal(t, userPointMgr, mgr.userPointMgr)
}
