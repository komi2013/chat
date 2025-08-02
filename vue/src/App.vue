<script setup>
import { RouterView } from 'vue-router';
import { onMounted } from 'vue';

onMounted(async () => { //for outsideApp
  window.advertisements = await getAllIDBs('advertisement');
})


</script>

<template>
  <RouterView />
</template>

<style scoped>

</style>
