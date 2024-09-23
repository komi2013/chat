import { useMessagesStore } from '../stores/messages.js';
import { getIDB, upsertData, deleteData, updOne } from '../my/indexDB.js';

export async function emoji(pushData) {
  const messageID = pushData[2];
  const aliasName = pushData[3];
  const editType = pushData[4];
  const emojiValue = pushData[5];
  const parentID = pushData[6];
  const table = messageID == parentID ? 'threadHead' : 'thread';
  const idb = await getIDB(table, messageID);
  // let idb;
  // if (messageID == parentID) {
    
  // } else {
  //   idb = await getIDB('thread', messageID);
  // }
  

  const messagesStore = useMessagesStore();

  let emojis = idb.emojis // [['kom1','/me.jpg'],['kom2','✋']]
  if (editType == 1) {
    if (emojis) {
      emojis.push([aliasName, emojiValue]);
    } else {
      emojis = [[aliasName, emojiValue]];
    }
  } else {
    const is = emojis.findIndex(d => d[0] === aliasName && d[1] === emojiValue);
    emojis.splice(is, 1);
  }
  updOne(table, messageID, 'emojis', emojis)
    .catch((error) => {
      console.error(error);
    });
  messagesStore.upOne(messageID, 'emojis', emojis);

}


