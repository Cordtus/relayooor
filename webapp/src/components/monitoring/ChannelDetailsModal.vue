<template>
  <TransitionRoot :show="true" as="template" @click="$emit('close')">
    <Dialog as="div" class="relative z-50" @close="$emit('close')">
      <TransitionChild
        as="template"
        enter="ease-out duration-300"
        enter-from="opacity-0"
        enter-to="opacity-100"
        leave="ease-in duration-200"
        leave-from="opacity-100"
        leave-to="opacity-0"
      >
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
      </TransitionChild>

      <div class="fixed inset-0 z-10 overflow-y-auto">
        <div class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
          <TransitionChild
            as="template"
            enter="ease-out duration-300"
            enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
            enter-to="opacity-100 translate-y-0 sm:scale-100"
            leave="ease-in duration-200"
            leave-from="opacity-100 translate-y-0 sm:scale-100"
            leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
          >
            <DialogPanel 
              class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-2xl sm:p-6"
              @click.stop
            >
              <div>
                <div class="flex items-center justify-between">
                  <DialogTitle as="h3" class="text-lg font-semibold leading-6 text-gray-900">
                    Channel Details
                  </DialogTitle>
                  <button
                    type="button"
                    class="rounded-md bg-white text-gray-400 hover:text-gray-500 focus:outline-none"
                    @click="$emit('close')"
                  >
                    <span class="sr-only">Close</span>
                    <XMarkIcon class="h-6 w-6" aria-hidden="true" />
                  </button>
                </div>
                
                <div class="mt-5 space-y-4">
                  <!-- Channel Info -->
                  <div class="bg-gray-50 rounded-lg p-4">
                    <h4 class="text-sm font-medium text-gray-900 mb-3">Channel Information</h4>
                    <div class="grid grid-cols-2 gap-4 text-sm">
                      <div>
                        <p class="text-gray-500">Source</p>
                        <p class="font-medium">{{ channel.srcChain }} / {{ channel.srcChannel }}</p>
                      </div>
                      <div>
                        <p class="text-gray-500">Destination</p>
                        <p class="font-medium">{{ channel.dstChain }} / {{ channel.dstChannel }}</p>
                      </div>
                      <div>
                        <p class="text-gray-500">Port</p>
                        <p class="font-medium">{{ channel.srcPort }} ↔ {{ channel.dstPort }}</p>
                      </div>
                      <div>
                        <p class="text-gray-500">Status</p>
                        <p class="font-medium">
                          <span :class="[
                            'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                            channel.status === 'active' ? 'bg-green-100 text-green-800' :
                            channel.status === 'congested' ? 'bg-yellow-100 text-yellow-800' :
                            'bg-gray-100 text-gray-800'
                          ]">
                            {{ channel.status }}
                          </span>
                        </p>
                      </div>
                    </div>
                  </div>

                  <!-- Performance Metrics -->
                  <div class="bg-gray-50 rounded-lg p-4">
                    <h4 class="text-sm font-medium text-gray-900 mb-3">Performance Metrics</h4>
                    <div class="grid grid-cols-2 gap-4 text-sm">
                      <div>
                        <p class="text-gray-500">Total Packets (24h)</p>
                        <p class="font-medium text-lg">{{ formatNumber(channel.volume24h) }}</p>
                      </div>
                      <div>
                        <p class="text-gray-500">Success Rate</p>
                        <p class="font-medium text-lg" :class="getSuccessRateClass(channel.successRate)">
                          {{ channel.successRate.toFixed(1) }}%
                        </p>
                      </div>
                      <div>
                        <p class="text-gray-500">Avg Processing Time</p>
                        <p class="font-medium text-lg">{{ channel.avgProcessingTime }}s</p>
                      </div>
                      <div>
                        <p class="text-gray-500">Effected/Uneffected</p>
                        <p class="font-medium text-lg">
                          {{ formatNumber(channel.effectedPackets || 0) }} / {{ formatNumber(channel.uneffectedPackets || 0) }}
                        </p>
                      </div>
                    </div>
                  </div>

                  <!-- Recent Activity -->
                  <div class="bg-gray-50 rounded-lg p-4">
                    <h4 class="text-sm font-medium text-gray-900 mb-3">Recent Activity</h4>
                    <div class="space-y-2">
                      <div v-for="i in 3" :key="i" class="flex items-center justify-between text-sm">
                        <span class="text-gray-600">{{ new Date(Date.now() - i * 300000).toLocaleTimeString() }}</span>
                        <span class="font-medium text-green-600">✓ Packet relayed successfully</span>
                      </div>
                    </div>
                  </div>

                  <!-- Actions -->
                  <div class="flex justify-end gap-3 pt-2">
                    <button
                      type="button"
                      class="inline-flex justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                      @click="$emit('close')"
                    >
                      Close
                    </button>
                    <button
                      type="button"
                      class="inline-flex justify-center rounded-md bg-blue-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-500"
                      @click="viewInExplorer"
                    >
                      View in Explorer
                    </button>
                  </div>
                </div>
              </div>
            </DialogPanel>
          </TransitionChild>
        </div>
      </div>
    </Dialog>
  </TransitionRoot>
</template>

<script setup lang="ts">
import { 
  TransitionRoot,
  TransitionChild,
  Dialog,
  DialogPanel,
  DialogTitle,
} from '@headlessui/vue'
import { XMarkIcon } from '@heroicons/vue/24/outline'

const props = defineProps<{
  channel: any
}>()

const emit = defineEmits<{
  close: []
}>()

function formatNumber(num: number): string {
  if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M'
  if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
  return num.toString()
}

function getSuccessRateClass(rate: number): string {
  if (rate >= 95) return 'text-green-600'
  if (rate >= 85) return 'text-yellow-600'
  return 'text-red-600'
}

function viewInExplorer() {
  // Open in Mintscan or similar explorer
  const explorerUrl = `https://www.mintscan.io/${props.channel.srcChain}/relayers/${props.channel.srcChannel}`
  window.open(explorerUrl, '_blank')
}
</script>