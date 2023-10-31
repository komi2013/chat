<script setup>
import { ref } from 'vue';

const userID = ref('');

const onClick = () => { console.log(userID.value) };

</script>

<template>
  <input v-model="userID" />
  <button @click="onClick">click</button>
</template>