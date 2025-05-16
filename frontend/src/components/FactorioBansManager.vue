<template>
  <UsernameListManager
    title="Banned Users"
    :load="loadFactorioBans"
    :add="addFactorioBan"
    :remove="removeFactorioBan"
  />
</template>

<script setup lang="ts">
import UsernameListManager from './UsernameListManager.vue'
import { loadFactorioBans, addFactorioBan, removeFactorioBan } from '@/api'
</script>