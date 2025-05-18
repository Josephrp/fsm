<template>
  <section class="space-y-2 rounded border p-4">
    <div v-if="loading" class="space-y-2">
      <div v-for="n in 5" :key="n" class="flex items-center justify-between p-2 bg-gray-100 rounded">
        <Skeleton width="30%" height="2rem" />
        <Skeleton width="6rem" height="2rem" />
      </div>
    </div>
    <ul v-else class="space-y-2">
      <li v-for="mod in mods" :key="mod.name"
        class="flex items-center justify-between p-2 bg-gray-100 dark:bg-gray-400 rounded">
        <span>{{ mod.name }}</span>
        <Button @click="onToggleMod(mod)" class="px-3 py-1" :severity="!mod.enabled ? 'success' : 'danger'">
          {{ mod.enabled ? 'Disable' : 'Enable' }}
        </Button>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue"
import { fetchMods, toggleMod } from '@/api'
import { Button } from "primevue"
import Skeleton from 'primevue/skeleton'
import { useAppToast } from '@/composables/useAppToast'
import type { Mod } from "@/types/mod"

const { showSuccess, showError } = useAppToast()

const mods = ref([] as Mod[])
const loading = ref(true)
const reload = async () => {
  loading.value = true
  mods.value = await fetchMods()
  loading.value = false
}

defineExpose({
  reload,
})

const onToggleMod = async (mod: Mod) => {
  mods.value = await toggleMod(mod)
}

onMounted(async () => {
  reload()
})

</script>