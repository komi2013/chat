<script setup>
import { ref } from 'vue'
import Advertisement from '@/components/Advertisement.vue'
import Drawer from '@/components/Drawer.vue'

document.title = 'ホーム'

// -------------------------
// ✔️ サインイン判定
// -------------------------
const signIn = !!localStorage.getItem('csrf')

// -------------------------
// ✔️ デフォルトリンク（signin あり／なし）
// -------------------------
const signedInLinks = [
  { topText: '組織・チャネル設定', topLink: '/channel/' },
  { topText: 'カレンダー', topLink: '/calendar/' },
  { topText: 'タイムスタンプ設定', topLink: '/timestampCode/' },
  { topText: '受付・予約機能', topLink: '/reception/' },
  { topText: 'チケット承認機能', topLink: '/tickets/' },
  { topText: 'ユーザー設定', topLink: '/user/' },
  { topText: '広告設定', topLink: '/adSetting/' },
  { topText: '設定', topLink: '/setting/' },
]

const guestLinks = [
  { topText: 'サインイン', topLink: '/sign/' },
  { topText: 'ツイート一覧', topLink: '/tweets/' },
  { topText: '規則', topLink: '/html/rule/' },
  { topText: '個人情報遵守', topLink: '/html/privacy/' }
]

// -------------------------
// ✔️ ユーザー編集された topLinks があるか確認
// -------------------------
const channelID = localStorage.getItem('channelID')
let storedLinks = null

if (channelID) {
  try {
    const parsed = JSON.parse(localStorage.getItem('topLinks' + channelID))
    if (Array.isArray(parsed) && parsed.length > 0) storedLinks = parsed
  } catch (e) {
    console.warn('topLinks parse error', e)
  }
}

// -------------------------
// ✔️ 最終的に表示するリンク
// -------------------------
const topLinks = ref(
  storedLinks
    ? storedLinks // ユーザー編集版
    : (signIn ? signedInLinks : guestLinks)
)
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
