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

.icon {
  max-width: 50px;
  max-height: 50px;
}

.icon_td {
  width: 50px;
}

.setting {
  text-align: right;
}

.box {
  display: flex;
}

.alias {
  margin: 2px;
}

.time {
  margin: 2px;
}

.emoji {
  margin: 2px;
}

.reply {
  margin: 2px;
}

.others {
  margin: 2px;
}

.msgBox textarea {
  width: 100%;
  border: none;
  height: 50px;
}

@media screen and (min-width : 701px) { 
  .msgBox {
    position: fixed;
    bottom: 10px;
    width: 300px;
  }
}

@media screen and (max-width : 700px) {
  .msgBox {
    position: fixed;
    bottom: 10px;
    width: 300px;
  }
  .headTitle {
    margin-left: 50px;
  }
}
</style>

