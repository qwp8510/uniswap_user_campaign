package service

import (
	"database/sql"
	iface "tradingAce/pkg/interface"
	"tradingAce/pkg/service/task"
	"tradingAce/pkg/service/transaction"
	"tradingAce/pkg/service/userpoint"
	"tradingAce/pkg/service/usertask"
)

type Service struct {
	Task        iface.TaskManager
	Transaction iface.TransactionManager
	UserTask    iface.UserTaskManager
	UserPoint   iface.UserPointManager
}

func NewService(db *sql.DB) *Service {
	s := &Service{}

	s.Task = task.NewManager(db)
	s.Transaction = transaction.NewManager(db)
	s.UserPoint = userpoint.NewManager(db)
	s.UserTask = usertask.NewManager(db, s.Task, s.Transaction, s.UserPoint)

	return s
}
