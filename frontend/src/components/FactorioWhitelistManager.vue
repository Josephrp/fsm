<template>
  <UsernameListManager
    title="Banned Users"
    :load="loadFactorioWhitelist"
    :add="addFactorioWhitelistUser"
    :remove="removeFactorioWhitelistUser"
  />
</template>

<script setup lang="ts">
import UsernameListManager from './UsernameListManager.vue'
import { loadFactorioWhitelist, addFactorioWhitelistUser, removeFactorioWhitelistUser } from '@/api'
</script>