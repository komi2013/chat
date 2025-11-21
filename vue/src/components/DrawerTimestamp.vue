<script setup>

import Advertisement from '@/components/Advertisement.vue';

</script>

<template>
  <label for="drawer_check" class="pc_disp_none for_drawer">≡</label>
  <input id="drawer_check" type="checkbox" class="pulling pc_disp_none">
  <table id="drawer">
    <tr><td><a href="/" > ホーム </a></td></tr>
    <tr><td style="text-align: center;"> <Advertisement /> </td></tr>
    <tr><td><a href="/timestampCode/" > 🖨️ QR発行 </a></td></tr>
    <tr><td><a href="/timestampReport/_/" > 📋 勤務表 </a></td></tr>
    <tr><td><a href="/sign/" > サインイン </a></td></tr>
  </table>
</template>

<style scoped>

#drawer td {
  background-color: #EEEEEE;
}

</style>
