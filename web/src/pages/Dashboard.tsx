import React, { useEffect, useState } from 'react'
import { Grid, Paper, Typography, Box, CircularProgress, Alert, Chip } from '@mui/material'
import { useQuery } from 'react-query'
import axios from 'axios'

interface StuckPacket {
  path: string
  channel: string
  count: number
}

interface RelayerPerformance {
  effected_packets: number
  uneffected_packets: number
  frontrun_events: number
  stuck_packets: number
  success_rate: number
}

interface ChainpulseMetrics {
  packet_flow: {
    effected: Record<string, number>
    uneffected: Record<string, number>
  }
  stuck_packets: StuckPacket[]
}

export default function Dashboard() {
  const [error, setError] = useState<string | null>(null)

  // Fetch relayer performance
  const { data: performance, isLoading: perfLoading } = useQuery<RelayerPerformance>(
    'relayerPerformance',
    async () => {
      const response = await axios.get('/api/metrics/relayer-performance')
      return response.data
    },
    { refetchInterval: 15000 } // Refresh every 15 seconds
  )

  // Fetch stuck packets
  const { data: stuckData, isLoading: stuckLoading } = useQuery(
    'stuckPackets',
    async () => {
      const response = await axios.get('/api/metrics/stuck-packets')
      return response.data
    },
    { refetchInterval: 30000 } // Refresh every 30 seconds
  )

  // Fetch chainpulse metrics
  const { data: chainpulse, isLoading: chainpulseLoading } = useQuery<ChainpulseMetrics>(
    'chainpulseMetrics',
    async () => {
      try {
        const response = await axios.get('/api/metrics/chainpulse')
        return response.data
      } catch (err) {
        setError('Failed to connect to Chainpulse. Make sure it is running.')
        throw err
      }
    },
    { refetchInterval: 15000 }
  )

  const isLoading = perfLoading || stuckLoading || chainpulseLoading

  if (error) {
    return (
      <Box>
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      </Box>
    )
  }

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        IBC Relayer Dashboard
      </Typography>
      
      {isLoading ? (
        <Box display="flex" justifyContent="center" p={4}>
          <CircularProgress />
        </Box>
      ) : (
        <Grid container spacing={3}>
          {/* Performance Overview */}
          <Grid item xs={12} md={6} lg={3}>
            <Paper sx={{ p: 2 }}>
              <Typography variant="h6" color="primary">Success Rate</Typography>
              <Typography variant="h3">
                {performance?.success_rate?.toFixed(1) || 0}%
              </Typography>
              <Typography variant="caption" color="text.secondary">
                Packet relay success
              </Typography>
            </Paper>
          </Grid>

          <Grid item xs={12} md={6} lg={3}>
            <Paper sx={{ p: 2 }}>
              <Typography variant="h6" color="success.main">Effected Packets</Typography>
              <Typography variant="h3">
                {performance?.effected_packets || 0}
              </Typography>
              <Typography variant="caption" color="text.secondary">
                Successfully relayed
              </Typography>
            </Paper>
          </Grid>

          <Grid item xs={12} md={6} lg={3}>
            <Paper sx={{ p: 2 }}>
              <Typography variant="h6" color="warning.main">Stuck Packets</Typography>
              <Typography variant="h3">
                {stuckData?.total || 0}
              </Typography>
              <Typography variant="caption" color="text.secondary">
                Requiring attention
              </Typography>
            </Paper>
          </Grid>

          <Grid item xs={12} md={6} lg={3}>
            <Paper sx={{ p: 2 }}>
              <Typography variant="h6" color="error">Frontrun Events</Typography>
              <Typography variant="h3">
                {performance?.frontrun_events || 0}
              </Typography>
              <Typography variant="caption" color="text.secondary">
                Lost to other relayers
              </Typography>
            </Paper>
          </Grid>

          {/* Active Channels */}
          <Grid item xs={12} md={6}>
            <Paper sx={{ p: 2 }}>
              <Typography variant="h6" gutterBottom>
                Active IBC Channels
              </Typography>
              <Box sx={{ maxHeight: 300, overflow: 'auto' }}>
                {chainpulse?.packet_flow?.effected && 
                  Object.entries(chainpulse.packet_flow.effected).map(([channel, count]) => (
                    <Box key={channel} sx={{ p: 1, borderBottom: '1px solid rgba(255,255,255,0.1)' }}>
                      <Typography variant="body2">{channel}</Typography>
                      <Typography variant="caption" color="text.secondary">
                        {count} packets relayed
                      </Typography>
                    </Box>
                  ))
                }
                {(!chainpulse?.packet_flow?.effected || 
                  Object.keys(chainpulse.packet_flow.effected).length === 0) && (
                  <Typography variant="body2" color="text.secondary">
                    No active channels detected
                  </Typography>
                )}
              </Box>
            </Paper>
          </Grid>

          {/* Stuck Packets */}
          <Grid item xs={12} md={6}>
            <Paper sx={{ p: 2 }}>
              <Typography variant="h6" gutterBottom>
                Stuck Packets
              </Typography>
              <Box sx={{ maxHeight: 300, overflow: 'auto' }}>
                {stuckData?.stuck_packets?.map((packet: StuckPacket, index: number) => (
                  <Box key={index} sx={{ p: 1, borderBottom: '1px solid rgba(255,255,255,0.1)' }}>
                    <Box display="flex" justifyContent="space-between" alignItems="center">
                      <Box>
                        <Typography variant="body2">{packet.path}</Typography>
                        <Typography variant="caption" color="text.secondary">
                          Channel: {packet.channel}
                        </Typography>
                      </Box>
                      <Chip 
                        label={`${packet.count} stuck`} 
                        color="warning" 
                        size="small"
                      />
                    </Box>
                  </Box>
                ))}
                {(!stuckData?.stuck_packets || stuckData.stuck_packets.length === 0) && (
                  <Typography variant="body2" color="text.secondary">
                    No stuck packets detected
                  </Typography>
                )}
              </Box>
            </Paper>
          </Grid>

          {/* Real-time Status */}
          <Grid item xs={12}>
            <Paper sx={{ p: 2 }}>
              <Typography variant="h6" gutterBottom>
                System Status
              </Typography>
              <Grid container spacing={2}>
                <Grid item xs={12} md={4}>
                  <Box display="flex" alignItems="center" gap={1}>
                    <Box 
                      width={12} 
                      height={12} 
                      borderRadius="50%" 
                      bgcolor={chainpulse ? "success.main" : "error.main"}
                    />
                    <Typography>Chainpulse Monitor</Typography>
                  </Box>
                </Grid>
                <Grid item xs={12} md={4}>
                  <Box display="flex" alignItems="center" gap={1}>
                    <Box 
                      width={12} 
                      height={12} 
                      borderRadius="50%" 
                      bgcolor="warning.main"
                    />
                    <Typography>Hermes Relayer (Not Connected)</Typography>
                  </Box>
                </Grid>
                <Grid item xs={12} md={4}>
                  <Box display="flex" alignItems="center" gap={1}>
                    <Box 
                      width={12} 
                      height={12} 
                      borderRadius="50%" 
                      bgcolor="warning.main"
                    />
                    <Typography>Go Relayer (Not Connected)</Typography>
                  </Box>
                </Grid>
              </Grid>
            </Paper>
          </Grid>
        </Grid>
      )}
    </Box>
  )
}