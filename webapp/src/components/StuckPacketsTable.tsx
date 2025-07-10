import { useQuery } from '@tanstack/react-query'
import { fetchStuckPackets } from '@/services/api'

export default function StuckPacketsTable() {
  const { data: packets, isLoading } = useQuery({
    queryKey: ['stuck-packets-summary'],
    queryFn: () => fetchStuckPackets(),
  })

  if (isLoading) {
    return <div className="animate-pulse">Loading...</div>
  }

  const topStuckPackets = packets?.slice(0, 5) || []

  if (topStuckPackets.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500">
        No stuck packets detected
      </div>
    )
  }

  return (
    <div className="overflow-hidden">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Channel
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Sequence
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Duration
            </th>
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {topStuckPackets.map((packet) => (
            <tr key={packet.id}>
              <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                {packet.sourceChain} → {packet.destinationChain}
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {packet.sequence}
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {packet.stuckDuration}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {packets && packets.length > 5 && (
        <div className="bg-gray-50 px-6 py-3 text-sm text-gray-500 text-center">
          And {packets.length - 5} more...
        </div>
      )}
    </div>
  )
}