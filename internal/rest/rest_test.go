package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"tradingAce/internal/testutils"
	"tradingAce/pkg/model"
	"tradingAce/pkg/model/option"
	"tradingAce/pkg/service/task"
	"tradingAce/pkg/service/transaction"
	"tradingAce/pkg/service/userpoint"
	"tradingAce/pkg/service/usertask"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func Test_GetUserTasks(t *testing.T) {
	godotenv.Load("../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	r := gin.Default()

	taskMgr := task.NewManager(d)
	userPointMgr := userpoint.NewManager(d)
	server := &RestServer{
		TaskMgr:      taskMgr,
		UserPointMgr: userPointMgr,
		UserTaskMgr:  usertask.NewManager(d, taskMgr, transaction.NewManager(d), userPointMgr),
	}

	// Register the endpoint
	r.GET("/userTasks/:address", server.GetUserTasks)

	tests := []struct {
		name       string
		address    string
		expected   []option.GetUserTaskPoint
		statusCode int
	}{
		{
			name:    "Valid request",
			address: "0x12345",
			expected: []option.GetUserTaskPoint{
				{
					UserAddress: "0x12345",
					State:       "completed",
					Amount:      decimal.NewFromInt(10),
					Point:       1,
					TaskName:    "share_pool",
					PairAddress: "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc",
				},
			},
			statusCode: http.StatusOK,
		},
	}

	ctx := context.TODO()

	if err := taskMgr.CreateSharePoolTask(ctx, "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc", time.Now()); err != nil {
		t.Errorf("create share pool task err: %v", err)
		return
	}
	tasks, err := taskMgr.GetSharePoolTask(ctx)
	if err != nil {
		t.Errorf("get share pool task err: %v", err)
		return
	}
	if err := server.UserTaskMgr.Upsert(ctx, "0x12345", tasks[0].ID, "completed", decimal.NewFromInt(10)); err != nil {
		t.Errorf("upsert user task err: %v", err)
		return
	}
	if err := userPointMgr.UpsertForUserTask(ctx, "0x12345", tasks[0].ID, 1); err != nil {
		t.Errorf("upsert user point err: %v", err)
		return
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new HTTP request
			req, err := http.NewRequest(http.MethodGet, "/userTasks/"+tt.address, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Create a response recorder
			w := httptest.NewRecorder()

			// Serve the request
			r.ServeHTTP(w, req)

			// Assert the status code
			assert.Equal(t, tt.statusCode, w.Code)

			// Assert the response body
			var responseBody []option.GetUserTaskPoint
			err = json.Unmarshal(w.Body.Bytes(), &responseBody)
			if err != nil {
				t.Fatalf("Failed to unmarshal response body: %v", err)
			}

			assert.Equal(t, tt.expected[0].UserAddress, responseBody[0].UserAddress)
			assert.Equal(t, tt.expected[0].State, responseBody[0].State)
			assert.Equal(t, tt.expected[0].Point, responseBody[0].Point)
			assert.Equal(t, tt.expected[0].TaskName, responseBody[0].TaskName)
			assert.Equal(t, tt.expected[0].PairAddress, responseBody[0].PairAddress)
			assert.True(t, tt.expected[0].Amount.Equal(responseBody[0].Amount))
		})
	}
}

func Test_GetUserPoints(t *testing.T) {
	godotenv.Load("../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	r := gin.Default()

	taskMgr := task.NewManager(d)
	userPointMgr := userpoint.NewManager(d)
	server := &RestServer{
		TaskMgr:      taskMgr,
		UserPointMgr: userPointMgr,
		UserTaskMgr:  usertask.NewManager(d, taskMgr, transaction.NewManager(d), userPointMgr),
	}

	// Register the endpoint
	r.GET("/userPoints/*taskId", server.GetUserPoints)

	tests := []struct {
		name       string
		taskID     string
		expected   []model.UserPoint
		statusCode int
	}{
		{
			name:   "Valid request",
			taskID: "123",
			expected: []model.UserPoint{
				{
					UserAddress: "0xabc",
					TaskID:      "123",
					Point:       100,
				},
			},
			statusCode: http.StatusOK,
		},
	}

	if err := userPointMgr.UpsertForUserTask(context.TODO(), "0xabc", "123", 100); err != nil {
		t.Errorf("UpsertForUserTask() error = %v", err)
		return
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/userPoints/"+tt.taskID, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Create a response recorder
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			// Assert the status code
			assert.Equal(t, tt.statusCode, w.Code)

			var responseBody []model.UserPoint
			err = json.Unmarshal(w.Body.Bytes(), &responseBody)
			if err != nil {
				t.Fatalf("Failed to unmarshal response body: %v", err)
			}

			assert.Equal(t, tt.expected[0].UserAddress, responseBody[0].UserAddress)
			assert.Equal(t, tt.expected[0].TaskID, responseBody[0].TaskID)
			assert.Equal(t, tt.expected[0].Point, responseBody[0].Point)
		})
	}
}

func Test_CreateSharePoolTask(t *testing.T) {
	godotenv.Load("../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	r := gin.Default()

	taskMgr := task.NewManager(d)
	userPointMgr := userpoint.NewManager(d)
	server := &RestServer{
		TaskMgr:      taskMgr,
		UserPointMgr: userPointMgr,
		UserTaskMgr:  usertask.NewManager(d, taskMgr, transaction.NewManager(d), userPointMgr),
	}

	// Register the endpoint
	r.POST("/sharePoolTask", server.CreateSharePoolTask)

	tests := []struct {
		name       string
		body       map[string]interface{}
		statusCode int
	}{
		{
			name: "Valid request",
			body: map[string]interface{}{
				"address": "0x12345",
				"startAt": "2024-08-25",
			},
			statusCode: http.StatusOK,
		},
		{
			name: "Invalid date format",
			body: map[string]interface{}{
				"address": "0x12345",
				"startAt": "invalid-date",
			},
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			// Create a new HTTP request
			req, err := http.NewRequest(http.MethodPost, "/sharePoolTask", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			// Create a response recorder
			w := httptest.NewRecorder()

			// Serve the request
			r.ServeHTTP(w, req)

			// Assert the status code
			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}
