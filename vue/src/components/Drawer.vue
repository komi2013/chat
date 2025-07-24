<script setup>

import Advertisement from '@/components/Advertisement.vue';

</script>

<template>
  <div id="drawer_column">
    <label for="drawer_check" class="pc_disp_none for_drawer">≡</label>
    <input id="drawer_check" type="checkbox" class="pulling pc_disp_none">
    <table id="drawer">
      <tr><td><a href="/" > ホーム </a></td></tr>
      <tr><td style="text-align: center;"> <Advertisement /> </td></tr>
      <tr><td><a href="/channel/" > 組織・チャネル設定 </a></td></tr>
      <tr><td><a href="/calendar/"> カレンダー </a></td></tr>
      <tr><td><a href="/timestampCode/"> タイムスタンプ設定 </a></td></tr>
      <tr><td><a href="/reception/" > 受付・予約機能 </a></td></tr>
      <tr><td><a href="/tickets/" > チケット承認機能 </a></td></tr>
      <tr><td><a href="/user/" > ユーザー設定 </a></td></tr>
      <tr><td><a href="/adSetting/" > 広告設定 </a></td></tr>
      <tr><td><a href="/setting/" > 設定 </a></td></tr>
      <tr><td><a href="/sign/" > サインイン </a></td></tr>
    </table>
  </div>
</template>

<style scoped>

#drawer td {
  background-color: #EEEEEE;
}

</style>
