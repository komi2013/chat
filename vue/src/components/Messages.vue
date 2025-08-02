<script setup>
import { ref, computed, onMounted } from 'vue';
import EditBox from '@/components/EditBox.vue';
import EmojiModal from '@/components/EmojiModal.vue';
import EmojiedModal from '@/components/EmojiedModal.vue';
import { useMessagesStore } from '@/stores/messages';
import { useChannelsStore } from '@/stores/channels';
import { isEmojiOpen, selectedMessageId, openEmoji, closeEmoji, selectEmoji, calcEmoji, emojiPath, isEmojiedOpen, openEmojied, closeEmojied, rotateEmoji } from '@/my/emoji';
import { toggleEdit, toggleBookmark } from '@/my/toggle.js';
import { markdownToHtml } from '@/my/markdown';
import { userIDsByName } from '@/my/channelFunc';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  messages: Object,
  channel: Object,
  aliases: Array,
  groups: Array,
  threadHead: Object,
  copyable: Boolean
});

function tF(a, b = null){ return timeFormat(a, b) }

const messagesStore = useMessagesStore();
const messages = computed(() => {
  return messagesStore.messages;
});

const limit = 20;
let offset = 0;
let more = false;
const moreMessages = async () => {
  const parentMessageID = props.threadHead ? props.threadHead.parentID : props.channel.channelID;
  const addPosition = props.threadHead ? 1 : 0;
  const thread = await getIDBs('thread', 'parentIDIndex', parentMessageID, limit, offset);
  thread.forEach(message => {
  	if (addPosition === 1) {
  		messagesStore.unshift(message, addPosition);
  	} else {
  		messagesStore.insert(message);
  	}
  });
  if (thread.length === limit) {
    offset += limit;
    more = true;
  } else {
    more = false;
  }
};

function editable(myname, message) {
  const now = Date.now();
  const createdAtTimestamp = new Date(message.createdAt).getTime();
  const within10min = Math.floor((now - createdAtTimestamp) / (1000 * 60)) < 10;
  if (within10min && myname === message.aliasName) {
    return true;
  }
  const isInSameGroup = props.groups.some(group => 
    within10min && group.aliasNames.includes(myname) && group.groupName === message.aliasName
  );
  return isInSameGroup;
}

function replyable (message) {
	if (message.messageID !== message.parentID && !message.backID) {
		return true;
	} else {
		return false;
	}
}

onMounted(async () => {
  await moreMessages();
});

function reply(message) {
  const channelID = props.channel.channelID;
  const messageID = message.messageID;
  const secondPart = messageID.replace(channelID, '');
  let URL = `/thread/${channelID}/${secondPart}/`;
  let queryString = message.parentID && message.parentID !== message.messageID && !message.reply
    ? `?${createGetParams({ backID: message.parentID })}` 
    : '';
  location.href = URL + queryString;
}

const clickEmoji = async (message, emoji) => {
  const fd = new FormData();
  fd.append('channelID', props.channel.channelID);
  fd.append('updatedBy', props.channel.myname);
  fd.append('userIDs', JSON.stringify(userIDsByName(props.aliases, props.threadHead.aliasNames)));
  fd.append('pushTitle', 'emoji');
  const contents = [message.messageID, emoji.emoji, message.parentID];
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
  <div class="messages" >
    <div v-if="more" @click="moreMessages" class="more"> - - more - - </div>
    <template v-for="(message, k) in messages" :key="message.messageID" >
      <div v-if="!more || k > 0">
        <!-- <tr> -->
          <div class="msg-header">
            <div v-if="!copyable" class="icon_td">
              <a :href="'/people/' + message.channelID + '/' + message.aliasName + '/'">
                <img v-if="message.aliasImg && message.aliasImg.charAt(0) != ','" 
                  :src="message.aliasImg" class="icon-img">
                <span v-if="message.aliasImg && message.aliasImg.charAt(0) == ','"
                  class="icon-span" 
                  :style="'background-color:' + message.aliasImg.split(',')[2] ">
                  {{message.aliasImg.split(',')[1]}}</span>
              </a>
            </div>
            <div class="name_time">
              <span class="aliasName">{{ message.aliasName }} {{message.spentMinute}}</span>
              <span class="dateTime" :id="'msg_'+message.messageID">{{ tF('MM-DD hh:mm', message.createdAt) }}</span>
            </div>
            <div v-if="!copyable" class="setting">
              <span
                v-if="editable(channel.myname, message)"
                :class="{ 'selected': message.editFlg }"
                @click="toggleEdit(message, message.messageID, $event)"> ✏️ </span>
              <span v-if="replyable(message)"
              			:class="{ 'selected': message.reply }">
              	<a @click="reply(message)"> 💬 </a>
              </span>
              <span
                :class="{ 'selected': message.bookmark }"
                @click="toggleBookmark(message, channel, aliases, threadHead)"> 🔖 </span>
              <span @click="openEmoji(message.messageID)"> 😄 </span>
            </div>
          </div>
<!--         </tr>
        <tr> -->
          <div v-if="message.editFlg" colspan="3" class="editText" :id="'for_content_' + messageID">
			      <EditBox
			        :channel="channel"
			        :aliases="aliases"
			        :groups="groups"
			        :message="message"
			        :threadHead="threadHead" />
          </div>
          <div v-if="!message.editFlg" colspan="3" class="ql-container ql-snow" >
            <div
              v-html="markdownToHtml(message.messageTxt, channel)"
              class="ql-editor"></div>
            <template v-for="emoji in calcEmoji(message.emojis, channel.myname)">
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
            <span v-if="calcEmoji(message.emojis, channel.myname).length"
              class="emojied"
              @click="openEmojied(message.messageID)" >&nbsp;⋮&nbsp;
            </span>
          </div>
        <!-- </tr> -->
      </div>
      <EmojiedModal
        :key="message.messageID"
        v-if="isEmojiedOpen && selectedMessageId === message.messageID"
        @closeEmojied="closeEmojied"
        :channelID="channel.channelID"
        :messageID="message.messageID"
        :myname="channel.myname"
        :parentID="message.parentID"
        :emojis="message.emojis" />
      <EmojiModal
        :key="message.messageID"
        v-if="isEmojiOpen && selectedMessageId === message.messageID"
        @selectEmoji="selectEmoji"
        @closeEmoji="closeEmoji"
        :channel="channel"
        :aliases="aliases"
        :groups="groups"
        :messageID="message.messageID"
        :parentID="message.parentID"
        :threadHead="threadHead" />

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

.name_time {
  display: flex;
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

/*.threads a {
  cursor: pointer;
  display: flex;
  align-items: center;
}*/
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
</style>
