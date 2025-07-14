// Mock data generator for development
export const mockData = {
  generateChannels() {
    return [
      {
        channelId: 'channel-0',
        counterpartyChannelId: 'channel-141',
        srcChain: 'osmosis-1',
        dstChain: 'cosmoshub-4',
        srcChannel: 'channel-0',
        dstChannel: 'channel-141',
        srcPort: 'transfer',
        dstPort: 'transfer',
        state: 'OPEN',
        volume24h: 1234567,
        successRate: 94.3,
        avgProcessingTime: 12.5,
        totalPackets: 15234,
        effectedPackets: 14321,
        uneffectedPackets: 913,
        pendingPackets: 0,
        status: 'active' as const,
        stuck_count: 0,
        total_value: {
          uosmo: '1234567890',
          uatom: '987654321'
        }
      },
      {
        channelId: 'channel-141',
        counterpartyChannelId: 'channel-0',
        srcChain: 'cosmoshub-4',
        dstChain: 'osmosis-1',
        srcChannel: 'channel-141',
        dstChannel: 'channel-0',
        srcPort: 'transfer',
        dstPort: 'transfer',
        state: 'OPEN',
        volume24h: 987654,
        successRate: 91.2,
        avgProcessingTime: 15.3,
        totalPackets: 18765,
        effectedPackets: 17102,
        uneffectedPackets: 1663,
        pendingPackets: 2,
        status: 'active' as const,
        stuck_count: 2,
        oldest_stuck_age_seconds: 3600,
        total_value: {
          uatom: '2345678901',
          uosmo: '876543210'
        }
      },
      {
        channelId: 'channel-10',
        counterpartyChannelId: 'channel-874',
        srcChain: 'neutron-1',
        dstChain: 'osmosis-1',
        srcChannel: 'channel-10',
        dstChannel: 'channel-874',
        srcPort: 'transfer',
        dstPort: 'transfer',
        state: 'OPEN',
        volume24h: 456789,
        successRate: 88.7,
        avgProcessingTime: 18.2,
        totalPackets: 8234,
        effectedPackets: 7302,
        uneffectedPackets: 932,
        pendingPackets: 1,
        status: 'congested' as const,
        stuck_count: 1,
        oldest_stuck_age_seconds: 1800,
        total_value: {
          untrn: '456789012'
        }
      }
    ]
  },

  generateRelayers() {
    const addresses = [
      'osmo1f269n4mrg0s8tqveny9huulyamvdv97n87xa7f',
      'osmo1evdjzy3w9t2yu54w4dhc2cvrlc2fvnptc9nqa2',
      'osmo1ks0qeq9vyt9l7vgasaajd49ff0k8klur3p2jrp',
      'osmo19kzuzfmmy9wjr3cl0ss8wjzjup9g49hqwnkfuk',
      'osmo1nna7k5lywn99cd63elcfqm6p8c5c4qcuqwwflx',
      'osmo1p7d8mnjttcszv34pk2a5yyug3474mhff4twwa6'
    ]
    
    const relayers = addresses.map((address, index) => ({
      address,
      memo: `Relayer ${index + 1} | hermes 1.13.${index}`,
      software: 'hermes',
      version: `1.13.${index}`,
      totalPackets: Math.floor(Math.random() * 10000) + 1000,
      effectedPackets: Math.floor(Math.random() * 9000) + 900,
      uneffectedPackets: Math.floor(Math.random() * 1000) + 100,
      successRate: 85 + Math.random() * 10,
      marketShare: 0,
      frontrunCount: Math.floor(Math.random() * 50),
      isActive: Math.random() > 0.2
    }))
    
    // Calculate market shares
    const totalPackets = relayers.reduce((sum, r) => sum + r.totalPackets, 0)
    relayers.forEach(r => {
      r.marketShare = (r.totalPackets / totalPackets) * 100
    })
    
    return relayers.sort((a, b) => b.totalPackets - a.totalPackets)
  },

  generateMetrics() {
    const channels = this.generateChannels()
    const relayers = this.generateRelayers()
    
    return {
      system: {
        totalChains: 3,
        totalChannels: channels.length,
        totalRelayers: relayers.length,
        totalPackets: channels.reduce((sum, c) => sum + c.totalPackets, 0),
        totalErrors: Math.floor(Math.random() * 100),
        totalTransactions: Math.floor(Math.random() * 50000) + 10000,
        uptime: 99.9,
        lastSync: new Date()
      },
      chains: [
        { 
          chainId: 'osmosis-1', 
          chainName: 'Osmosis',
          status: 'connected' as const,
          totalPackets: 23499,
          totalTxs: 112543,
          errors: 45,
          reconnects: 2,
          timeouts: 8,
          lastUpdate: new Date()
        },
        { 
          chainId: 'cosmoshub-4', 
          chainName: 'Cosmos Hub',
          status: 'connected' as const,
          totalPackets: 18765,
          totalTxs: 98234,
          errors: 23,
          reconnects: 1,
          timeouts: 5,
          lastUpdate: new Date()
        },
        { 
          chainId: 'neutron-1', 
          chainName: 'Neutron',
          status: 'connected' as const,
          totalPackets: 8234,
          totalTxs: 45123,
          errors: 67,
          reconnects: 5,
          timeouts: 12,
          lastUpdate: new Date()
        }
      ],
      channels,
      relayers,
      stuckPackets: [],
      frontrunEvents: [],
      recentPackets: [],
      timestamp: new Date()
    }
  },

  generateMonitoringData() {
    const channels = this.generateChannels()
    return {
      chains: [
        { 
          chainId: 'osmosis-1', 
          chainName: 'Osmosis', 
          status: 'active' as const,
          channels: channels.filter(c => c.srcChain === 'osmosis-1' || c.dstChain === 'osmosis-1'),
          lastUpdate: new Date()
        },
        { 
          chainId: 'cosmoshub-4', 
          chainName: 'Cosmos Hub', 
          status: 'active' as const,
          channels: channels.filter(c => c.srcChain === 'cosmoshub-4' || c.dstChain === 'cosmoshub-4'),
          lastUpdate: new Date()
        },
        { 
          chainId: 'neutron-1', 
          chainName: 'Neutron', 
          status: 'syncing' as const,
          channels: channels.filter(c => c.srcChain === 'neutron-1' || c.dstChain === 'neutron-1'),
          lastUpdate: new Date()
        }
      ],
      channels,
      totalPackets24h: 52498,
      successRate: 92.4,
      activeRelayers: 15,
      timestamp: new Date()
    }
  }
}