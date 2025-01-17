// if bad guy change userIDs and post?
// create userIDs from session
// update check from aliasName and session
// only write alias group aliasName
// when community join, decide alias
// aliasImg is changeble
// aliasName is not changeble

import { fetchChannel, fetchAliases, fetchGroups, userIDsByName } from '@/my/channelFunc';

export async function channelEdit(pushData) {
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
  const channelName = pushData[4][0];
  const channelDescription = pushData[4][1];
  let channel = await fetchChannel(channelID);
  if (channel) {
    channel.channelName = channelName;
    channel.myname = updatedBy;
    channel.channelDescription = channelDescription;    
  } else {
    channel = {
      channelID: channelID,
      myname: updatedBy,
      channelName: channelName,
      channelDescription: channelDescription
    }
  }
  const editLogs = channel.editLogs ?? [];
  const editLog = {
    updatedBy: updatedBy,
    updatedAt: timeFormat()
  }
  editLogs.push(editLog);
  channel.editLogs = editLogs;
  upsertIDB(channel, 'channel', 'channelID', channel.channelID)
    .catch((error) => {
      console.error(error);
    });
}

