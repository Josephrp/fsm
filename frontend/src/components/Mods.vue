<template>
  <section class="space-y-4">
    <Tabs value="mods" v-model="activeTab" @update:value="onTabChange">
      <TabList>
        <Tab value="mods">Mods</Tab>
        <Tab value="yourmods">Your Mods</Tab>
      </TabList>
      <TabPanels>
        <TabPanel value="mods">
          <ModList ref="modListRef" />
        </TabPanel>
        <TabPanel value="yourmods">
          <BookmarkedMods ref="bookmarkedRef" />
        </TabPanel>
      </TabPanels>
    </Tabs>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import Tabs from 'primevue/tabs';
import TabList from 'primevue/tablist';
import Tab from 'primevue/tab';
import TabPanels from 'primevue/tabpanels';
import TabPanel from 'primevue/tabpanel';
import ModList from './ModList.vue'
import BookmarkedMods from './BookmarkedMods.vue';

const activeTab = ref('mods')
const modListRef = ref()
const bookmarkedRef = ref()

function onTabChange(val: string | number) {
  const tab = String(val)
  if (tab === 'mods') {
    modListRef.value?.reload()
  } else if (tab === 'yourmods') {
    bookmarkedRef.value?.reload()
  }
}

onMounted(() => {
  onTabChange(activeTab.value)
})
</script>