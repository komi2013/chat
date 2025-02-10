export async function group(pd) {
  const channelID = pd[2];
  const updatedBy = pd[3];
  const groupName = pd[4][0];
  const aliasNames = pd[4][1];
  const groupImg = pd[5];
  const timestamp = timeFormat();
  const pre = await getIDB('group', channelID + groupName);
  let group;
  if (pre) {
    group = JSON.parse(JSON.stringify(pre));
    group.aliasNames = aliasNames;
    group.groupImg = groupImg;
  } else {
    group = {
      groupID: channelID + groupName,
      channelID: channelID,
      groupName: groupName,
      groupImg: groupImg,
      aliasNames: aliasNames,
      createdBy: updatedBy,
      createdAt: timestamp
    };
  }
  let action;
  if (aliasNames) {
    upsertIDB(group, 'group', 'group', group.groupID);
    action = 'update';
  } else {
    deleteIDB('group', 'groupID', group.groupID);
    action = 'delete';
  }
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

