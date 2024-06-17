import { useMessagesStore } from '../stores/messages.js';
import { getIDB, upsertData, deleteData, updOne } from '../my/indexDB.js';

export async function channelJoin(pushData) {
  const pushID = pushData[0];
  const channelID = pushData[2];
  const obj = {
    aliasName: pushData[3],
    aliasImg: pushData[4],
    userID: pushData[5]
  };
  let alias;
  async function fetchAlias() {
    try {
      alias = getIDB('alias', obj.aliasName);
      console.log('alias', alias);
      alias.channelIDs.push(channelID);
    } catch (error) {
      console.log('aliass not found.', error);
      alias = obj;
      alias.channelIDs = [channelID];
    }
  }
  await fetchAlias();
  console.log(alias);
  upsertData(alias, 'alias', 'aliasName', alias.aliasName)
    .catch((error) => {
      console.error(error);
    });

  const fd = new FormData()
  fd.append('pushID', pushID);
  const request = new Request('/PushGet/', {
      method: 'POST',
      body: fd,
  });
  fetch(request)
  .then((response)=>{
    if(!response.ok){
      throw new Error();
    }
    return response.json()
  })
  .then((channel)=>{
    channel.aliasName = obj.aliasName;
    const aliases = [
      alias.aliasName, alias.aliasImg, alias.userID
    ];
    channel.allAliases.push(aliases);
    upsertData(channel, 'channel', 'channelID', channel.channelID)
      .catch((error) => {
        console.error(error);
      });
  })
  .catch((reason)=>{
    console.log(reason)
  });
}
