package usertask

import (
	"database/sql"
	iface "tradingAce/pkg/interface"
)

func NewManager(
	db *sql.DB,
	taskMgr iface.TaskManager,
	transactionMgr iface.TransactionManager,
	userPointMgr iface.UserPointManager,
) iface.UserTaskManager {

	return &Manager{
		db,
		taskMgr,
		transactionMgr,
		userPointMgr,
	}
}
