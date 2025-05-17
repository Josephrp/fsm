<template>
  <section class="flex-1 h-[calc(90vh-11rem)] overflow-y-auto">
    <div v-if="settings === null" class="p-6 bg-yellow-100 border border-yellow-400 text-yellow-800 rounded mb-4">
      <p class="mb-2 font-semibold">Server settings are unavailable</p>
      <p class="mb-4">To view or change settings, please start the server.</p>
    </div>
    <form v-else @submit.prevent="saveSettings" class="space-y-4 rounded border px-4 py-6">
      <div v-for="(value, key) in settings" :key="key" class="grid grid-cols-[16rem_1fr] gap-4 items-start w-full">
        <label class="capitalize text-md flex items-center justify-between gap-1 w-64 relative">
          <span class="truncate">{{ key }}</span>
          <span v-if="comments[key]" class="cursor-pointer flex-shrink-0 relative" @click.stop="toggleTooltip(key)"
            @mouseenter="visibleTooltip = key" @mouseleave="visibleTooltip = null">
            <span class="inline-block w-4 h-4 text-center border rounded-full text-xs leading-4">
              ?
            </span>
            <div v-if="visibleTooltip === key"
              class="absolute top-full mt-1 right-0 z-50 bg-gray-800 text-white text-xs rounded px-2 py-1 whitespace-pre-wrap max-w-xl shadow pointer-events-auto">
              {{ comments[key] }}
            </div>
          </span>
        </label>
        <div class="flex-1">
          <input v-if="typeof value === 'string'" type="text" v-model="settings[key]"
            class="w-full p-1 border rounded" />
          <input v-else-if="typeof value === 'number'" type="number" v-model.number="settings[key]"
            class="w-32 p-1 border rounded" />
          <select v-else-if="typeof value === 'boolean'" v-model="settings[key]" class="p-1 border rounded">
            <option :value="true">true</option>
            <option :value="false">false</option>
          </select>
          <div v-else-if="Array.isArray(value) && key === 'tags'"
            class="w-full flex flex-wrap items-center gap-2 border p-2 rounded">
            <span v-for="(tag, i) in settings[key]" :key="i"
              class="bg-blue-100 text-blue-800 px-2 py-1 text-xs rounded flex items-center gap-1">
              {{ tag }}
              <button @click="removeTag(i)" class="text-blue-500 hover:text-red-600 font-bold text-xs">&times;</button>
            </span>
            <input v-model="newTag" @keydown.space.prevent="addTag" class="flex-1 min-w-[8rem] outline-none"
              placeholder="Add tag..." />
          </div>
          <div v-else-if="Array.isArray(value)" class="w-full">
            <textarea v-model="jsonArrayFields[key]" class="w-full p-1 border rounded font-mono" rows="2"></textarea>
          </div>
          <div
            v-else-if="typeof value === 'object' && value !== null && Object.values(value).every(v => typeof v === 'boolean')"
            class="w-full flex flex-wrap gap-4">
            <div v-for="(subVal, subKey) in value" :key="subKey" class="flex items-center gap-2">
              <input type="checkbox" :id="`${key}.${subKey}`" v-model="settings[key][subKey]" class="form-checkbox" />
              <label :for="`${key}.${subKey}`" class="text-sm">{{ subKey }}</label>
            </div>
          </div>
          <span v-else class="text-gray-500 text-sm italic">Unsupported type</span>
        </div>
      </div>

      <div class="pb-4">
        <Button label="Save Settings" type="submit" icon="pi pi-save" class="w-full justify-center mt-4" />
      </div>
    </form>
  </section>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useAppToast } from '@/composables/useAppToast'
import Button from 'primevue/button'
import { API_BASE, fetchWithAuth } from '@/api'
import { useTabState } from '@/composables/useTabState'

const { activeTab } = useTabState()

const settings = ref({})
const comments = ref({})
const jsonArrayFields = ref({})
const newTag = ref('')
const visibleTooltip = ref(null)
const { showSuccess, showError } = useAppToast()

function toggleTooltip(key) {
  visibleTooltip.value = visibleTooltip.value === key ? null : key
}

function addTag() {
  const trimmed = newTag.value.trim()
  if (trimmed && !settings.value.tags.includes(trimmed)) {
    settings.value.tags.push(trimmed)
  }
  newTag.value = ''
}

function removeTag(index) {
  settings.value.tags.splice(index, 1)
}

async function fetchSettings() {
  const res = await fetchWithAuth(`${API_BASE}/factorio-settings`)
  if (!res.ok) {
    settings.value = null
    return
  }

  const data = await res.json()
  const temp = {}

  for (const key in data) {
    if (key.startsWith('_comment_')) {
      const settingKey = key.replace('_comment_', '')
      comments.value[settingKey] = data[key]
    } else {
      temp[key] = data[key]
      if (Array.isArray(data[key])) {
        jsonArrayFields.value[key] = JSON.stringify(data[key], null, 2)
      }
    }
  }

  settings.value = temp
}

onMounted(fetchSettings)

watch(jsonArrayFields, () => {
  for (const key in jsonArrayFields.value) {
    try {
      const parsed = JSON.parse(jsonArrayFields.value[key])
      settings.value[key] = parsed
    } catch (e) {
    }
  }
}, { deep: true })

async function saveSettings() {
  const res = await fetchWithAuth(`${API_BASE}/factorio-settings`, {
    method: 'PUT',
    body: JSON.stringify(settings.value),
  })

  if (res.ok) {
    showSuccess('Settings saved!')
  } else {
    showError('Failed to save settings')
  }
}

document.addEventListener('click', () => {
  visibleTooltip.value = null
})

watch(activeTab, (val) => {
  if (val === 'serversettings') {
    fetchSettings()
  }
})

</script>