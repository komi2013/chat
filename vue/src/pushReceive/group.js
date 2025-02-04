export async function group(pushData) {
  // const pushID = pushData[0];
  // const fd = new FormData();
  // fd.append('pushID', pushID);
  // const request = new Request('/PushResponse/', {
  //   method: 'POST',
  //   body: fd,
  // });
  // fetch(request);
  const channelID = pushData[2];
  const updatedBy = pushData[3];
  const groupName = pushData[4][0];
  const aliasNames = pushData[4][1];
  const groupImg = pushData[5];

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
    upsertIDB(group, 'group', 'group', group.groupID).catch((error) => {
      console.error(error);
    });
    action = 'update';
  } else {
    deleteIDB('group', 'groupID', group.groupID);
    action = 'delete';
  }
  const storeName = 'group';
  if (pre) {
    const logID = storeName + updatedBy + action + timestamp;
    const log = {
      logID: logID,
      storeName: storeName,
      updatedBy: updatedBy,
      action: action,
      updatedAt: timestamp,
      preContents: pre
    }
    upsertIDB(log, 'log', 'log', log.logID).catch((error) => {
      console.error(error);
    });
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

