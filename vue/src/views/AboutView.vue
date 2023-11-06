<!-- About.vue -->
<template>
  <div>
    <h1>About</h1>
    <p>{{ message }}</p>
  </div>
</template>

<script setup>
import { defineProps } from 'vue';

const message = defineProps('message');
</script>