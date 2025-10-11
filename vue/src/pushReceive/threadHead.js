// import { fetchChannel, fetchAliases, fetchGroups, userIDsByName } from '@/my/channelFunc';
// import { useMessagesStore } from '@/stores/messages.js';

export async function threadHead(pd) {
  const channelID = pd[2];
  const updatedBy = pd[3];
  let editThreadHead = pd[4];
  const pre = await getIDB('threadHead', editThreadHead.parentID) || editThreadHead;
  let threadHead;
  if (pre) {
    threadHead = JSON.parse(JSON.stringify(pre))
  }
  console.log(editThreadHead)
  threadHead.title = editThreadHead.title;
  threadHead.description = editThreadHead.description;
  threadHead.aliasNames = editThreadHead.aliasNames;
  threadHead.adminNames = editThreadHead.adminNames;
  threadHead.broadcastFlag = editThreadHead.broadcastFlag;
  const channel = await getIDB('channel', channelID)
  const parentID = threadHead.parentID;
  if (parentID) {
    if (parentID.startsWith('@')) {
      const nameTail = parentID.slice(1);
      if (channel.myname !== updatedBy) {
        threadHead.title = nameTail;
      }
    }
    else if (parentID.includes('@')) {
      const [namePre, nameTail] = parentID.split('@');
      if (channel.myname === namePre) {
        threadHead.title = nameTail;
      } else if (channel.myname === nameTail) {
        threadHead.title = namePre;
      } else {
        threadHead.title = editThreadHead.title
      }
    }
  }
  if (channel.myname === updatedBy) {
    threadHead.displayStatus = editThreadHead.displayStatus
  }
  console.log(threadHead)
  // if (editThreadHead.newThread) {
  //   const messagesStore = useMessagesStore()
  //   if (messagesStore.currentDisplay(editThreadHead.parentID)) {
  //     messagesStore.insert(editThreadHead)
  //   }
  // }
  delete editThreadHead.newThread
  console.log(threadHead)
  upsertIDB(threadHead, 'threadHead', 'parentID', threadHead.parentID);
  if (pre) {
    const logID = pd[1] + pd[2] + pd[3] + pd[0];
    const log = {
      logID: logID,
      pushID: pd[0],
      pushTitle: pd[1],
      channelID: pd[2],
      updatedBy: pd[3],
      updatedAt: timeFormat(),
      preContents: pre
    }
    upsertIDB(log, 'log', 'log', log.logID);
  }
}

