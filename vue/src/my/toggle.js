import { ref } from 'vue';
import { useMessagesStore } from '@/stores/messages.js';
import { userIDsByName } from '@/my/channelFunc';

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

export const toggleBookmark = (message, channel, aliases, threadHead) => {
  const backID = message.parentID ?? threadHead.backID;
  const fd = new FormData();
  fd.append('channelID', channel.channelID);
  fd.append('updatedBy', channel.myname);
  fd.append('userIDs', JSON.stringify(userIDsByName(aliases, [channel.myname])));
  fd.append('pushTitle', 'bookmark');
  const contents = [message.messageID, backID];
  fd.append('contents', JSON.stringify(contents));
  sendRequest('/ContentsPush/', fd);
};
