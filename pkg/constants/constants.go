package constants

import "github.com/shopspring/decimal"

// Uniswap V2 Swap event ABI
const UniswapSwapEventABI = `
[
    {
        "anonymous": false,
        "inputs": [
            {
                "indexed": true,
                "name": "sender",
                "type": "address"
            },
            {
                "indexed": false,
                "name": "amount0In",
                "type": "uint256"
            },
            {
                "indexed": false,
                "name": "amount1In",
                "type": "uint256"
            },
            {
                "indexed": false,
                "name": "amount0Out",
                "type": "uint256"
            },
            {
                "indexed": false,
                "name": "amount1Out",
                "type": "uint256"
            },
            {
                "indexed": true,
                "name": "to",
                "type": "address"
            }
        ],
        "name": "Swap",
        "type": "event"
    }
]
`

var (
	// USDC and ETH price
	UsdcPrice = decimal.NewFromFloat(1.0)
	EthPrice  = decimal.NewFromFloat(2000.0)
	// USDC and ETH precision
	UsdcPrecision = decimal.NewFromFloat(1e6)
	EthPrecision  = decimal.NewFromFloat(1e18)
)

var PointsPerWeek = decimal.NewFromInt(10000)
