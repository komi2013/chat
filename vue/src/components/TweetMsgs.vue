<script setup>
import { ref, computed, onMounted } from 'vue';
import TweetEmojiModal from '@/components/TweetEmojiModal.vue';
import EmojiedModal from '@/components/EmojiedModal.vue';
import { useMessagesStore } from '@/stores/messages';
import { isEmojiOpen, selectedMessageId, openEmoji, closeEmoji, selectEmoji, calcEmoji, emojiPath, isEmojiedOpen, openEmojied, closeEmojied, rotateEmoji } from '@/my/emoji';
import { toggleEdit, toggleBookmark } from '@/my/toggle.js';
import { markdownToHtml } from '@/my/markdown';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  messages: Object,
  threadHead: Object,
  copyable: Boolean,
  messageID: String
});

const myname = localStorage.getItem('myname') || ''
function tF(a, b = null){ return timeFormat(a, b) }

const messagesStore = useMessagesStore();
const messages = computed(() => {
  return messagesStore.messages
});

const limit = 20;
let offset = 0
let offsetNew = 0
const more = ref(false)
const moreNew = ref(false)
const parentMessageID = props.threadHead ? props.threadHead.parentID : ''
const moreMessages = async (later = false) => {
  let threads = []
  if (props.messageID) {
    // threads = await 
  } else {
    // threads = await 
  }
  threads.forEach((message, index) => {
    message.href = `/thread//${message.parentID}/?messageID=${message.messageID}`
    if (later) {
      if (index < limit || offset === 0) {
        messagesStore.insert(message)
      }
    } else {
      if (index < limit) {
        messagesStore.unshift(message, 1)
      }
    }
  })
  if (later) {
    offsetNew = offsetNew + limit
    moreNew.value = threads.length > limit
  } else {
    offset = offset + limit
    more.value = threads.length > limit
  }
}

function editable(myname, message) {
  if (props.threadHead.inquirerFlag) return false
  const now = Date.now()
  const createdAtTimestamp = new Date(message.createdAt).getTime();
  const within10min = Math.floor((now - createdAtTimestamp) / (1000 * 60)) < 10;
  if (within10min && myname === message.nickname) {
    return true
  }
}

function replyable (message) {
  if (message.messageID !== message.parentID && !message.backID) {
    return true;
  } else {
    return false;
  }
}

let threadPosition
let threadPositionNew
let threadCount
onMounted(async () => {
  // if (props.messageID) {
  //   moreNew.value = true
  //   more.value = true
  //   await moreMessages(true)
  // } else {
  //   await moreMessages()
  // }
})

function reply(message) {
  const messageID = message.messageID;
  let URL = `/thread//${secondPart}/`;
  let queryString = message.parentID && message.parentID !== message.messageID && !message.reply
    ? `?${createGetParams({ backID: message.parentID })}` 
    : '';
  location.href = URL + queryString;
}

const clickEmoji = async (message, emoji) => {
  const del = message.emojis.some(e => e.nickname === localStorage.get('nickname') && e.emoji === emoji.emoji)
  const fd = new FormData();
  const contents = [message.messageID, emoji.emoji, message.parentID, del]
  fd.append('contents', JSON.stringify(contents));
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ContentsPush/', fd);
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
  rotateEmoji(emoji.emoji);
};

</script>

<template>
  <div v-if="more" @click="moreMessages(false)" class="more"> - - more - - </div>
  <div class="messages" :contenteditable="copyable">
    <template v-for="(message, k) in messages" :key="message.messageID" >
      <div v-if="!more || k > 0">
        <div class="msg-header">
          <div v-if="!copyable" class="icon_td">
            <a :href="'/people/' + '' + '/' + message.nickname + '/'">
              <img v-if="message.nickImg && message.nickImg.charAt(0) != ','" 
                :src="message.nickImg" class="icon-img">
              <span v-if="message.nickImg && message.nickImg.charAt(0) == ','"
                class="icon-span" 
                :style="'background-color:' + message.nickImg.split(',')[2] ">
                {{message.nickImg.split(',')[1]}}</span>
            </a>
          </div>
          <div v-if="!copyable">
            <span class="nickname">{{ message.nickname }}</span>
            <span class="dateTime" :id="'msg_'+message.messageID">
              <a :href="message.href">
                {{ tF('MM-DD hh:mm', message.createdAt) }}
              </a>
            </span>
          </div>
          <div v-if="copyable" class="name-time">{{ message.nickname }} {{ tF('MM-DD hh:mm', message.createdAt) }}</div>
          <div v-if="!copyable" class="setting">
            <span
              v-if="editable(myname, message)"
              :class="{ 'selected': message.editFlg }"
              @click="toggleEdit(message, message.messageID, $event)"> ✏️ </span>
            <span v-if="replyable(message)"
                  :class="[{ 'selected': message.reply }, 'message-wrapper']">
              <a :class="{'message-button': message.reply}" @click="reply(message)"> 💬 </a>
              <span v-if="message.threadCount" @click="reply(message)" class="badge">{{ message.threadCount}}</span>
            </span>
            <span
              :class="{ 'selected': message.bookmark }"
              @click="toggleBookmark(message, threadHead)"> 🔖 </span>
            <span @click="openEmoji(message.messageID)"> 😄 </span>
          </div>
        </div>
        <div v-if="!message.editFlg" colspan="3" class="ql-container ql-snow" >
          <div
            v-html="markdownToHtml(message.messageTxt)"
            class="ql-editor"></div>
          <template v-for="emoji in calcEmoji(message.emojis, myname)">
            <template v-if="emojiPath(emoji.emoji)">
              <span class="img-stamp"
                :class="{ 'selected': emoji.selected }">
                  <img :src="emoji.emoji" 
                    class="emoji-img" 
                    @click="clickEmoji(message, emoji)" />{{emoji.count}}
              </span>
            </template>
            <template v-if="!emojiPath(emoji.emoji)">
              <span class="emoji-stamp"
                :class="{ 'selected': emoji.selected }"
                @click="clickEmoji(message, emoji)" >
                {{ emoji.emoji }}{{emoji.count}}
              </span>
            </template>
          </template>
          <span v-if="calcEmoji(message.emojis, myname).length"
            class="emojied"
            @click="openEmojied(message.messageID)" >&nbsp;⋮&nbsp;
          </span>
        </div>
      </div>
      <EmojiedModal
        :key="message.messageID"
        v-if="isEmojiedOpen && selectedMessageId === message.messageID"
        @closeEmojied="closeEmojied"
        :messageID="message.messageID"
        :parentID="message.parentID"
        :emojis="message.emojis" />
      <TweetEmojiModal
        :key="message.messageID"
        v-if="isEmojiOpen && selectedMessageId === message.messageID"
        @selectEmoji="selectEmoji"
        @closeEmoji="closeEmoji"
        :messageID="message.messageID"
        :parentID="message.parentID"
        :threadHead="threadHead" />
    </template>
  </div>
  <div v-if="moreNew" @click="moreMessages(true)" class="more"> - - more - - </div>
</template>

<style>

blockquote {
  margin: 0;
  padding: 3px; 
  margin-left: 5px;
  border-left: 3px solid #ccc;
  background-color: #f9f9f9;
}
code {
  display: block;
  background-color: #c0c0c029;
  border: 1px solid silver;
  margin: 3px;
}

.msg-header {
  display: flex;
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

.setting {
  display: flex;
  margin-left: auto;
}

.setting span {
  margin: 2px;
  padding: 0px 2px;
  cursor: pointer;
  display: inline-block;
}

.dateTime {
  margin: 2px;
  font-size: 12px;
}

.name-time {
  font-size: 11px;
}

.emoji-stamp {
  display: inline-flex;
  align-items: center;
  padding: 1px 2px;
  border-radius: 5px;
  margin: 1px 2px;
  vertical-align: text-bottom;
  font-size: 14px;
  height: 26px;
}

.selected {
  background-color: #92a7b54a;
  border-radius: 5px;
  border: 1px solid #3498db;
}

.img-stamp {
  display: inline-flex;
  align-items: center;
  padding: 1px 2px;
  border-radius: 5px;
  margin: 1px 2px;
}

.emoji-img {
  max-width: 23px;
  max-height: 23px;
  padding: 2px;
}

.emojied {
  vertical-align: bottom;
  background-color: aliceblue;
  border-radius: 4px;
  padding: 0px 4px;
  cursor: pointer;
}

.ql-editor {
  padding: 0;
}

.ql-container.ql-snow {
  border: none;
  font-size: 14px;
  padding: 5px;
}

.ql-snow .ql-editor a {
  text-decoration: none;
  color: #06c;
}

.mentioned {
  background-color: #a7cad63d;
  color: blue;
}
.mentionme {
  background-color: #ffee00;
  color: blue;
}
.more {
  background-color: silver;
/*  color: #fff;*/
  margin: 8px;
  border-radius: 4px;
  cursor: pointer;
  text-align: center;
}

.message-wrapper {
  position: relative;
  width: 20px;
}
.message-button {
  position: absolute;
  opacity: 0.3;
}
.badge {
  position: absolute;
  font-size: 12px;
  width: max-content;
  right: 0px;
}
</style>
