<script setup>
import { RouterLink, RouterView } from 'vue-router';
import { ref, onUpdated } from 'vue';
import { pushReceive } from './pushReceive/pushReceive.js';

navigator.serviceWorker.addEventListener('message', async (event) => {
  pushReceive(event.data.notificationData);
});

</script>

<template>
  <RouterView />
</template>

<style scoped>

</style>
