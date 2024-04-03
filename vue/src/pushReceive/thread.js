import { useMessagesStore } from '../stores/messages.js';
import { getIDB, upsertData, deleteData, getAllIDBs } from '../my/indexDB.js';
import { getSubstring, removeHtmlTags } from '../my/strings.js';
import { removeMark } from '../my/markdown.js';
export async function thread(pushData) {
  const messagesStore = useMessagesStore();
  const obj = {
    messageID: pushData[1],
    parentID: pushData[2],
    messageTxt: pushData[3],
    aliasName: pushData[4],
    aliasImg: pushData[5],
    createdAt: pushData[6],
    channelID: pushData[7],
    backID: pushData[8]
  };
  upsertData(obj, 'thread', 'messageID', obj.messageID)
    .catch((error) => {
      console.error(error);
    });
  let threadHead = {
    parentID: obj.parentID,
    messageTxt: obj.messageTxt,
    channelID: obj.channelID,
    displayStatus: 1
  };
  let table = 'message';
  if (obj.backID) {
    table = 'thread';
    threadHead.backID = obj.backID;
  }
  let title = 'edit your own title as you like';
  let parent = {};
  try {
    parent = await getIDB(table, obj.parentID);
    parent.threadCount = parent.threadCount ? parent.threadCount + 1 : 1;
    parent.threadImgs = parent.threadImgs || [];
    if (!parent.threadImgs.includes(obj.aliasImg)) {
        parent.threadImgs.push(obj.aliasImg);
    }
    title = getSubstring(removeMark(parent.messageTxt), 0, 8);
  } catch (error) {
    parent = {
      messageID: obj.parentID,
      channelID: obj.channelID,
      messageTxt: 'new to this thread',
      aliasName: obj.aliasName,
      aliasImg: obj.aliasImg,
      createdAt: obj.createdAt,
      emoji: [],
      threadCount: 1,
      threadImgs: [obj.aliasImg]
    };
  }
  upsertData(parent, table, 'messageID', obj.parentID)
    .catch((error) => {
      console.error('parent, table', error);
    });
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
  try {
    threadHead = await getIDB('threadHead', obj.parentID);
    if (threadHead.displayStatus != 3 || notify) {
      threadHead.displayStatus = displayStatus;
    }
    threadHead.updatedAt = obj.createdAt;
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
  }
  upsertData(threadHead, 'threadHead', 'parentID', obj.parentID)
    .catch((error) => {
      console.error(error);
    });

  if (notify) {
    new Notification(pushTitle, { body: getSubstring(removeMark(obj.messageTxt), 0, 8), icon: obj.aliasImg });
  }
  if (messagesStore.currentDisplay(obj.parentID)) {
    messagesStore.insert(obj);
  }
}
