<script setup>
import { ref, onMounted } from 'vue';
import Advertisement from '@/components/Advertisement.vue';
import DrawerTweet from '@/components/DrawerTweet.vue'
import TweetMsgs from '@/components/TweetMsgs.vue';
import TweetForm from '@/components/TweetForm.vue';

import { useMessagesStore } from '@/stores/messages.js'

import { removeMark } from '@/my/markdown.js'
import { pushReceive } from '@/pushReceive/pushReceive.js'

const props = defineProps({
  date: String,
  parentID: String,
  backID: String,
  messageID: String
})

const parentID = props.parentID.replace(/\.html$/, '')
const msg = {
  messageTxt: '',
  messageID: '',
  parentID: parentID
}
// console.log('parentID', parentID)
const threadHead = ref({ parentID: '', title: '' })
const fetched = ref(false)
const errorMessage = ref('')
const copyable = ref(false)
const nickname = ref('')
const messagesStore = useMessagesStore()
const postable = ref(true)
const messagesCount = ref(0)
async function fetchThreadHead() {
  if (!parentID) {
    threadHead.value = { parentID: '', title: '新規スレッド', nicknames: [] }
    return
  }
  const fd = new FormData();
  fd.append('parentID', parentID)
  fd.append('backID', props.backID ?? '')
  fd.append('csrf', localStorage.getItem('csrf'))
  const res = await sendRequest('/TweetGet/', fd)
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
  nickname.value = res.nickname
  const head = res.tweet.tweetHead
  const tweets = res.tweet.tweets
  checkTweetPostLimit(res.tweetPosts, parentID)
  // if (props.backID) {
  //   // const message = await getIDB('thread', props.parentID);
  //   threadHead.value = message
  //   threadHead.value.parentID = props.parentID;
  //   threadHead.value.title = getSubstring(removeMark(message.messageTxt), 0, 12);
  //   threadHead.value.messageTxt = message.messageTxt;
  //   // threadHead.value.threadType = 0;
  //   threadHead.value.backID = props.backID
  //   threadHead.value.newReply = true
  //   const originalThreadHead = await getIDB('threadHead', props.backID)
  //   threadHead.value.aliasNames = [
  //     ...new Set([
  //       ...originalThreadHead.aliasNames,
  //       ...threadHeadValue.aliasNames
  //     ])
  //   ]
  //   threadHead.value.adminNames = originalThreadHead.adminNames
  //   messagesStore.insert(message)
  //   return
  // }

  threadHead.value = head
  threadHead.value.title = getSubstring(removeMark(threadHead.value.messageTxt), 0, 12)
  let message = threadHead.value
  message.messageID = threadHead.value.parentID
  message.createdAt = threadHead.value.updatedAt
  messagesStore.insert(message)
  // console.log('tweets', tweets)
  messagesCount.value = tweets ? tweets.length : 0
  if (Array.isArray(tweets)) {
    for (const tweet of tweets) {
      if (props.backID) {
        tweet.href = `/tweet/${props.date}/${message.parentID}.html?backID=${props.backID}messageID=${message.messageID}`
      } else {
        tweet.href = `/tweet/${props.date}/${message.parentID}.html?messageID=${message.messageID}`
      }

      messagesStore.insert(tweet)
    }
  }


  // threadHead.value = heads.length > 0 ? heads[0] : { parentID, title: 'headなし' }
}

function checkTweetPostLimit(tweetPosts, parentID) {
  if (!Array.isArray(tweetPosts)) {
    postable.value = true
    return
  }

  const now = new Date()
  const TWENTY_HOURS = 20 * 60 * 60 * 1000

  // 20時間以内の投稿のみ抽出
  const recentPosts = tweetPosts.filter(post => {
    if (!post.postedAt) return false
    const postedAt = new Date(post.postedAt)
    return now - postedAt < TWENTY_HOURS
  })

  // === 条件①：同スレッドで20時間以内に3回投稿済み（管理者除く） ===
  const sameThread = recentPosts.find(
    p => p.parentID === parentID && !p.postAdminFlag && p.postCount >= 3
  )
  if (sameThread) {
    postable.value = false
    console.log('❌ 投稿制限: このスレッドでは20時間以内に3回投稿済み')
    errorMessage.value = '投稿回数制限を超えました。明日投稿できます'
    return
  }

  // === 条件②：全体で20時間以内のスレッド投稿が3件ある ===
  const limitedThreads = recentPosts.filter(
    p => !p.postAdminFlag && p.postCount >= 3
  )
  if (limitedThreads.length >= 3) {
    postable.value = false
    console.log('❌ 投稿制限: 20時間以内に3スレッドで上限投稿済み')
    errorMessage.value = '投稿回数制限を超えました。明日投稿できます'
    return
  }
  // 制限なし → 投稿可能
  postable.value = true
}


onMounted(async () => {
  await fetchThreadHead()
  document.title = threadHead.value.title
  const content = await document.getElementById('content')
  content.scrollTop = await content.scrollHeight
  await window.scrollTo(0, content.scrollHeight)
  // readStatus()
  fetched.value = true
});

function backTo() {
  location.href = `/tweet/${props.date}/${threadHead.value.parentID}.html`
}

</script>

<template>
  <DrawerTweet />
  <div id="content">
    <div v-if="!fetched"><br><br>Loading…</div>
    <div v-if="fetched">
      <div class="headTitle" v-if="threadHead">
        <a :href="`/tweet/${date}/${threadHead.parentID}.html`">{{ threadHead.title }}</a>
      </div>
      <div class="headIcon">
        <span v-if="backID"><a @click="backTo">⬅</a></span>&nbsp;
        <span :class="[{ 'selected': copyable, 'emoji-stamp': copyable }]">
          <a @click="copyable = !copyable">📄</a>
        </span>
      </div>

      <TweetMsgs
        :date="date"
        :parentID="parentID"
        :threadHead="threadHead"
        :copyable="copyable"
        :messageID="messageID"
        :nickname="nickname"
        :messagesCount="messagesCount"
      />

      <div v-if="postable" class="editText">
        <TweetForm
          :message="msg"
          :threadHead="threadHead"
          :backID="backID"
          :tweetPosts="tweetPosts"
        />
      </div>

      <div v-if="errorMessage">
        <p style="color:red;">{{ errorMessage }}</p>
      </div>
    </div>
  </div>
  <div id="ad_right"><Advertisement /><Advertisement /><Advertisement /></div>
</template>

<style>

#content {
  height: 100%;
  overflow-y: auto;
}

@media screen and (min-width : 701px) { 
  .headTitle {
    width: 50%;
    min-height: 40px;
    display: inline-block;
  }
  .headIcon {
    width: 40%;
    text-align: right;
    display: inline-block;
  }

}

@media screen and (max-width : 700px) {
  .headTitle {
    width: 45%;
    min-height: 40px;
    margin-left: 50px;
    display: inline-block;
  }
  .headIcon {
    width: 35%;
    text-align: right;
    display: inline-block;
  }

}
</style>
