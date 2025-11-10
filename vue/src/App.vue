<script setup>
import { RouterView } from 'vue-router'
import { ref, onMounted } from 'vue'
import { loadAdvertisements } from '@/my/advertisement'

const initialized = ref(false)

onMounted(async () => {
  await loadAdvertisements()
  initialized.value = true // ✅ 完了後に RouterView を描画
})
</script>

<template>
  <template v-if="initialized">
    <RouterView />
  </template>
</template>
