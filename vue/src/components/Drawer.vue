<script setup>
import { ref, defineProps, onMounted } from 'vue'

import Advertisement from '@/components/Advertisement.vue'

const props = defineProps({
  channel: Object,
  aliases: Array,
})

const groups = ref([])
const iamGuest = ref(false)
onMounted(async () => {
  if (props.channel) {
    iamGuest.value = props.aliases.some(
      alias => alias.accessRight === "guest" && alias.aliasName === localStorage.getItem('myname')
    )
  } else if (localStorage.getItem('csrf')) {
    const aliases = await getIDBs('alias', 'channelIDIndex', localStorage.getItem("channelID"), 10000)
    iamGuest.value = aliases.some(
      alias => alias.accessRight === "guest" && alias.aliasName === localStorage.getItem('myname')
    )
  }
})

</script>

<template>
  <div id="drawer_column">
    <label for="drawer_check" class="pc_disp_none for_drawer">≡</label>
    <input id="drawer_check" type="checkbox" class="pulling pc_disp_none">
    <table id="drawer">
      <tr><td><a href="/" > ホーム </a></td></tr>
      <tr><td style="text-align: center;"> <Advertisement /> </td></tr>
      <tr v-if="!iamGuest"><td><a href="/channel/" > 組織・チャネル設定 </a></td></tr>
      <tr><td><a href="/calendar/"> カレンダー </a></td></tr>
      <tr v-if="!iamGuest"><td><a href="/timestampCode/"> タイムスタンプ設定 </a></td></tr>
      <tr v-if="!iamGuest"><td><a href="/reception/" > 受付・予約機能 </a></td></tr>
      <tr><td><a href="/tickets/" > チケット承認機能 </a></td></tr>
      <tr><td><a href="/user/" > ユーザー設定 </a></td></tr>
      <tr v-if="!iamGuest"><td><a href="/entryFormEdit/" > フォーム編集 </a></td></tr>
      <tr v-if="!iamGuest"><td><a href="/topEdit/" > ホームページ編集 </a></td></tr>
      <tr><td><a href="/tweets/" > ツイート </a></td></tr>
      <tr><td><a href="/adSetting/" > 広告設定 </a></td></tr>
      <tr><td><a href="/setting/" > 設定 </a></td></tr>
      <tr><td><a href="/sign/" > サインイン </a></td></tr>
      <tr><td><a href="/html/rule/" > 規則 </a></td></tr>
      <tr><td><a href="/html/privacy/" > 個人情報遵守 </a></td></tr>
    </table>
  </div>
</template>

<style scoped>

#drawer td {
  background-color: #EEEEEE;
}

</style>
