import { useMessagesStore } from '../stores/messages.js';

export async function rookie(pushData) {
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
  const aliasImg = pushData[4];
  const userID = pushData[5];

  const alias = {
    aliasID: channelID + aliasName, // should be userID for unique per
    channelID: channelID,
    aliasName: aliasName,
    aliasImg: aliasImg,
    userID: userID
  }
  upsertIDB(alias, 'alias', 'aliasID', alias.aliasID)
    .catch((error) => {
      console.error(error);
    });

  // let channel;
  // async function fetchChannel() {
  //   try {
  //     channel =　await getIDB('channel', obj.channelID);
  //     console.log('channel', channel);
  //     addAliases(channel, obj);
  //   } catch (error) {
  //     channel = null;
  //     console.log('channel not found', error);
  //   }
  // }
  // fetchChannel();
  // console.log('channel after wait', channel);
}

// function addAliases (channel, alias) {
//   if (channel.allAliases && Array.isArray(channel.allAliases)) {
//     channel.allAliases.push([alias.aliasName, alias.aliasImg, alias.userID]);
//   } else {
//     channel.allAliases = [[alias.aliasName, alias.aliasImg, alias.userID]];
//   }
//   upsertIDB(channel, 'channel', 'channelID', channel.channelID)
//     .catch((error) => {
//       console.error(error);
//     });
// }

