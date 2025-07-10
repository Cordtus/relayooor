import { useQuery } from '@tanstack/react-query'
import { fetchChannels } from '@/services/api'

export default function Channels() {
  const { data: channels, isLoading } = useQuery({
    queryKey: ['channels'],
    queryFn: fetchChannels,
  })

  if (isLoading) {
    return <div className="animate-pulse">Loading channels...</div>
  }

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-gray-900">IBC Channels</h2>
      
      <div className="bg-white shadow overflow-hidden sm:rounded-md">
        <ul className="divide-y divide-gray-200">
          {channels?.map((channel) => (
            <li key={channel.channelId}>
              <div className="px-4 py-4 sm:px-6">
                <div className="flex items-center justify-between">
                  <div className="flex items-center">
                    <div className="flex-shrink-0">
                      <div className={`h-2 w-2 rounded-full ${
                        channel.state === 'OPEN' ? 'bg-green-400' : 'bg-red-400'
                      }`} />
                    </div>
                    <div className="ml-4">
                      <div className="text-sm font-medium text-gray-900">
                        {channel.sourceChain} → {channel.destinationChain}
                      </div>
                      <div className="text-sm text-gray-500">
                        {channel.channelId} / {channel.counterpartyChannelId}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center space-x-4">
                    <div className="text-sm text-gray-500">
                      <span className="font-medium">{channel.pendingPackets}</span> pending
                    </div>
                    <div className="text-sm text-gray-500">
                      <span className="font-medium">{channel.totalPackets}</span> total
                    </div>
                  </div>
                </div>
              </div>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}