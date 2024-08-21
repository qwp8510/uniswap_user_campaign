package listener

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"
	"tradingAce/pkg/constants"
	iface "tradingAce/pkg/interface"
	"tradingAce/pkg/model"
	"tradingAce/pkg/model/option"
	"tradingAce/pkg/utils"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type SwapEventTask struct {
	TaskMgr        iface.TaskManager
	TransactionMgr iface.TransactionManager
	UserTaskMgr    iface.UserTaskManager
}

type swapEvent struct {
	Amount0In  *big.Int
	Amount1In  *big.Int
	Amount0Out *big.Int
	Amount1Out *big.Int
}

func (t *SwapEventTask) Listen() {
	ctx := context.Background()

	subscribeUrl := "wss://mainnet.infura.io/ws/v3/" + os.Getenv("API_KEY")
	if os.Getenv("SUBSCRIBE_MODE") == "http" {
		subscribeUrl = "https://mainnet.infura.io/v3/" + os.Getenv("API_KEY")
	}

	client, err := ethclient.Dial(subscribeUrl)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	tasks, listErr := t.TaskMgr.GetSharePoolTask(ctx)
	if listErr != nil {
		log.Fatalf("list task pair address failed: %s", listErr)
	}

	// Parse ABI for Swap event
	contractABI, err := abi.JSON(strings.NewReader(constants.UniswapSwapEventABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	var wg sync.WaitGroup

	for _, task := range tasks {
		wg.Add(1)

		go func(task model.Task) {
			defer wg.Done()
			if err := t.subscribeToPool(ctx, client, contractABI, task); err != nil {
				log.Printf("subscribeToPool error: %s", err)
				return
			}
		}(task)
	}

	wg.Wait()
}

func (t *SwapEventTask) subscribeToPool(
	ctx context.Context, client *ethclient.Client, contractABI abi.ABI, task model.Task,
) error {

	log.Println("syncing history event...")
	latestBlockNum, syncErr := t.syncHistoryEvent(ctx, contractABI, task)
	if syncErr != nil {
		return syncErr
	}
	log.Println("finish syncing history")

	if t.getTaskEndAt(task.StartAt).Before(time.Now()) {
		// task was finished
		return nil
	}

	if os.Getenv("SUBSCRIBE_MODE") == "http" {
		t.subscribeByHTTP(ctx, client, contractABI, task, latestBlockNum)
	} else {
		t.subscribeByWS(ctx, client, contractABI, task, latestBlockNum)
	}

	return nil
}

func (t *SwapEventTask) subscribeByWS(
	ctx context.Context,
	client *ethclient.Client,
	contractABI abi.ABI,
	task model.Task,
	startBlock *big.Int,
) {

	endAt := t.getTaskEndAt(task.StartAt)

	// Filter query for Swap events in the Uniswap pool
	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(task.PairAddress.String)},
		Topics:    [][]common.Hash{{contractABI.Events["Swap"].ID}},
		FromBlock: startBlock,
	}

	// Subscribe to Swap events
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		log.Fatalf("Failed to subscribe to logs: %v", err)
	}

	log.Printf("Listening Swap events for target pool: %s", task.PairAddress.String)

	count := 0
	for {
		select {
		case err := <-sub.Err():
			log.Fatalf("Subscription error: %v", err)
		case vLog := <-logs:
			count += 1
			if stop, err := t.isStopTask(ctx, client, vLog, endAt); stop || err != nil {
				log.Fatalf("Subscription isStopTask error: %v", err)
			}

			t.handleEvent(ctx, vLog, contractABI)
			fmt.Println("count", count)
		}
	}
}

func (t *SwapEventTask) subscribeByHTTP(
	ctx context.Context,
	client *ethclient.Client,
	contractABI abi.ABI,
	task model.Task,
	startBlock *big.Int,
) {

	for {
		latestBlock, err := client.BlockByNumber(context.Background(), nil)
		if err != nil {
			log.Fatalf("Failed to get latest block: %v", err)
		}

		endBlock := latestBlock.Number()

		query := ethereum.FilterQuery{
			FromBlock: startBlock,
			// ToBlock:   endBlock,
			Addresses: []common.Address{common.HexToAddress(task.PairAddress.String)},
			Topics:    [][]common.Hash{{contractABI.Events["Swap"].ID}},
		}

		logs, err := client.FilterLogs(context.Background(), query)
		if err != nil {
			log.Fatalf("Failed to filter logs: %v", err)
		}

		for _, vLog := range logs {
			t.handleEvent(ctx, vLog, contractABI)
		}

		// Update startBlock to the latest block number, so that the next query will continue to query new events
		if len(logs) != 0 {
			startBlock = endBlock.Add(endBlock, big.NewInt(1))
		}

		time.Sleep(1 * time.Second)
	}
}

func (t *SwapEventTask) syncHistoryEvent(
	ctx context.Context, contractABI abi.ABI, task model.Task,
) (latestBlockNum *big.Int, err error) {

	fmt.Println("https://mainnet.infura.io/v3/" + os.Getenv("API_KEY"))
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/" + os.Getenv("API_KEY"))
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	poolAddress := common.HexToAddress(task.PairAddress.String)

	fromBlock := big.NewInt(20504485)
	endBlock := big.NewInt(20514485)

	// log.Printf("seaching start block number for start time: %s", task.StartAt)
	// fromBlock, startBlockErr := t.getBlockByTimestamp(ctx, client, task.StartAt)
	// if startBlockErr != nil {
	// 	return big.NewInt(0), startBlockErr
	// }

	// endAt := t.getTaskEndAt(task.StartAt)
	// endBlock := big.NewInt(0)
	// if endAt.After(time.Now()) {
	// 	latestBlock, err := client.BlockByNumber(ctx, nil)
	// 	if err != nil {
	// 		return big.NewInt(0), err
	// 	}

	// 	endBlock = latestBlock.Number()

	// } else {
	// 	log.Printf("seaching end block number for end time: %s", endAt)
	// 	b, endBlockErr := t.getBlockByTimestamp(ctx, client, endAt)
	// 	if endBlockErr != nil {
	// 		return big.NewInt(0), endBlockErr
	// 	}
	// 	endBlock = b
	// }

	// Filter query for Swap events in the Uniswap pool
	batch := int64(10000)
	for fromBlock.Cmp(endBlock) < 0 {
		toBlock := new(big.Int).Add(fromBlock, big.NewInt(batch))

		if toBlock.Cmp(endBlock) > 0 {
			toBlock.Set(endBlock)
		}

		query := ethereum.FilterQuery{
			Addresses: []common.Address{poolAddress},
			Topics:    [][]common.Hash{{contractABI.Events["Swap"].ID}},
			FromBlock: fromBlock,
			ToBlock:   toBlock,
		}

		log.Printf("sync history event from blockNum: %d~%d", fromBlock, toBlock)
		logs, err := client.FilterLogs(ctx, query)
		if err != nil {
			return endBlock, fmt.Errorf("failed to filter logs: %v", err)
		}

		for _, vLog := range logs {
			fmt.Println("vvvvveeeee", vLog.Address)
			t.handleEvent(ctx, vLog, contractABI)
		}

		fromBlock.Set(toBlock)
	}

	if err := t.UserTaskMgr.CheckSharePoolTasks(ctx); err != nil {
		return endBlock, err
	}

	return endBlock, nil
}

func (t *SwapEventTask) getTaskEndAt(startAt time.Time) time.Time {
	fourWeeks := 4 * 7
	return startAt.AddDate(0, 0, fourWeeks)
}

func (t *SwapEventTask) isStopTask(
	ctx context.Context, client *ethclient.Client, vLog types.Log, endAt time.Time,
) (bool, error) {

	b := big.NewInt(int64(vLog.BlockNumber))
	block, err := client.BlockByNumber(ctx, b)
	if err != nil {
		return false, err
	}

	blockTime := int64(block.Time())
	if blockTime >= endAt.Unix() {
		return true, nil
	}

	return false, nil
}

func (t *SwapEventTask) handleEvent(ctx context.Context, vLog types.Log, contractABI abi.ABI) {
	sender := common.HexToAddress(vLog.Topics[1].Hex())
	to := common.HexToAddress(vLog.Topics[2].Hex())

	// Parse the non-indexed fields (amount0In, amount1In, amount0Out, amount1Out) from Log.Data
	event := swapEvent{}

	err := contractABI.UnpackIntoInterface(&event, "Swap", vLog.Data)
	if err != nil {
		log.Printf("Failed to unpack log: %v", err)
		return
	}

	amount0In, err := utils.BigIntToDecimal(event.Amount0In)
	if err != nil {
		log.Printf("Failed to big int to decimal: %v", err)
		return
	}
	amount1In, err := utils.BigIntToDecimal(event.Amount1In)
	if err != nil {
		log.Printf("Failed to big int to decimal: %v", err)
		return
	}
	amount0Out, err := utils.BigIntToDecimal(event.Amount0Out)
	if err != nil {
		log.Printf("Failed to big int to decimal: %v", err)
		return
	}
	amount1Out, err := utils.BigIntToDecimal(event.Amount1Out)
	if err != nil {
		log.Printf("Failed to big int to decimal: %v", err)
		return
	}

	opt := option.TransactionUpsertOptions{
		BlockNum:        vLog.BlockNumber,
		PairAddress:     vLog.Address.Hex(),
		SenderAddress:   sender.Hex(),
		Amount0In:       amount0In,
		Amount1In:       amount1In,
		Amount0Out:      amount0Out,
		Amount1Out:      amount1Out,
		ReceiverAddress: to.Hex(),
	}
	err = t.TransactionMgr.Upsert(ctx, opt)
	if err != nil {
		log.Printf("upsert transaction: %v", err)
		return
	}

	if err := t.UserTaskMgr.CheckOnboardingTask(ctx, sender.Hex()); err != nil {
		log.Printf("handle event CheckOnboardingTask fail: %v", err)
		return
	}
}

func (t *SwapEventTask) getBlockByTimestamp(ctx context.Context, client *ethclient.Client, taskTime time.Time) (*big.Int, error) {
	var blockNumber *big.Int

	start := big.NewInt(0)
	latestBlock, err := client.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}

	end := latestBlock.Number()

	targetTimestamp := taskTime.Unix()

	for start.Cmp(end) < 0 {
		mid := new(big.Int).Add(start, end)
		mid = mid.Div(mid, big.NewInt(2))

		block, err := client.BlockByNumber(ctx, mid)
		if err != nil {
			return nil, err
		}

		blockTime := int64(block.Time())

		if blockTime < targetTimestamp {
			start = mid.Add(mid, big.NewInt(1))
		} else {
			end = mid
		}
	}

	blockNumber = start
	return blockNumber, nil
}
