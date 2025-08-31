import { useBookmarksStore } from '@/stores/bookmarks.js';
import { useMessagesStore } from '@/stores/messages.js';
import { removeMark } from '@/my/markdown.js';
export async function bookmark(pushData) {
  const bookmarksStore = useBookmarksStore();
  const messagesStore = useMessagesStore();
  const channelID = pushData[2];
  const aliasName = pushData[3];
  const messageID = pushData[4][0]
  const backID = pushData[4][1]
  const toggle = pushData[4][2]
  const thread = await getIDB('thread', messageID);
  // const toggle = !thread.bookmark;
  updIDBone('thread', messageID, 'bookmark', toggle);
  const bm = {
    messageID: messageID,
    channelID: channelID,
    backID: backID,
    title: getSubstring(removeMark(thread.messageTxt), 0, 20),
    displayStatus: 1
  };
  if (messageID == backID) {
    updIDBone('threadHead', backID, 'bookmark', toggle);
  }
  if (toggle) {
    upsertIDB(bm, 'bookmark', 'messageID', messageID);
    bookmarksStore.insert(bm);
  } else {
    deleteIDB('bookmark', 'messageID', messageID);
    bookmarksStore.delete(messageID);
  }
  messagesStore.upOne(messageID, 'bookmark', toggle);
}


