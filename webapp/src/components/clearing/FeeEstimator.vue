<template>
  <div class="fee-estimator">
    <div class="bg-gray-50 rounded-lg p-6">
      <h3 class="text-lg font-medium mb-4">Fee Breakdown</h3>
      
      <div class="space-y-3">
        <!-- Service fee -->
        <div class="flex justify-between items-center">
          <div>
            <p class="text-sm font-medium text-gray-700">Base Service Fee</p>
            <p class="text-xs text-gray-500">Platform fee for clearing service</p>
          </div>
          <p class="font-mono text-sm">{{ formatBaseFee() }}</p>
        </div>

        <!-- Per packet fee -->
        <div class="flex justify-between items-center">
          <div>
            <p class="text-sm font-medium text-gray-700">Per Packet Fee</p>
            <p class="text-xs text-gray-500">{{ packetCount }} packets × {{ formatPerPacketFee() }}</p>
          </div>
          <p class="font-mono text-sm">{{ formatPacketFees() }}</p>
        </div>

        <div class="border-t pt-3">
          <!-- Gas fee -->
          <div class="flex justify-between items-center">
            <div>
              <p class="text-sm font-medium text-gray-700">Estimated Gas Fee</p>
              <p class="text-xs text-gray-500">Network transaction costs</p>
            </div>
            <p class="font-mono text-sm">{{ formatGasFee() }}</p>
          </div>
        </div>

        <div class="border-t pt-3">
          <!-- Total -->
          <div class="flex justify-between items-center">
            <p class="font-semibold text-gray-900">Total Required</p>
            <p class="font-mono font-semibold text-lg">{{ formatTotal() }}</p>
          </div>
        </div>
      </div>

      <!-- Additional info -->
      <div class="mt-6 bg-blue-50 rounded-md p-4">
        <div class="flex">
          <InfoIcon class="h-5 w-5 text-blue-400 mt-0.5" />
          <div class="ml-3">
            <p class="text-sm text-blue-800">
              This is an estimate. Actual gas costs may vary slightly based on network conditions.
            </p>
            <p class="text-sm text-blue-800 mt-1">
              Any overpayment will be automatically refunded minus network fees.
            </p>
          </div>
        </div>
      </div>

      <!-- Time estimate -->
      <div class="mt-4 text-center">
        <p class="text-sm text-gray-600">
          Estimated clearing time: 
          <span class="font-medium">{{ estimatedTime }}</span>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Info as InfoIcon } from 'lucide-vue-next'
import { clearingService, type ClearingToken } from '@/services/clearing'

const props = defineProps<{
  token: ClearingToken | null
  packetCount: number
}>()

const BASE_FEE = BigInt('1000000') // 1 TOKEN
const PER_PACKET_FEE = BigInt('100000') // 0.1 TOKEN

const formatBaseFee = (): string => {
  if (!props.token) return '1.0 TOKEN'
  return clearingService.formatAmount(BASE_FEE.toString(), props.token.acceptedDenom)
}

const formatPerPacketFee = (): string => {
  if (!props.token) return '0.1 TOKEN'
  return clearingService.formatAmount(PER_PACKET_FEE.toString(), props.token.acceptedDenom)
}

const formatPacketFees = (): string => {
  if (!props.token) return `${props.packetCount * 0.1} TOKEN`
  const total = PER_PACKET_FEE * BigInt(props.packetCount)
  return clearingService.formatAmount(total.toString(), props.token.acceptedDenom)
}

const formatGasFee = (): string => {
  if (!props.token) return '~0.5 TOKEN'
  return clearingService.formatAmount(props.token.estimatedGasFee, props.token.acceptedDenom)
}

const formatTotal = (): string => {
  if (!props.token) return '~2.0 TOKEN'
  return clearingService.formatAmount(props.token.totalRequired, props.token.acceptedDenom)
}

const estimatedTime = computed(() => {
  const baseTime = 5 // seconds
  const perPacketTime = 0.5 // seconds
  const totalSeconds = baseTime + (props.packetCount * perPacketTime)
  
  if (totalSeconds < 60) {
    return `${Math.round(totalSeconds)} seconds`
  } else {
    return `${Math.round(totalSeconds / 60)} minutes`
  }
})
</script>