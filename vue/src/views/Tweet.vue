<script setup>
import { ref, onMounted } from 'vue';
import Advertisement from '@/components/Advertisement.vue';
import DrawerThread from '@/components/DrawerThread.vue';
import TweetMsgs from '@/components/TweetMsgs.vue';
import TweetForm from '@/components/TweetForm.vue';

import { useMessagesStore } from '@/stores/messages.js'

import { removeMark } from '@/my/markdown.js'
import { pushReceive } from '@/pushReceive/pushReceive.js'

const props = defineProps({
  parentID: String,
  backID: String,
  messageID: String
})

const parentID = props.parentID
const threadHead = ref({ parentID: '', title: '' });
const fetched = ref(false);
const errorMessage = ref('');
const copyable = ref(false);

const messagesStore = useMessagesStore()
async function fetchThreadHead() {
  if (!parentID) {
    threadHead.value = { parentID: '', title: '新規スレッド' }
    return
  }
  const fd = new FormData();
  fd.append('parentID', parentID);
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
  const heads = res.tweet.tweetHeads
  const tweets = res.tweet.tweets
  if (heads.length > 0) {
    threadHead.value = heads[0]
    let message = threadHead.value;
    message.messageID = threadHead.value.parentID
    message.createdAt = threadHead.value.updatedAt
    messagesStore.insert(message)
    if (Array.isArray(tweets)) {
      for (const tweet of tweets) {
        messagesStore.insert(tweet)
      }
    }
    return
  }
  if (props.backID) {
    const message = await getIDB('thread', props.parentID);
    threadHead.value = message;
    threadHead.value.parentID = props.parentID;
    threadHead.value.title = getSubstring(removeMark(message.messageTxt), 0, 12);
    threadHead.value.messageTxt = message.messageTxt;
    // threadHead.value.threadType = 0;
    threadHead.value.backID = props.backID
    threadHead.value.newReply = true
    const originalThreadHead = await getIDB('threadHead', props.backID)
    threadHead.value.aliasNames = [
      ...new Set([
        ...originalThreadHead.aliasNames,
        ...threadHeadValue.aliasNames
      ])
    ]
    threadHead.value.adminNames = originalThreadHead.adminNames
    messagesStore.insert(message)
    return
  }
  // threadHead.value = heads.length > 0 ? heads[0] : { parentID, title: 'headなし' }
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
  location.href = '/';
}

const msg = {
  messageTxt: '',
  messageID: '',
  parentID: props.parentID
}

</script>

<template>
  <DrawerThread />
  <div id="content">
    <div v-if="!fetched"><br><br>Loading…</div>
    <div v-if="fetched">
      <div class="headTitle" v-if="threadHead">
        <a :href="'/tweet/' + threadHead.parentID + '/'">{{ threadHead.title }}</a>
      </div>
      <div class="headIcon">
        <span><a @click="backTo">⬅</a></span>
        <span :class="[{ 'selected': copyable, 'emoji-stamp': copyable }]">
          <a @click="copyable = !copyable">📄</a>
        </span>
      </div>

      <TweetMsgs
        :parentID="parentID"
        :threadHead="threadHead"
        :copyable="copyable"
        :messageID="messageID"
      />

      <div class="editText">
        <TweetForm
          :message="msg"
          :threadHead="threadHead"
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
