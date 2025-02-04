import { useMessagesStore } from '@/stores/messages.js';

export async function emoji(pushData) {
  // const pushID = pushData[0];
  // const fd = new FormData();
  // fd.append('pushID', pushID);
  // const request = new Request('/PushResponse/', {
  //   method: 'POST',
  //   body: fd,
  // });
  // fetch(request);
  const channelID = pushData[2];
  const aliasName = pushData[3];
  const messageID = pushData[4][0];
  const emojiValue = pushData[4][1];
  const parentID = pushData[4][2];
  // const parentID = pushData[4][3];

  const table = messageID == parentID ? 'threadHead' : 'thread';
  const idb = await getIDB(table, messageID);
  const messagesStore = useMessagesStore();
  // console.log('idb', idb.emojis, aliasName, emojiValue);
  let emojis = idb.emojis;
  const is = emojis.findIndex(d => d.aliasName === aliasName && d.emoji === emojiValue);
  // console.log('is', is);
  if (is > -1) {
    emojis.splice(is, 1);
  } else {
    emojis = emojis || [];
    emojis.push({ aliasName, emoji: emojiValue });
  }

  updIDBone(table, messageID, 'emojis', emojis);
  messagesStore.upOne(messageID, 'emojis', emojis);
}


