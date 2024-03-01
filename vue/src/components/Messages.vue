<script setup>
import { ref, computed, onBeforeMount } from 'vue'
import DrawerColumn from '../components/DrawerColumn.vue'
import EditBox from '../components/EditBox.vue'
import { useMessagesStore } from '../stores/messages.js';
import { useChannelsStore } from '../stores/channels.js';
import { get_formated_time } from '../my/get_formated_time.js';
import { getIDB, getIDBs, upsertData } from '../my/indexDB.js';
import { isEmojiOpen, selectedMessageId, openEmoji, closeEmoji, selectEmoji, calcEmoji, emojiPath } from '../my/emoji.js';
import { isOtherOpen, otherMessageId, openOther, closeOther, selectOther, activeEdit, adjustHeight, textareaRefs } from '../my/other.js';
import { textToHtml } from '../my/textToHtml.js';
import EmojiModal from '../components/EmojiModal.vue';
import OtherModal from '../components/OtherModal.vue';

const props = defineProps({
  messages: Object,
  channel: Object
});

const messages = props.messages;
const channel = props.channel;

const clickEmoji = (messageId, emoji) => {
  const fd = new FormData();
  fd.append('parentMessageID', props.message_id);
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


</script>


<template>
  <div class="messages" >
    <table v-for="message in messages" :key="message.messageID">
      <tr>
        <td rowspan="2" class="icon_td">
          <img v-if="message.aliasImg" :src="message.aliasImg" class="icon">
        </td>
        <td>
          <span class="emoji">{{ channel.aliasName }}</span>
          <span class="emoji">{{ get_formated_time('MM/DD hh:mm', message.createdAt) }}</span>
        </td>
        <td class="setting">
          <span class="emoji" @click="activeEdit(message, message.messageID, $event)"> 🖋 </span>
          <span class="emoji" @click="openEmoji(message.messageID)"> 😄 </span>
          <EmojiModal :key="message.messageID" v-if="isEmojiOpen && selectedMessageId === message.messageID" @selectEmoji="selectEmoji" @closeEmoji="closeEmoji" :channelID="channel.channelID" :messageId="message.messageID" :aliasName="channel.aliasName" :parent_id="message.parentID" />

          <span> <RouterLink :to="'/thread/' + channel.channelID + '/' + message.messageID + '/'"> 💬 </RouterLink> </span>
          <span class="emoji" @click="openOther(message.messageID)"> &nbsp; ⋮ &nbsp; </span>
          <OtherModal :key="message.messageID" v-if="isOtherOpen && otherMessageId === message.messageID" @selectOther="selectOther" @closeOther="closeOther" @activeEdit="activeEdit" :channelID="channel.channelID" :messageId="message.messageID" :aliasName="channel.aliasName" :message="message" />
        </td>
      </tr>
      <tr>
        <td v-if="message.editFlg" colspan="2" >
          <EditBox :channel="channel" :message="message" :parent_id="message.parentID" />
        </td>
        <td v-else colspan="2" >
          <div v-html="textToHtml(message.messageTxt)"></div>
          <template v-for="d in calcEmoji(message.emojis, channel.aliasName)">
            <template v-if="emojiPath(d[0])">
              <span class="img-stamp" :class="{ 'selected': d[2] }"> <img :src="d[0]" class="emoji-img" @click="clickEmoji(message.messageID, d)" /> {{d[1]}} </span>
            </template>
            <template v-else>
              <span class="emoji-stamp" :class="{ 'selected': d[2] }" @click="clickEmoji(message.messageID, d)" >{{ d[0] }} {{d[1]}} </span>
            </template>
          </template>
        </td>
      </tr>
    </table>
  </div>
</template>

<style>

.icon {
  max-width: 50px;
  max-height: 50px;
}
.icon_td {
  width: 50px;
  vertical-align: top;
}

.setting {
  text-align: right;
}

.emoji {
  margin: 2px;
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


</style>
