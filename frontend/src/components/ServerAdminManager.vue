<template>
  <div class="rounded border px-4 py-4">
    <ConfirmDialog />
    <section class="space-y-4">
      <h3 class="text-lg font-semibold">Server Admins</h3>

      <DataTable :value="Object.entries(admins)" class="p-datatable-sm" responsiveLayout="scroll"
        :paginator="Object.keys(admins).length > 5" :rows="5" :rowsPerPageOptions="[5, 10, 20, 50]">
        <Column header="Username">
          <template #body="slotProps">
            <span class="font-mono text-sm">{{ slotProps.data[0] }}</span>
          </template>
        </Column>

        <Column header="Password">
          <template #body="slotProps">
            <form @submit.prevent="addAdmin" class="flex gap-2 flex-col">
              <InputText type="text" :value="slotProps.data[0]" autocomplete="username" tabindex="-1" aria-hidden="true"
                class="sr-only" readonly />
              <Password v-model="admins[slotProps.data[0]]" class="w-75"
                :placeholder="slotProps.data[0] === currentUser ? '●●●●●●' : 'Enter new password'"
                :autocomplete="slotProps.data[0] === currentUser ? 'current-password' : 'new-password'" />
              <p v-if="admins[slotProps.data[0]] && admins[slotProps.data[0]].length < MIN_PASSWORD_LENGTH"
                class="text-red-600 text-xs mt-1">
                Password must be at least 6 characters.
              </p>
            </form>
          </template>
        </Column>

        <Column header="">
          <template #body="slotProps">
            <div class="text-right space-x-2">
              <Button v-if="hasChanges(slotProps.data[0]) && isPasswordValid(slotProps.data[0])"
                @click="updateAdmin(slotProps.data[0])" class="p-button-text" :loading="saving === slotProps.data[0]"
                :loadingIcon="'pi pi-spinner pi-spin'" icon="pi pi-save" title="Save Password" />
              <Button v-if="slotProps.data[0] !== currentUser" @click="confirmDelete(slotProps.data[0])"
                class="p-button-text p-danger" icon="pi pi-trash" title="Delete" />
            </div>
          </template>
        </Column>
        <template
          #paginatorcontainer="{ first, last, page, pageCount, prevPageCallback, nextPageCallback, totalRecords }">
          <div
            class="flex items-center gap-4 border border-primary bg-transparent rounded-full w-full py-1 px-2 justify-between">
            <Button icon="pi pi-chevron-left" rounded text @click="prevPageCallback" :disabled="page === 0" />
            <div class="text-color font-medium">
              <span class="hidden sm:block">Showing {{ first }} to {{ last }} of {{ totalRecords }}</span>
              <span class="block sm:hidden">Page {{ page + 1 }} of {{ pageCount }}</span>
            </div>
            <Button icon="pi pi-chevron-right" rounded text @click="nextPageCallback"
              :disabled="page === pageCount - 1" />
          </div>
        </template>
      </DataTable>

      <div class="mt-4">
        <form @submit.prevent="addAdmin" class="flex items-center gap-2">
          <InputText v-model="newUsername" placeholder="New username" autocomplete="username" class="p-1" />
          <Password v-model="newPassword" placeholder="New password" autocomplete="new-password" class="p-1" />
          <Button type="submit" class="p-button-text p-success"
            :disabled="!isValidFormat || usernameExists || !newUsername || newPassword.length < MIN_PASSWORD_LENGTH"
            :loading="saving === 'new'" :loadingIcon="'pi pi-spinner pi-spin'" icon="pi pi-save" title="Add Admin" />
        </form>
        <p v-if="newPassword && newPassword.length < MIN_PASSWORD_LENGTH" class="text-red-600 text-xs mt-1 ml-1">
          Password must be at least {{ MIN_PASSWORD_LENGTH }} characters.
        </p>
        <div v-if="!isValidFormat && newUsername" class="text-red-600 text-xs mt-1 ml-1">
          Username must be 3–30 characters using letters, numbers, hyphens, or underscores.
        </div>
        <div v-if="usernameExists" class="text-red-600 text-xs mt-1 ml-1">
          That username is already in the list.
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue'
import {
  fetchAdmins as loadAdmins,
  addAdmin as apiAddAdmin,
  updateAdmin as apiUpdateAdmin,
  deleteAdmin as apiDeleteAdmin,
} from '@/api'
import { useAppToast } from '@/composables/useAppToast'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import ConfirmDialog from 'primevue/confirmdialog'
import { useConfirm } from 'primevue/useconfirm'

const MIN_PASSWORD_LENGTH = 6

const admins = ref({})
const originalAdmins = ref({})
const newUsername = ref('')
const newPassword = ref('')
const currentUser = localStorage.getItem('username')
const showAddForm = ref(false)
const saving = ref(null)
const { showSuccess, showError } = useAppToast()
const confirm = useConfirm()

const usernameExists = computed(() =>
  Object.keys(admins.value || {}).some(u => u.toLowerCase() === newUsername.value.trim().toLowerCase()) ?? false
)

const isValidFormat = computed(() =>
  /^[a-zA-Z0-9_-]{3,30}$/.test(newUsername.value.trim())
)

const isPasswordValid = (user) => {
  const value = admins.value[user]
  return value && value.length >= MIN_PASSWORD_LENGTH
}

const load = async () => {
  const loaded = await loadAdmins()
  admins.value = loaded
  originalAdmins.value = { ...loaded }
}

const hasChanges = (user) => {
  return admins.value[user] !== originalAdmins.value[user]
}

const updateAdmin = async (user) => {
  saving.value = user
  try {
    await apiUpdateAdmin(user, admins.value[user])
    originalAdmins.value[user] = admins.value[user]
    showSuccess('Admin updated!')
  } catch (e) {
    showError(e instanceof Error ? e.message : String(e))
  }
  saving.value = null
}

const deleteAdmin = async (user) => {
  try {
    await apiDeleteAdmin(user)
    showSuccess('Admin deleted!')
    load()
  } catch (e) {
    showError(e instanceof Error ? e.message : String(e))
  }
}

const confirmDelete = (user) => {
  confirm.require({
    message: `Are you sure you want to delete admin "${user}"?`,
    header: 'Confirm Deletion',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    accept: () => deleteAdmin(user)
  })
}

const addAdmin = async () => {
  saving.value = 'new'
  try {
    await apiAddAdmin(newUsername.value, newPassword.value)
    showSuccess('Admin added!')
    newUsername.value = ''
    newPassword.value = ''
    showAddForm.value = false
    load()
  } catch (e) {
    showError(e instanceof Error ? e.message : String(e))
  }
  saving.value = null
}

onMounted(load)
</script>