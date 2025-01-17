export async function alias(pushData) {
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
  const aliasImg = pushData[4][0];
  const userID = pushData[4][1];

  const alias = {
    aliasID: channelID + aliasName, // should not? be userID for unique per
    channelID: channelID,
    aliasName: aliasName,
    aliasImg: aliasImg,
    userID: userID,
  }
  console.log(alias);
  upsertIDB(alias, 'alias', 'aliasID', alias.aliasID)
    .catch((error) => {
      console.error(error);
    });

}
