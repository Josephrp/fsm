<template>
  <div class="rounded border px-4 py-4">
    <h2 class="text-lg font-semibold mb-4">Factorio User Settings</h2>
    <form @submit.prevent="submit">
      <div class="mb-4">
        <label class="block text-sm font-medium mb-1" for="username">Username</label>
        <InputText id="username" v-model="username" class="w-75" autocomplete="off" />
      </div>
      <div class="mb-4">
        <label class="block text-sm font-medium mb-1" for="token">Token</label>
        <Password id="token" v-model="token" toggleMask class="w-75" inputClass="w-full" autocomplete="off" />
      </div>
      <Button label="Update" type="submit" :disabled="!formValid" class="bg-blue-600 text-white" />
    </form>
  </div>
</template>

<script setup lang="ts">
import { getFactorioUser, updateFactorioUser } from '@/api'
import { useAppToast } from '@/composables/useAppToast'
import { ref, computed, onMounted } from 'vue'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Password from 'primevue/password'

const username = ref('')
const token = ref('')
const originalUsername = ref('')
const originalToken = ref('')
const { showError, showSuccess } = useAppToast()

const formValid = computed(() => {
  return (
    username.value &&
    token.value &&
    (username.value !== originalUsername.value || token.value !== originalToken.value)
  )
})

const loadSettings = async () => {
  try {
    const data = await getFactorioUser()
    username.value = data.username
    token.value = data.token
    originalUsername.value = data.username
    originalToken.value = data.token
  } catch (e) {
    showError(e instanceof Error ? e.message : String(e))
  }
}

const submit = async () => {
  try {
    await updateFactorioUser(username.value, token.value)
    originalUsername.value = username.value
    originalToken.value = token.value
    showSuccess('Factorio user updated')
  } catch (e) {
    showError(e instanceof Error ? e.message : String(e))
  }
}

onMounted(loadSettings)

defineExpose({ InputText, Button, Password })
</script>