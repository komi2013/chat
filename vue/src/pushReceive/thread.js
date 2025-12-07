import { useBookmarksStore } from '@/stores/bookmarks.js';
import { useMessagesStore } from '@/stores/messages.js';
import { removeMark } from '@/my/markdown.js';

export async function thread(pushData) {
  const bookmarksStore = useBookmarksStore();
  const messagesStore = useMessagesStore();
  const channelID = pushData[2];
  const updatedBy = pushData[3];
  const messageID = pushData[4][1]
  const directPush = pushData.directPush
  const unixtime = base62Decode(messageID.slice(0, -1));
	let filelinks = "";
	// if (Array.isArray(pushData[5])) {
 //    pushData[5].forEach(filelink => {
 //      filelinks += `＊f＊${filelink}・＊f＊ `;
 //    });
	// }
  if (Array.isArray(pushData[5])) {
    pushData[5].forEach(filelink => {
      const lower = filelink.toLowerCase();
      if (lower.endsWith('.jpg') || lower.endsWith('.jpeg') || lower.endsWith('.png')) {
        filelinks += `＊img＊${filelink}・＊img＊ `;
      } else {
        filelinks += `＊f＊${filelink}・＊f＊ `;
      }
    });
  }
  const thread = await getIDB('thread', messageID)
  if (thread) return
  const pushThread = {
    messageID: messageID,
    parentID: pushData[4][0],
    messageTxt: pushData[4][2] + filelinks,
    aliasName: pushData[4][7] || updatedBy,
    aliasImg: pushData[4][3],
    createdAt: timeFormat('YYYY/MM/DD hh:mm:ss', unixtime),
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

  const parent = await getIDB('thread', pushThread.parentID);
  if (parent) {  // more than 2nd generation thread
    parent.threadCount = parent.threadCount ? parent.threadCount + 1 : 1;
    parent.reply = true
  }
  let displayStatus = 1;
  let notify = false;
  if (pushThread.messageTxt.includes('＠＠' + channel.myname + '・＠＠')) {
    displayStatus = 2;
    notify = true;
  }
  let pushTitle = getSubstring(removeMark(pushThread.messageTxt.replace(/＠＠[^・＠]+・＠＠/g, '')), 0, 10)

  if (displayStatus == 2 && pushThread.emojis.length > 0) {
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
  upsertIDB(pushThread, 'thread', 'messageID', pushThread.messageID)
  if (parent) {
    upsertIDB(parent, 'thread', 'messageID', pushThread.parentID)
  }
  updIDBone('threadHead', pushThread.parentID, 'displayStatus', displayStatus)
  if (notify && directPush) {
    new Notification(pushTitle, {
      body: getSubstring(removeMark(pushThread.messageTxt), 0, 30), icon: pushThread.aliasImg
    })
    addFaviconBadge()
  }
  if (messagesStore.currentDisplay(pushThread.parentID)) {
    messagesStore.insert(pushThread);
  }
}
