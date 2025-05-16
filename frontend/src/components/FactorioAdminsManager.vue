<template>
  <UsernameListManager
    title="Server Admins"
    :load="loadFactorioAdmins"
    :add="addFactorioAdmin"
    :remove="removeFactorioAdmin"
  />
</template>

<script setup lang="ts">
import UsernameListManager from './UsernameListManager.vue'
import { loadFactorioAdmins, addFactorioAdmin, removeFactorioAdmin } from '@/api'
</script>