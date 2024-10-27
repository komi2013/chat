import { useBookmarksStore } from '../stores/bookmarks.js';
import { useMessagesStore } from '../stores/messages.js';
import { getSubstring } from '../my/strings.js';
import { removeMark } from '../my/markdown.js';
export async function bookmark(pushData) {
  // arr = append(arr, "bookmark")
  // arr = append(arr, r.FormValue("messageID"))
  // arr = append(arr, r.FormValue("channelID"))
  // arr = append(arr, r.FormValue("parentID"))
  // arr = append(arr, r.FormValue("toggle"))

  const messageID = pushData[1];
  const channelID = pushData[2];
  const parentID = pushData[3];
  const toggle = pushData[4] == '1' ? 1 : 0;
  const idb = await getIDB('thread', messageID);
  const bookmarksStore = useBookmarksStore();
  const messagesStore = useMessagesStore();
  updIDBone('thread', messageID, 'bookmark', toggle)
    .catch((error) => {
      console.error(error);
    });
  const bm = {
    messageID: messageID,
    channelID: channelID,
    title: getSubstring(removeMark(idb.messageTxt), 0, 20),
    displayStatus: 1
  };
  if (parentID) {
    bm.parentID = parentID
  }
  messagesStore.upOne(messageID, 'bookmark', toggle);
  if (toggle) {
    upsertIDB(bm, 'bookmark', 'messageID', messageID)
      .catch((error) => {
        console.error(error);
      });
    bookmarksStore.insert(bm);
  } else {
    deleteIDB('bookmark', 'messageID', messageID)
      .catch((error) => {
        console.error(error);
      });
    bookmarksStore.delete(messageID);
  }
}


