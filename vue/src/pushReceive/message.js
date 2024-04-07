import { useMessagesStore } from '../stores/messages.js';
import { getIDB, upsertData, deleteData, getAllIDBs } from '../my/indexDB.js';
import { getSubstring, removeHtmlTags } from '../my/strings.js';
import { removeMark } from '../my/markdown.js';
export async function message(pushData) {
  const obj = {
    messageID: pushData[1],
    channelID: pushData[2],
    messageTxt: pushData[3],
    aliasName: pushData[4],
    aliasImg: pushData[5],
    createdAt: pushData[6]
  };
  const messagesStore = useMessagesStore();
  upsertData(obj, 'message', 'messageID', obj.messageID)
    .catch((error) => {
      console.error(error);
    });

  const alias = await getAllIDBs('alias');
  let displayStatus = 1;
  let notify = false;
  alias.forEach(d => {
    const atName = '＠＠' + d.aliasName + '・＠＠';
    if (obj.messageTxt.includes(atName) && d.groupFlg == 1) {
      displayStatus = 2;
    } else if (obj.messageTxt.includes(atName)) {
      displayStatus = 2;
      notify = true;
      return;
    }
  });
  let channel;
  try {
    channel = await getIDB('channel', obj.channelID);
    if (channel.displayStatus != 3 || notify) {
      channel.displayStatus = displayStatus;
    }
    channel.updatedAt = obj.createdAt;
  } catch (error) {
    console.log('this device dont have this channel but receive message', obj);
  }
  upsertData(channel, 'channel', 'channelID', obj.channelID)
    .catch((error) => {
      console.error(error);
    });

  if (notify) {
    new Notification(channel.channelName, { body: getSubstring(removeMark(obj.messageTxt), 0, 8), icon: obj.aliasImg });
  }
  if (messagesStore.currentDisplay(obj.channelID)) {
    messagesStore.insert(obj);
  }
}


