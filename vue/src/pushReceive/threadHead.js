// import { fetchChannel, fetchAliases, fetchGroups, userIDsByName } from '@/my/channelFunc';
// import { useMessagesStore } from '@/stores/messages.js';

export async function threadHead(pd) {
  const channelID = pd[2]
  const updatedBy = pd[3]
  let threadHead = pd[4]
  const pre = await getIDB('threadHead', threadHead.parentID)
  const channel = await getIDB('channel', channelID)
  const parentID = threadHead.parentID
  if (parentID && parentID.startsWith('@')) {
    const nameTail = parentID.slice(1);
    if (channel.myname !== updatedBy) {
      threadHead.title = nameTail;
    }
  } else if (parentID && parentID.includes('@')) {
    const [namePre, nameTail] = parentID.split('@');
    if (channel.myname === namePre) {
      threadHead.title = nameTail
    } else if (channel.myname === nameTail) {
      threadHead.title = namePre
    }
  }
  if (channel.myname === updatedBy) {
    threadHead.displayStatus = 1
  }
  delete threadHead.newThread
  console.log('editThreadHead', threadHead)
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

