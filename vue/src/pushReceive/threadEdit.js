import { useMessagesStore } from '../stores/messages.js';
import { getIDB, upsertData, deleteData } from '../my/indexDB.js';

export async function threadEdit(pushData) {
  console.log('other file', pushData);
  const idb = await getIDB('thread', pushData[1]);
  console.log('idb', idb);
  const messagesStore = useMessagesStore();
  if(pushData[3] == 3){
    deleteData('thread', 'messageID', idb.messageID)
      .then((message) => {
        console.log(message);
      })
      .catch((error) => {
        console.error(error);
      });
      messagesStore.delete(idb.messageID);
  } else {
    let emojis = idb.emojis // [['kom1','/me.jpg'],['kom2','✋']]
    let messageTxt = idb.messageTxt
    if (pushData[3] == 2) {
      messageTxt = pushData[4];
    }
    if (pushData[3] == 1) {
      if (emojis) {
        const is = emojis.findIndex(item => item[0] === pushData[2]);
        if (is !== -1) {
          emojis.splice(is, 1);
        } else {
          emojis.push([pushData[2], pushData[4]]);
        }
      } else {
        emojis = [[pushData[2], pushData[4]]];
      }

    }

    const obj = {
      messageID: idb.messageID,
      channelID: idb.channelID,
      messageTxt: messageTxt,
      aliasName: idb.aliasName,
      aliasImg: idb.aliasImg,
      createdAt: idb.createdAt,
      emojis: emojis
    };
    upsertData(obj, 'thread', 'messageID', idb.messageID)
      .then((message) => {
        console.log(message);
      })
      .catch((error) => {
        console.error(error);
      });
    messagesStore.update(obj, idb.messageID);

  }

}


