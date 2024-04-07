import { useMessagesStore } from '../stores/messages.js';
import { getIDB, upsertData, deleteData } from '../my/indexDB.js';

export async function threadEdit(pushData) {
  const rcv = {
    messageID: pushData[1],
    messageTxt: pushData[2],
    task: pushData[3]
  };
  const messagesStore = useMessagesStore();
  if(rcv.messageTxt == ''){
    deleteData('thread', 'messageID', rcv.messageID)
      .catch((error) => {
        console.error(error);
      });
      messagesStore.delete(rcv.messageID);
  } else {
    const idb = await getIDB('thread', rcv.messageID);
    const obj = {
      parentID: idb.parentID,
      messageID: idb.messageID,
      channelID: idb.channelID,
      messageTxt: rcv.messageTxt,
      aliasName: idb.aliasName,
      aliasImg: idb.aliasImg,
      createdAt: idb.createdAt,
      emojis: idb.emojis
    };
    upsertData(obj, 'thread', 'messageID', idb.messageID)
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

    let threadHead;
    try {
      threadHead = await getIDB('threadHead', obj.parentID);
      if (threadHead.displayStatus != 3 || notify) {
        threadHead.displayStatus = displayStatus;
      }
      threadHead.updatedAt = obj.createdAt;
    } catch (error) {
      console.log('this device dont have this threadHead but receive message', obj);
    }
    upsertData(threadHead, 'threadHead', 'parentID', obj.parentID)
      .catch((error) => {
        console.error(error);
      });
    if (notify) {
      new Notification(pushTitle, { body: getSubstring(removeMark(obj.messageTxt), 0, 8), icon: obj.aliasImg });
    }
    if (messagesStore.currentDisplay(obj.parentID)) {
      messagesStore.insert(obj);
    }
  }

}


