<script setup>
import { ref, computed, onMounted } from 'vue'
import Advertisement from '@/components/Advertisement.vue'

// === IndexedDB から TweetHead を取得する想定関数 ===
// ※ 実際は `idb.js` などにある関数を利用してください。
// import { getAllIDBs } from '@/my/idb.js'

// TweetHead の一覧を保持
const tweetHeads = ref([])

// === IndexedDBからTweetHeadを取得 ===
const fetchTweetHeads = async () => {
  const data = await getAllIDBs('tweetHead')
  console.log('data', data)
  if (Array.isArray(data) && data.length > 0) {
    // displayStatusの順に並べ替え（mention → unread → read → mute）
    const sorted = [
      ...data.filter(d => d.displayStatus === 2), // mention
      ...data.filter(d => d.displayStatus === 1), // unread
      ...data.filter(d => d.displayStatus === 0), // read
      ...data.filter(d => d.displayStatus === 3)  // mute
    ]
    tweetHeads.value = sorted
  }
}

// ステータス別の見た目を指定
function getStatusClass(status) {
  return {
    read: status === 0,
    unread: status === 1,
    mention: status === 2,
    mute: status === 3
  }
}

// === 遷移先 ===
function goThread(head) {
  const date = timeFormat('YYYYMMDD', head.createdAt)
  location.href = `/tweet/${date}/${head.parentID}.html`
}

onMounted(async () => {
  await fetchTweetHeads()
})
</script>

<template>
  <div id="drawer_column">
    <label for="drawer_check" class="pc_disp_none for_drawer">≡</label>
    <input id="drawer_check" type="checkbox" class="pulling pc_disp_none">
    <table id="drawer">
      <tr><td><a href="/" > ホーム </a></td></tr>
      <tr><td style="text-align: center;"> <Advertisement /> </td></tr>
      <tr><td><a href="/tweets/" > ツイート </a></td></tr>
      <tr v-for="head in tweetHeads" :key="head.parentID">
        <td :class="getStatusClass(head.displayStatus)">
          <a @click="goThread(head)">
            {{ head.title || '(無題)' }}
          </a>
        </td>
      </tr>

      <tr v-if="tweetHeads.length === 0">
        <td>スレッドがありません。</td>
      </tr>

    </table>
  </div>
</template>

<style scoped>
#tweethead_table {
  width: 100%;
  border-collapse: collapse;
}

#tweethead_table td {
  background-color: #f8f8f8;
  padding: 6px 10px;
  border-bottom: 1px solid #ddd;
}

/* ステータスごとの色分け */
.mention a {
  color: red;
  font-weight: bold;
}

.unread a {
  color: blueviolet;
}

.read a {
  color: blue;
  opacity: 0.8;
}

.mute a {
  color: gray;
  opacity: 0.4;
}

a {
  cursor: pointer;
  text-decoration: none;
}

a:hover {
  text-decoration: underline;
}
</style>
