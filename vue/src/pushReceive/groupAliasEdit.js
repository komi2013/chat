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
  try {
    const channel = await getIDB('channel', channelID);
    channel.groupAliases = pushData[3];
    channel.updatedBy = pushData[4];
    // channel.updatedAt = pushData[6];
    // channel.displayStatus = 1
    upsertIDB(channel, 'channel', 'channelID', channelID)
      .catch((error) => {
        console.error(error);
      });
  } catch (error) {
    console.log('channel not', error);
  }
}

