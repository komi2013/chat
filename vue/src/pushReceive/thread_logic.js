import { useBookmarksStore } from '../stores/bookmarks.js';
import { useMessagesStore } from '../stores/messages.js';
import { removeMark } from '../my/markdown.js';
import { fetchChannel, fetchAliases, fetchGroups } from '@/my/channelFunc';

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

const firstGenerationThreadHead = ["Q8vnKdUkfLW3","thread","I0JH1takLrQ","sOP","＊p＊1sthread・＊p＊","kom1",",🚦,#081522","",["kom1"],"",""]

const firstGenerationThread = ["ma5n7PSCSSfI","thread","I0JH1takMwX","sOP","＊p＊hiiii・＊p＊","kom1",",🚦,#081522","",["kom1"],"",""]

const secondGenerationThreadHead = ["f5u6E9qkqTjJ","thread","I0JH1takNUI","I0JH1takMwX","＊p＊ppp・＊p＊","kom1",",🚦,#081522","",["kom1"],"I0JHsOP",""];

const secondGenerationThread = ["wIOcc1aykQP1","thread","I0JH1takNtB","I0JH1takMwX","＊p＊ppp・＊p＊","kom1",",🚦,#081522","",["kom1"],"I0JHsOP",""]

  const unixtime = base62Decode(pushData[2].slice(4, 10));
  const channelID = pushData[2].slice(0, 4);
  const messageID = pushData[2];
  const parentID = pushData[3];
  const messageTxt = pushData[4];
  const createdBy = pushData[5];
  const aliasImg = pushData[6];
  const createdAt = timeFormat('YYYY/MM/DD hh:mm:ss', unixtime * 1000);
  const threadType = pushData[7] ?? '';
  const aliasNames = pushData[8];
  const backID = pushData[9] ?? '';
  const emojis = pushData[10] || [];

  // const obj = {
  //   messageID: pushData[2],
  //   parentID: pushData[3], // parentID = pushData[3]
  //   messageTxt: pushData[4],
  //   aliasName: pushData[5],
  //   aliasImg: pushData[6],
  //   createdAt: timeFormat('YYYY/MM/DD hh:mm:ss', unixtime * 1000),
  //   channelID: channelID,
  //   threadType: pushData[7] ?? '',
  //   aliasNames: pushData[8],
  //   backID: pushData[9] ?? '',
  //   emojis: pushData[10] || []
  // };

  const channel = await getIDB('channel', channelID);
  const aliases = await getIDBs('alias', 'channelIDIndex', channelID, 10000);
  const groups = await getIDBs('group', 'channelIDIndex', channelID, 10000);
  const threadHead = await getIDB('threadHead', parentID);
  const thread = await getIDB('thread', parentID);
  let title = getSubstring(removeMark(messageTxt), 0, 30);
  if (parentID.includes('@')) {
    const parts = parentID.split('@');
    const toWhom = parts[0] === channel.myname ? parts[1] : parts[0];
    title = getSubstring(toWhom, 0, 12);
  }

  let parent = {};
  if (!thread && !threadHead) { // become threadHead 1st gene
  } else if (!thread && threadHead) { // become thread 1st gene
  } else if (thread && !threadHead) { // become threadHead pre thread 2nd and insert this thread 
  } else if (thread && threadHead) { // become thread 2nd gene thread is null
    parent.threadCount = parent.threadCount ? parent.threadCount + 1 : 1;
    parent.threadImgs = parent.threadImgs || [];
    if (!parent.threadImgs.includes(aliasImg)) {
        parent.threadImgs.push(aliasImg);
    }
    title = getSubstring(removeMark(parent.messageTxt), 0, 30);
  
  } else if (!thread && threadHead) { // become thread 1st gene thread
  } else if (!thread && !threadHead) { 
  } else {
    parent = {
      messageID: parentID,
      channelID: channelID,
      messageTxt: messageTxt,
      aliasName: aliasName,
      aliasImg: aliasImg,
      createdAt: createdAt,
      emojis: emojis,
      threadCount: 1,
      threadImgs: [aliasImg]
    };
  }
  let displayStatus = 1;
  let notify = false;
  if (messageTxt.includes('＠＠' + channel.myname + '・＠＠')) {
    displayStatus = 2;
    notify = true;
    // return;
  }
  let pushTitle = title;
  let newThread = false;
  
  if (threadHead) { // thread or threadHead
    if (threadHead.displayStatus != 3 || notify) {
      threadHead.displayStatus = displayStatus;
    }
    threadHead.updatedAt = createdAt;
    threadHead.threadCount = parent.threadCount;
    pushTitle = threadHead.title;
  } else {
    threadHead = {};
    console.warn('jjj', threadHead.emojis);
    threadHead.parentID = obj.parentID;
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
    threadHead.channelID = channelID;
    newThread = true;

    // let threadHead = {
    //   parentID: obj.parentID,
    //   messageTxt: obj.messageTxt,
    //   channelID: obj.channelID,
    //   displayStatus: 1
    // };
    // console.warn('obj', obj);
    // if (obj.backID) {
    //   threadHead.backID = obj.backID;
    // }

 
  }
  console.warn('threadHead', threadHead, obj);

  // try {
  //   threadHead = await getIDB('threadHead', obj.parentID);
  //   if (threadHead.displayStatus != 3 || notify) {
  //     threadHead.displayStatus = displayStatus;
  //   }
  //   threadHead.updatedAt = obj.createdAt;
  //   threadHead.threadCount = parent.threadCount;
  //   pushTitle = threadHead.title;
  // } catch (error) {
  //   threadHead.emojis = parent.emojis;
  //   threadHead.title = title;
  //   threadHead.displayStatus = displayStatus;
  //   threadHead.messageTxt = parent.messageTxt;
  //   threadHead.aliasName = parent.aliasName;
  //   threadHead.aliasImg = parent.aliasImg;
  //   threadHead.createdAt = parent.createdAt;
  //   threadHead.updatedAt = obj.createdAt;
  //   threadHead.threadCount = parent.threadCount;
  //   threadHead.aliasNames = obj.aliasNames;
  //   threadHead.backID = obj.backID;
  //   threadHead.threadType = obj.threadType;
  //   newThread = true;
  // }
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
  console.log('newThread', newThread);
  if (!newThread) {
    console.log('obj', obj);
    upsertIDB(obj, 'thread', 'messageID', obj.messageID);
  }
  // if (obj.backID) {
  //   upsertIDB(parent, 'thread', 'messageID', obj.parentID);
  // }
  upsertIDB(threadHead, 'threadHead', 'parentID', threadHead.parentID);
  if (notify) {
    new Notification(pushTitle, {
      body: getSubstring(removeMark(obj.messageTxt), 0, 30), icon: obj.aliasImg
    });
  }
  if (messagesStore.currentDisplay(obj.parentID)) {
    messagesStore.insert(obj);
  }
  if (newThread) {
    // location.href = '';
  }
}
