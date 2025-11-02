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
      <div v-if="errorMessage">
        <p style="color:red;">{{ errorMessage }}</p>
      </div>      
    </div>

  </div>
</template>

<script setup>
import { ref, defineProps, defineEmits } from 'vue';

import { emojiPath } from '@/my/emoji.js';
import { validateEmoji, rotateEmoji, masterEmojis } from '@/my/emoji';
import { pushReceive } from '@/pushReceive/pushReceive.js';


import { useMessagesStore } from '@/stores/messages'

const props = defineProps([
  'messageID',
  'parentID'
]);

const errorMessage = ref('')

const emit = defineEmits()
const messagesStore = useMessagesStore()
const selectEmoji = async (emoji) => {
  const fd = new FormData()
  fd.append('parentID', props.parentID)
  fd.append('messageID', props.messageID)
  fd.append('emoji', emoji);
  fd.append('csrf', localStorage.getItem('csrf'))
  const res = await sendRequest('/TweetEmoji/', fd)
  res.csrf && localStorage.setItem('csrf', res.csrf)
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
  rotateEmoji(emoji)
  if ( res.error ) {
    errorMessage.value = res.error
  } else {
    messagesStore.upOne(res.message.messageID, 'emojis', res.message.emojis)
    emit('closeEmoji')
  }
};

const selectedEmoji = ref('');
const emojiValidErr = ref(false);
const inputEmoji = async () => {
  if (!validateEmoji(selectedEmoji.value)) {
    emojiValidErr.value = true;
    return;
  }
  const fd = new FormData()
  fd.append('parentID', props.parentID)
  fd.append('messageID', props.messageID)
  fd.append('emoji', emoji);
  fd.append('csrf', localStorage.getItem('csrf'))
  const res = await sendRequest('/TweetEmoji/', fd);
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  })

  rotateEmoji(selectedEmoji.value)
  if ( res.error ) {
    errorMessage.value = res.error
  } else {
    messagesStore.upOne(res.message.messageID, 'emojis', res.message.emojis)
    emit('closeEmoji')
  }
}

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
