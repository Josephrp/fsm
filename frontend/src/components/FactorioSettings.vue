<template>
  <section class="space-y-4">
    <Tabs value="0">
      <TabList>
        <Tab value="0">Server Admins</Tab>
        <Tab value="1">Ban List</Tab>
        <Tab value="2">White List</Tab>
        <Tab value="3">Server Settings</Tab>
      </TabList>
      <TabPanels>
        <TabPanel value="0">
          <FactorioAdminsManager />
        </TabPanel>
        <TabPanel value="1">
          <FactorioBansManager />
        </TabPanel>
        <TabPanel value="2">
          <FactorioWhitelistManager />
        </TabPanel>
        <TabPanel value="3">
          <FactorioServerSettings />
        </TabPanel>
      </TabPanels>
    </Tabs>
  </section>
</template>

<script setup>
import { ref } from 'vue'
import Tabs from 'primevue/tabs';
import TabList from 'primevue/tablist';
import Tab from 'primevue/tab';
import TabPanels from 'primevue/tabpanels';
import TabPanel from 'primevue/tabpanel';
import FactorioAdminsManager from './FactorioAdminsManager.vue'
import FactorioBansManager from './FactorioBansManager.vue'
import FactorioWhitelistManager from './FactorioWhitelistManager.vue'
import FactorioServerSettings from './FactorioServerSettings.vue'

const activeTab = ref('admins')
</script>