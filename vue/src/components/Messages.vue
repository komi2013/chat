<script setup>
import { ref, computed, onBeforeMount } from 'vue'

// import '@vueup/vue-quill/dist/vue-quill.snow.css';

import DrawerColumn from '../components/DrawerColumn.vue'
import EditBox from '../components/EditBox.vue'
import EmojiModal from '../components/EmojiModal.vue';
import OtherModal from '../components/OtherModal.vue';

import { useMessagesStore } from '../stores/messages.js';
import { useChannelsStore } from '../stores/channels.js';
import { get_formated_time } from '../my/get_formated_time.js';
import { getIDB, getIDBs, upsertData, getAllIDBs } from '../my/indexDB.js';
import { isEmojiOpen, selectedMessageId, openEmoji, closeEmoji, selectEmoji, calcEmoji, emojiPath } from '../my/emoji.js';
import { toggleEdit, toggleBookmark } from '../my/other.js';

import { markdownToHtml } from '../my/markdown.js';

const props = defineProps({
  messages: Object,
  channel: Object,
  threadHead: Object
});
const channel = props.channel;
const messagesStore = useMessagesStore();
const messages = computed(() => {
  return messagesStore.messages;
});
const limit = 5;
let offset = 0;
let more = false;
const fetchMessages = (table, index, parentMessageID) => {
  return new Promise((resolve, reject) => {
    getIDBs(table, index, parentMessageID, limit, offset)
      .then((data) => {
        const latest = data.reverse();
        latest.forEach(message => {
          messagesStore.insert(message);
        });
        resolve();
        if (latest.length == limit) {
          offset += limit;
          more = true;
        } else {
          more = false;
        }
      });
  });
};

const moreMessages = () => {
  return new Promise((resolve, reject) => {
    getIDBs(table, index, parentMessageID, limit, offset)
      .then((data) => {
        const latest = data;
        latest.forEach(message => {
          messagesStore.unshift(message, addPosition);
        });
        resolve();
        if (latest.length == limit) {
          offset += limit;
          more = true;
        } else {
          more = false;
        }
      });
  });
};


function nextURL (message) {
  let URL = '';
  URL = '/thread/' + channel.channelID + '/' + message.messageID + '/';
  const params = {
    backID: ''
  };
  if (message.parentID) {
    params.backID = message.parentID;
    const queryString = createGetParams(params);
    return URL + '?' + queryString;
  } else {
    return URL;
  }
}

function createGetParams(params) {
  const queryString = Object.keys(params).map(key =>
    `${encodeURIComponent(key)}=${encodeURIComponent(params[key])}`
    ).join('&');
  return queryString;
}

const clickEmoji = (message, emoji) => {
  const fd = new FormData();
  if (message.parentID) {
    fd.append('parentID', message.parentID);
  }
  fd.append('channelID', channel.channelID);
  fd.append('messageID', message.messageID);
  fd.append('emojiValue', emoji[0]);
  fd.append('clicked', emoji[2] ? 1 : 0);
  fd.append('aliasName', channel.aliasName);
  const request = new Request('/EmojiToggle/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      alert(reason)
    })
};


const aliass = ref([]);
let table = 'message';
let index = 'channelIDIndex';
let parentMessageID = channel.channelID;
let addPosition = 0;
if (props.threadHead) {
  table = 'thread';
  index = 'parentIDIndex';
  parentMessageID = props.threadHead.parentID;
  addPosition = 1;
}

onBeforeMount(async () => {
  await fetchMessages(table, index, parentMessageID);
  aliass.value = await getAllIDBs('alias');
});

</script>


<template>
  <div class="messages" >
    <div v-if="more" @click="moreMessages" class="more"> - - more - - </div>
    <template v-for="(message, k) in messages" :key="message.messageID" >
    <table v-if="!more || k > 0">
      <tr>
        <td rowspan="2" class="icon_td">
          <img v-if="message.aliasImg" :src="message.aliasImg" class="icon">
        </td>
        <td>
          <span class="aliasName">{{ message.aliasName }}</span>
          <span class="dateTime" :id="'msg_'+message.messageID">{{ get_formated_time('MM-DD hh:mm', message.createdAt) }}</span>
        </td>
        <td class="setting">
          <span
            v-if="channel.aliasName == message.aliasName"
            :class="{ 'selected': message.editFlg }"
            @click="toggleEdit(message, message.messageID, $event)"> 🖋 </span>
          <span v-if="!message.threadCount"> <a :href="nextURL(message)"> 💬 </a> </span>
          <span
            :class="{ 'selected': message.bookmark }"
            @click="toggleBookmark(message)"> 🔖 </span>
          <span @click="openEmoji(message.messageID)"> 😄 </span>
            <EmojiModal
              :key="message.messageID"
              v-if="isEmojiOpen && selectedMessageId === message.messageID"
              @selectEmoji="selectEmoji"
              @closeEmoji="closeEmoji"
              :channelID="channel.channelID"
              :messageID="message.messageID"
              :aliasName="channel.aliasName"
              :parentID="message.parentID" />
        </td>
      </tr>
      <tr>
        <td v-if="message.editFlg" colspan="2" class="editText" :id="'for_content_' + messageID">
          <EditBox :channel="channel" :message="message" :threadHead="threadHead" />
        </td>
        <td v-else colspan="2" class="ql-container ql-snow" >
          <div
            v-html="markdownToHtml(message.messageTxt, channel, aliass)"
            class="ql-editor"></div>
          <div class="threads" v-if="message.threadCount">
            <a :href="'/thread/' + channel.channelID + '/' + message.messageID + '/'">
              <span>{{message.threadCount}} messages &nbsp;</span>
              <template v-for="img in message.threadImgs"><img :src="img" class="iconMini"></template>
            </a>
          </div>
          <template v-for="d in calcEmoji(message.emojis, channel.aliasName)">
            <template v-if="emojiPath(d[0])">
              <span class="img-stamp"
                :class="{ 'selected': d[2] }">
                  <img :src="d[0]" 
                    class="emoji-img" 
                    @click="clickEmoji(message, d)" /> {{d[1]}}
              </span>
            </template>
            <template v-else>
              <span class="emoji-stamp"
                :class="{ 'selected': d[2] }"
                @click="clickEmoji(message, d)" >
                {{ d[0] }} {{d[1]}}
              </span>
            </template>
          </template>
        </td>
      </tr>
    </table>
    </template>
  </div>
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
.icon {
  max-width: 50px;
  max-height: 50px;
  border-radius: 10%;
}
.icon_td {
  width: 50px;
  vertical-align: top;
  text-align: center;
}

.setting {
  text-align: right;
}

.setting span {
  margin: 2px;
  padding: 0px 2px;
  cursor: pointer;
}

.aliasName {
  margin: 2px;
  font-size: 12px;
}

.dateTime {
  margin: 2px;
  font-size: 12px;
}

.emoji-stamp {
  display: inline-flex;
  align-items: center;
  padding: 1px 2px;
  border-radius: 5px;
  margin: 1px 2px;
  vertical-align: text-bottom;
  font-size: 14px;
}

.selected {
  display: inline-flex;
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

.ql-editor {
  padding: 0;
}
.iconMini {
  max-width: 26px;
  max-height: 26px;
}
.threads a {
  cursor: pointer;
  display: flex;
  align-items: center;
}
.ql-container.ql-snow {
  border: none;
  font-size: 14px;
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
</style>
