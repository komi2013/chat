// import { fetchChannel, fetchAliases, fetchGroups, userIDsByName } from '@/my/channelFunc';

export async function threadHead(pushData) {
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
  let editThreadHead = pushData[4];
  // const secondPartID = pushData[4][0];
  // const title = pushData[4][1];
  // const description = pushData[4][2];
  // const aliasNames = pushData[4][3];
  // const adminNames = pushData[4][4];
  // const broadcastFlag = pushData[4][5];
  // let channel = await fetchChannel(channelID);
  let threadHead = await getIDB('threadHead', editThreadHead.parentID) || editThreadHead;
  // console.log('threadHead', threadHead);
  // threadHead.parentID = channelID + secondPartID;
  // threadHead.channelID = channelID;
  threadHead.title = editThreadHead.title;
  threadHead.description = editThreadHead.description;
  threadHead.aliasNames = editThreadHead.aliasNames;
  threadHead.adminNames = editThreadHead.adminNames;
  threadHead.broadcastFlag = editThreadHead.broadcastFlag;
  // const editLogs = channel.editLogs ?? [];
  // const editLog = {
  //   updatedBy: updatedBy,
  //   updatedAt: timeFormat()
  // }
  // editLogs.push(editLog);
  // channel.editLogs = editLogs;
  console.log('threadHead', threadHead);
  upsertIDB(threadHead, 'threadHead', 'parentID', threadHead.parentID);
}

