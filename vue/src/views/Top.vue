<script setup>
import { ref, computed, onMounted } from 'vue'
import DrawerColumn from '../components/DrawerColumn.vue'
import { get_formated_time } from '../my/get_formated_time.js';
import QRCode from 'qrcode';
const props = defineProps({
  id: '',
})

</script>

<template>
<DrawerColumn />
<div id="content">
TOP
</div>
</template>

<style>

</style>

