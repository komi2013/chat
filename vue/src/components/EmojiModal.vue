<template>
  <div class="modal">
    <div class="modal-content">
      <template class="emoji-list" v-for="emoji in emojis">
        <template v-if="emojiPath(emoji)">
          <img :src="emoji" class="emoji-img" @click="selectEmoji(emoji)" />
        </template>
        <template v-else>
          <span class="emoji-img" @click="selectEmoji(emoji)" >{{ emoji }}</span>
        </template>
      </template>
 
      <button @click="closeModal">Close</button>
    </div>
  </div>
</template>

<script setup>
import { ref, defineProps, defineEmits } from 'vue';
import { emojiPath } from '../my/emoji.js';

const props = defineProps(['emojis', 'channelID', 'messageId', 'aliasName', 'parent_id']);
const emojis = ['😀','😁','/me.jpg','😂'];
const emit = defineEmits();

console.log('props.parent_id', props.parent_id);

const selectEmoji = (emoji) => {
  // emit('selectEmoji', emoji);
  const uri = props.parent_id ? '/ThreadEdit/' : '/MessageEdit/';
  const fd = new FormData();
  fd.append('channelID', props.channelID);
  fd.append('messageID', props.messageId);
  fd.append('emoji', emoji);
  fd.append('aliasName', props.aliasName);
  const request = new Request(uri, {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .then((response) => response.json())
    .then((json)=>{
      // when status not 1
    })
    .catch((reason)=>{
      console.log(reason)
    })
  emit('closeEmoji');
};

const closeModal = () => {
  emit('closeEmoji');
};

</script>

<style scoped>
/* Add your styles here */
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

.emoji-img {
  max-width: 20px;
  max-height: 20px;
  padding: 6px;
}

/* Add more styles as needed */
</style>
