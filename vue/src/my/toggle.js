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
