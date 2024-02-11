<script setup>
import { ref, computed, onMounted } from 'vue'
import DrawerColumn from '../components/DrawerColumn.vue'
import { useMessagesStore } from '../stores/messages.js';
import { useChannelsStore } from '../stores/channels.js';
import { get_formated_time } from '../my/get_formated_time.js';
import { getIDB, getIDBs, upsertData } from '../my/indexDB.js';
import { isEmojiOpen, selectedMessageId, openEmoji, closeEmoji, selectEmoji, calcEmoji, emojiPath } from '../my/emoji.js';
import { isOtherOpen, otherMessageId, openOther, closeOther, selectOther, activeEdit, cancelEdit, adjustHeight, textareaRefs } from '../my/other.js';
import EmojiModal from '../components/EmojiModal.vue';
import OtherModal from '../components/OtherModal.vue';

const props = defineProps({
  id: '',
})

const channel = ref('');
const messagesStore = useMessagesStore();
const messages = computed(() => {
  return messagesStore.messages;
})

async function fetchData() {
  try {
    const data = await getIDB('channel', props.id);
    channel.value = data;
  } catch (error) {
    channel.value = null;
  }
}

const fetchMessageData = () => {
  return new Promise((resolve, reject) => {
    getIDBs('message', 'channelIDIndex', props.id)
      .then((data) => {
        const latest = data.reverse();
        latest.forEach(message => {
          messagesStore.insert(message);
        });
        resolve();
      })
      .catch((error) => {
        reject(error);
      });
  });
};

const msgText = ref(null);
const msgUpsert = (messageTxt, messageId) => {
  const messageData = messageTxt ? messageTxt : document.getElementById('editable_' + messageId).innerText;
  let uri = messageId ? '/MessageEdit/' : '/MessagePost/';
  const fd = new FormData();
  fd.append('channelID', props.id);
  fd.append('messageID', messageId);
  fd.append('messageTxt', messageData);
  fd.append('messageType', 1);
  fd.append('editFlg', 1);
  fd.append('aliasName', channel.value.aliasName);
  const request = new Request(uri, {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      alert(reason)
    })
}

const contentRef = ref(null);
onMounted(() => {
  fetchData();
  fetchMessageData()
  .then(() => {
      const content = contentRef.value;
      content.scrollTop = content.scrollHeight;
    });
});

const clickEmoji = (messageId, emoji) => {
  const fd = new FormData();
  fd.append('channelID', props.id);
  fd.append('messageID', messageId);
  fd.append('emoji', emoji[0]);
  fd.append('clicked', emoji[2] ? 1 : 0);
  fd.append('aliasName', channel.value.aliasName);
  const request = new Request('/MessageEdit/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      alert(reason)
    })
};

// const handleInput = (event, message) => {
//   const newText = event.target.innerText; // テキストの変更を取得
//   message.updateTxt = newText; // updateTxt に新しいテキストを代入
//   // message.updateTxt = message.messageTxt;
//   messagesStore.update(message, message.messageID);
// };

</script>



<template>
<DrawerColumn />
<div id="content" ref="contentRef">
<div class="headTitle">
  <RouterLink :to="'/channelInfo/' + channel.channelID"> {{ channel.channelName }} </RouterLink>
</div>

  <div v-for="message in messages" :key="message.messageID">
    <table>
      <tr>
        <td rowspan="2" class="icon_td">
          <img v-if="message.aliasImg" :src="message.aliasImg" class="icon">
        </td>
        <td>
          <span class="alias">{{ message.aliasName }}</span>
          <span class="time">{{ get_formated_time('hh:mm', message.createdAt) }}</span>
        </td>
        <td class="setting">
          <span class="emoji" @click="activeEdit(message, message.messageID, $event)"> 🖋 </span>
          <span class="emoji" @click="openEmoji(message.messageID)"> 😄 </span>
          <EmojiModal :key="message.messageID" v-if="isEmojiOpen && selectedMessageId === message.messageID" @selectEmoji="selectEmoji" @closeEmoji="closeEmoji" :channelID="channel.channelID" :messageId="message.messageID" :aliasName="channel.aliasName" />

          <span class="reply"> <RouterLink :to="'/reply/' + message.messageID"> 💬 </RouterLink> </span>
          <span class="others" @click="openOther(message.messageID)"> &nbsp; ⋮ &nbsp; </span>
          <OtherModal :key="message.messageID" v-if="isOtherOpen && otherMessageId === message.messageID" @selectOther="selectOther" @closeOther="closeOther" @activeEdit="activeEdit" :channelID="channel.channelID" :messageId="message.messageID" :aliasName="channel.aliasName" :message="message" />
        </td>
      </tr>
      <tr>
        <td v-if="message.editFlg" colspan="2" class="msg">
          <div>
            <div><span>📎</span><span style="font-weight: bold;">B</span></div>
            <div contenteditable="true"
              class="chgble"
              :id="'editable_' + message.messageID"
            >{{ message.updateTxt }}</div>
            <div style="text-align: right">
              <button @click="cancelEdit(message, message.messageID)">⬅</button>
              <button @click="msgUpsert(null, message.messageID)">▶️</button>
            </div>
          </div>
        </td>
        <td v-else colspan="2" class="msg">
          <div>{{ message.messageTxt }}</div>
          <div>
          <template v-for="d in calcEmoji(message.emojis, channel.aliasName)">
            <template v-if="emojiPath(d[0])">
              <span class="img-stamp" :class="{ 'selected-class': d[2] }"> <img :src="d[0]" class="emoji-img" @click="clickEmoji(message.messageID, d)" /> {{d[1]}} </span>
            </template>
            <template v-else>
              <span class="emoji-stamp" :class="{ 'selected-class': d[2] }" @click="clickEmoji(message.messageID, d)" >{{ d[0] }} {{d[1]}} </span>
            </template>
          </template>
          </div>
        </td>
      </tr>
    </table>
  </div>
<!-- <RouterLink to="/channel/abc/" >channel abc</RouterLink> -->

<div class="msgBox">
  <div><span>📎</span><span style="font-weight: bold;">B</span></div>
  <div contenteditable="true" ref="msgData" class="chgble" ></div>
  <div style="text-align: right"><button @click="msgUpsert($refs.msgData.innerText)">▶️</button></div>
</div>

</div>
</template>

<style>

textarea {
  resize: none;
  height: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
}

#content {
  height: 100%;  /*コンテナの高さを固定 */
  overflow-y: auto; /* 縦方向のスクロールを有効にする */
}

#content > div {
  display: flex;
  flex-direction: column-reverse; /* コンテンツを下から上に並べる */
}

/* コンテンツのスタイル */
#content table {
  /* テーブルのスタイルを適用 */
}

.icon {
  max-width: 50px;
  max-height: 50px;
}

.icon_td {
  width: 50px;
}

.setting {
  text-align: right;
}

.box {
  display: flex;
}

.alias {
  margin: 2px;
}

.time {
  margin: 2px;
}

.emoji {
  margin: 2px;
}

.reply {
  margin: 2px;
}

.others {
  margin: 2px;
}

.msgBox textarea {
  width: 100%;
  border: none;
  height: 50px;
}

.emoji-stamp {
  display: inline-flex;
  align-items: center;
  padding: 1px 2px;
  border-radius: 5px;
  margin: 1px 2px;
  vertical-align: text-bottom;
  font-size: 14px;
  background-color: #92a7b54a;
}

.img-stamp {
  display: inline-flex;
  align-items: center;
  padding: 1px 2px;
  border-radius: 5px;
  margin: 1px 2px;
  background-color: #92a7b54a;
}

.emoji-img {
  max-width: 20px;
  max-height: 20px;
  padding: 2px;
}

.selected-class {
  border: 1px solid #3498db;
}

.chgble {
  white-space: pre-wrap;
  word-wrap: break-word;
  overflow-y: hidden;
  width: 90%;
}
.msgBox {
  /*position: fixed;*/
  /*bottom: 10px;*/
  width: 90%;
}
@media screen and (min-width : 701px) { 

}

@media screen and (max-width : 700px) {
  .headTitle {
    margin-left: 50px;
  }
}
</style>

