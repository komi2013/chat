// if bad guy change userIDs and post?
// create userIDs from session
// update check from aliasName and session
// only write alias group aliasName
// when community join, decide alias
// aliasImg is changeble
// aliasName is not changeble
// groupLockUntilDate prevent join group 
// because invitation URL is open for anybody
// changing group is possible by anybody who join

export async function channelEdit(pd) {
  const channelID = pd[2];
  const updatedBy = pd[3];
  const channelName = pd[4][0];
  const channelDescription = pd[4][1];
  const groupLockUntilDate = pd[4][2] ?? null;
  let pre = await getIDB('channel', channelID);
  let channel;
  if (pre) {
    channel = JSON.parse(JSON.stringify(pre));
    channel.channelName = channelName;
    channel.channelDescription = channelDescription;
    channel.groupLockUntilDate = groupLockUntilDate;
  } else {
    channel = {
      channelID: channelID,
      myname: updatedBy,
      channelName: channelName,
      channelDescription: channelDescription
    }
  }
  localStorage.setItem('channelID', channelID)
  localStorage.setItem('myname', myname)
  upsertIDB(channel, 'channel', 'channelID', channel.channelID);
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

