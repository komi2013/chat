<template>
  <div>
    <p>Count: {{ count }}</p>
    <p>Count: {{ countString }}</p>
    <button @click="incrementAction">Incrementこ</button>
    <button @click="decrementAction">Decrementま</button>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useCounterStore } from '../stores/counter.js';

const counterStore = useCounterStore();
const count = counterStore.count;

const incrementAction = () => {
  counterStore.increment();
};

const decrementAction = () => {
  counterStore.decrement();
};

const countString = computed(() => {
  return counterStore.count
})
</script>