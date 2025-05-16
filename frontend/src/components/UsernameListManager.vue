<template>
  <section class="space-y-4 rounded border">
    <DataTable :value="users" class="p-datatable-sm" :paginator="users.length > 5" :rows="5"
      :rowsPerPageOptions="[5, 10, 20, 50]">
      <Column field="username" header="Username" sortable bodyClass="pl-6" />
      <Column header="Actions" style="width: 100px">
        <template #body="{ data }">
          <Button icon="pi pi-trash" severity="danger" class="p-button-text" @click="confirmRemove(data.username)" />
        </template>
      </Column>

      <template
        #paginatorcontainer="{ first, last, page, pageCount, prevPageCallback, nextPageCallback, totalRecords }: {first: number, last: number, page: number, pageCount?: number,  prevPageCallback: (value: MouseEvent) => void, nextPageCallback: (value: MouseEvent) => void, totalRecords?: number}">
        <div
          class="flex items-center gap-4 border border-primary bg-transparent rounded-full w-full py-1 px-2 justify-between">
          <Button icon="pi pi-chevron-left" rounded text @click="prevPageCallback" :disabled="page === 0" />
          <div class="text-color font-medium">
            <span class="hidden sm:block">Showing {{ first }} to {{ last }} of {{ totalRecords }}</span>
            <span class="block sm:hidden">Page {{ page + 1 }} of {{ pageCount }}</span>
          </div>
          <Button icon="pi pi-chevron-right" rounded text @click="nextPageCallback"
            :disabled="pageCount === undefined || page === pageCount - 1" />
        </div>
      </template>
    </DataTable>

    <div class="flex gap-2 m-4">
      <InputText v-model="newUser" name="input" placeholder="username" class="w-[30ch]" autocomplete="off" />
      <Button icon="pi pi-user-plus" severity="success" @click="addUser"
        :disabled="!newUser || usernameExists || !isValidFormat" />
    </div>
    <div v-if="!isValidFormat && newUser" class="text-sm text-red-600 m-4">
      Username must be 3–30 characters using letters, numbers, hyphens, or underscores.
    </div>
    <div v-if="usernameExists" class="text-sm text-red-600 m-4">
      That username is already in the list.
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppToast } from '@/composables/useAppToast'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import { useConfirm } from 'primevue/useconfirm'

const props = defineProps<{
  title: string
  load: () => Promise<string[]>
  add: (username: string) => Promise<void>
  remove: (username: string) => Promise<void>
}>()

const confirm = useConfirm()
const users = ref<{ username: string }[]>([])
const newUser = ref('')
const { showError } = useAppToast()

const usernameExists = computed(() =>
  users.value.some(u => u.username.toLowerCase() === newUser.value.trim().toLowerCase())
)

const isValidFormat = computed(() =>
  /^[a-zA-Z0-9_-]{3,30}$/.test(newUser.value.trim())
)

const loadUsers = async () => {
  try {
    const loadedUsers = await props.load()
    users.value = loadedUsers.map(username => ({ username }))
  } catch (e) {
    showError(e instanceof Error ? e.message : String(e))
  }
}

const addUser = async () => {
  try {
    await props.add(newUser.value)
    newUser.value = ''
    await loadUsers()
  } catch (e) {
    showError(e instanceof Error ? e.message : String(e))
  }
}

const removeUser = async (user: string) => {
  try {
    await props.remove(user)
    await loadUsers()
  } catch (e) {
    showError(e instanceof Error ? e.message : String(e))
  }
}

const confirmRemove = (name: string) => {
  confirm.require({
    message: `Are you sure you want to remove "${name}"?`,
    header: 'Confirm Remove',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    accept: () => removeUser(name)
  })
}

onMounted(() => {
  loadUsers()
})
</script>
