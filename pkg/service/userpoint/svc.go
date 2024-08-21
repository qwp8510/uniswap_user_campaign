package userpoint

import (
	"database/sql"
	iface "tradingAce/pkg/interface"
)

func NewManager(db *sql.DB) iface.UserPointManager {
	return &Manager{
		db,
	}
}
