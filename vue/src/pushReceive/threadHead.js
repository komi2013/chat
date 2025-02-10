// import { fetchChannel, fetchAliases, fetchGroups, userIDsByName } from '@/my/channelFunc';

export async function threadHead(pd) {
  const channelID = pd[2];
  const updatedBy = pd[3];
  let editThreadHead = pd[4];
  const pre = await getIDB('threadHead', editThreadHead.parentID) || editThreadHead;
  let threadHead;
  if (pre) {
    threadHead = JSON.parse(JSON.stringify(pre));
  }
  threadHead.title = editThreadHead.title;
  threadHead.description = editThreadHead.description;
  threadHead.aliasNames = editThreadHead.aliasNames;
  threadHead.adminNames = editThreadHead.adminNames;
  threadHead.broadcastFlag = editThreadHead.broadcastFlag;
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

