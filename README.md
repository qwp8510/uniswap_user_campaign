## Support
- [v] API: Get user tasks status 
- [v] API: Get user points history for distributed tasks
- [v] [Test coverage: 53.1%](https://github.com/qwp8510/uniswap_user_campaign/actions/runs/10548759653)
- [v] Support Onboarding Task
- [v] Support Share Pool Task

- [v] Support both subscriptions over WebSockets or HTTP API
- [v] Support real-time calculation when action happens(for onboarding task)
- [v] Support dynamic adding Share pool task based on different pairs
- [v] Github action CI pipeline (run test on PR, build image, etc.)

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

