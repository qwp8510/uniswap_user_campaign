package transaction

import (
	"database/sql"
	iface "tradingAce/pkg/interface"
)

func NewManager(db *sql.DB) iface.TransactionManager {
	return &Manager{
		db,
	}
}
