<template>
  <div class="modal">
    <div class="modal-content">
      <table>
      <tr class="emoji-list" v-for="emoji in emojis">
        <td>
        <template v-if="emojiPath(emoji.emoji)">
          <img :src="emoji.emoji" class="emoji-img" />
        </template>
        <template v-else>
          <span class="emoji-img" >{{ emoji.emoji }}</span>
        </template>
        </td>
        <td>{{ emoji.aliasName }}</td>
      </tr>
      </table>
      <button @click="closeModal">Close</button>
    </div>
  </div>
</template>

<script setup>
import { ref, defineProps, defineEmits } from 'vue';
import { emojiPath } from '@/my/emoji.js';

const props = defineProps(['emojis', 'channelID', 'messageID', 'aliasName', 'parentID']);
// const emojis = props.emojis;
const emit = defineEmits();

const closeModal = () => {
  emit('closeEmojied');
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
  z-index: 3;
}

.modal-content {
  background: #fff;
  padding: 20px;
  border-radius: 8px;
  max-height: 80%;
  overflow-y: auto; 
}

/*.emoji-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}*/

.emoji-img {
  max-width: 30px;
  max-height: 30px;
  padding: 6px;
}

/* Add more styles as needed */
</style>
