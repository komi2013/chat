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
  <h2>ホーム</h2>
  <div class="block"><a href="/">チャット</a></div>
  <div class="block"><a href="/channel/" > 👪 組織・チャネル </a></div>
  <div class="block"><a href="/timestampCode/">タイムスタンプ</a></div>
  <div class="block"><a href="/workflow/">ワークフロー</a></div>
  <div class="block"><a href="/calendar/">カレンダー</a></div>
  <div class="block"><a href="/bookPattern/">bookPattern</a></div>
  <div class="block"><a href="/setting/">setting</a></div>
</div>

<div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>

</template>

<style>

h2 {
  position: relative;
  left: 50px;
  top: -10px;
}

.block {
  display: inline-block;
  width: 90px;
  margin: 10px;
}

</style>

