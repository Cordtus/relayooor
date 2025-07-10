export default function Settings() {
  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-gray-900">Settings</h2>
      
      <div className="bg-white shadow sm:rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h3 className="text-lg leading-6 font-medium text-gray-900">
            Monitoring Configuration
          </h3>
          <div className="mt-2 max-w-xl text-sm text-gray-500">
            <p>Configure monitoring preferences and notification settings.</p>
          </div>
          <div className="mt-5 space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">
                Refresh Interval (seconds)
              </label>
              <input
                type="number"
                defaultValue={30}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">
                Stuck Packet Threshold (minutes)
              </label>
              <input
                type="number"
                defaultValue={60}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
              />
            </div>
          </div>
        </div>
      </div>

      <div className="bg-white shadow sm:rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h3 className="text-lg leading-6 font-medium text-gray-900">
            Connected Services
          </h3>
          <div className="mt-2 max-w-xl text-sm text-gray-500">
            <p>Status of connected monitoring and relayer services.</p>
          </div>
          <div className="mt-5 space-y-3">
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium text-gray-700">Chainpulse</span>
              <span className="text-sm text-green-600">Connected</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium text-gray-700">Prometheus</span>
              <span className="text-sm text-green-600">Connected</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium text-gray-700">Relayer Middleware</span>
              <span className="text-sm text-gray-500">Not configured</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}