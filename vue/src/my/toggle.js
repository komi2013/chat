import { ref } from 'vue';
import { useMessagesStore } from '@/stores/messages.js';
import { userIDsByName } from '@/my/channelFunc';
import { pushReceive } from '@/pushReceive/pushReceive.js';

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

export const toggleBookmark = async (message, channel, aliases, threadHead) => {
  const backID = message.parentID ?? threadHead.backID;
  const fd = new FormData();
  fd.append('channelID', channel.channelID);
  fd.append('updatedBy', channel.myname);
  fd.append('userIDs', JSON.stringify(userIDsByName(aliases, [channel.myname])));
  fd.append('pushTitle', 'bookmark');
  const contents = [message.messageID, backID];
  fd.append('contents', JSON.stringify(contents));
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ContentsPush/', fd);
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
};
