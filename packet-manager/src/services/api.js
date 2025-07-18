import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000
})

export const packetService = {
  // Get stuck packets from Chainpulse
  async getStuckPackets(chainId, channelId) {
    const params = new URLSearchParams()
    if (chainId) params.append('chain', chainId)
    if (channelId) params.append('channel', channelId)
    
    const response = await api.get(`/ibc/packets/stuck?${params}`)
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
    const response = await api.post('/relayer/hermes/clear', {
      chain: chainId,
      channel: channelId,
      port: port || 'transfer',
      sequences: [sequence]
    })
    return response.data
  },

  // Clear multiple packets
  async clearPackets(chainId, channelId, port, sequences) {
    const response = await api.post('/relayer/hermes/clear', {
      chain: chainId,
      channel: channelId,
      port: port || 'transfer',
      sequences: sequences
    })
    return response.data
  },

  // Get available chains
  async getChains() {
    const response = await api.get('/ibc/chains')
    return response.data
  },

  // Get channels for a chain
  async getChannels(chainId) {
    const response = await api.get(`/ibc/chains/${chainId}/channels`)
    return response.data
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