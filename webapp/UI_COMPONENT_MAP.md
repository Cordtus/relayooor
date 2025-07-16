# UI Component Map - Relayooor Webapp

## Overview
This document provides a comprehensive map of all UI components in the webapp/src directory, identifying data usage, duplications, hardcoded values, and relationships.

## Navigation Structure (App.vue)

### Main Navigation
- **File**: `/webapp/src/App.vue`
- **Components Used**:
  - `ConnectionStatus` - Shows WebSocket connection status
  - `WalletConnect` - Keplr wallet connection button
- **Navigation Items** (hardcoded):
  1. Dashboard (`/`)
  2. Monitoring (`/monitoring`)
  3. Channels (`/channels`)
  4. Relayers (`/relayers`)
  5. Packet Clearing (`/packet-clearing`)
  6. Analytics (`/analytics`)
  7. Settings (`/settings`)

## Page/View Analysis

### 1. Dashboard.vue
**File**: `/webapp/src/views/Dashboard.vue`

#### Components Used:
- None (all UI is inline)

#### Data Displayed:
- **Quick Stats** (4 cards):
  - Total Chains: `stats.chains`
  - Active Relayers: `stats.relayers`
  - 24h Packets: `stats.packets`
  - Success Rate: `stats.successRate`
  
- **Recent Activity Section**:
  - Chain names (using `getChainName()` helper)
  - Channel IDs
  - Status (success/failed)
  
- **Top Relayers Section**:
  - Relayer addresses (truncated with `formatAddress()`)
  - Success rates
  - Total packets
  
- **Top Routes Section**:
  - Channel pairs
  - Stuck packet counts
  - Stuck packet age
  - Total value by denomination

#### API Calls:
1. `metricsService.getMonitoringData()` - Main monitoring data
2. `api.get('/api/monitoring/metrics')` - Comprehensive metrics
3. `api.get('/api/channels/congestion')` - Channel congestion

#### Hardcoded Values:
- Chain name mappings: `osmosis-1` → 'Osmosis', `cosmoshub-4` → 'Cosmos Hub', `neutron-1` → 'Neutron'
- Refresh interval: 10000ms (10 seconds)
- Channel inference logic (e.g., channel-141 → cosmoshub-4)

#### Data Sources:
- `comprehensiveMetrics` - Primary data source
- `monitoringData` - Fallback data source
- `channelCongestion` - Channel-specific data

### 2. Monitoring.vue
**File**: `/webapp/src/views/Monitoring.vue`

#### Components Used:
- `MetricCard` (4 instances) - Displays key metrics
- `RefreshControl` - Auto-refresh toggle
- `LastUpdate` - Shows last update timestamp
- `PacketFlowChart` - Time series chart
- `NetworkHealthMatrix` - Chain health grid
- `ChannelUtilizationHeatmap` - Channel usage heatmap
- `ChainCard` - Individual chain status cards
- `ChainComparisonChart` - Chain performance comparison
- `RelayerLeaderboard` - Top relayers list
- `RelayerMarketShare` - Market share pie chart
- `RelayerEfficiencyMatrix` - Relayer performance matrix
- `SoftwareDistribution` - Relayer software breakdown
- `ChannelPerformanceTable` - Channel metrics table
- `ChannelFlowSankey` - Flow visualization
- `CongestionAnalysis` - Channel congestion analysis
- `FrontrunTimeline` - Frontrun events timeline
- `CompetitionHeatmap` - Relayer competition visualization
- `StuckPacketsAlert` - Alert for stuck packets
- `ConnectionIssues` - Chain connection problems
- `PerformanceAlerts` - Performance issue alerts
- `ErrorLog` - Recent errors display
- `InsightCard` - Insight display cards
- `PredictiveChart` - Prediction visualizations
- `RelayerChurn` - Relayer entry/exit analysis

#### Data Displayed:
- **System Overview**:
  - Active chains count
  - Total packets (24h)
  - Global success rate
  - Active relayers count
  
- **Tab-based Content**:
  1. Overview: Packet flow, network health, channel utilization
  2. Chains: Chain cards, performance comparison
  3. Relayers: Leaderboard, market share, efficiency, software
  4. Channels: Performance table, flow diagram, congestion
  5. Frontrun: Frontrun stats, timeline, competition heatmap
  6. Alerts: Stuck packets, connection issues, errors

#### API Calls:
- `metricsService.getMonitoringMetrics()` - Main metrics endpoint
- `analyticsService.getHistoricalTrends('24h')` - Historical data

#### Hardcoded Values:
- Tab definitions with icons
- Refresh interval options
- Peak activity period default: '14:00-18:00 UTC'
- Chart generation patterns (hourly multipliers)
- Chain name mappings (duplicated from Dashboard)

#### Computed Properties (Inferred Data):
- `globalSuccessRate` - Average across all channels
- `activeRelayersCount` - Filtered by packet count > 0
- `peakActivityPeriod` - Analyzed from packet timestamps
- `mostReliableRoute` - Highest success rate with volume > 100
- `emergingRelayer` - First relayer in sorted list
- `congestionRisk` - Based on average errors

### 3. Analytics.vue
**File**: `/webapp/src/views/Analytics.vue`

#### Components Used:
- `InsightCard` (4 instances) - Key insights display
- `VolumePredictionChart` - Volume predictions
- `SuccessRateTrendChart` - Success rate trends
- `NetworkFlowSankey` - Network flow visualization
- `HHIChart` - Market concentration chart
- `ChurnRateChart` - Relayer churn visualization
- `RecommendationCard` - Optimization recommendations

#### Data Displayed:
- **Key Insights**:
  - Peak activity time
  - Optimal route
  - Rising relayer
  - Congestion risk
  
- **Predictive Analytics**:
  - Volume prediction (7 days)
  - Success rate trend
  - Expected totals and confidence
  
- **Network Analysis**:
  - Flow diagrams
  - Top routes
  - Bottlenecks
  
- **Competition Analysis**:
  - Market concentration (HHI)
  - Relayer churn rates
  
- **Anomaly Detection**:
  - Anomaly alerts with severity
  - Timestamps and descriptions

#### API Calls:
- `clearingService.getPlatformStatistics()` - Platform stats
- `analyticsService.getChannelCongestion()` - Channel congestion
- `analyticsService.getStuckPacketsAnalytics()` - Stuck packets
- `analyticsService.getRelayerPerformance()` - Relayer performance
- `analyticsService.getNetworkFlows()` - Network flows

#### Hardcoded Values:
- Time range options: '24h', '7d', '30d', '90d'
- Default anomalies (2 examples)
- Default network flows (Osmosis ↔ Cosmos Hub, Neutron)
- Projection growth rate: 2% daily
- Success rate improvement: 0.1% per hour

### 4. PacketClearing.vue
**File**: `/webapp/src/views/PacketClearing.vue`

#### Components Used:
- `WalletConnect` - Wallet connection
- `ClearingWizard` - Main clearing flow
- `UserStatistics` - User-specific stats
- `PlatformStats` - Platform-wide statistics

#### Data Displayed:
- Wallet connection status
- User authentication state
- Platform statistics (delegated to PlatformStats)

#### API Calls:
- `clearingService.generateAuthMessage()` - Auth message generation

#### Hardcoded Values:
- Session token key: 'clearing_session_token'
- Info banner text

### 5. Settings.vue
**File**: `/webapp/src/views/Settings.vue`

#### Components Used:
- `Card` - Generic card wrapper

#### Data Displayed:
- **Monitoring Configuration**:
  - Refresh interval (5-300 seconds)
  - Stuck packet threshold (15-1440 minutes)
  - Notification toggle
  
- **Connected Services**:
  - Chainpulse status
  - API Backend status
  - WebSocket status
  
- **Chain Configuration**:
  - Supported chains list
  - Clearing fees
  - Denominations
  
- **Advanced Settings**:
  - Cache TTL (60-3600 seconds)
  - Max packets per clear (1-100)
  - Developer mode toggle

#### API Calls:
- `apiClient.get('/health')` - API health check
- `fetch(chainpulseUrl + '/api/health')` - Chainpulse health

#### Hardcoded Values:
- Default settings object
- Supported chains: 'osmosis-1', 'cosmoshub-4', 'neutron-1', 'noble-1'
- Settings storage key: 'relayooor_settings'
- API URLs from environment variables

## Data Duplication Analysis

### 1. Chain Name Mappings
**Duplicated in**:
- `Dashboard.vue` - `getChainName()` function
- `Monitoring.vue` - `getChainName()` function
- `services/api.ts` - `metricsService.getChainName()` and `chainRegistryService.getChainName()`
- `config/chains.ts` - `CHAIN_CONFIG` and `getChainName()`

**Recommendation**: Use centralized `config/chains.ts` everywhere

### 2. Number Formatting
**Duplicated in**:
- `Dashboard.vue` - `formatNumber()`, `formatPacketCount()`
- `Monitoring.vue` - `formatNumber()`
- `Analytics.vue` - `formatNumber()`
- `clearing/PlatformStats.vue` - `formatNumber()`

**Recommendation**: Create shared utility function

### 3. Success Rate Calculations
**Duplicated in**:
- `Dashboard.vue` - Inline calculation from channels
- `Monitoring.vue` - `globalSuccessRate` computed property
- `services/api.ts` - Channel and relayer success rate calculations

### 4. Stuck Packet Filtering
**Duplicated logic in**:
- `Dashboard.vue` - Channel congestion display
- `Monitoring.vue` - `enrichedChannels` computed property
- Multiple API calls for stuck packet data

### 5. Time/Duration Formatting
**Duplicated in**:
- `Dashboard.vue` - `formatDuration()`
- Various components use different formatting approaches

## Component Relationships

### Data Flow Hierarchy:
1. **API Services** (`services/api.ts`)
   - `metricsService` - Raw metrics parsing
   - `analyticsService` - Analytics endpoints
   - `chainRegistryService` - Chain information
   
2. **Stores** (Pinia)
   - `wallet` - Wallet connection state
   - `settings` - User preferences
   - `connection` - WebSocket status
   
3. **Views** consume data via:
   - Direct API calls using `useQuery`
   - Computed properties for derived data
   - Props passed to child components

### Shared Components Usage:
- `MetricCard` - Used in Monitoring for system stats
- `InsightCard` - Used in Monitoring and Analytics
- `WalletConnect` - Used in App.vue and PacketClearing
- `Card` - Generic wrapper used in Settings

## Recommendations

### 1. Centralize Common Functions
Create a `utils/formatting.ts` file for:
- `formatNumber()`
- `formatDuration()`
- `formatAddress()`
- `formatAmount()`

### 2. Use Chain Config Consistently
- Remove hardcoded chain mappings
- Always use `config/chains.ts` functions
- Consider loading chain data once at app startup

### 3. Reduce API Duplication
- Some views fetch similar data (monitoring metrics)
- Consider using shared composables or stores
- Cache responses appropriately

### 4. Standardize Component Props
- Many components display similar data differently
- Create consistent prop interfaces
- Use TypeScript interfaces consistently

### 5. Extract Reusable Components
- Success rate displays
- Chain selector dropdowns
- Packet status badges
- Time range selectors

### 6. Remove Hardcoded Values
- Move refresh intervals to config
- Extract threshold values to settings
- Centralize route definitions