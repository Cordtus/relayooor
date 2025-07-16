# Monitoring Dashboard Quick Start

## Prerequisites
- Chainpulse is running on http://localhost:3000/metrics
- Node.js and Go installed

## Start the Services

### 1. Start the API Backend
```bash
cd api
go run cmd/server/main.go
# API will start on http://localhost:8080
```

### 2. Start the Frontend
```bash
cd webapp
yarn install  # if not already done
yarn dev
# Frontend will start on http://localhost:5173
```

## Access the Dashboard

1. Open http://localhost:5173/monitoring
2. You should see:
   - **System Overview**: 3 active chains (Cosmos Hub, Osmosis, Neutron)
   - **Packet Flow Chart**: Real-time packet activity
   - **Channel Performance**: Active IBC channels with success rates
   - **Relayer Leaderboard**: Top relayers by packet count

## What You're Seeing

The dashboard displays real-time data from Chainpulse:
- **Effected Packets**: Successfully relayed packets (first to relay)
- **Uneffected Packets**: Packets that were frontrun by other relayers
- **Success Rate**: Percentage of packets successfully relayed vs frontrun
- **Active Relayers**: Currently active relayers with their memos

## Troubleshooting

### No data showing?
1. Check Chainpulse is running: `curl http://localhost:3000/metrics`
2. Check API is running: `curl http://localhost:8080/api/metrics/chainpulse`
3. Check browser console for errors (F12)

### CORS errors?
- Make sure the API backend is running (it proxies Chainpulse metrics)

### Data not updating?
- Check auto-refresh is enabled (bottom right of monitoring page)
- Default refresh: 5 seconds

## Live Data Features

With Chainpulse running, you'll see:
- Real packet counts from Osmosis transactions
- Active relayer competition metrics
- Channel-specific performance data
- Actual relayer addresses and their service memos

The dashboard automatically updates every 5 seconds to show the latest blockchain activity.