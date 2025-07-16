# Relayooor API Data Sources Catalog

This document provides a comprehensive catalog of all available data sources in the Relayooor application.

## Table of Contents
- [Backend API Endpoints](#backend-api-endpoints)
  - [Clearing Service Endpoints](#clearing-service-endpoints)
  - [Payment Service Endpoints](#payment-service-endpoints)
  - [Help Service Endpoints](#help-service-endpoints)
  - [Chainpulse Integration Endpoints](#chainpulse-integration-endpoints)
  - [Packet Stream Endpoints](#packet-stream-endpoints)
  - [Authentication Endpoints](#authentication-endpoints)
  - [IBC Service Endpoints](#ibc-service-endpoints)
  - [Relayer Management Endpoints](#relayer-management-endpoints)
  - [Metrics and Monitoring Endpoints](#metrics-and-monitoring-endpoints)
- [WebSocket Endpoints](#websocket-endpoints)
- [Frontend Service Mappings](#frontend-service-mappings)
- [Data Duplication Analysis](#data-duplication-analysis)

## Backend API Endpoints

### Clearing Service Endpoints

#### POST /api/v1/clearing/request-token
- **Method**: POST
- **Parameters**: 
  ```json
  {
    "walletAddress": "string",
    "chainId": "string",
    "packetIdentifiers": [
      {
        "chain": "string",
        "channel": "string",
        "port": "string",
        "sequence": "number"
      }
    ]
  }
  ```
- **Response Type**: `TokenResponse`
  ```json
  {
    "token": {
      "token": "string",
      "version": 1,
      "requestType": "string",
      "walletAddress": "string",
      "chainId": "string",
      "issuedAt": "timestamp",
      "expiresAt": "timestamp",
      "serviceFee": "string",
      "estimatedGasFee": "string",
      "totalRequired": "string",
      "acceptedDenom": "string",
      "signature": "string"
    },
    "paymentAddress": "string",
    "memo": "string"
  }
  ```
- **Data Provided**: Clearing token with payment requirements
- **Usage Notes**: Generate token for packet clearing request

#### POST /api/v1/clearing/verify-payment
- **Method**: POST
- **Parameters**: 
  ```json
  {
    "token": "string",
    "txHash": "string"
  }
  ```
- **Response Type**: `PaymentVerificationResponse`
  ```json
  {
    "operationID": "string",
    "success": true,
    "message": "string"
  }
  ```
- **Data Provided**: Payment verification status
- **Usage Notes**: Verify on-chain payment for clearing

#### GET /api/v1/clearing/status/:token
- **Method**: GET
- **Parameters**: token (path parameter)
- **Response Type**: `ClearingStatus`
  ```json
  {
    "token": "string",
    "status": "pending | paid | executing | completed | failed",
    "payment": {
      "received": true,
      "txHash": "string",
      "amount": "string"
    },
    "execution": {
      "startedAt": "timestamp",
      "completedAt": "timestamp",
      "packetsCleared": 0,
      "packetsFailed": 0,
      "txHashes": ["string"],
      "error": "string"
    }
  }
  ```
- **Data Provided**: Real-time clearing operation status
- **Usage Notes**: Also supports SSE (Server-Sent Events) with Accept: text/event-stream

#### GET /api/v1/clearing/operations
- **Method**: GET
- **Authentication**: Required (Bearer token)
- **Parameters**: 
  - page (query): number
  - pageSize (query): number  
  - sortBy (query): string
  - sortDir (query): asc|desc
- **Response Type**: Paginated operations list
- **Data Provided**: User's clearing operation history
- **Usage Notes**: Protected endpoint requiring session

### Payment Service Endpoints

#### GET /api/v1/payments/uri
- **Method**: GET
- **Parameters**: token (query parameter)
- **Response Type**: Payment URI information
  ```json
  {
    "uri": "cosmos:address?amount=X&denom=Y&memo=Z",
    "qr_code": "base64_data_url",
    "payment_address": "string",
    "amount": "string",
    "denom": "string",
    "memo": "string",
    "expires_at": "timestamp",
    "chain_id": "string"
  }
  ```
- **Data Provided**: Cosmos payment URI for wallet integration
- **Usage Notes**: Generate payment URIs for wallet apps

#### GET /api/v1/prices/:denom
- **Method**: GET
- **Parameters**: denom (path parameter)
- **Response Type**: Price information
  ```json
  {
    "denom": "string",
    "price": 0.0,
    "timestamp": "timestamp",
    "expires_at": "timestamp"
  }
  ```
- **Data Provided**: USD price for token denomination
- **Usage Notes**: Cached for 5 minutes

#### GET /api/v1/clearing/simple-status
- **Method**: GET
- **Parameters**: wallet (query parameter)
- **Response Type**: Simplified stuck packet summary
  ```json
  {
    "stuck_count": 0,
    "total_value": "string",
    "primary_denom": "string",
    "chains": [...],
    "estimated_fees": {...},
    "potential_savings": "string",
    "last_updated": "timestamp"
  }
  ```
- **Data Provided**: Simplified view for non-technical users
- **Usage Notes**: User-friendly packet summary

#### GET /api/v1/fees/breakdown
- **Method**: GET
- **Parameters**: 
  - packets (query): number
  - chain (query): string
- **Response Type**: Detailed fee breakdown
  ```json
  {
    "service_fee": {
      "amount": "string",
      "denom": "string",
      "usd_value": 0.0,
      "breakdown": {...}
    },
    "gas_fee": {
      "amount": "string",
      "denom": "string",
      "usd_value": 0.0,
      "is_estimate": true
    },
    "total": {...},
    "comparison": {...},
    "price_info": {...}
  }
  ```
- **Data Provided**: Detailed fee calculation with USD values
- **Usage Notes**: Shows savings compared to manual retry

#### POST /api/v1/payments/validate-memo
- **Method**: POST
- **Parameters**: 
  ```json
  {
    "memo": "string"
  }
  ```
- **Response Type**: Memo validation result
- **Data Provided**: Validates payment memo format
- **Usage Notes**: Ensures correct memo format before payment

### Help Service Endpoints

#### GET /api/v1/help/terms
- **Method**: GET
- **Response Type**: List of available terms
- **Data Provided**: All available glossary terms
- **Usage Notes**: Returns term names without definitions

#### GET /api/v1/help/terms/:term
- **Method**: GET
- **Parameters**: term (path parameter)
- **Response Type**: Term definition
  ```json
  {
    "term": "string",
    "definition": "string",
    "examples": ["string"],
    "related": ["string"]
  }
  ```
- **Data Provided**: Detailed term explanation
- **Usage Notes**: Educational content for users

#### GET /api/v1/help/glossary
- **Method**: GET
- **Parameters**: category (query parameter, optional)
- **Response Type**: Categorized glossary
- **Data Provided**: Terms grouped by category
- **Usage Notes**: Categories: IBC Basics, Problems, Solutions, Payments, General

### Chainpulse Integration Endpoints

#### GET /api/v1/chainpulse/packets/by-user
- **Method**: GET
- **Parameters**: address (query parameter)
- **Response Type**: Array of user packets
- **Data Provided**: All packets for a specific wallet address
- **Usage Notes**: Integrates with Chainpulse monitoring

#### GET /api/v1/chainpulse/packets/stuck
- **Method**: GET
- **Parameters**: min_stuck_minutes (query, optional, default: 30)
- **Response Type**: Array of stuck packets
- **Data Provided**: Currently stuck packets across all chains
- **Usage Notes**: Global view of stuck packets

#### GET /api/v1/chainpulse/packets/:chain/:channel/:sequence
- **Method**: GET
- **Parameters**: chain, channel, sequence (path parameters)
- **Response Type**: Detailed packet information
- **Data Provided**: Complete packet details including status
- **Usage Notes**: Deep dive into specific packet

#### GET /api/v1/chainpulse/channels/congestion
- **Method**: GET
- **Response Type**: Array of channel congestion data
  ```json
  [{
    "channelId": "string",
    "counterpartyChannelId": "string",
    "sourceChain": "string",
    "destinationChain": "string",
    "pendingPackets": 0,
    "avgClearTime": 0,
    "congestionLevel": "string"
  }]
  ```
- **Data Provided**: Channel congestion metrics
- **Usage Notes**: Helps identify problematic channels

#### GET /api/v1/chainpulse/metrics
- **Method**: GET
- **Response Type**: text/plain (Prometheus metrics)
- **Data Provided**: Raw Prometheus metrics from Chainpulse
- **Usage Notes**: For integration with monitoring systems

#### GET /api/v1/chainpulse/health
- **Method**: GET
- **Response Type**: Health status
- **Data Provided**: Chainpulse service health
- **Usage Notes**: Service availability check

### Packet Stream Endpoints

#### GET /api/v1/packets/stuck/stream
- **Method**: GET
- **Parameters**: 
  - wallet (query): required
  - cursor (query): optional
  - limit (query): optional (default: 50)
- **Response Type**: Cursor-based paginated response
  ```json
  {
    "packets": [...],
    "next_cursor": "string",
    "has_more": true,
    "count": 0
  }
  ```
- **Data Provided**: Stuck packets for wallet with pagination
- **Usage Notes**: Supports ETag caching and cursor-based pagination

#### GET /api/v1/packets/channel/stream
- **Method**: GET
- **Parameters**: 
  - src_chain (query): required
  - src_channel (query): required
  - cursor (query): optional
  - limit (query): optional
- **Response Type**: Cursor-based paginated response
- **Data Provided**: Packets for specific channel
- **Usage Notes**: Channel-specific packet stream

### Authentication Endpoints

#### POST /api/v1/auth/login
- **Method**: POST
- **Parameters**: Username/password credentials
- **Response Type**: JWT token
- **Data Provided**: Authentication token
- **Usage Notes**: Legacy authentication (when AUTH_ENABLED=true)

#### POST /api/v1/auth/refresh
- **Method**: POST
- **Parameters**: Refresh token
- **Response Type**: New JWT token
- **Data Provided**: Refreshed authentication token
- **Usage Notes**: Token refresh mechanism

#### POST /api/v1/auth/wallet-sign
- **Method**: POST
- **Parameters**: 
  ```json
  {
    "walletAddress": "string",
    "signature": "string",
    "message": "string",
    "chain": "string",
    "timestamp": 0
  }
  ```
- **Response Type**: Session token
  ```json
  {
    "sessionToken": "string",
    "expiresAt": "timestamp",
    "wallet": "string"
  }
  ```
- **Data Provided**: Wallet-based authentication session
- **Usage Notes**: Modern wallet authentication

### IBC Service Endpoints (Protected)

#### GET /api/v1/ibc/chains
- **Method**: GET
- **Authentication**: May be required
- **Response Type**: List of configured chains
- **Data Provided**: All IBC chains from relayers
- **Usage Notes**: Aggregates Hermes and Go relayer data

#### GET /api/v1/ibc/chains/:chain_id
- **Method**: GET
- **Parameters**: chain_id (path parameter)
- **Response Type**: Chain details
- **Data Provided**: Specific chain information
- **Usage Notes**: Chain configuration and status

#### GET /api/v1/ibc/chains/:chain_id/status
- **Method**: GET
- **Parameters**: chain_id (path parameter)
- **Response Type**: Chain status
- **Data Provided**: Chain connectivity status
- **Usage Notes**: Real-time chain health

#### GET /api/v1/ibc/channels
- **Method**: GET
- **Response Type**: All IBC channels
- **Data Provided**: Complete channel list
- **Usage Notes**: Currently returns placeholder

#### GET /api/v1/ibc/chains/:chain_id/channels
- **Method**: GET
- **Parameters**: chain_id (path parameter)
- **Response Type**: Chain-specific channels
- **Data Provided**: Channels for specific chain
- **Usage Notes**: Chain-filtered channel list

#### GET /api/v1/ibc/packets/pending
- **Method**: GET
- **Response Type**: Pending packets
- **Data Provided**: Packets awaiting relay
- **Usage Notes**: Aggregates from both relayers

#### POST /api/v1/ibc/packets/clear
- **Method**: POST
- **Parameters**: 
  ```json
  {
    "chain_id": "string",
    "channel_id": "string",
    "use_hermes": true
  }
  ```
- **Response Type**: Clearing result
- **Data Provided**: Packet clearing execution result
- **Usage Notes**: Legacy clearing endpoint

#### GET /api/v1/ibc/packets/stuck
- **Method**: GET
- **Response Type**: Stuck packets
- **Data Provided**: Packets stuck for extended time
- **Usage Notes**: Similar to chainpulse endpoint

### Relayer Management Endpoints (Protected)

#### GET /api/v1/relayer/status
- **Method**: GET
- **Response Type**: Relayer status
- **Data Provided**: Hermes and Go relayer status
- **Usage Notes**: Overall relayer health

#### POST /api/v1/relayer/hermes/start
- **Method**: POST
- **Response Type**: Operation result
- **Data Provided**: Hermes start confirmation
- **Usage Notes**: Start Hermes relayer

#### POST /api/v1/relayer/hermes/stop
- **Method**: POST
- **Response Type**: Operation result
- **Data Provided**: Hermes stop confirmation
- **Usage Notes**: Stop Hermes relayer

#### GET /api/v1/relayer/config
- **Method**: GET
- **Response Type**: Relayer configuration
- **Data Provided**: Current relayer config
- **Usage Notes**: View configuration

#### PUT /api/v1/relayer/config
- **Method**: PUT
- **Parameters**: New configuration
- **Response Type**: Update result
- **Data Provided**: Configuration update status
- **Usage Notes**: Update relayer config

### Metrics and Monitoring Endpoints (Protected)

#### GET /api/v1/metrics/summary
- **Method**: GET
- **Response Type**: Metrics summary
- **Data Provided**: High-level metrics overview
- **Usage Notes**: Dashboard summary data

#### GET /api/v1/metrics/packets
- **Method**: GET
- **Response Type**: Packet metrics
- **Data Provided**: Packet-specific metrics
- **Usage Notes**: Packet statistics

#### GET /api/v1/metrics/channels
- **Method**: GET
- **Response Type**: Channel metrics
- **Data Provided**: Channel performance metrics
- **Usage Notes**: Channel health data

#### GET /api/v1/metrics/chainpulse
- **Method**: GET
- **Response Type**: Chainpulse metrics
- **Data Provided**: Chainpulse monitoring data
- **Usage Notes**: Integration metrics

#### GET /api/v1/metrics/packet-flow
- **Method**: GET
- **Response Type**: Packet flow data
- **Data Provided**: Network flow visualization data
- **Usage Notes**: For flow diagrams

#### GET /api/v1/metrics/stuck-packets
- **Method**: GET
- **Response Type**: Stuck packet analytics
- **Data Provided**: Stuck packet analysis
- **Usage Notes**: Problem identification

#### GET /api/v1/metrics/relayer-performance
- **Method**: GET
- **Response Type**: Relayer performance metrics
- **Data Provided**: Relayer efficiency data
- **Usage Notes**: Performance monitoring

#### GET /api/v1/monitoring/data
- **Method**: GET
- **Response Type**: Structured monitoring data
- **Data Provided**: Comprehensive monitoring snapshot
- **Usage Notes**: Alternative to raw metrics

#### GET /api/v1/monitoring/metrics
- **Method**: GET
- **Response Type**: Monitoring metrics
- **Data Provided**: System monitoring data
- **Usage Notes**: Structured metrics

### Platform Statistics Endpoints

#### GET /api/v1/statistics/platform
- **Method**: GET
- **Response Type**: Platform-wide statistics
  ```json
  {
    "global": {
      "totalPacketsCleared": 0,
      "totalUsers": 0,
      "totalFeesCollected": "string",
      "avgClearTime": 0,
      "successRate": 0.0
    },
    "daily": {...},
    "topChannels": [...],
    "peakHours": [...]
  }
  ```
- **Data Provided**: Platform analytics
- **Usage Notes**: Public endpoint, cached for 5 minutes

#### GET /api/v1/users/statistics
- **Method**: GET
- **Authentication**: Required
- **Response Type**: User-specific statistics
- **Data Provided**: Personal clearing history and stats
- **Usage Notes**: Protected endpoint, cached for 1 minute

### Health Check Endpoint

#### GET /api/v1/health
- **Method**: GET
- **Response Type**: Service health status
  ```json
  {
    "status": "healthy",
    "database": "connected",
    "redis": "connected",
    "hermes": "available",
    "timestamp": "timestamp"
  }
  ```
- **Data Provided**: System health check
- **Usage Notes**: Used for monitoring and uptime checks

## WebSocket Endpoints

### /ws (Legacy WebSocket)
- **Protocol**: WebSocket
- **Data Provided**: Real-time system updates
- **Message Types**: Various system events
- **Usage Notes**: Legacy endpoint for backward compatibility

### /api/v1/ws (Clearing Updates WebSocket)
- **Protocol**: WebSocket
- **Authentication**: Optional
- **Message Types**:
  - subscribe: `{ type: "subscribe", token: "string" }`
  - unsubscribe: `{ type: "unsubscribe", token: "string" }`
  - clearing_update: Status updates for clearing operations
  - payment_verified: Payment confirmation events
- **Data Provided**: Real-time clearing operation updates
- **Usage Notes**: Modern WebSocket for clearing status

## Frontend Service Mappings

### api.ts (analyticsService)
- `getPlatformStatistics()` → GET /api/v1/statistics/platform
- `getNetworkFlows()` → GET /api/v1/metrics/packet-flow
- `getChannelCongestion()` → GET /api/v1/chainpulse/channels/congestion
- `getStuckPacketsAnalytics()` → GET /api/v1/metrics/stuck-packets
- `getRelayerPerformance()` → GET /api/v1/metrics/relayer-performance
- `getHistoricalTrends(timeRange)` → GET /api/v1/metrics/trends

### api.ts (metricsService)
- `getRawMetrics()` → GET /api/v1/chainpulse/metrics
- `getMonitoringMetrics()` → GET /api/monitoring/metrics
- `getMonitoringData()` → GET /api/v1/monitoring/data

### clearing.ts (ClearingService)
- `requestToken()` → POST /api/v1/clearing/request-token
- `verifyPayment()` → POST /api/v1/clearing/verify-payment
- `getStatus()` → GET /api/v1/clearing/status/:token
- `authenticateWallet()` → POST /api/v1/auth/wallet-sign
- `getUserStatistics()` → GET /api/v1/users/statistics
- `getPlatformStatistics()` → GET /api/v1/statistics/platform
- WebSocket connection → /api/v1/ws/clearing-updates

### packets.ts (packetsService)
- `getUserTransfers()` → GET /api/user/:wallet/transfers
- `getUserStuckPackets()` → GET /api/user/:wallet/stuck
- `getAllStuckPackets()` → GET /api/packets/stuck
- `clearPackets()` → POST /api/packets/clear
- `getPacketDetails()` → GET /api/v1/chainpulse/packets/:chain/:channel/:sequence
- `getChannelCongestion()` → GET /api/v1/chainpulse/channels/congestion
- `subscribeToStuckPackets()` → EventSource /api/v1/packets/stuck/stream

## Data Duplication Analysis

### Stuck Packets Data
Multiple endpoints provide stuck packet information:
1. **GET /api/v1/chainpulse/packets/stuck** - Global view
2. **GET /api/v1/ibc/packets/stuck** - Legacy endpoint
3. **GET /api/v1/metrics/stuck-packets** - Analytics view
4. **GET /api/v1/packets/stuck/stream** - Real-time stream
5. **GET /api/packets/stuck** - Frontend endpoint

**Recommendation**: Consolidate to use Chainpulse as primary source

### Channel Data
Channel information available from:
1. **GET /api/v1/ibc/channels** - Basic channel list
2. **GET /api/v1/metrics/channels** - Channel metrics
3. **GET /api/v1/chainpulse/channels/congestion** - Congestion data

**Recommendation**: Merge into comprehensive channel endpoint

### Platform Statistics
Statistics available from:
1. **GET /api/v1/statistics/platform** - Primary endpoint
2. **GET /api/v1/metrics/summary** - Similar summary data
3. **GET /api/v1/monitoring/data** - Overlapping metrics

**Recommendation**: Use /statistics/platform as canonical source

### User-Specific Data
User data endpoints:
1. **GET /api/v1/users/statistics** - Authenticated user stats
2. **GET /api/v1/chainpulse/packets/by-user** - User packets
3. **GET /api/user/:wallet/transfers** - User transfers
4. **GET /api/user/:wallet/stuck** - User stuck packets

**Recommendation**: Standardize under /api/v1/users/* namespace

## Usage Recommendations

1. **For Packet Clearing**: Use the clearing service endpoints with token-based flow
2. **For Monitoring**: Use Chainpulse integration for real-time data
3. **For Analytics**: Use platform statistics and metrics endpoints
4. **For User Data**: Authenticate with wallet signing for personalized data
5. **For Real-time Updates**: Use WebSocket connections or SSE endpoints
6. **For Caching**: Respect cache headers and use cursor-based pagination for large datasets