import { useBookmarksStore } from '../stores/bookmarks.js';
import { useMessagesStore } from '../stores/messages.js';
import { getIDB, upsertData, deleteData, getAllIDBs } from '../my/indexDB.js';
import { getSubstring, removeHtmlTags } from '../my/strings.js';
import { removeMark } from '../my/markdown.js';
export async function thread(pushData) {
  const bookmarksStore = useBookmarksStore();
  const messagesStore = useMessagesStore();
  const pushID = pushData[0];
  const fd = new FormData();
  fd.append('pushID', pushID);
  const request = new Request('/PushResponse/', {
    method: 'POST',
    body: fd,
  });
  fetch(request);
  const obj = {
    messageID: pushData[2],
    parentID: pushData[3],
    messageTxt: pushData[4],
    aliasName: pushData[5],
    aliasImg: pushData[6],
    createdAt: pushData[7],
    channelID: pushData[8],
    threadType: pushData[9],
    backID: pushData[10],
    emojis: pushData[11]
  };
  let threadHead = {
    parentID: obj.parentID,
    messageTxt: obj.messageTxt,
    channelID: obj.channelID,
    displayStatus: 1
  };
  if (obj.backID) {
    threadHead.backID = obj.backID;
  }
  let title = getSubstring(removeMark(obj.messageTxt), 0, 30);
  let parent = {};
  try {
    parent = await getIDB('thread', obj.parentID);
    parent.threadCount = parent.threadCount ? parent.threadCount + 1 : 1;
    parent.threadImgs = parent.threadImgs || [];
    if (!parent.threadImgs.includes(obj.aliasImg)) {
        parent.threadImgs.push(obj.aliasImg);
    }
    title = getSubstring(removeMark(parent.messageTxt), 0, 30);
  } catch (error) {
    parent = {
      messageID: obj.parentID,
      channelID: obj.channelID,
      messageTxt: obj.messageTxt,
      aliasName: obj.aliasName,
      aliasImg: obj.aliasImg,
      createdAt: obj.createdAt,
      emoji: obj.emojis,
      threadCount: 1,
      threadImgs: [obj.aliasImg]
    };
  }
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
  let pushTitle = title;
  let newThread = false;
  try {
    threadHead = await getIDB('threadHead', obj.parentID);
    if (threadHead.displayStatus != 3 || notify) {
      threadHead.displayStatus = displayStatus;
    }
    threadHead.updatedAt = obj.createdAt;
    threadHead.threadCount = parent.threadCount;
    pushTitle = threadHead.title;
  } catch (error) {
    threadHead.emojis = parent.emojis;
    threadHead.title = title;
    threadHead.displayStatus = displayStatus;
    threadHead.messageTxt = parent.messageTxt;
    threadHead.aliasName = parent.aliasName;
    threadHead.aliasImg = parent.aliasImg;
    threadHead.createdAt = parent.createdAt;
    threadHead.updatedAt = obj.createdAt;
    threadHead.threadCount = parent.threadCount;
    newThread = true;
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
  console.log(newThread);
  if (!newThread) {
    upsertData(obj, 'thread', 'messageID', obj.messageID);
  }
  if (obj.backID) {
    upsertData(parent, 'thread', 'messageID', obj.parentID);
  }
  upsertData(threadHead, 'threadHead', 'parentID', obj.parentID);
  if (notify) {
    new Notification(pushTitle, {
      body: getSubstring(removeMark(obj.messageTxt), 0, 30), icon: obj.aliasImg
    });
  }
  if (messagesStore.currentDisplay(obj.parentID)) {
    messagesStore.insert(obj);
  }
}
