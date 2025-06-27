import { useBookmarksStore } from '@/stores/bookmarks.js';
import { useMessagesStore } from '@/stores/messages.js';
import { removeMark } from '@/my/markdown.js';

export async function thread(pushData) {
  const bookmarksStore = useBookmarksStore();
  const messagesStore = useMessagesStore();
  const channelID = pushData[2];
  const updatedBy = pushData[3];
  const secondPartMsgID = pushData[4][1];
  const unixtime = base62Decode(secondPartMsgID.slice(0, -1));
	let filelinks = "";
	if (Array.isArray(pushData[5])) {
    pushData[5].forEach(filelink => {
      filelinks += `＊f＊${filelink}・＊f＊ `;
    });
	}
  const pushThread = {
    messageID: channelID + secondPartMsgID,
    parentID: pushData[4][0],
    messageTxt: pushData[4][2] + filelinks,
    aliasName: pushData[4][7] || updatedBy,
    aliasImg: pushData[4][3],
    createdAt: timeFormat('YYYY/MM/DD hh:mm:ss', unixtime * 1000),
    channelID: channelID,
    aliasNames: pushData[4][4],
    backID: pushData[4][5] || '',
    emojis: pushData[4][6] || []
  };

  const channel = await getIDB('channel', channelID);
  const aliases = await getIDBs('alias', 'channelIDIndex', channelID, 10000);
  const groups = await getIDBs('group', 'channelIDIndex', channelID, 10000);
  let title = getSubstring(removeMark(pushThread.messageTxt), 0, 30);
  if (pushData[3].includes('@')) {
    const parts = pushData[3].split('@');
    const toWhom = parts[0] === channel.myname ? parts[1] : parts[0];
    title = getSubstring(toWhom, 0, 12);
  }

  let parent = {};
  parent = await getIDB('thread', pushThread.parentID);
  let second = false;
  if (parent) {  // more than 2nd generation thread
    parent.threadCount = parent.threadCount ? parent.threadCount + 1 : 1;
    title = getSubstring(removeMark(parent.messageTxt), 0, 30);
    second = true;
  } else {
    parent = {
      messageID: pushThread.parentID,
      channelID: pushThread.channelID,
      messageTxt: pushThread.messageTxt,
      aliasName: pushThread.aliasName,
      aliasImg: pushThread.aliasImg,
      createdAt: pushThread.createdAt,
      emojis: pushThread.emojis,
      threadCount: 1
    };    
  }
  let displayStatus = 1;
  let notify = false;
  if (pushThread.messageTxt.includes('＠＠' + channel.myname + '・＠＠')) {
    displayStatus = 2;
    notify = true;
  }
  let pushTitle = find;
  let newThreadHeadFlag = false;
  let toWhom;
  let threadHead = await getIDB('threadHead', pushThread.parentID);
  if (threadHead) {
    if (threadHead.displayStatus != 3 || notify) {
      threadHead.displayStatus = displayStatus;
    }
    threadHead.updatedAt = pushThread.createdAt;
    threadHead.threadCount = threadHead.threadCount + 1;
  } else { // from reply first message
    threadHead = {};
    threadHead.parentID = pushThread.parentID;
    threadHead.emojis = parent.emojis;
    if (pushThread.parentID.includes('@')) {
      const parentSecondPart = pushThread.parentID.replace(channelID, '');
      const parts = parentSecondPart.split('@');
      const matchedGroup = groups.find(group => parts.includes(group.groupName));
      if (matchedGroup) {
        toWhom = parts.find(part => part !== matchedGroup.groupName);
      }
      if (parts.includes(channel.myname)) {
        toWhom = parts.find(part => part !== channel.myname);
      }
    }
    threadHead.title = toWhom || title;
    threadHead.displayStatus = displayStatus;
    threadHead.messageTxt = parent.messageTxt;
    threadHead.aliasName = parent.aliasName;
    threadHead.aliasImg = parent.aliasImg;
    threadHead.createdAt = parent.createdAt;
    threadHead.updatedAt = pushThread.createdAt;
    threadHead.threadCount = 0;
    threadHead.aliasNames = pushThread.aliasNames;
    threadHead.adminNames = pushThread.aliasNames;
    threadHead.backID = pushThread.backID;
    threadHead.channelID = channelID;
    newThreadHeadFlag = true;
  }
  if (displayStatus == 2 && pushThread.emojis.length > 0) {
    console.log('pushThread', pushThread);
    const bm = {
      messageID: pushThread.messageID,
      channelID: pushThread.channelID,
      title: getSubstring(removeMark(pushThread.messageTxt), 0, 20),
      displayStatus: 1,
      backID: pushThread.parentID
    };
    upsertIDB(bm, 'bookmark', 'messageID', pushThread.messageID);
    bookmarksStore.insert(bm);
    pushThread.bookmark = 1;
  }
  if (!newThreadHeadFlag || second) {
    upsertIDB(pushThread, 'thread', 'messageID', pushThread.messageID);
  }
  if (newThreadHeadFlag && second) {
    updIDBone('thread', pushThread.parentID, 'reply', true);
  }
  upsertIDB(threadHead, 'threadHead', 'parentID', threadHead.parentID);
  if (notify) {
    new Notification(pushTitle, {
      body: getSubstring(removeMark(pushThread.messageTxt), 0, 30), icon: pushThread.aliasImg
    });
  }
  if (messagesStore.currentDisplay(pushThread.parentID)) {
    messagesStore.insert(pushThread);
  }
  if (newThreadHeadFlag) {
    location.href = '';
  }
}
