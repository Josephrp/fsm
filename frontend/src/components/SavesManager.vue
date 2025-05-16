<template>
  <section class="space-y-4 p-4 rounded border">
    <div class="pt-4">
      <FileUpload mode="basic" :auto="true" name="save" @select="uploadSave($event)" accept=".zip"
        :chooseLabel="uploading ? 'Uploading...' : 'Upload'" :disabled="uploading" />
    </div>

    <ul class="space-y-2 pt-4">
      <li v-for="save in saves" :key="save.name"
        class="group flex items-center justify-between p-2 bg-gray-100 rounded">
        <div class="flex items-center gap-2">
          <span>{{ save.name }} ({{ formatSize(save.size) }})</span>
          <Tag v-if="save.name === selectedSave" value="Selected" severity="info" class="ml-2 text-xs" />
        </div>
        <div class="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity duration-150">
          <Button v-if="save.name !== selectedSave" label="Select" icon="pi pi-check"
            class="p-button-sm p-button-warning" @click="updateSave(save.name)" />
          <Button :label="'Download'" icon="pi pi-download" class="p-button-sm p-button-success"
            @click="downloadSave(save.name)" download />
          <Button label="Delete" icon="pi pi-trash" class="p-button-sm p-button-danger"
            @click="confirmDelete(save.name)" />
        </div>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { API_BASE, fetchWithAuth, fetchSaves as apiFetchSaves, uploadSave as apiUploadSave, deleteSave as apiDeleteSave, fetchCurrentSave as apiFetchCurrentSave, updateSave as apiUpdateSave } from '../api'
import FileUpload from 'primevue/fileupload'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import ConfirmDialog from 'primevue/confirmdialog'
import { useConfirm } from 'primevue/useconfirm'
import { useAppToast } from '@/composables/useAppToast'
import type { SaveFile } from '@/types/save_file';

const saves = ref([] as SaveFile[])
const fileInput = ref(null)
const selectedSave = ref('')
const uploading = ref(false)
const fileSelected = ref(false)
const confirm = useConfirm()
const { showSuccess, showError } = useAppToast()

const onFileChange = () => {
  fileSelected.value = true
}

const fetchCurrentSave = async () => {
  const data = await apiFetchCurrentSave()
  selectedSave.value = data.save
}

const updateSave = async (save: string) => {
  await apiUpdateSave(save)
  fetchCurrentSave()
  fetchSaves()
}

const downloadSave = async (name: string) => {
  try {
    const res = await fetchWithAuth(`${API_BASE}/saves/${name}`)
    if (!res.ok) {
      if (!res.ok) {
        const message = (await res.json())?.message || null
        throw Error('Download failed.' + (message ? `\n${message}` : ''))
      }
    }

    const blob = await res.blob()
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = name
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  } catch (e) {
    showError(e instanceof Error ? e.message : String(e))
  }
}

const fetchSaves = async () => {
  saves.value = await apiFetchSaves()
}

const uploadSave = async (event: any) => {
  try {
    const file = event.files[0]
    if (!file) return
    uploading.value = true
    await apiUploadSave(file)
    await fetchSaves()
  } catch (e) {
    showError(e instanceof Error ? e.message : String(e))
  }
  fileSelected.value = false
  uploading.value = false
}

const deleteSave = async (name: string) => {
  await apiDeleteSave(name)
  fetchSaves()
}

const confirmDelete = (name: string) => {
  confirm.require({
    message: `Are you sure you want to delete save "${name}"?`,
    header: 'Confirm Deletion',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    accept: () => deleteSave(name)
  })
}

const formatSize = (bytes: number) => {
  const kb = bytes / 1024
  const mb = kb / 1024
  return mb >= 1 ? `${mb.toFixed(2)} MB` : `${kb.toFixed(2)} KB`
}

onMounted(() => {
  fetchCurrentSave()
  fetchSaves()
})
</script>
