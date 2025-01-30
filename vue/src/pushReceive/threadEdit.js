import { useBookmarksStore } from '../stores/bookmarks.js';
import { useMessagesStore } from '../stores/messages.js';
import { removeMark } from '../my/markdown.js';

export async function threadEdit(pushData) {
  const pushID = pushData[0];
  const fd = new FormData();
  fd.append('pushID', pushID);
  const request = new Request('/PushResponse/', {
    method: 'POST',
    body: fd,
  });
  fetch(request);
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
  const editThread = {
    messageID: channelID + secondPartMsgID,
    parentID: pushData[4][0],
    messageTxt: pushData[4][2] + filelinks,
    aliasImg: pushData[4][3],
    aliasNames: pushData[4][4],
    emojis: pushData[4][6] || []
  };
  const messagesStore = useMessagesStore();
  if(editThread.aliasImg == ''){
    deleteIDB('thread', 'messageID', editThread.messageID);
    messagesStore.delete(editThread.messageID);
    return;
  }

  const channel = await getIDB('channel', channelID);
  const aliases = await getIDBs('alias', 'channelIDIndex', channelID, 10000);
  const groups = await getIDBs('group', 'channelIDIndex', channelID, 10000);
  let thread = await getIDB('thread', editThread.messageID);
  let threadHead = await getIDB('threadHead', editThread.parentID);

	const mergedEmojis = [
	  ...thread.emojis,
	  ...editThread.emojis.filter(
	    editEmoji => !thread.emojis.some(threadEmoji => threadEmoji.aliasName === editEmoji.aliasName)
	  )
	];
	thread.emojis = mergedEmojis;
	thread.messageTxt = editThread.messageTxt;
	thread.aliasNames = editThread.aliasNames;
	thread.updatedAt = editThread.updatedAt;
	if (threadHead) {
		threadHead.emojis = thread.emojis;
		threadHead.messageTxt = thread.messageTxt;
		threadHead.aliasNames = thread.aliasNames;
		threadHead.updatedAt = thread.updatedAt;
		upsertIDB(threadHead, 'threadHead', 'parentID', threadHead.parentID);
	}
  let displayStatus = 1;
  let notify = false;
  groups.forEach(d => {
    const atName = '＠＠' + d.groupName + '・＠＠';
    if (thread.messageTxt.includes(atName)) {
    	if (d.aliasNames.includes(channel.myname)) {
	      displayStatus = 2;
    	}
    }
  });
  if (editThread.messageTxt.includes('＠＠' + channel.myname + '・＠＠')) {
    displayStatus = 2;
    notify = true;
  }
  const bookmarksStore = useBookmarksStore();
  if (displayStatus == 2 && editThread.emojis) {
    const bm = {
      messageID: thread.messageID,
      channelID: thread.channelID,
      title: getSubstring(removeMark(thread.messageTxt), 0, 20),
      displayStatus: 1
    };
    upsertIDB(bm, 'bookmark', 'messageID', bm.messageID);
    bookmarksStore.insert(bm);
  }
  upsertIDB(thread, 'thread', 'messageID', thread.messageID);
  if (notify) {
    new Notification(getSubstring(removeMark(thread.messageTxt), 0, 20), {
      body: getSubstring(removeMark(thread.messageTxt), 0, 30), icon: thread.aliasImg
    });
  }
  if (messagesStore.currentDisplay(thread.parentID)) {
    messagesStore.update(thread, thread.messageID);
  }
}
