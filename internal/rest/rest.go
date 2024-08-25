package rest

import (
	"context"
	"net/http"
	"time"
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

func (s *RestServer) CreateSharePoolTask(c *gin.Context) {
	type body struct {
		Address string `json:"address"`
		StartAt string `json:"startAt"`
	}
	ctx := context.Background()

	var b body
	if err := c.BindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startAt, parseErr := time.Parse("2006-01-02", b.StartAt)
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseErr.Error()})
		return
	}

	if err := s.TaskMgr.CreateSharePoolTask(ctx, b.Address, startAt); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "ok")
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
