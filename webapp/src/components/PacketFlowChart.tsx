import { useEffect, useRef } from 'react'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js'
import { Line } from 'react-chartjs-2'
import { useQuery } from '@tanstack/react-query'
import { fetchPacketFlow } from '@/services/metrics'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
)

export default function PacketFlowChart() {
  const { data: flowData } = useQuery({
    queryKey: ['packet-flow'],
    queryFn: fetchPacketFlow,
    refetchInterval: 30000,
  })

  const options = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'top' as const,
      },
      title: {
        display: false,
      },
    },
    scales: {
      y: {
        beginAtZero: true,
        title: {
          display: true,
          text: 'Packets/min',
        },
      },
    },
  }

  const data = {
    labels: flowData?.labels || [],
    datasets: [
      {
        label: 'Osmosis → Cosmos Hub',
        data: flowData?.osmosisToCosmoshub || [],
        borderColor: 'rgb(59, 130, 246)',
        backgroundColor: 'rgba(59, 130, 246, 0.5)',
      },
      {
        label: 'Osmosis → Sei',
        data: flowData?.osmosisToSei || [],
        borderColor: 'rgb(34, 197, 94)',
        backgroundColor: 'rgba(34, 197, 94, 0.5)',
      },
    ],
  }

  return (
    <div className="h-64">
      <Line options={options} data={data} />
    </div>
  )
}