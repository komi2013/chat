<script setup>
import { ref, computed, onMounted } from 'vue'

import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue'

const props = defineProps({
  id: '',
})

</script>

<template>
<Drawer />
<div id="content">
  <h2 class="sp_head">ホーム</h2>
  <div class="block"><a href="/channel/" > 組織・チャネル設定 </a></div>
  <div class="block"><a href="/calendar/"> カレンダー </a></div>
  <div class="block"><a href="/timestampCode/"> タイムスタンプ設定 </a></div>
  <div class="block"><a href="/reception/" > 受付・予約機能 </a></div>
  <div class="block"><a href="/tickets/" > チケット承認機能 </a></div>
  <div class="block"><a href="/user/" > ユーザー設定 </a></div>
  <div class="block"><a href="/adSetting/" > 広告設定 </a></div>
  <div class="block"><a href="/setting/" > 設定 </a></div>
  <div class="block"><a href="/sign/" > サインイン </a></div>
</div>

<div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>

</template>

<style>
.block {
  /*display: inline-block;*/
  width: 200px;
  margin: 10px;
}

</style>

