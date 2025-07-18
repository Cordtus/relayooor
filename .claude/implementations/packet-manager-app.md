# Packet Manager Application

## Overview
A Vue.js 3 web application for managing stuck IBC packets, created as a simplified interface for operators to query and clear stuck packets across multiple chains.

## Technical Stack
- Vue.js 3 with Composition API
- Vite for build tooling
- Axios for API communication
- CSS Grid for layout

## Features

### 1. Chain Selection
- Dropdown selector for supported chains:
  - Cosmos Hub
  - Osmosis
  - Noble
  - Stride
  - Jackal
  - Axelar

### 2. Channel Selection
- Dynamic channel list based on selected chain
- Shows source and destination channels

### 3. Packet Query
- Fetches stuck packets from Chainpulse API
- Displays:
  - Packet sequence number
  - Stuck duration
  - Amount and denomination
  - Sender/receiver addresses
  - Transaction hash

### 4. Hermes Metrics
- Real-time metrics from Hermes relayer
- Shows pending packets per channel
- Helps validate Chainpulse data

### 5. Packet Clearing
- Individual packet clearing with checkbox selection
- "Clear All" functionality
- Simulated clearing in current implementation

## API Integration

### Endpoints Used
- `GET /api/packets/stuck?chain={chainId}` - Query stuck packets
- `GET /api/metrics/chainpulse` - Get Prometheus metrics
- `POST /api/relayer/hermes/clear` - Clear packets (simulated)

### Data Flow
1. User selects chain and channel
2. App queries Chainpulse for stuck packets
3. App fetches Hermes metrics for validation
4. User can select and clear packets
5. Clear requests sent to API

## Configuration

### Chains and Channels
Currently hardcoded in `/packet-manager/src/services/api.js`. Includes known problematic channels:
- Osmosis channel-0 ↔ Cosmos Hub channel-141
- Osmosis channel-750 ↔ Noble channel-1
- Axelar channel-208 ↔ Osmosis channel-208

### Environment
- Development: `npm run dev` (port 5174)
- Production: Built as static files, served via nginx

## User Interface

### Layout
- Header with title
- Control panel with dropdowns and query button
- Two-column data display:
  - Left: Chainpulse stuck packets
  - Right: Hermes metrics
- Action buttons at bottom

### Styling
- Clean, professional design
- Responsive layout
- Color coding for different states
- Loading indicators

## Known Limitations

1. **Mock Clearing**: Current implementation simulates packet clearing
2. **Static Channels**: Channel configuration is hardcoded
3. **No Real-time Updates**: Requires manual refresh
4. **Limited Error Details**: Basic error messages

## Future Enhancements

1. **Real Packet Clearing**: Connect to actual Hermes instance
2. **Dynamic Channels**: Fetch from API based on chain activity
3. **WebSocket Support**: Real-time packet updates
4. **Wallet Integration**: Keplr/Leap wallet for signing
5. **Batch Operations**: Clear multiple channels at once
6. **Historical Data**: Show clearing history and success rates