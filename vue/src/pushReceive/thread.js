import { useBookmarksStore } from '../stores/bookmarks.js';
import { useMessagesStore } from '../stores/messages.js';
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
  const unixtime = base62Decode(pushData[2].slice(4, 10));

  const channelID = pushData[2].slice(0, 4);
  const obj = {
    messageID: pushData[2],
    parentID: channelID + pushData[3],
    messageTxt: pushData[4],
    aliasName: pushData[5],
    aliasImg: pushData[6],
    createdAt: timeFormat('YYYY/MM/DD hh:mm:ss', unixtime * 1000),
    channelID: channelID,
    threadType: pushData[7] ?? '',
    aliasNames: pushData[8],
    backID: pushData[9] ?? '',
    emojis: pushData[10]
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
  const channel = await getIDB('channel', channelID);
  let title = getSubstring(removeMark(obj.messageTxt), 0, 30);
  // const dm = props.threadHead.parentID.includes('@');
  if (pushData[3].includes('@')) {
    const parts = pushData[3].split('@');
    const toWhom = parts[0] === channel.aliasName ? parts[1] : parts[0];
    title = getSubstring(toWhom, 0, 12);
  }
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
  let displayStatus = 1;
  let notify = false;
  channel.allAliases.forEach(d => {
    const atName = '＠＠' + d[0] + '・＠＠';
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
    threadHead.aliasNames = obj.aliasNames;
    threadHead.backID = obj.backID;
    threadHead.threadType = obj.threadType;
    newThread = true;
  }
  if (displayStatus == 2 && obj.emojis) {
    const bm = {
      messageID: obj.messageID,
      channelID: obj.channelID,
      title: getSubstring(removeMark(obj.messageTxt), 0, 20),
      displayStatus: 1
    };
    upsertIDB(bm, 'bookmark', 'messageID', obj.messageID);
    bookmarksStore.insert(bm);
    obj.bookmark = 1;
  }
  // console.log(newThread);
  if (!newThread) {
    upsertIDB(obj, 'thread', 'messageID', obj.messageID);
  }
  // if (obj.backID) {
  //   upsertIDB(parent, 'thread', 'messageID', obj.parentID);
  // }
  upsertIDB(threadHead, 'threadHead', 'parentID', obj.parentID);
  if (notify) {
    new Notification(pushTitle, {
      body: getSubstring(removeMark(obj.messageTxt), 0, 30), icon: obj.aliasImg
    });
  }
  if (messagesStore.currentDisplay(obj.parentID)) {
    messagesStore.insert(obj);
  }
  if (newThread) {
    location.href = '';
  }
}
