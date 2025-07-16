const express = require('express');
const cors = require('cors');
const app = express();

app.use(cors());
app.use(express.json());

// Chainpulse metrics endpoint
app.get('/api/v1/chainpulse/metrics', (req, res) => {
  res.setHeader('Content-Type', 'text/plain');
  res.send(`# HELP chainpulse_chains Number of chains being monitored
# TYPE chainpulse_chains gauge
chainpulse_chains 3

# HELP chainpulse_txs Total number of transactions processed
# TYPE chainpulse_txs counter
chainpulse_txs{chain_id="cosmoshub-4"} 1250000
chainpulse_txs{chain_id="osmosis-1"} 2340000
chainpulse_txs{chain_id="neutron-1"} 560000

# HELP chainpulse_packets Total number of packets processed
# TYPE chainpulse_packets counter
chainpulse_packets{chain_id="cosmoshub-4"} 450000
chainpulse_packets{chain_id="osmosis-1"} 890000
chainpulse_packets{chain_id="neutron-1"} 234000

# HELP ibc_effected_packets IBC packets effected (successfully relayed)
# TYPE ibc_effected_packets counter
ibc_effected_packets{chain_id="cosmoshub-4",src_channel="channel-141",src_port="transfer",dst_channel="channel-0",dst_port="transfer",signer="cosmos1xyz...abc",memo="hermes/1.7.3"} 118125
ibc_effected_packets{chain_id="osmosis-1",src_channel="channel-0",src_port="transfer",dst_channel="channel-141",dst_port="transfer",signer="osmo1abc...xyz",memo="rly/2.4.2"} 90454

# HELP ibc_uneffected_packets IBC packets relayed but not effected (frontrun)
# TYPE ibc_uneffected_packets counter
ibc_uneffected_packets{chain_id="cosmoshub-4",src_channel="channel-141",src_port="transfer",dst_channel="channel-0",dst_port="transfer",signer="cosmos1xyz...abc",memo=""} 6875
ibc_uneffected_packets{chain_id="osmosis-1",src_channel="channel-0",src_port="transfer",dst_channel="channel-141",dst_port="transfer",signer="osmo1abc...xyz",memo=""} 7546

# HELP ibc_stuck_packets Number of stuck packets on an IBC channel
# TYPE ibc_stuck_packets gauge
ibc_stuck_packets{src_chain="cosmoshub-4",dst_chain="osmosis-1",src_channel="channel-141"} 3
ibc_stuck_packets{src_chain="osmosis-1",dst_chain="cosmoshub-4",src_channel="channel-0"} 1
`);
});

// Monitoring data endpoint
app.get('/api/v1/monitoring/data', (req, res) => {
  res.json({
    timestamp: new Date(),
    chains: [
      {
        chainId: 'cosmoshub-4',
        chainName: 'Cosmos Hub',
        totalTxs: 1250000,
        totalPackets: 450000,
        reconnects: 2,
        timeouts: 15,
        errors: 3,
        status: 'connected',
        lastUpdate: new Date()
      },
      {
        chainId: 'osmosis-1',
        chainName: 'Osmosis',
        totalTxs: 2340000,
        totalPackets: 890000,
        reconnects: 1,
        timeouts: 8,
        errors: 1,
        status: 'connected',
        lastUpdate: new Date()
      },
      {
        chainId: 'neutron-1',
        chainName: 'Neutron',
        totalTxs: 560000,
        totalPackets: 234000,
        reconnects: 3,
        timeouts: 22,
        errors: 5,
        status: 'connected',
        lastUpdate: new Date()
      }
    ],
    top_relayers: [
      {
        address: 'cosmos1xyz...abc',
        totalPackets: 125000,
        successRate: 94.5,
        effectedPackets: 118125,
        uneffectedPackets: 6875,
        software: 'Hermes',
        version: '1.7.3'
      },
      {
        address: 'osmo1abc...xyz',
        totalPackets: 98000,
        successRate: 92.3,
        effectedPackets: 90454,
        uneffectedPackets: 7546,
        software: 'Go Relayer',
        version: '2.4.2'
      }
    ],
    recent_activity: [
      {
        from_chain: 'osmosis-1',
        to_chain: 'cosmoshub-4',
        channel: 'channel-0',
        status: 'success',
        timestamp: new Date(Date.now() - 5 * 60 * 1000)
      },
      {
        from_chain: 'cosmoshub-4',
        to_chain: 'osmosis-1',
        channel: 'channel-141',
        status: 'success',
        timestamp: new Date(Date.now() - 10 * 60 * 1000)
      }
    ],
    channels: [
      {
        src_channel: 'channel-0',
        dst_channel: 'channel-141',
        stuck_count: 3,
        oldest_stuck_age_seconds: 3600,
        total_value: {
          uosmo: '1250000000',
          uatom: '450000000'
        }
      },
      {
        src_channel: 'channel-141',
        dst_channel: 'channel-0',
        stuck_count: 1,
        oldest_stuck_age_seconds: 1800,
        total_value: {
          uatom: '890000000'
        }
      }
    ],
    system: {
      totalChains: 3,
      totalPackets: 1574000,
      totalErrors: 9,
      uptime: 99.8,
      lastSync: new Date()
    }
  });
});

// Channel congestion endpoint
app.get('/api/channels/congestion', (req, res) => {
  res.json({
    channels: [
      {
        src_channel: 'channel-0',
        dst_channel: 'channel-141',
        stuck_count: 3,
        oldest_stuck_age_seconds: 3600,
        total_value: {
          uosmo: '1250000000',
          uatom: '450000000'
        }
      },
      {
        src_channel: 'channel-141',
        dst_channel: 'channel-0',
        stuck_count: 1,
        oldest_stuck_age_seconds: 1800,
        total_value: {
          uatom: '890000000'
        }
      }
    ]
  });
});

// Platform statistics endpoint
app.get('/api/v1/statistics/platform', (req, res) => {
  res.json({
    global: {
      totalPacketsCleared: 1574000,
      totalUsers: 523,
      totalFeesCollected: '125000',
      avgClearTime: 45,
      successRate: 94.5
    },
    daily: {
      packetsCleared: 8500,
      activeUsers: 87,
      feesCollected: '1250'
    },
    topChannels: [
      {
        channel: 'channel-0 → channel-141',
        packetsCleared: 125000,
        avgClearTime: 42
      },
      {
        channel: 'channel-141 → channel-0',
        packetsCleared: 98000,
        avgClearTime: 48
      }
    ],
    peakHours: [
      { hour: 14, activity: 1250 },
      { hour: 15, activity: 1180 },
      { hour: 16, activity: 1320 }
    ]
  });
});

// Other endpoints
app.get('/api/v1/metrics/packet-flow', (req, res) => {
  res.json([
    {
      sourceChain: 'Osmosis',
      targetChain: 'Cosmos Hub',
      packetCount: 125000,
      volume: 2500000000,
      avgPacketSize: 20000
    },
    {
      sourceChain: 'Cosmos Hub',
      targetChain: 'Osmosis',
      packetCount: 98000,
      volume: 1960000000,
      avgPacketSize: 20000
    },
    {
      sourceChain: 'Neutron',
      targetChain: 'Osmosis',
      packetCount: 45000,
      volume: 900000000,
      avgPacketSize: 20000
    }
  ]);
});

app.get('/api/v1/metrics/stuck-packets', (req, res) => {
  res.json([
    {
      channelId: 'channel-0',
      sourceChain: 'osmosis-1',
      destinationChain: 'cosmoshub-4',
      stuckCount: 3,
      totalValue: 125000000,
      avgStuckTime: 3600,
      oldestPacketAge: 7200
    },
    {
      channelId: 'channel-141',
      sourceChain: 'cosmoshub-4',
      destinationChain: 'osmosis-1',
      stuckCount: 1,
      totalValue: 45000000,
      avgStuckTime: 1800,
      oldestPacketAge: 1800
    }
  ]);
});

app.get('/api/v1/metrics/relayer-performance', (req, res) => {
  res.json([
    {
      address: 'cosmos1xyz...abc',
      packetCount: 125000,
      successRate: 94.5,
      avgRelayTime: 45,
      frontrunRate: 0.096,
      gasEfficiency: 92.3,
      uptime: 99.8,
      isNew: false
    },
    {
      address: 'osmo1abc...xyz',
      packetCount: 98000,
      successRate: 92.3,
      avgRelayTime: 52,
      frontrunRate: 0.082,
      gasEfficiency: 89.5,
      uptime: 99.5,
      isNew: false
    }
  ]);
});

const PORT = 8080;
app.listen(PORT, () => {
  console.log(`Simple API server running on http://localhost:${PORT}`);
});