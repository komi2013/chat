<template>
  <div class="modal" @click.self="closeModal">
    <div class="modal-content">
      <template class="emoji-list" v-for="emoji in masterEmojis">
        <template v-if="emojiPath(emoji)">
          <span>
            <img class="emoji-img" :src="emoji" @click="selectEmoji(emoji)" />
          </span>
        </template>
        <template v-else>
          <span class="emoji" @click="selectEmoji(emoji)" >{{ emoji }}</span>
        </template>
      </template>

      <input 
        type="text" 
        v-model="selectedEmoji" 
        maxlength="2" 
        class="emoji-input"
        @change="inputEmoji"
      />
 
      <button @click="closeModal"> x </button>
      <br>
      <span v-if="emojiValidErr" class="emoji-valid-err">絵文字か1文字にしてください</span>
    </div>
  </div>
</template>

<script setup>
import { ref, defineProps, defineEmits } from 'vue';

import { emojiPath } from '@/my/emoji.js';
import { userIDsByName } from '@/my/channelFunc';
import { validateEmoji, rotateEmoji, masterEmojis } from '@/my/emoji';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps([
  'channel',
  'aliases',
  'groups',
  'messageID',
  'parentID',
  'threadHead'
]);

const emit = defineEmits();

const selectEmoji = async (emoji) => {
  const fd = new FormData();
  fd.append('channelID', localStorage.getItem('channelID'));
  fd.append('updatedBy', localStorage.getItem('myname'));
  fd.append('userIDs', JSON.stringify(userIDsByName(props.aliases, props.threadHead.aliasNames)));
  fd.append('pushTitle', 'emoji');
  const contents = [props.messageID, emoji, props.parentID];
  fd.append('contents', JSON.stringify(contents));
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ContentsPush/', fd);
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });

  rotateEmoji(emoji);
  emit('closeEmoji');
};

const selectedEmoji = ref('');
const emojiValidErr = ref(false);
const inputEmoji = async () => {
  if (!validateEmoji(selectedEmoji.value)) {
    emojiValidErr.value = true;
    return;
  }
  const fd = new FormData();
  fd.append('channelID', localStorage.getItem('channelID'));
  fd.append('updatedBy', localStorage.getItem('myname'));
  fd.append('userIDs', JSON.stringify(userIDsByName(props.aliases, props.threadHead.aliasNames)));
  fd.append('pushTitle', 'emoji');
  const contents = [props.messageID, selectedEmoji.value, props.parentID];
  fd.append('contents', JSON.stringify(contents));
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ContentsPush/', fd);
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
  rotateEmoji(selectedEmoji.value);
  emit('closeEmoji');
};


const closeModal = () => {
  emit('closeEmoji');
};

</script>

<style scoped>

.modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 20;
}

.modal-content {
  background: #fff;
  padding: 20px;
  border-radius: 8px;
}

.emoji-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.emoji {
  max-width: 20px;
  max-height: 20px;
  padding: 6px;
}

.emoji-img {
  max-width: 20px;
  max-height: 20px;
  vertical-align: middle;
  padding: 6px;
}
.emoji-input {
  width: 24px;
}

</style>
