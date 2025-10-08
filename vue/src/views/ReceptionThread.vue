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
  nickname: String,
  backID: String,
  messageID: String,
  code: String
})

let myname = localStorage.getItem('myname') || 'お客様';
const reception = ref(null);
const threadHead = ref({
  parentID: '',
  title: '',
});
const fetched = ref(false);
const copyable = ref(false);
const errorMessage = ref('');

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
  return res.reception;
}

onMounted(async () => {
  reception.value = await findReception()
  if (reception.value) {
    const parentID = '@@' + props.nickname;
    threadHead.value = await getIDB('threadHead', parentID);
    await makeThreadHead()
    document.title = threadHead.value.title
    const content = await document.getElementById('content');
    content.scrollTop = await content.scrollHeight;
    await window.scrollTo(0, content.scrollHeight);
    readStatus()
  }
  fetched.value = true
});

const message = ref('');
const parentID = ref('');
const messagesStore = useMessagesStore();
const messages = computed(() => {
  return messagesStore.messages;
});

const msg = {
  messageTxt: '',
  messageID: '',
  parentID: '@@' + props.nickname
};

async function makeThreadHead() {
  const parentID = '@@' + props.nickname;
  if (threadHead.value) {
    // 問い合わせ対応では、スタッフかどうかを判定
    if (reception.value && reception.value.joinNames && reception.value.joinNames.includes(myname)) {
      threadHead.value.edit = true;
    }
    let message = threadHead.value;
    message.messageID = threadHead.value.parentID
    message.createdAt = threadHead.value.updatedAt
    messagesStore.insert(message)
  } else {
    const threadHeadValue = {
      receptionID: reception.value.receptionID,
      parentID: parentID,
      title: '新規問い合わせ',
      messageTxt: '',
      aliasName: reception.value.customerName || 'お客様',
      aliasNames: [reception.value.customerName || 'お客様'],
      joinNames: reception.value.joinNames || [reception.value.customerName || 'お客様'],
      displayStatus: 0,
      newThread: true,
      threadType: 'reception' // 問い合わせ用の識別子
    }
    
    if (parentID.startsWith('@@')) {
      const nickname = parentID.replace('@@', '');
      threadHeadValue.title = getSubstring(nickname, 0, 12);
      threadHeadValue.messageTxt = reception.value.receptionTitle || '問い合わせ';
      threadHeadValue.aliasNames = [reception.value.customerName || 'お客様', nickname]
      threadHeadValue.aliasNames = [...new Set(threadHeadValue.aliasNames)]
      threadHeadValue.joinNames = threadHeadValue.aliasNames
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
      threadHead.value.joinNames = [reception.value.customerName || 'お客様', originalThreadHead.aliasName]
      messagesStore.insert(message);
    }
  }
}

function readStatus () {
  if (threadHead.value.displayStatus && threadHead.value.displayStatus == 1 || threadHead.value.displayStatus == 2) {
    threadHead.value.displayStatus = 0;
    updIDBone('threadHead', '@@' + props.nickname, 'displayStatus', 0);
    revertFaviconBadge()
  }
}

function backTo() {
  const backID = threadHead.value.backID;
  if (backID) {
    location.href = '/receptionThread/' + props.channelID + '/@@' + props.nickname + '/' + backID + '/?code=' + props.code;
  } else {
    location.href = '/reception/' + props.nickname + '/?code=' + props.code;
  }
}

// 問い合わせ用のシンプルなchannelオブジェクトを作成
const channel = computed(() => {
  if (!reception.value) return null;
  return {
    channelID: reception.value.receptionID,
    myname: reception.value.customerName || 'お客様',
    channelName: reception.value.facilityName || '問い合わせ対応'
  };
});

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
