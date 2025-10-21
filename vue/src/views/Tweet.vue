<script setup>
import { ref, onMounted, computed } from 'vue';
// import { useRoute } from 'vue-router';

import Advertisement from '@/components/Advertisement.vue';
import DrawerThread from '@/components/DrawerThread.vue';
import TweetForm from '@/components/TweetForm.vue';
import TweetMsgs from '@/components/TweetMsgs.vue';

import { removeMark } from '@/my/markdown.js';
import { useMessagesStore } from '@/stores/messages.js';

const props = defineProps({
  parentID: String,
  backID: String,
  messageID: String
})

// const route = useRoute();
const parentID = props.parentID // URLからparent_idを取得

const threadHead = ref({
  parentID: '',
  title: '',
  aliasName: '',
});
const tweets = ref([]);
const fetched = ref(false);
const copyable = ref(false);
const errorMessage = ref('');
const messagesStore = useMessagesStore();

const msg = ref({
  messageTxt: '',
  messageID: '',
  parentID: parentID,
});

// =====================================
// 🔹 API呼び出し関数
// =====================================
async function fetchTweets() {
  if (!parentID) {
    // parentIDがない場合（新規作成）
    threadHead.value = {
      parentID: '',
      title: '新規スレッド',
      messageTxt: '',
      aliasName: localStorage.getItem('myname') || 'unknown',
      aliasNames: [],
      newThread: true,
    };
    fetched.value = true;
    return;
  }
  const fd = new FormData();
  fd.append('parentID', parentID);
  fd.append('csrf', localStorage.getItem('csrf'));

  const res = await sendRequest('/TweetGet/', fd);

  if (!res || res.error) {
    errorMessage.value = res?.error || 'データ取得に失敗しました';
    return;
  }

  if (res.csrf) localStorage.setItem('csrf', res.csrf);

  // Tweetデータ構造を展開
  const tweetData = res.tweet || {};
  threadHead.value = tweetData.tweetHeads?.[0] || { parentID, title: '無題' };
  tweets.value = tweetData.tweets || [];
  console.log('threadHead', threadHead.value.messageTxt)
  messagesStore.insert(threadHead.value.messageTxt)
  // storeに反映
  tweets.value.forEach(t => messagesStore.insert(t));

  fetched.value = true;
}

// =====================================
// 🔹 スクロール時に追加取得（ページング）
// =====================================
async function loadOlderTweets() {
  const offset = tweets.value.length;
  const limit = 20;

  const fd = new FormData();
  fd.append('parentID', parentID);
  fd.append('offset', offset);
  fd.append('limit', limit);
  fd.append('csrf', localStorage.getItem('csrf'));

  const res = await sendRequest('/TweetGet/', fd);
  if (res && res.tweet?.tweets?.length) {
    const older = res.tweet.tweets;
    tweets.value = [...older, ...tweets.value];
  }
}

// =====================================
// 🔹 初期ロード
// =====================================
onMounted(async () => {
  await fetchTweets();
  document.title = threadHead.value.title || 'Tweet';
  const content = document.getElementById('content');
  if (content) content.scrollTop = content.scrollHeight;
});

function backTo() {
  location.href = '/';
}
</script>

<template>
<DrawerThread />
<div id="content">
  <div v-if="fetched">
    <div v-if="threadHead">
      <div class="headTitle">
        <a :href="'/tweet/' + threadHead.parentID + '/'">
          {{ threadHead.title }}
        </a>
      </div>
      <div class="headIcon">
        <span>
          <a @click="backTo"> ⬅ </a>
        </span>
        <span :class="[{ 'selected': copyable, 'emoji-stamp': copyable }]">
          <a @click="copyable = !copyable"> 📄 </a>
        </span>
      </div>
    </div>

    <TweetMsgs
      :messages="tweets"
      :threadHead="threadHead"
      :copyable="copyable"
      :messageID="parentID"
    />

    <div class="editText">
      <TweetForm
        :message="msg"
        :threadHead="threadHead"
      />
    </div>
  </div>
  <div v-if="!fetched"><br><br> Loading... </div>
  <div v-if="errorMessage">
    <p style="color:red;">{{ errorMessage }}</p>
  </div>
</div>

<div id="ad_right">
  <Advertisement /> <Advertisement /> <Advertisement />
</div>
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
