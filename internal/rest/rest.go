package rest

import (
	"context"
	"net/http"
	iface "tradingAce/pkg/interface"

	"github.com/gin-gonic/gin"
)

type RestServer struct {
	TaskMgr      iface.TaskManager
	UserPointMgr iface.UserPointManager
	UserTaskMgr  iface.UserTaskManager
}

func (s *RestServer) GetUserTasks(c *gin.Context) {
	ctx := context.Background()
	address := c.Param("address")

	result, err := s.UserTaskMgr.GetUserTasks(ctx, address)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}

	c.JSON(http.StatusOK, result)
}

func (s *RestServer) GetUserPoints(c *gin.Context) {
	ctx := context.Background()
	taskID := c.Param("taskID")

	result, err := s.UserPointMgr.GetUserPointsForTask(ctx, taskID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}

	c.JSON(http.StatusOK, result)
}

func NewRestServer(
	taskMgr iface.TaskManager,
	userPointMgr iface.UserPointManager,
	userTaskMgr iface.UserTaskManager,
) *RestServer {

	return &RestServer{
		TaskMgr:      taskMgr,
		UserPointMgr: userPointMgr,
		UserTaskMgr:  userTaskMgr,
	}
}
