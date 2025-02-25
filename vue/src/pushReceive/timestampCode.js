export async function timestampCode(pushData) {
  const pushID = pushData[0];
  const fd = new FormData();
  fd.append('pushID', pushID);
  const request = new Request('/PushResponse/', {
    method: 'POST',
    body: fd,
  });
  fetch(request);
  const channelID = pushData[2];
  const stampCodes = pushData[4];
  for (let stampCode of stampCodes) {
    if (stampCode.isDel) {
      deleteIDB('timestampCode', 'code', stampCode.code);
    }
    upsertIDB(stampCode, 'timestampCode', 'code', stampCode.code);
  }
}

