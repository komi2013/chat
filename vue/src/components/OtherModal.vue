<template>
  <div class="modal">
    <div class="modal-content">
        <div @click="deleteMessage"> delete </div>
        <div @click="editMessage"> edit </div>

      <button @click="closeModal">Close</button>
    </div>
  </div>
</template>

<script setup>
import { ref, defineProps, defineEmits } from 'vue';

const props = defineProps(['channelID', 'messageId', 'aliasName', 'message']);
const emojis = ['😀','😁','/me.jpg','😂'];
const emit = defineEmits();

console.log('props', props.message, props.aliasName);

const deleteMessage = () => {
  const fd = new FormData();
  fd.append('channelID', props.channelID);
  fd.append('messageID', props.messageId);
  fd.append('aliasName', props.aliasName);
  const request = new Request('/MessageEdit/', {
    method: 'POST',
    body: fd,
  });
  fetch(request).catch((reason)=>{alert(reason);})
  emit('closeOther');
};

const editMessage = () => {
  console.log('activeEdit', props.message, props.messageId);
  emit('activeEdit', props.message, props.messageId);
};

const closeModal = () => {
  emit('closeOther');
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
