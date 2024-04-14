import { useBookmarksStore } from '../stores/bookmarks.js';
import { useMessagesStore } from '../stores/messages.js';
import { getIDB, upsertData, deleteData, getAllIDBs } from '../my/indexDB.js';
import { getSubstring, removeHtmlTags } from '../my/strings.js';
import { removeMark } from '../my/markdown.js';
export async function message(pushData) {
  // arr = append(arr, insertedID)
  // arr = append(arr, channelID)
  // arr = append(arr, message.MessageTxt)
  // arr = append(arr, aliasName)
  // arr = append(arr, aliasImg)
  // arr = append(arr, time.Now())
  const bookmarksStore = useBookmarksStore();
  const messagesStore = useMessagesStore();
  const obj = {
    messageID: pushData[1],
    channelID: pushData[2],
    messageTxt: pushData[3],
    aliasName: pushData[4],
    aliasImg: pushData[5],
    createdAt: pushData[6],
    emojis: pushData[7]
  };
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
    console.log('this device dont have this channel but receive message', obj);
  }
  if (displayStatus == 2) {
    const bm = {
      messageID: obj.messageID,
      channelID: obj.channelID,
      title: getSubstring(removeMark(obj.messageTxt), 0, 20),
      displayStatus: 1
    };
    upsertData(bm, 'bookmark', 'messageID', obj.messageID);
    bookmarksStore.insert(bm);
    obj.bookmark = 1;
  }
  upsertData(obj, 'message', 'messageID', obj.messageID);
  upsertData(channel, 'channel', 'channelID', obj.channelID);

  if (notify) {
    new Notification(channel.channelName, { body: getSubstring(removeMark(obj.messageTxt), 0, 20), icon: obj.aliasImg });
  }
  if (messagesStore.currentDisplay(obj.channelID)) {
    messagesStore.insert(obj);
  }
}


