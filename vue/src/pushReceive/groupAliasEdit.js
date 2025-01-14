export async function groupAliasEdit(pushData) {
  const pushID = pushData[0];
  const fd = new FormData();
  fd.append('pushID', pushID);
  const request = new Request('/PushResponse/', {
    method: 'POST',
    body: fd,
  });
  fetch(request);
  const channelID = pushData[2];

  const aliasName = pushData[3];
  const data = pushData[4];
  const groupName = pushData[4][0];
  const groupImg = pushData[4][1];
  const aliasNames = pushData[4][2];

  const editLog = {
    updatedBy: aliasName,
    updatedAt: timeFormat()
  }
  const group = {
    groupID: channelID + groupName,
    channelID: channelID,
    groupName: groupName,
    groupImg: groupImg,
    aliasNames: aliasNames,
    editLogs: [editLog]
  }
  upsertIDB(group, 'group', 'group', group.groupID)
    .catch((error) => {
      console.error(error);
    });

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

