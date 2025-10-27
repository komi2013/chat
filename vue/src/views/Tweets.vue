<script setup>
import { ref, onMounted } from 'vue'
// import { sendRequest } from '@/my/common.js'
import DrawerTweet from '@/components/DrawerTweet.vue'

import { pushReceive } from '@/pushReceive/pushReceive.js'
import { removeMark } from '@/my/markdown.js'

function tF(a, b = null){ return timeFormat(a, b) }

// === 状態 ===
const tweets = ref([]) // TweetStruct配列
const errorMessage = ref('')
const fetched = ref(false)
// const nickname = ref('')

// === 最新Tweet取得 ===
async function fetchLatestTweets() {
  const fd = new FormData();
  fd.append('csrf', localStorage.getItem('csrf'))
  const res = await sendRequest('/TweetGetLatest/', fd)
  if (res.error) {
    errorMessage.value = res.error
    return
  }
  res.csrf && localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  if (Array.isArray(res.tweets)) tweets.value = res.tweets
  fetched.value = true
  // try {
  //   const fd = new FormData()
  //   fd.append('csrf', localStorage.getItem('csrf') || '')

  //   const res = await sendRequest('/TweetGetLatest/', fd)

  //   if (res.error) {
  //     errorMessage.value = res.error
  //     return
  //   }

  //   if (res.csrf) localStorage.setItem('csrf', res.csrf)
  //   nickname.value = res.nickname

  //   if (Array.isArray(res.tweets)) {
  //     tweets.value = res.tweets
  //   } else {
  //     errorMessage.value = '取得データが不正です'
  //   }
  // } catch (e) {
  //   console.error('TweetGetLatest error:', e)
  //   errorMessage.value = 'サーバーとの通信に失敗しました'
  // } finally {
  //   fetched.value = true
  // }
}

// === 日付フォーマット ===
// function formatDate(isoStr) {
//   if (!isoStr) return ''
//   const date = new Date(isoStr)
//   return date.toLocaleString('ja-JP', {
//     month: '2-digit',
//     day: '2-digit',
//     hour: '2-digit',
//     minute: '2-digit'
//   })
// }

// === タイトル生成 ===
function getTitle(text) {
  if (!text) return '(無題)'
  const plain = removeMark(text).replace(/<[^>]+>/g, '')
  return plain.length > 30 ? plain.slice(0, 30) + '…' : plain
}

onMounted(async () => {
  await fetchLatestTweets()
})
</script>

<template>
  <DrawerTweet />
  <div id="content">
    <h2>最新スレッド一覧</h2>

    <div v-if="!fetched"><br><br>Loading…</div>
    <div v-else>
      <div v-if="errorMessage" class="error">
        <p style="color:red;">{{ errorMessage }}</p>
      </div>

      <div v-if="tweets.length === 0 && !errorMessage">
        <p>スレッドがありません。</p>
      </div>

      <ul v-else class="tweet-list">
        <li v-for="t in tweets" :key="t.id" class="tweet-item">
          <a :href="`/tweet/${tF('YYYYMMDD', t.tweetHead.updatedAt)}/${t.tweetHead.parentID}.html`" class="tweet-link">
            <div class="tweet-title">{{ getTitle(t.tweetHead.messageTxt) }}</div>
            <div class="tweet-meta">
              <div class="icon_td">
                <a :href="'/nickname/' + '' + '/' + t.tweetHead.nickname + '/'">
                  <img v-if="t.tweetHead.nickImg && t.tweetHead.nickImg.charAt(0) != ','" 
                    :src="t.tweetHead.nickImg" class="icon-img">
                  <span v-if="t.tweetHead.nickImg && t.tweetHead.nickImg.charAt(0) == ','"
                    class="icon-span" 
                    :style="'background-color:' + t.tweetHead.nickImg.split(',')[2] ">
                    {{t.tweetHead.nickImg.split(',')[1]}}</span>
                </a>
              </div>
              <span class="tweet-nick">{{ t.tweetHead.nickname || '匿名' }}</span>
              <span class="tweet-date">{{ tF('MM/DD hh:mm', t.tweetHead.updatedAt) }}</span>
            </div>
          </a>
        </li>
      </ul>
    </div>
  </div>
</template>

<style scoped>
#content {
  padding: 20px;
  max-width: 720px;
  margin: auto;
  font-family: "Hiragino Kaku Gothic ProN", Meiryo, sans-serif;
}

h2 {
  text-align: center;
  margin-bottom: 16px;
  color: #333;
}

.icon-span {
  border-radius: 10%;
  display: inline-block;
  height: 28px;
  width: 28px;
}
.icon-img {
  border-radius: 10%;
  display: inline-block;
  max-height: 30px;
  max-width: 30px;
}
.icon_td {
  width: 50px;
  vertical-align: top;
  text-align: center;
  display: inline-block;
}

.tweet-list {
  list-style: none;
  padding: 0;
}

.tweet-item {
  border-bottom: 1px solid #ddd;
  padding: 10px 0;
}

.tweet-link {
  display: block;
  text-decoration: none;
  color: inherit;
}

.tweet-title {
  font-weight: bold;
  color: #222;
  margin-bottom: 4px;
}

.tweet-meta {
  font-size: 0.9em;
  color: #666;
}

.tweet-nick {
  margin-right: 10px;
}
</style>
