import { ref } from 'vue';
import { useMessagesStore } from '../stores/messages.js';

const messagesStore = useMessagesStore();

const toggleEdit = (message, messageID) => {
  if (message.editFlg) {
    message.editFlg = false;
  } else {
    message.editFlg = true;
  }
  message.updateTxt = message.messageTxt;
  messagesStore.update(message, message.messageID);
};

// export const isOtherOpen = ref(false);

// export const selectedOther = ref('');

// export const otherMessageId = ref(null);

// const openOther = (messageId) => {
//   otherMessageId.value = messageId;
//   isOtherOpen.value = true;
// };

// const closeOther = () => {
//   isOtherOpen.value = false;
//   otherMessageId.value = null;
// };

// const selectOther = (emoji) => {
//   selectedOther.value = emoji;
//   console.log(selectedOther.value);
//   closeOther();
// };

// const adjustHeight = () => {
//   const textarea = document.querySelector('.chgble');
//   textarea.style.height = 'auto';
//   textarea.style.height = `${textarea.scrollHeight}px`;
// };

export { toggleEdit };