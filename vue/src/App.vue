<script setup>
import { RouterView } from 'vue-router'
import { ref, onMounted } from 'vue'
import { loadAdvertisements } from '@/my/advertisement'

const initialized = ref(false)

onMounted(async () => {
  await loadAdvertisements()
  console.log('in onMounted')
  initialized.value = true // ✅ 完了後に RouterView を描画
})

console.log('out onMounted')
</script>

<template>
  <RouterView />
</template>
