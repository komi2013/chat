<template>
  <div v-if="notice" class="popup">
    {{ notice }}
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useNoticesStore } from '../stores/notices.js';

const noticesStore = useNoticesStore();

const notice = computed(() => noticesStore.notice);
</script>

<style scoped>
.popup {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background-color: rgba(0, 0, 0, 0.8);
  color: #fff;
  padding: 20px;
  border-radius: 10px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.5);
  z-index: 20;
  text-align: center;
}
</style>
