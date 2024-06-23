import { useMessagesStore } from '../stores/messages.js';
import { getIDB, upsertData, deleteData } from '../my/indexDB.js';

// if bad guy change userIDs and post?
// create userIDs from session
// update check from aliasName and session
// only write alias group aliasName
// when community join, decide alias
// aliasImg is changeble

export async function channelUpsert(pushData) {
  const pushID = pushData[0];
  const fd = new FormData();
  fd.append('pushID', pushID);
  const request = new Request('/PushResponse/', {
    method: 'POST',
    body: fd,
  });
  fetch(request);
  const obj = {
    channelID: pushData[2],
    channelName: pushData[3],
    channelDescription: pushData[4],
    updatedAt: pushData[7],
    updatedBy: pushData[6],
    displayStatus: 1
  };
  console.log(pushData[8]);
  if (pushData[8] == 1) { // 1 = add 
    obj.aliasName = pushData[6];
    obj.ownerName = pushData[6];
    addAliasIfNotExists(pushData[5], obj.channelID);
    obj.allAliases = [pushData[5]];
  } else {
    obj.allAliases = pushData[5];
  }
  upsertData(obj, 'channel', 'channelID', obj.channelID)
    .catch((error) => {
      console.error(error);
    });
}

function addAliasIfNotExists(newAlias, channelID) {
  const obj = {
    aliasName: newAlias[0],
    aliasImg: newAlias[1]
  };
  let alias;
  async function fetchAlias() {
    try {
      alias = await getIDB('alias', obj.aliasName);
      console.log('alias', alias);
    } catch (error) {
      alias = null;
      console.log('aliass not');
    }
  }
  fetchAlias();
  if (alias && alias.channelIDs) {
    alias.channelIDs.push(channelID);
  } else {
    alias = obj;
    alias.channelIDs = [channelID];
  }
  upsertData(alias, 'alias', 'aliasName', obj.aliasName)
    .catch((error) => {
      console.error(error);
    });
}

