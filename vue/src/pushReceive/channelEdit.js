// if bad guy change userIDs and post?
// create userIDs from session
// update check from aliasName and session
// only write alias group aliasName
// when community join, decide alias
// aliasImg is changeble
// aliasName is not changeble

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
  const updatedBy = pushData[6];
  const aliasImg = pushData[5];
  try {
    const channel = await getIDB('channel', channelID);
    let allAliases = channel.allAliases;
    for (let i = 0; i < allAliases.length; i++) {
      if (updatedBy == allAliases[i][0]) {
        allAliases[i][1] = aliasImg;
      }
    }
    channel.allAliases = allAliases;
    channel.channelName = pushData[3];
    channel.channelDescription = pushData[4];
    channel.updatedBy = updatedBy;
    channel.updatedAt = pushData[7];
    console.log('channel', channel);
    // channel.groupAliases = pushData[8];
    // channel.displayStatus = 1
    upsertIDB(channel, 'channel', 'channelID', channelID)
      .catch((error) => {
        console.error(error);
      });
  } catch (error) {
    const channel = {
      channelID: channelID,
      channelName: pushData[3],
      channelDescription: pushData[4],
      updatedBy: updatedBy,
      updatedAt: pushData[7],
      displayStatus: 1,
      aliasName: updatedBy,
      allAliases: [[updatedBy, aliasImg, localStorage.userID]]
    };
    upsertIDB(channel, 'channel', 'channelID', channelID)
      .catch((error) => {
        console.error(error);
      });
  }
}

