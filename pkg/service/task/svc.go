package task

import (
	"database/sql"
	iface "tradingAce/pkg/interface"
)

func NewManager(db *sql.DB) iface.TaskManager {
	return &Manager{
		db,
	}
}
