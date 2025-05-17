<template>
  <section class="space-y-4">
    <Tabs value="admins" v-model:value="activeTab">
      <TabList>
        <Tab value="admins">Server Admins</Tab>
        <Tab value="bans">Ban List</Tab>
        <Tab value="whitelist">White List</Tab>
        <Tab value="serversettings">Server Settings</Tab>
      </TabList>
      <TabPanels>
        <TabPanel value="admins">
          <FactorioAdminsManager />
        </TabPanel>
        <TabPanel value="bans">
          <FactorioBansManager />
        </TabPanel>
        <TabPanel value="whitelist">
          <FactorioWhitelistManager />
        </TabPanel>
        <TabPanel value="serversettings">
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
import { useTabState } from '@/composables/useTabState'

const { activeTab } = useTabState()
</script>