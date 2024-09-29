import { getIDB, upsertData, deleteData } from '../my/indexDB.js';

export async function timestampReport(pushData) {
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
  const timestamps = pushData[4];

  // timestamps.forEach(value => {
  //   value.channelID = "hQKP";
  //   value.aliasName = "いけるか";
  // });
  // console.log()
  for (let d of timestamps) {
    d.channelID = channelID;
    d.aliasName = aliasName;
    d.stampStatus = 1;
    if (d.timestampID == null) {
      d.timestampID = `${d.timeIn}${d.aliasName}`;
    }
    upsertData(d, 'timestamp', 'timestampID', d.timestampID)
      .catch((error) => {
        console.error(error);
      });
  }
}

