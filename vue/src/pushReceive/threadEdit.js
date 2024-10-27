import { useBookmarksStore } from '../stores/bookmarks.js';
import { useMessagesStore } from '../stores/messages.js';
import { getSubstring } from '../my/strings.js';
import { removeMark } from '../my/markdown.js';

export async function threadEdit(pushData) {
  const rcv = {
    messageID: pushData[1],
    messageTxt: pushData[2],
    yets: pushData[3]
  };
  const bookmarksStore = useBookmarksStore();
  const messagesStore = useMessagesStore();
  if(rcv.messageTxt == ''){
    deleteIDB('thread', 'messageID', rcv.messageID);
    messagesStore.delete(rcv.messageID);
  } else {
    const idb = await getIDB('thread', rcv.messageID);
    const obj = {
      parentID: idb.parentID,
      messageID: idb.messageID,
      channelID: idb.channelID,
      messageTxt: rcv.messageTxt,
      aliasName: idb.aliasName,
      aliasImg: idb.aliasImg,
      createdAt: idb.createdAt
    };
    let emojis;
    if (rcv.yets) {
      if (idb.emojis) {
        const filtered = idb.emojis.filter(
          ([name, url]) => !rcv.yets.some(([n, u]) => n === name && u === url)
        );
        emojis = rcv.yets.concat(filtered);
      } else {
        emojis = rcv.yets;
      }
    } else {
      emojis = idb.emojis;
    }
    obj.emojis = emojis;
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
      if (obj.messageTxt.includes(atName)) {
        displayStatus = 2;
        if (d.groupFlg != 1) {
          notify = true;
        }
      }
    });

    let threadHead;
    try {
      threadHead = await getIDB('threadHead', obj.parentID);
      if (threadHead.displayStatus != 3 || notify) {
        threadHead.displayStatus = displayStatus;
      }
      threadHead.updatedAt = obj.createdAt;
    } catch (error) {
      console.log('this device dont have this threadHead but receive message', obj);
    }
    if (displayStatus == 2) {
      const bm = {
        messageID: idb.messageID,
        channelID: idb.channelID,
        title: getSubstring(removeMark(rcv.messageTxt), 0, 20),
        displayStatus: 1
      };
      upsertIDB(bm, 'bookmark', 'messageID', idb.messageID);
      bookmarksStore.insert(bm);
      obj.bookmark = 1;
    }
    upsertIDB(obj, 'thread', 'messageID', idb.messageID);
    upsertIDB(threadHead, 'threadHead', 'parentID', obj.parentID);
    if (notify) {
      new Notification(threadHead.title, {
        body: getSubstring(removeMark(obj.messageTxt), 0, 30), icon: obj.aliasImg
      });
    }
    if (messagesStore.currentDisplay(obj.parentID)) {
      messagesStore.update(obj, idb.messageID);
    }
  }

}
