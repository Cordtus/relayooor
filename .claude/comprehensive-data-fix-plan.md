# Comprehensive Data Fix Plan

## Phase 1: Data Source Inventory

### A. Chainpulse Data Sources
1. **Prometheus Metrics** (http://localhost:3001/metrics)
   - `chainpulse_chains` - Number of monitored chains
   - `chainpulse_txs{chain_id}` - Total transactions per chain
   - `chainpulse_packets{chain_id}` - Total packets per chain
   - `chainpulse_errors{chain_id}` - Errors per chain
   - `ibc_effected_packets{...}` - Successfully relayed packets
   - `ibc_uneffected_packets{...}` - Frontrun packets
   - `ibc_stuck_packets{...}` - Stuck packet gauge
   - `ibc_frontrun{...}` - Frontrun events

2. **Chainpulse API** (if available)
   - `/api/v1/packets/stuck` - Stuck packets with details
   - `/api/v1/packets/by-user` - User-specific packets
   - `/api/v1/channels/congestion` - Channel congestion data

### B. Backend API Data Sources
1. **Monitoring Endpoints**
   - `/api/monitoring/metrics` - Comprehensive metrics snapshot
   - `/api/monitoring/data` - Structured monitoring data
   - `/api/metrics/chainpulse` - Raw Prometheus metrics proxy
   - `/api/config` - Chain registry and configuration

2. **Packet Endpoints**
   - `/api/packets/stuck` - Global stuck packets
   - `/api/user/{wallet}/transfers` - User transfers
   - `/api/user/{wallet}/stuck` - User stuck packets
   - `/api/packets/{chain}/{channel}/{sequence}` - Packet details

3. **Analytics Endpoints**
   - `/api/statistics/platform` - Platform-wide statistics
   - `/api/channels/congestion` - Channel congestion
   - `/api/chains/registry` - Known chains configuration

## Phase 2: UI Component Mapping

### Dashboard Page
1. **Top Stats Cards**
   - Total Chains → `metrics.system.totalChains`
   - Active Relayers → `metrics.relayers.length`
   - 24h Packets → `metrics.system.totalPackets`
   - Success Rate → Calculate from channels data

2. **Recent Activity**
   - Source: `metrics.recentPackets`
   - Currently: Using computed mock data
   - Fix: Use real recent packets from API

3. **Top Relayers**
   - Source: `metrics.relayers` sorted by totalPackets
   - Currently: Showing mock relayers
   - Fix: Display real relayer data

4. **Top Routes**
   - Source: `metrics.channels` sorted by volume
   - Currently: Hardcoded routes
   - Fix: Use real channel data

### IBC Monitoring Page

#### Overview Tab
1. **Packet Flow Chart**
   - Source: Should aggregate from `metrics.recentPackets` over time
   - Currently: Generated pattern based on total
   - Fix: Store historical data or use time-series from Chainpulse

2. **Network Health Matrix**
   - Source: `metrics.chains` status
   - Currently: Working correctly
   - Verify: Chain status colors and indicators

3. **Channel Utilization Heatmap**
   - Source: `metrics.channels` utilization
   - Currently: May be using mock data
   - Fix: Calculate from real channel packet counts

#### Chains Tab
1. **Chain Cards**
   - Source: `metrics.chains` + calculated packet stats
   - Currently: Fixed to show real totalPackets
   - Verify: Success rates calculation

2. **Chain Performance Comparison**
   - Source: `metrics.chains` comparative data
   - Fix: Ensure using real metrics

#### Relayers Tab
1. **Top Relayers Leaderboard**
   - Source: `metrics.relayers`
   - Fix: Sort by real performance metrics

2. **Market Share Chart**
   - Source: Calculate from relayer packet counts
   - Fix: Use real relayer data

3. **Software Distribution**
   - Source: `metrics.relayers` software field
   - Fix: Aggregate real software types

#### Channels Tab
1. **Channel Performance Table**
   - Source: `enrichedChannels` computed
   - Currently: Seems correct
   - Verify: All columns show real data

2. **Channel Flow Sankey**
   - Source: `metrics.channels`
   - Fix: Ensure proper data flow visualization

#### Alerts Tab
1. **Stuck Packets Alert**
   - Source: `metrics.stuckPackets`
   - Currently: Empty array (no stuck packets)
   - Fix: Add test data or wait for real stuck packets

2. **Connection Issues**
   - Source: `metrics.chains` with errors/disconnected
   - Fix: Show chains with issues

3. **Error Log**
   - Source: `recentErrors` computed from chains
   - Currently: Shows chains with errors > 0
   - Verify: Error display format

### Packet Clearing Page
1. **User's Stuck Packets**
   - Source: `/api/user/{wallet}/stuck`
   - Fix: Ensure wallet connection and data fetch

2. **Clearing Wizard**
   - Fix: Test full flow with mock wallet

### Analytics Page
1. **Platform Statistics**
   - Source: `/api/statistics/platform`
   - Currently: Returns mock data
   - Fix: Calculate from real metrics

2. **Network Flow**
   - Fix: Use real channel data

3. **Channel Congestion**
   - Source: `/api/channels/congestion`
   - Fix: Forward real Chainpulse data

### Settings Page
1. **Chain Configuration**
   - Source: `configService`
   - Currently: Fixed to load from API
   - Verify: All chains display correctly

## Phase 3: Data Consistency Checks

### Common Data Points
1. **Chain IDs**
   - Used in: Chain cards, dropdowns, channel displays
   - Source: Single source of truth in chain registry
   - Ensure: Consistent naming and formatting

2. **Packet Counts**
   - 24h total per chain
   - Channel-specific counts
   - Relayer-specific counts
   - Ensure: All derive from same base metrics

3. **Success Rates**
   - Chain-level: Aggregate from channels
   - Channel-level: effected/total packets
   - Relayer-level: Individual performance
   - Ensure: Consistent calculation method

4. **Stuck Packets**
   - Global view in monitoring
   - User-specific in packet clearing
   - Ensure: Same identification criteria

## Phase 4: Implementation Steps

1. **Fix Data Sources**
   - Update API endpoints to return real data
   - Remove all hardcoded/mock responses
   - Add data aggregation where needed

2. **Update Components**
   - Replace mock data references
   - Fix data transformations
   - Add proper error handling

3. **Add Missing Data**
   - Historical packet flow
   - Real-time updates via WebSocket
   - Persistent metrics storage

4. **Testing**
   - Verify each component displays real data
   - Check data consistency across views
   - Test with different chain states

## Phase 5: Validation Checklist

- [ ] Dashboard stats show real totals
- [ ] Recent activity shows actual packets
- [ ] Chain cards show correct packet counts
- [ ] Relayer leaderboard uses real data
- [ ] Channel performance reflects actual metrics
- [ ] Success rates calculated consistently
- [ ] All graphs/charts use real data
- [ ] No placeholder text remains
- [ ] Data updates properly on refresh
- [ ] Cross-page data consistency