import { useQuery } from '@tanstack/react-query'
import { fetchMetrics } from '@/services/api'
import MetricCard from '@/components/MetricCard'
import PacketFlowChart from '@/components/PacketFlowChart'
import StuckPacketsTable from '@/components/StuckPacketsTable'

export default function Dashboard() {
  const { data: metrics, isLoading } = useQuery({
    queryKey: ['metrics'],
    queryFn: fetchMetrics,
  })

  if (isLoading) {
    return <div className="animate-pulse">Loading metrics...</div>
  }

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-gray-900">IBC Monitoring Dashboard</h2>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard
          title="Stuck Packets"
          value={metrics?.stuckPackets || 0}
          trend={metrics?.stuckPacketsTrend}
          color="red"
        />
        <MetricCard
          title="Active Channels"
          value={metrics?.activeChannels || 0}
          color="green"
        />
        <MetricCard
          title="Packet Flow Rate"
          value={`${metrics?.packetFlowRate || 0}/s`}
          color="blue"
        />
        <MetricCard
          title="Success Rate"
          value={`${metrics?.successRate || 0}%`}
          trend={metrics?.successRateTrend}
          color="green"
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white p-6 rounded-lg shadow">
          <h3 className="text-lg font-semibold mb-4">Packet Flow</h3>
          <PacketFlowChart />
        </div>
        
        <div className="bg-white p-6 rounded-lg shadow">
          <h3 className="text-lg font-semibold mb-4">Stuck Packets</h3>
          <StuckPacketsTable />
        </div>
      </div>
    </div>
  )
}