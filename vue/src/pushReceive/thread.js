import { useMessagesStore } from '../stores/messages.js';
import { getIDB, upsertData, deleteData } from '../my/indexDB.js';
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
    aliasName: obj.aliasName,
    aliasImg: obj.aliasImg,
    createdAt: obj.createdAt,
    channelID: obj.channelID,
    unreadFlg: true
  };
  let table = 'message';
  if (obj.backID) {
    table = 'thread';
    // parent = await getIDB('thread', obj.parentID);
    // parent.threadCount = parent.threadCount ? parent.threadCount + 1 : 1;
    // upsertData(parent, 'thread', 'messageID', obj.parentID)
    //   .catch((error) => {
    //     console.error(error);
    //   });
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
    console.log('parent', error);
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
  console.log('clone??', parent, table, obj);
  upsertData(parent, table, 'messageID', obj.parentID)
    .catch((error) => {
      console.error('parent, table', error);
    });
  try {
    threadHead = await getIDB('threadHead', obj.parentID);
    threadHead.unreadFlg = true;
    console.log('threadHead..', threadHead);
  } catch (error) {
    console.log('no threadHead', error);
    threadHead.emojis = parent.emojis;
    threadHead.title = title;
    threadHead.messageTxt = parent.messageTxt;
    threadHead.aliasName = parent.aliasName;
    threadHead.createdAt = parent.createdAt;
  }
  upsertData(threadHead, 'threadHead', 'parentID', obj.parentID)
    .catch((error) => {
      console.error(error);
    });

  // messagesStore.insert(obj);

}
