import { useMessagesStore } from '../stores/messages.js';
import { getIDB, upsertData, deleteData } from '../my/indexDB.js';

export async function messageEdit(pushData) {
  const rcv = {
    messageID: pushData[1],
    messageTxt: pushData[2],
    task: pushData[3]
  };
  const messagesStore = useMessagesStore();
  if(rcv.messageTxt == ''){
    deleteData('message', 'messageID', rcv.messageID)
      .catch((error) => {
        console.error(error);
      });
      messagesStore.delete(rcv.messageID);
  } else {
    const idb = await getIDB('message', rcv.messageID);
    const obj = {
      messageID: idb.messageID,
      channelID: idb.channelID,
      messageTxt: rcv.messageTxt,
      aliasName: idb.aliasName,
      aliasImg: idb.aliasImg,
      createdAt: idb.createdAt,
      emojis: idb.emojis
    };
    upsertData(obj, 'message', 'messageID', idb.messageID)
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
      messagesStore.update(obj, idb.messageID);
    }
  }
}


