<script setup>
import { ref, computed, onMounted } from 'vue';

import Advertisement from '@/components/Advertisement.vue';
import DrawerReception from '@/components/DrawerReception.vue';
import ReceptionEditBox from '@/components/ReceptionEditBox.vue';
import Messages from '@/components/Messages.vue';

import { useMessagesStore } from '@/stores/messages.js';

import { isEmojiOpen, selectedMessageId, openEmoji, closeEmoji, selectEmoji, calcEmoji, emojiPath } from '@/my/emoji.js';
import { removeMark } from '@/my/markdown.js';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  channelID: String,
  code: String,
  backID: String
  // messageID: String,
  
})

let myname = localStorage.getItem('myname')
const reception = ref(null);
const threadHead = ref(null)
const fetched = ref(false);
const copyable = ref(false);
const errorMessage = ref('');
let nickname
// receptionBook.vueのfindReception()を参考にした関数
async function findReception() {
  const fd = new FormData();
  fd.append('receptionID', props.channelID);
  fd.append('channelID', props.channelID);
  fd.append('aliasName', myname);
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ReceptionGet/', fd);
  if (!res.csrf) { errorMessage.value = res; return }
  if (res.error) {
    errorMessage.value = res.error
    return 
  }
  res.csrf && localStorage.setItem('csrf', res.csrf);
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  nickname = res.nickname
  return res.reception;
}

// 問い合わせ用のシンプルなchannelオブジェクトを作成
const groups = ref([])
const channel = ref({channelID: '', myname: '', channelName: ''})
let msg = {messageTxt: '',messageID: '',parentID: '@'}
onMounted(async () => {
  reception.value = await findReception()
  const parentID = '@' + nickname
  threadHead.value = await getIDB('threadHead', parentID)
  msg.parentID = parentID
  await makeThreadHead()
  document.title = threadHead.value.title
  channel.value.channelID = props.channelID
  channel.value.myname = nickname
  channel.value.channelName = threadHead.value.title
  const content = await document.getElementById('content');
  content.scrollTop = await content.scrollHeight;
  await window.scrollTo(0, content.scrollHeight);
  readStatus()
  fetched.value = true
});

const message = ref('');
const parentID = ref('');
const messagesStore = useMessagesStore();
const messages = computed(() => {
  return messagesStore.messages;
});

async function makeThreadHead() {
  const parentID = '@' + nickname
  if (threadHead.value) {
    threadHead.value.inquirerFlag = true
    if (reception.value && reception.value.joinNames && reception.value.joinNames.includes(myname)) {
      threadHead.value.edit = true;
    }
    let message = threadHead.value;
    message.messageID = threadHead.value.parentID
    message.createdAt = threadHead.value.updatedAt
    messagesStore.insert(message)
  } else {
    const threadHeadValue = {
      channelID: props.channelID,
      receptionID: reception.value.receptionID,
      parentID: parentID,
      title: reception.value.receptionTitle + 'への問い合わせ',
      messageTxt: '',
      aliasName: nickname,
      aliasNames: [nickname],
      displayStatus: 0,
      newThread: true,
      // threadType: 'reception' // 問い合わせ用の識別子
    }
    threadHead.value = threadHeadValue;
    if (props.backID) {
      const message = await getIDB('thread', parentID);
      threadHead.value = message;
      threadHead.value.parentID = parentID;
      threadHead.value.title = getSubstring(removeMark(message.messageTxt), 0, 12);
      threadHead.value.messageTxt = message.messageTxt;
      threadHead.value.threadType = 'reception';
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
    }
  }
}

function readStatus () {
  if (threadHead.value.displayStatus && threadHead.value.displayStatus == 1 || threadHead.value.displayStatus == 2) {
    threadHead.value.displayStatus = 0;
    updIDBone('threadHead', '@' + nickname, 'displayStatus', 0);
    revertFaviconBadge()
  }
}

function backTo() {
  const backID = threadHead.value.backID;
  if (backID) {
    location.href = '/receptionThread/' + props.channelID + '/' + props.code + '/?backID=' + backID
  }
}

</script>

<template>
<DrawerReception/>
<div id="content">
  <div v-if="fetched">
    <div v-if="threadHead">
      <div class="headTitle">
        <a :href="'/receptionThreadHead/' + channelID + '/' + threadHead.parentID + '/?code=' + code">
          {{threadHead.title}}
        </a>
      </div>
      <div class="headIcon">
        <span>
          <a @click="backTo"> ⬅ </a>
        </span>
        <span :class="[{ 'selected': copyable, 'emoji-stamp': copyable }]">
          <a @click="copyable = !copyable" > 📄 </a>
        </span>
      </div>
    </div>
    <Messages 
      :channel="channel"
      :groups="groups"
      :messages="messages"
      :threadHead="threadHead"
      :copyable="copyable"
      :messageID="messageID" />

    <div class="editText">
      <ReceptionEditBox 
        :channel="channel"
        :message="msg"
        :threadHead="threadHead" />
    </div>
    <br>
  </div>
  <div v-if="!fetched"><br><br> Loading... or Something Went </div>
  <div v-if="errorMessage"> 
    <div class="errorMessage">{{errorMessage}}<br>データ取得に失敗しました。</div>
    <a href="/setting/"> データ設定ページ </a><br>
    <a href="/sign/"> サインインページ </a>
  </div>
</div>
<div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>

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
