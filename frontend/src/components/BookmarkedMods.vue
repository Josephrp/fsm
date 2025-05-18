<template>
    <div>
        <div v-if="loading" class="p-4">
            <div class="flex flex-col gap-3">
                <div v-for="n in 5" :key="n" class="flex justify-between items-center gap-3">
                    <Skeleton width="20%" height="2rem" />
                    <Skeleton width="15%" height="2rem" />
                    <Skeleton width="25%" height="2rem" />
                    <Skeleton width="20%" height="2rem" />
                </div>
            </div>
        </div>
        <div v-if="!loading">
            <Message v-if="authError" severity="info" class="p-2">
                <i class="pi pi-exclamation-triangle text-yellow-500"
                    title="Installed version is not the latest available" /> You must configure your Factorio username
                and token to view your mods.
            </Message>
            <DataTable v-if="!authError" :value="mods" dataKey="name" removableSort scrollable scrollHeight="600px">
                <Column field="title" header="Mod Name" sortable />
                <Column field="owner" header="Author" sortable />
                <Column header="Installed">
                    <template #body="{ data }">
                        <span class="flex items-center gap-2">
                            <span>{{ data.installed_version || 'Not Installed' }}</span>
                            <i v-if="data.installed_version && data.is_outdated"
                                class="pi pi-exclamation-triangle text-yellow-500"
                                title="Installed version is not the latest available" />
                        </span>
                    </template>
                </Column>
                <Column header="Version">
                    <template #body="{ data }">
                        <Select v-model="selectedVersions[data.name]"
                            :options="data.releases.map((r: ModRelease) => r.version)" placeholder="Select"
                            class="w-full" :disabled="Boolean(data.installed_version)"
                            :pt="{ root: { title: data.installed_version ? 'Uninstall to change version' : '' } }">
                            <template #option="slotProps">
                                <div class="flex items-center justify-between w-full">
                                    <span>{{ slotProps.option }}</span>
                                    <i v-if="availableMap[data.name]?.includes(slotProps.option)"
                                        class="pi pi-check text-green-500" />
                                </div>
                            </template>
                        </Select>
                    </template>
                </Column>
                <Column header="Action">
                    <template #body="{ data }">
                        <Button v-if="data.installed_version" label="Uninstall" severity="warn"
                            @click="onUninstallMod(data.name, selectedVersions[data.name])" />
                        <div v-else-if="data.available_versions.includes(selectedVersions[data.name])"
                            class="flex gap-2">
                            <Button v-if="data.available_versions.includes(selectedVersions[data.name])" label="Install"
                                severity="success" @click="onInstallMod(data.name, selectedVersions[data.name])"
                                :disabled="!selectedVersions[data.name]" />
                            <Button
                                v-if="data.available_versions.includes(selectedVersions[data.name]) && selectedVersions[data.name] !== data.installed_version"
                                icon="pi pi-trash" severity="danger" aria-label="Delete downloaded mod"
                                @click="confirmDelete(data.name, selectedVersions[data.name])"
                                :disabled="!selectedVersions[data.name]" />
                        </div>
                        <Button v-else label="Download" @click="onDownloadMod(data.name, selectedVersions[data.name])"
                            :disabled="!selectedVersions[data.name]" />
                    </template>
                </Column>
            </DataTable>
        </div>
    </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { API_BASE, deleteMod, downloadMod, fetchBookmarkedMods, installMod, uninstallMod } from '@/api';
import { useAppToast } from '@/composables/useAppToast'
import { useConfirm } from 'primevue/useconfirm'
import Select from 'primevue/select';
import DataTable from 'primevue/datatable';
import Column from 'primevue/column';
import Button from 'primevue/button';
import Message from 'primevue/message';
import Skeleton from 'primevue/skeleton';

const { showSuccess, showError } = useAppToast()
const loading = ref<boolean>(true)
const working = ref<boolean>(false)
const authError = ref(false)
const confirm = useConfirm()

interface ModRelease {
    version: string
    download_url: string
}

interface ModInfo {
    name: string
    title: string
    owner: string
    installed_version?: string
    is_outdated?: boolean
    available_versions: string[]
    releases: ModRelease[]
}

interface BookmarkedModsResponse {
    available: Record<string, string[]>[]
    installed: Record<string, string[]>[]
    downloadable: ModInfo[]
    code?: number
}

const mods = ref<ModInfo[]>([])
const selectedVersions = ref<Record<string, string>>({})
const availableMap = ref<Record<string, string[]>>({})

const reload = async () => {
    await loadMods()
}

defineExpose({
    reload,
})

const loadMods = async () => {
    loading.value = true
    authError.value = false
    try {
        const response = await fetchBookmarkedMods() as BookmarkedModsResponse
        if (response?.code === 403 || response?.code === 502) {
            authError.value = true;
            return
        }
        const installedMap: Record<string, string[]> = {}

        for (const item of response.available) {
            for (const [modName, versions] of Object.entries(item)) {
                availableMap.value[modName] = versions
            }
        }
        for (const item of response.installed) {
            for (const [modName, versions] of Object.entries(item)) {
                installedMap[modName] = versions
            }
        }

        mods.value = response.downloadable.map((mod: ModInfo) => {
            const installed = installedMap[mod.name]?.[0] || undefined

            const availableVersions = availableMap.value[mod.name] || []

            let selected: string | undefined = undefined
            if (installed && mod.releases.some(r => r.version === installed)) {
                selected = installed
            } else {
                const availableRelease = mod.releases.find(r => availableVersions.includes(r.version))
                if (availableRelease) {
                    selected = availableRelease.version;
                } else if (availableVersions.length > 0) {
                    selected = availableVersions[0]
                } else {
                    selected = mod.releases[0]?.version || undefined
                }
            }

            if (selected) {
                selectedVersions.value[mod.name] = selected
            }

            return {
                ...mod,
                available_versions: availableVersions,
                installed_version: installed,
                is_outdated: Boolean(installed && mod.releases.length > 0 && installed !== mod.releases[0].version),
            }
        });
    } catch (error: any) {
        showError(error instanceof Error ? error.message : String(error))
    } finally {
        loading.value = false
    }
};

const onDownloadMod = async (modName: string, version: string) => {
    try {
        await downloadMod(modName, version)
        showSuccess(`Downloaded ${modName} ${version}`)

        const index = mods.value.findIndex(m => m.name === modName)
        if (index !== -1) {
            const mod = mods.value[index]

            if (!mod.releases.find(r => r.version === version)) {
                mod.releases.push({ version, download_url: '' })
            }

            if (!mod.available_versions.includes(version)) {
                mod.available_versions.push(version)
            }

            mod.is_outdated = Boolean(mod.installed_version && mod.available_versions.length > 0 &&
                mod.installed_version !== mod.available_versions[0])

            mods.value[index] = { ...mod }
            selectedVersions.value[modName] = version
        }
    } catch (e) {
        showError(e instanceof Error ? e.message : String(e))
    }
}

const onInstallMod = async (modName: string, version: string) => {
    try {
        await installMod(modName, version)
        showSuccess(`Installed ${modName} ${version}`)

        const index = mods.value.findIndex(m => m.name === modName)
        if (index !== -1) {
            const mod = mods.value[index]
            mod.installed_version = version
            mod.is_outdated = mod.releases.length > 0 && version !== mod.releases[0].version
            mods.value[index] = { ...mod }
        }
    } catch (e) {
        showError(e instanceof Error ? e.message : String(e))
    }
}

const onUninstallMod = async (modName: string, version: string) => {
    try {
        await uninstallMod(modName, version)
        showSuccess(`Uninstalled ${modName} ${version}`)
        const mod = mods.value.find(m => m.name === modName)
        if (mod) {
            mod.installed_version = undefined
            mod.is_outdated = false
        }
    } catch (e) {
        showError(e instanceof Error ? e.message : String(e))
    }
}

const onDeleteMod = async (modName: string, version: string) => {
    try {
        await deleteMod(modName, version)
        showSuccess(`Stub: deleted ${modName} ${version}`)
    } catch (e) {
        showError(e instanceof Error ? e.message : String(e))
        return
    }

    const index = mods.value.findIndex(m => m.name === modName)
    if (index !== -1) {
        const mod = mods.value[index]

        // Remove the version from available_versions
        mod.available_versions = mod.available_versions.filter(v => v !== version)
        // Also update availableMap for dropdown rendering consistency
        availableMap.value[modName] = mod.available_versions

        const hasOtherMatchingRelease = mod.releases.some(r => mod.available_versions.includes(r.version))
        if (!hasOtherMatchingRelease) {
            mod.releases = mod.releases.filter(r => r.version !== version)
        }

        mod.is_outdated = !!mod.installed_version &&
            mod.available_versions.length > 0 &&
            mod.installed_version !== mod.available_versions[0]

        if (selectedVersions.value[modName] === version) {
            const fallback = mod.releases.find(r => mod.available_versions.includes(r.version))?.version
                || mod.available_versions[0]
                || mod.releases[0]?.version
                || undefined
            if (fallback) {
                selectedVersions.value[modName] = fallback
            } else {
                delete selectedVersions.value[modName]
            }
        }

        mods.value[index] = { ...mod }
    }
}

const confirmDelete = (modName: string, version: string) => {
    confirm.require({
        message: `Are you sure you want to delete "${modName} ${version}"?`,
        header: 'Confirm Delete',
        icon: 'pi pi-exclamation-triangle',
        acceptClass: 'p-button-danger',
        accept: () => onDeleteMod(modName, version)
    })
}
</script>