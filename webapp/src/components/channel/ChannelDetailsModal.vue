<template>
  <v-dialog
    :model-value="modelValue"
    @update:model-value="$emit('update:modelValue', $event)"
    max-width="600px"
  >
    <v-card>
      <v-card-title class="text-h5">
        Channel Details
        <v-spacer></v-spacer>
        <v-btn
          icon
          variant="text"
          @click="$emit('update:modelValue', false)"
        >
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-card-title>
      <v-card-text>
        <v-container class="text-center py-8">
          <v-icon size="64" color="grey-lighten-1" class="mb-4">
            mdi-connection
          </v-icon>
          <h3 class="text-h6 text-grey-darken-1 mb-2">Coming Soon</h3>
          <p class="text-body-2 text-grey">
            Channel details and configuration will be available here.
          </p>
        </v-container>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup>
defineProps({
  modelValue: {
    type: Boolean,
    default: false
  }
})

defineEmits(['update:modelValue'])
</script>