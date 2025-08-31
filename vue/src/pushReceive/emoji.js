import { useMessagesStore } from '@/stores/messages.js';

export async function emoji(pushData) {
  const channelID = pushData[2];
  const aliasName = pushData[3];
  const messageID = pushData[4][0];
  const emojiValue = pushData[4][1];
  const parentID = pushData[4][2]
  const del = pushData[4][3] ? true : false
  // console.log('del', del)
  const table = messageID == parentID ? 'threadHead' : 'thread';
  const idb = await getIDB(table, messageID);
  const messagesStore = useMessagesStore();
  let emojis = idb.emojis;


  const position = emojis.findIndex(d => d.aliasName === aliasName && d.emoji === emojiValue);
  if (del) {
    emojis.splice(position, 1)
  } else {
    emojis = emojis || []
    // emojis.push({ aliasName, emoji: emojiValue })
    addEmoji(emojis, aliasName, emojiValue)
  }
  updIDBone(table, messageID, 'emojis', emojis);
  messagesStore.upOne(messageID, 'emojis', emojis);
}

function addEmoji(emojis, aliasName, emojiValue) {
  const exists = emojis.some(e => e.aliasName === aliasName && e.emoji === emojiValue);
  if (!exists) {
    emojis.push({ aliasName, emoji: emojiValue });
  }
  return emojis;
}

