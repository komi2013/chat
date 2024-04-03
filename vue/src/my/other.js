import { ref } from 'vue';
import { useMessagesStore } from '../stores/messages.js';

const messagesStore = useMessagesStore();

export const toggleEdit = (message, messageID) => {
  if (message.editFlg) {
    message.editFlg = false;
  } else {
    message.editFlg = true;
  }
  console.log(message);
  message.updateTxt = message.messageTxt;
  messagesStore.update(message, message.messageID);
};

export const toggleBookmark = (message) => {
  let toggle = 1;
  if (message.bookmark) {
    toggle = 0;
  }
  let parentID;
  let threadFlg = false;
  const fd = new FormData();
  if (message.parentID) {
    parentID = message.parentID;
    fd.append('parentID', parentID);
  }
  fd.append('messageID', message.messageID);
  fd.append('channelID', message.channelID);
  fd.append('toggle', toggle);
  const request = new Request('/BookmarkToggle/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      alert(reason)
    })
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

// export { toggleEdit };