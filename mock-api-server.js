const express = require('express')
const cors = require('cors')
const app = express()

app.use(cors())
app.use(express.json())

// Health check
app.get('/api/health', (req, res) => {
  res.json({ status: 'ok' })
})

// Metrics
app.get('/api/metrics', (req, res) => {
  res.json({
    stuckPackets: 100,
    activeChannels: 77,
    packetFlowRate: 12.5,
    successRate: 87.3
  })
})

// Stuck packets
app.get('/api/packets/stuck', (req, res) => {
  res.json([
    {
      id: 'packet-1',
      channelId: 'channel-165',
      sequence: 12345,
      sourceChain: 'osmosis-1',
      destinationChain: 'cosmoshub-4',
      stuckDuration: '2h',
      amount: '1000000',
      denom: 'uosmo',
      sender: 'osmo1abc123',
      receiver: 'cosmos1xyz789',
      timestamp: new Date(Date.now() - 2 * 60 * 60 * 1000)
    }
  ])
})

// User transfers
app.get('/api/user/:wallet/transfers', (req, res) => {
  res.json([
    {
      id: 'transfer-1',
      channelId: 'channel-0',
      sequence: 12345,
      sourceChain: 'osmosis-1',
      destinationChain: 'cosmoshub-4',
      amount: '1000000',
      denom: 'uosmo',
      sender: req.params.wallet,
      receiver: 'cosmos1xyz789',
      status: 'stuck',
      timestamp: new Date(Date.now() - 2 * 60 * 60 * 1000),
      txHash: 'ABC123DEF456',
      stuckDuration: '2h'
    }
  ])
})

// Clear packets
app.post('/api/packets/clear', (req, res) => {
  res.json({
    status: 'success',
    txHash: 'CLEAR_TX_' + Date.now(),
    cleared: req.body.packetIds || [],
    failed: [],
    message: 'Packets cleared successfully'
  })
})

const PORT = 8080
app.listen(PORT, () => {
  console.log(`Mock API server running on http://localhost:${PORT}`)
})