<template>
  <section class="space-y-2 rounded border px-4 py-4">
    <form @submit.prevent="send" class="flex gap-2 py-2">
      <InputText v-model="input" placeholder="/help" class="flex-grow" />
      <Button type="submit" label="Send" class="p-button-primary" />
    </form>
    <pre ref="outputRef" class="p-4 bg-black text-green-400 rounded overflow-y-auto max-h-60 whitespace-pre-wrap">
      {{ output }}
    </pre>
  </section>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { sendRconCommand } from '../api'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'

defineEmits(['send'])

const input = ref('')
const output = ref('')
const outputRef = ref<HTMLElement | null>(null)

const send = async () => {
  try {
    const result = await sendRconCommand(input.value)
    output.value += result + '\n'
    input.value = ''
  } catch (e) {
    output.value += `Error: ${e instanceof Error ? e.message : String(e)}\n`
  }

  nextTick(() => {
    if (outputRef.value) {
      outputRef.value.scrollTop = outputRef.value.scrollHeight
    }
  })
}
</script>