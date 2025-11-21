<script setup>
import { ref } from 'vue'
import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue'

document.title = 'ホーム'

const defaultLinks = [
  { topText: '組織・チャネル設定', topLink: '/channel/' },
  { topText: 'カレンダー', topLink: '/calendar/' },
  { topText: 'タイムスタンプ設定', topLink: '/timestampCode/' },
  { topText: '受付・予約機能', topLink: '/reception/' },
  { topText: 'チケット承認機能', topLink: '/tickets/' },
  { topText: 'ユーザー設定', topLink: '/user/' },
  { topText: '広告設定', topLink: '/adSetting/' },
  { topText: '設定', topLink: '/setting/' },
  { topText: 'サインイン', topLink: '/sign/' }
]

const channelID = localStorage.getItem('channelID')
let storedLinks = null
if (channelID) {
  try {
    const parsed = JSON.parse(localStorage.getItem('topLinks' + channelID))
    if (Array.isArray(parsed) && parsed.length > 0) {
      storedLinks = parsed
    }
  } catch (e) {
    console.warn('topLinks parse error', e)
  }
}

const topLinks = ref(storedLinks || defaultLinks)
</script>

<template>
<div id="drawer_column"><Drawer /></div>
<div id="content">
  <h2 class="sp_head">ホーム</h2>
  <div v-for="(link, i) in topLinks" :key="i" class="block">
    <a :href="link.topLink">{{ link.topText }}</a>
  </div>
</div>

<div id="ad_right">
  <Advertisement /> <Advertisement /> <Advertisement />
</div>
</template>

<style>
.block {
  width: 200px;
  margin: 10px;
}
</style>
