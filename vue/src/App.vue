<script setup>
import { RouterLink, RouterView } from 'vue-router';
import { ref, onMounted } from 'vue';
import { pushReceive } from './pushReceive/pushReceive.js';

navigator.serviceWorker.addEventListener('message', async (event) => {
  pushReceive(event.data.notificationData);
});

onMounted(async () => {
  window.advertisements = await getAllIDBs('advertisement');
  // await findAdPrices();
})


</script>

<template>
  <RouterView />
</template>

<style scoped>

</style>
