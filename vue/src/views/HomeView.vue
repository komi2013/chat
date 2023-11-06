<template>
  <div>
    <h1>Home</h1>
    <button @click="emitEvent">Emit Event to About</button>
  </div>
</template>

<script setup>
import { defineEmits } from 'vue';

const { emit } = defineEmits();
console.log(emit)

const emitEvent = () => {
  // defineEmits('custom-event', 'Hello from Home');
  console.log('emmmmiitttt')
};
</script>