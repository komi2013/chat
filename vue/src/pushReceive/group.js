export async function group(pd) {
  // const pushID = pushData[0];
  // const fd = new FormData();
  // fd.append('pushID', pushID);
  // const request = new Request('/PushResponse/', {
  //   method: 'POST',
  //   body: fd,
  // });
  // fetch(request);
  const channelID = pd[2];
  const updatedBy = pd[3];
  const groupName = pd[4][0];
  const aliasNames = pd[4][1];
  const groupImg = pd[5];

  const timestamp = timeFormat();
  // const editLogKey = `editLog${timestamp}`;
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



  // let editLog = {
  //   updatedBy: aliasName,
  //   updatedAt: timeFormat()
  // }
  // let group;
  // const pre = await getIDB('group', group.groupID);
  // if (pre) {
  //   group = pre;
  //   group.groupImg = groupImg;
  //   group.aliasNames = aliasNames;
  //   editLog.pre = pre;
  //   group.editLogs.push(editLog);
  // } else {
  //   group = {
  //     groupID: channelID + groupName,
  //     channelID: channelID,
  //     groupName: groupName,
  //     groupImg: groupImg,
  //     aliasNames: aliasNames,
  //     editLogs: [editLog]
  //   }
  // }
  // upsertIDB(group, 'group', 'group', group.groupID)
  //   .catch((error) => {
  //     console.error(error);
  //   });

  // const contents = [
  //   group.groupName, group.aliasNames
  // ]

  // try {
  //   const channel = await getIDB('channel', channelID);
  //   channel.groupAliases = pushData[3];
  //   channel.updatedBy = pushData[4];
  //   // channel.updatedAt = pushData[6];
  //   // channel.displayStatus = 1
  //   upsertIDB(channel, 'channel', 'channelID', channelID)
  //     .catch((error) => {
  //       console.error(error);
  //     });
  // } catch (error) {
  //   console.log('channel not', error);
  // }
}

