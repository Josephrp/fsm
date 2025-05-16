<template>
  <section class="flex flex-col h-[calc(100vh-14rem)] space-y-4 overflow-hidden">
    <div class="p-3 rounded border">
      <div class="flex items-center gap-2">
        <span class="font-medium">Status:</span>
        <span :class="props.status.running ? 'text-green-600' : 'text-red-600'">
          {{ props.status.running ? 'Running' : 'Stopped' }}
        </span>
      </div>
      <div class="flex items-center gap-2 pt-2">
        <span class="font-medium">Version:</span>
        <Dropdown v-model="selectedVersion" :options="versions" optionLabel="label" optionValue="value"
          :disabled="!versions.length || props.status.running" class="text-sm w-72" placeholder="Select version" />
        <Button v-if="isInstalled" @click="confirmUninstall" icon="pi pi-trash"
          class="ml-2 p-button-text p-button-danger" :disabled="loading || working || props.status.running" />
      </div>
    </div>
    <div class="flex flex-col flex-1 overflow-hidden gap-4">
      <div class="flex gap-4 items-center">
        <div v-if="buttonAction && (props.status.can_download || buttonAction.label !== 'Download')">
          <Button :label="buttonAction.label" :icon="`pi pi-${buttonAction.icon}`" :severity="buttonAction.severity"
            class="w-36" @click="buttonAction.handler" :disabled="loading || working" />
        </div>
        <div v-else-if="buttonAction && buttonAction.label === 'Download' && !props.status.can_download"
          class="text-sm text-red-600">
          Downloads are unavailable until a Factorio username and token are configured.
        </div>
      </div>
      <div v-if="downloadProgress !== null" class="w-full">
        <div class="relative w-full">
          <ProgressBar :value="downloadProgress" style="height: 2rem;" />
          <div class="absolute inset-0 flex items-center justify-center font-semibold pointer-events-none">
            {{ downloadStage === 'download' ? 'Downloading...' : 'Unpacking...' }}
          </div>
        </div>
      </div>
      <LogStream ref="logStreamRef" class="flex-1 overflow-y-auto max-h-full" />
    </div>
  </section>
</template>

<script setup lang="ts">
import { useAppToast } from '@/composables/useAppToast'
import LogStream from '@/components/LogStream.vue'
import { ref, watch, onMounted, computed, nextTick } from 'vue'
import { API_BASE, fetchWithAuth, startServer, stopServer } from '@/api'
import Dropdown from 'primevue/dropdown'
import Button from 'primevue/button'
import ProgressBar from 'primevue/progressbar';
import { useConfirm } from 'primevue/useconfirm'

const props = defineProps({
  status: {
    type: Object,
    required: true
  }
})
const { showSuccess, showError } = useAppToast()
const confirm = useConfirm()

onMounted(async () => {
  await fetchVersions()
  await nextTick()
  loading.value = false
})

const WS_BASE = API_BASE.replace(/^http/, 'ws')

const emit = defineEmits(['refreshStatus'])
const logStreamRef = ref<InstanceType<typeof LogStream> | null>(null)
const versions = ref<{ label: string, value: string }[]>([])
const installed = ref<Record<string, string[]>>({})
const selectedVersion = ref<string>('')

const downloadProgress = ref<number | null>(null)
const downloadStage = ref<string>('')
const loading = ref<boolean>(true)
const working = ref<boolean>(false)
let socket: WebSocket | null = null

const hasSyncedVersion = ref(false)

const confirmUninstall = () => {
  confirm.require({
    message: `Are you sure you want to delete "${selectedVersion.value}"?`,
    header: 'Confirm Deletion',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    accept: () => handleUninstall()
  })
}

const buttonAction = computed(() => {
  const [branch, version] = selectedVersion.value.split('/')
  const current = props.status
  const installedList = installed.value[branch] || []
  const installedFlag = installedList.includes(version)

  if (!version || !branch) return null

  if (!installedFlag) {
    return {
      label: 'Download',
      icon: 'download',
      severity: 'info',
      handler: downloadVersion
    }
  }

  const isCurrent =
    current?.version?.version === version &&
    current?.version?.branch === branch

  if (!isCurrent) {
    return {
      label: 'Use',
      icon: 'check-square',
      severity: 'warn',
      handler: async () => {
        working.value = true
        try {
          const res = await fetchWithAuth(`${API_BASE}/factorio-versions/${branch}/${version}`, {
            method: 'PUT'
          })
          if (!res.ok) throw new Error('Switch version failed')
        } catch (e) {
          showError(e instanceof Error ? e.message : String(e))
        }
        emit('refreshStatus')
        working.value = false
      }
    }
  }

  if (isCurrent && !current.running && props.status.running === false) {
    return {
      label: 'Start',
      icon: 'play',
      severity: 'success',
      handler: async () => {
        onStart()
        await nextTick()
        if (logStreamRef.value && typeof logStreamRef.value.clear === 'function') {
          logStreamRef.value.clear()
        }
      }
    }
  }

  return {
    label: 'Stop',
    icon: 'stop',
    severity: 'danger',
    handler: () => onStop()
  }
})

const handleUninstall = async () => {
  working.value = true
  try {
    const [branch, version] = selectedVersion.value.split('/')
    const res = await fetchWithAuth(`${API_BASE}/factorio-versions/${branch}/${version}`, {
      method: 'DELETE'
    })
    if (!res.ok) throw new Error('Uninstall failed')
    await fetchVersions()
    emit('refreshStatus')
  } catch (err) {
    showError('Uninstall failed')
  }
  working.value = false
}

const syncSelectedVersionWithStatus = () => {
  const statusVersion = props.status.version.version
  const statusBranch = props.status.version.branch

  if (!statusVersion || !statusBranch) return

  const combined = `${statusBranch}/${statusVersion}`

  const validOptions = versions.value.map(v => v.value)
  if (!validOptions.includes(selectedVersion.value) || selectedVersion.value !== combined) {
    selectedVersion.value = combined
  }
}

const fetchVersions = async () => {
  const res = await fetchWithAuth(`${API_BASE}/factorio-versions`)
  const data = await res.json()

  const stable = data?.available?.stable?.headless
  const experimental = data?.available?.experimental?.headless

  installed.value = data.installed || {}

  versions.value = []

  Object.entries(installed.value).forEach(([branch, versionsArr]) => {
    versionsArr.forEach(ver => {
      versions.value.push({ label: `${ver} (${branch}, installed)`, value: `${branch}/${ver}` })
    })
  })

  if (stable) {
    if (!(installed.value.stable || []).includes(stable)) {
      versions.value.push({ label: `${stable} (stable, available)`, value: `stable/${stable}` })
    }
  }

  if (experimental) {
    if (!(installed.value.experimental || []).includes(experimental)) {
      versions.value.push({ label: `${experimental} (experimental, available)`, value: `experimental/${experimental}` })
    }
  }
}

const isInstalled = computed(() => {
  const [branch, version] = selectedVersion.value.split('/')
  return installed.value[branch]?.includes(version)
})

const needsSwitch = computed(() => {
  const [branch, version] = selectedVersion.value.split('/')
  if (!version || !branch) return false
  if (!props.status.version && !props.status.branch) return false
  return (
    props.status.version !== version ||
    props.status.branch !== branch
  )
})

const downloadVersion = async () => {
  working.value = true
  downloadProgress.value = 0
  downloadStage.value = ''

  const [branch, version] = selectedVersion.value.split('/')
  socket = new WebSocket(`${WS_BASE}/ws/download/${branch}/${version}`)
  socket.onmessage = async (event) => {
    const data = JSON.parse(event.data)
    if (data.type === 'progress') {
      downloadProgress.value = data.percent
      downloadStage.value = data.stage

      if (data.stage === 'done') {
        socket?.close()
        socket = null
        downloadProgress.value = null
        downloadStage.value = ''
        await fetchVersions()
        working.value = false
        showSuccess('Server download completed')
      }
    }
  }
  socket.onclose = async () => {
  }

  try {
    const res = await fetchWithAuth(`${API_BASE}/factorio-versions/${branch}/${version}/download`, {
      method: 'GET'
    })
    if (!res.ok) throw new Error('Download failed')
    emit('refreshStatus')
  } catch (e) {
    showError(e instanceof Error ? e.message : String(e))
    downloadProgress.value = null
    downloadStage.value = ''
    if (socket) socket.close()
  }
}

const onStart = async () => {
  try {
    await startServer()
    emit('refreshStatus')
  } catch (e) {
    showError(e instanceof Error ? e.message : String(e))
  }

  await nextTick()
  if (logStreamRef.value && typeof logStreamRef.value.clear === 'function') {
    logStreamRef.value.clear()
  }
}

const onStop = async () => {
  try {
    await stopServer()
    emit('refreshStatus')
  } catch (e) {
    showError(e instanceof Error ? e.message : String(e))
  }
}

watch(
  () => props.status,
  (newStatus) => {
    if (
      !hasSyncedVersion.value &&
      newStatus?.version?.version &&
      newStatus?.version?.branch
    ) {
      syncSelectedVersionWithStatus()
      hasSyncedVersion.value = true
    }
  },
  { immediate: true }
)
</script>