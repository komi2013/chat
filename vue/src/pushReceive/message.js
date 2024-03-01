import { useMessagesStore } from '../stores/messages.js';
import { getIDB, upsertData, deleteData } from '../my/indexDB.js';

export async function message(pushData) {
  const messagesStore = useMessagesStore();
  const obj = {
    messageID: pushData[1],
    channelID: pushData[2],
    messageTxt: pushData[3],
    aliasName: pushData[4],
    aliasImg: pushData[5],
    createdAt: pushData[6]
  };
  upsertData(obj, 'message', 'messageID', pushData[1])
    .then((message) => {
      console.log(message);
    })
    .catch((error) => {
      console.error(error);
    });
  messagesStore.update(obj);
}


