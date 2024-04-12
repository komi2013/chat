import { useMessagesStore } from '../stores/messages.js';
import { getIDB, upsertData, deleteData, getAllIDBs } from '../my/indexDB.js';
import { getSubstring } from '../my/strings.js';
import { removeMark } from '../my/markdown.js';

export async function messageEdit(pushData) {
  const rcv = {
    messageID: pushData[1],
    messageTxt: pushData[2],
    yets: pushData[3]
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
    let emojis;
    if (rcv.yets) {
      if (idb.emojis) {
        const filtered = idb.emojis.filter(
          ([name, url]) => !rcv.yets.some(([n, u]) => n === name && u === url)
        );
        emojis = rcv.yets.concat(filtered);
      } else {
        emojis = rcv.yets;
      }
    } else {
      emojis = idb.emojis;
    }
    const obj = {
      messageID: idb.messageID,
      channelID: idb.channelID,
      messageTxt: rcv.messageTxt,
      aliasName: idb.aliasName,
      aliasImg: idb.aliasImg,
      createdAt: idb.createdAt,
      emojis: emojis
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
      console.error('this device dont have this channel but receive message', obj);
    }
    upsertData(channel, 'channel', 'channelID', obj.channelID)
      .catch((error) => {
        console.error(error);
      });
    if (notify) {
      new Notification(channel.channelName, {
        body: getSubstring(removeMark(obj.messageTxt), 0, 30), icon: obj.aliasImg
      });
    }
    if (messagesStore.currentDisplay(obj.channelID)) {
      messagesStore.update(obj, idb.messageID);
    }
  }
}
