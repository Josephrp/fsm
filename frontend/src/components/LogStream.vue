<template>
  <section class="flex flex-col h-full space-y-2">
    <h2 class="text-xl font-semibold">Live Log</h2>
    <div ref="logContainer"
      class="flex-1 p-3 bg-gray-500 text-white rounded overflow-y-auto whitespace-pre-wrap font-mono text-sm">
      <div v-for="(line, i) in logLines" :key="i">{{ line }}</div>
    </div>
  </section>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
const logLines = ref([])
const logContainer = ref(null)

import { API_BASE } from '@/api'
const WS_BASE = API_BASE.replace(/^http/, 'ws')

onMounted(() => {
  const socket = new WebSocket(`${WS_BASE}/ws/logs`)
  socket.onmessage = (event) => {
    logLines.value.push(event.data)
    nextTick(() => {
      if (logContainer.value) {
        logContainer.value.scrollTop = logContainer.value.scrollHeight
      }
    })
  }
})

defineExpose({
  clear: () => {
    logLines.value = []
  }
})
</script>