<template>
  <div id="tweet-content">
    <div v-if="fetched">
      <div v-if="tweetBase">
        <h2>{{ tweetBase.title || 'スレッド' }}</h2>
        <div v-for="t in tweetBase.tweets" :key="t.messageID" class="tweet-msg">
          <div class="tweet-header">
            <span class="tweet-alias">{{ t.aliasName }}</span>
            <span class="tweet-date">{{ formatDate(t.createdAt) }}</span>
          </div>
          <div class="tweet-body" v-html="t.messageTxt"></div>

          <div class="tweet-emojis">
            <span v-for="(e, i) in t.emojis" :key="i">
              {{ e.emoji }} <small>({{ e.aliasName }})</small>
            </span>
            <button @click="openEmojiModal(t.messageID)">😀</button>
          </div>
        </div>
      </div>
      <TweetEmojiModal
        v-if="emojiModalOpen"
        :messageID="selectedMessageID"
        :parentID="props.parentID"
        @closeEmoji="emojiModalOpen = false"
      />
    </div>
    <div v-else>Loading...</div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import TweetEmojiModal from '@/components/TweetEmojiModal.vue';
// import { sendRequest } from '@/my/sendRequest.js';

const props = defineProps({
  parentID: String
});

const tweetBase = ref(null);
const fetched = ref(false);
const emojiModalOpen = ref(false);
const selectedMessageID = ref('');

onMounted(async () => {
  await fetchTweets();
});

const fetchTweets = async () => {
  if (!props.parentID) fetched.value = true; return
  const fd = new FormData();
  fd.append('parentID', props.parentID);
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/TweetGet/', fd);
  if (res.error) {
    console.error(res.error);
    return;
  }
  tweetBase.value = res;
  fetched.value = true;
};

const openEmojiModal = (messageID) => {
  selectedMessageID.value = messageID;
  emojiModalOpen.value = true;
};

const formatDate = (dateStr) => {
  return new Date(dateStr).toLocaleString('ja-JP');
};
</script>

<style scoped>
#tweet-content {
  padding: 16px;
}
.tweet-msg {
  border-bottom: 1px solid #ddd;
  padding: 10px 0;
}
.tweet-header {
  display: flex;
  justify-content: space-between;
  font-weight: bold;
}
.tweet-emojis {
  margin-top: 6px;
}
.tweet-emojis button {
  margin-left: 10px;
}
</style>
