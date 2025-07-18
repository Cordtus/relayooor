import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000
})

export const packetService = {
  // Get stuck packets from Chainpulse
  async getStuckPackets(chainId, channelId) {
    const params = new URLSearchParams()
    if (chainId) params.append('chain', chainId)
    if (channelId) params.append('channel', channelId)
    
    const response = await api.get(`/packets/stuck?${params}`)
    return response.data
  },

  // Get pending packets from Hermes
  async getPendingPackets(chainId, channelId) {
    const response = await api.get(`/chain/${chainId}`)
    return response.data
  },

  // Get channel details
  async getChannelDetails(chainId, channelId) {
    const response = await api.get(`/ibc/channels/${channelId}`)
    return response.data
  },

  // Clear individual packet
  async clearPacket(chainId, channelId, port, sequence) {
    // For simple API, simulate clearing
    console.log(`Would clear packet ${sequence} on ${chainId}/${channelId}`)
    return {
      success: true,
      message: `Packet ${sequence} clearing initiated (simulated)`
    }
  },

  // Clear multiple packets
  async clearPackets(chainId, channelId, port, sequences) {
    // For simple API, simulate clearing
    console.log(`Would clear packets ${sequences.join(', ')} on ${chainId}/${channelId}`)
    return {
      success: true,
      message: `${sequences.length} packets clearing initiated (simulated)`
    }
  },

  // Get available chains
  async getChains() {
    // For simple API, return hardcoded chains
    return [
      { id: 'cosmoshub-4', name: 'Cosmos Hub' },
      { id: 'osmosis-1', name: 'Osmosis' },
      { id: 'noble-1', name: 'Noble' },
      { id: 'stride-1', name: 'Stride' },
      { id: 'jackal-1', name: 'Jackal' },
      { id: 'axelar-dojo-1', name: 'Axelar' }
    ]
  },

  // Get channels for a chain
  async getChannels(chainId) {
    // For simple API, return common IBC channels
    const channels = {
      'cosmoshub-4': [
        { id: 'channel-0', state: 'OPEN', port: 'transfer' },
        { id: 'channel-141', state: 'OPEN', port: 'transfer' },
        { id: 'channel-207', state: 'OPEN', port: 'transfer' }
      ],
      'osmosis-1': [
        { id: 'channel-0', state: 'OPEN', port: 'transfer' },
        { id: 'channel-1', state: 'OPEN', port: 'transfer' },
        { id: 'channel-42', state: 'OPEN', port: 'transfer' }
      ],
      'noble-1': [
        { id: 'channel-0', state: 'OPEN', port: 'transfer' },
        { id: 'channel-1', state: 'OPEN', port: 'transfer' },
        { id: 'channel-2', state: 'OPEN', port: 'transfer' }
      ],
      'stride-1': [
        { id: 'channel-0', state: 'OPEN', port: 'transfer' },
        { id: 'channel-1', state: 'OPEN', port: 'transfer' },
        { id: 'channel-2', state: 'OPEN', port: 'transfer' }
      ],
      'jackal-1': [
        { id: 'channel-0', state: 'OPEN', port: 'transfer' },
        { id: 'channel-1', state: 'OPEN', port: 'transfer' },
        { id: 'channel-2', state: 'OPEN', port: 'transfer' },
        { id: 'channel-11', state: 'OPEN', port: 'transfer' },
        { id: 'channel-12', state: 'OPEN', port: 'transfer' }
      ],
      'axelar-dojo-1': [
        { id: 'channel-0', state: 'OPEN', port: 'transfer' },
        { id: 'channel-1', state: 'OPEN', port: 'transfer' },
        { id: 'channel-2', state: 'OPEN', port: 'transfer' },
        { id: 'channel-3', state: 'OPEN', port: 'transfer' },
        { id: 'channel-4', state: 'OPEN', port: 'transfer' },
        { id: 'channel-5', state: 'OPEN', port: 'transfer' },
        { id: 'channel-208', state: 'OPEN', port: 'transfer' }
      ]
    }
    return channels[chainId] || []
  },

  // Get Hermes metrics for validation
  async getHermesMetrics() {
    try {
      const response = await axios.get('http://localhost:3010/metrics')
      return response.data
    } catch (error) {
      console.error('Failed to fetch Hermes metrics:', error)
      return null
    }
  },

  // Parse Hermes metrics for pending packets
  parseHermesMetrics(metricsText, chainId, channelId) {
    if (!metricsText) return []
    
    const lines = metricsText.split('\n')
    const pendingPackets = []
    
    lines.forEach(line => {
      if (line.includes('hermes_pending_packets') && !line.startsWith('#')) {
        // Example: hermes_pending_packets{chain_id="cosmoshub-4",channel_id="channel-0",port_id="transfer"} 5
        const match = line.match(/hermes_pending_packets{chain_id="([^"]+)",channel_id="([^"]+)",port_id="([^"]+)"} (\d+)/)
        if (match) {
          const [_, chain, channel, port, count] = match
          if ((!chainId || chain === chainId) && (!channelId || channel === channelId)) {
            pendingPackets.push({
              chain_id: chain,
              channel_id: channel,
              port_id: port,
              pending_count: parseInt(count)
            })
          }
        }
      }
    })
    
    return pendingPackets
  }
}

export default api