import axios from 'axios'

export interface PacketFlowData {
  labels: string[]
  osmosisToCosmoshub: number[]
  osmosisToSei: number[]
}

export const fetchPacketFlow = async (): Promise<PacketFlowData> => {
  // Mock data for now - would be replaced with real API call
  const now = new Date()
  const labels = []
  const osmosisToCosmoshub = []
  const osmosisToSei = []
  
  for (let i = 29; i >= 0; i--) {
    const time = new Date(now.getTime() - i * 60000)
    labels.push(time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }))
    osmosisToCosmoshub.push(Math.floor(Math.random() * 100) + 50)
    osmosisToSei.push(Math.floor(Math.random() * 80) + 30)
  }
  
  return { labels, osmosisToCosmoshub, osmosisToSei }
}