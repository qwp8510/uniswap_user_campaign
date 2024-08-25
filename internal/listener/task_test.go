package listener

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"
	"tradingAce/internal/testutils"
	"tradingAce/pkg/constants"
	"tradingAce/pkg/core/db"
	"tradingAce/pkg/service/task"
	"tradingAce/pkg/service/transaction"
	"tradingAce/pkg/service/userpoint"
	"tradingAce/pkg/service/usertask"
	"tradingAce/pkg/utils"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestSwapEventTask_handleEvent(t *testing.T) {
	godotenv.Load("../../.env/.env")

	d, err := testutils.GetTestDb(t, "../../migrations")
	if err != nil {
		t.Errorf("setup db err: %v", err)
		return
	}
	defer d.Close()

	contractABI, err := abi.JSON(strings.NewReader(constants.UniswapSwapEventABI))
	if err != nil {
		t.Errorf("contractABI err: %v", err)
		return
	}

	trMgr := transaction.NewManager(d)
	listener := SwapEventTask{
		TransactionMgr: transaction.NewManager(d),
		UserTaskMgr:    usertask.NewManager(d, task.NewManager(d), trMgr, userpoint.NewManager(d)),
	}
	listener.newClient()

	eventSignature := []byte("Swap(address,uint256,uint256,uint256,uint256,address)")
	eventSigHash := crypto.Keccak256Hash(eventSignature)

	// Log topics
	sender := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	to := common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdef")

	topics := []common.Hash{
		eventSigHash,
		common.BytesToHash(sender.Bytes()),
		common.BytesToHash(to.Bytes()),
	}

	// Non-indexed data (packed into log.Data)
	amount0In := big.NewInt(100000000)
	amount1In := big.NewInt(0)
	amount0Out := big.NewInt(0)
	amount1Out := big.NewInt(222222222)

	// Pack the non-indexed parameters into log.Data using the ABI encoder
	swapEvent := contractABI.Events["Swap"]
	data, err := abi.Arguments{
		{Type: swapEvent.Inputs[1].Type},
		{Type: swapEvent.Inputs[2].Type},
		{Type: swapEvent.Inputs[3].Type},
		{Type: swapEvent.Inputs[4].Type},
	}.Pack(amount0In, amount1In, amount0Out, amount1Out)
	if err != nil {
		t.Errorf("data err: %v", err)
		return
	}

	log := types.Log{
		Topics:      topics,
		Data:        data,
		BlockNumber: uint64(20504485),
	}

	ctx := context.TODO()
	block, err := listener.client.BlockByNumber(ctx, big.NewInt(int64(log.BlockNumber)))
	if err != nil {
		t.Errorf("failed to get block: %v", err)
		return
	}

	db.Upgrade(d, "../../migrations")
	if _, err := d.Exec(
		`INSERT INTO task("id", "createdAt", "name", "pairAddress", "startAt")
		SELECT $1, $2, $3, $4, $5
		WHERE NOT EXISTS (SELECT 1 FROM task WHERE name = 'onboarding');`,
		utils.GenDBID(), time.Now(), "onboarding", nil, "2024-06-02",
	); err != nil {
		t.Errorf("insert onboarding task err: %v", err)
		return
	}

	if err := listener.handleEvent(ctx, log, block, contractABI); err != nil {
		t.Errorf("handleEvent err: %v", err)
	}
}

func TestSwapEventTask_isStopTask(t *testing.T) {
	listener := SwapEventTask{}
	listener.newClient()

	ctx := context.TODO()
	block, err := listener.client.BlockByNumber(ctx, big.NewInt(int64(20504485)))
	if err != nil {
		t.Errorf("failed to get block: %v", err)
		return
	}

	result1, err := listener.isStopTask(ctx, block, time.Now())
	if err != nil {
		t.Errorf("isStopTask err: %v", err)
		return
	}
	assert.True(t, !result1)

	endAt, parseErr := time.Parse("2006-01-02", "2024-06-02")
	if parseErr != nil {
		t.Errorf("parse time err: %v", parseErr)
		return
	}
	result2, err := listener.isStopTask(context.Background(), block, endAt)
	if err != nil {
		t.Errorf("isStopTask err: %v", err)
		return
	}
	assert.True(t, result2)
}

func TestSwapEventTask_getTaskEndAt(t *testing.T) {
	listener := SwapEventTask{}

	endAt, parseErr := time.Parse("2006-01-02", "2024-06-02")
	if parseErr != nil {
		t.Errorf("parse time err: %v", parseErr)
		return
	}

	expected, parseErr := time.Parse("2006-01-02", "2024-06-30")
	if parseErr != nil {
		t.Errorf("parse time err: %v", parseErr)
		return
	}

	result := listener.getTaskEndAt(endAt)

	assert.True(t, result.Equal(expected))
}
