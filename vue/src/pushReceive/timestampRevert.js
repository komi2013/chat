export async function timestampRevert(pushData) {
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
  const code = pushData[4][0];
  const action = pushData[4][1];
  const now = pushData[4][2];
  const name = pushData[4][3];
  const timestampID = now + aliasName;
  switch (action) {
    case 'startWork':
      deleteIDB('timestamp', 'timestampID', timestampID);
      break;
    case 'endWork':
      timestamp = getIDB('timestamp', timestampID);
      delete timestamp.timeOut;
      upsertIDB(timestamp, 'timestamp', 'timestampID', timestampID)
        .catch((error) => {
          console.error(error);
        });
    case 'deletePrevious':
      const timeIn = pushData[4][0];
      const timestamps = await getIDBbyMulti('timestamp', ['channelID', 'aliasName'], 
        [channelID, aliasName], 60, 0, 'desc');
      timestamps.forEach(timestamp => {
        if (timestamp.timeIn <= timeIn) {
          // deleteData('timestamp', 'timestampID', timestamp.timestampID);
          if (timestamp.stampStatus) {
            timestamp.stampStatus += 20;
          } else {
            timestamp.stampStatus = 20;
          }
          upsertIDB(timestamp, 'timestamp', 'timestampID', timestamp.timestampID)
            .catch((error) => {
              console.error(error);
            });
        }
      });
      break;
  }
}

