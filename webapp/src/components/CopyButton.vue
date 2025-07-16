<template>
  <button
    @click.stop="copyToClipboard"
    class="p-1 text-gray-400 hover:text-gray-600 transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 rounded"
    :title="copied ? 'Copied!' : 'Copy to clipboard'"
  >
    <Check v-if="copied" class="w-3 h-3 text-green-600" />
    <Copy v-else class="w-3 h-3" />
  </button>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Copy, Check } from 'lucide-vue-next'
import { useToast } from 'vue-toastification'

const props = defineProps<{
  text: string
}>()

const toast = useToast()
const copied = ref(false)

const copyToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(props.text)
    copied.value = true
    toast.success('Copied to clipboard', { timeout: 2000 })
    
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch (error) {
    toast.error('Failed to copy')
  }
}
</script>