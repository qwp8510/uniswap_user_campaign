## Overview

This project is a backend service that powers a **user incentive campaign on Uniswap**, designed to track on-chain activity and reward users based on their trading behavior.

The system monitors swaps on a target Uniswap pool and translates user activity into **points**, which are later used for reward distribution. It simulates how real-world Web3 growth campaigns (e.g. liquidity mining, trading incentives) are implemented from a backend perspective.

### Key Ideas

* **On-chain activity tracking**
  Listen to Uniswap V2 pool events and aggregate user swap volume in USD.

* **Task-based campaign system**
  Users complete predefined tasks to earn points:

  * Onboarding task (first milestone)
  * Weekly share-pool tasks (volume-based rewards)

* **Points & reward distribution**
  Points are calculated based on user contribution and distributed proportionally.

* **Flexible campaign lifecycle**
  Supports both:

  * Backtesting historical data
  * Running in real-time mode for active campaigns

## 📌 Notes

This project is inspired by real-world trading incentive campaigns commonly used in DeFi ecosystems.

## Setup Env
1. copy environment file
Start by copying the example environment configuration file to create your own `.env` file:
```bash
cp .env/.env.example .env/.env
```
2. update `API_KEY` (infura) in your .env/.env  
infura: https://app.infura.io/  
if you really need the API_KEY to test the service, please mail the developer

## How To Run
docker-compose up -d

## HoW To Play
### API: Get user tasks status by address
```bash
# sample api: http://0.0.0.0:8080/userTasks/<address>
curl --location 'http://0.0.0.0:8080/userTasks/0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D'
```
### API: Get user points history for distributed tasks
```bash
curl --location 'http://0.0.0.0:8080/userPoints/'
```
```bash
# sample api: http://0.0.0.0:8080/userPoints/<task id>
curl --location 'http://0.0.0.0:8080/userPoints/8cc05973606147b883bb9da5ccb9c0c1'
```

### API: Dynamic adding Share pool task based on different pairs
only support adding pair address for USDC/ETH
```bash
curl --location 'http://0.0.0.0:8080/sharePoolTask/' \
--header 'Content-Type: application/json' \
--data '{
    "address": "0x8ad599c3A0ff1De082011EFDDc58f1908eb6e6D8",
    "startAt": "2024-08-15"
}'
```
### CLI: Check share pool task
Run it in your container environment
```bash
/home/nonroot/app checkSharePoolTask
```

## Task Processing Overview
For each new swap event received, the system checks if it meets the criteria for an onboarding task.    
If the `share_pool` task started before today, after synchronizing historical events, the service will check the weekly `share_pool` tasks. The service also provides a CLI that allows you to manually check `share_pool` tasks at any time.

