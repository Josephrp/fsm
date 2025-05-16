<template>
  <div class="space-y-6">
    <div class="">
      <FactorioUserSettings />
    </div>

    <div class="">
      <ServerAdminManager />
    </div>
  </div>
</template>

<script setup>
  import ServerAdminManager from '@/components/ServerAdminManager.vue'
  import FactorioUserSettings from '@/components/FactorioUserSettings.vue'
</script>