import { useMessagesStore } from '../stores/messages.js';
import { getIDB, upsertData, deleteData, updOne } from '../my/indexDB.js';

export async function bookmark(pushData) {
  // arr = append(arr, "bookmark")
  // arr = append(arr, r.FormValue("messageID"))
  // arr = append(arr, r.FormValue("channelID"))
  // arr = append(arr, r.FormValue("parentID"))
  // arr = append(arr, r.FormValue("toggle"))

  const messageID = pushData[1];
  const channelID = pushData[2];
  const parentID = pushData[3];
  const toggle = pushData[4];

  let table = 'message';
  if (parentID) {
    table = 'thread';
  }
  console.log('before idb', table, messageID);
  const idb = await getIDB(table, messageID);
  console.log('idb', idb);
  const messagesStore = useMessagesStore();
  // async function updOne(table, key, columnName, columnValue) {
  updOne(table, messageID, 'bookmark', toggle)
    .catch((error) => {
      console.error(error);
    });
  const bm = {
    messageID: messageID,
    channelID: channelID
  };
  if (parentID) {
    bm.parentID = parentID
  }
  if (toggle) {
    upsertData(bm, 'bookmark', 'messageID', messageID)
      .catch((error) => {
        console.error(error);
      });
    messagesStore.upOne(messageID, 'bookmark', 1);
  } else {
    deleteData('bookmark', 'messageID', messageID)
      .catch((error) => {
        console.error(error);
      });
    messagesStore.delete(messageID);
  }
}


