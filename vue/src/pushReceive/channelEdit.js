import { useMessagesStore } from '../stores/messages.js';
import { getIDB, upsertData, deleteData } from '../my/indexDB.js';

// if bad guy change userIDs and post?
// create userIDs from session
// update check from aliasName and session
// only write alias group aliasName
// when community join, decide alias
// aliasImg is changeble

export async function channelEdit(pushData) {
  const pushID = pushData[0];
  const fd = new FormData();
  fd.append('pushID', pushID);
  const request = new Request('/PushResponse/', {
    method: 'POST',
    body: fd,
  });
  fetch(request);

  // arr = append(arr, r.FormValue("channelID")) // 2
  // arr = append(arr, r.FormValue("channelName"))
  // arr = append(arr, r.FormValue("description"))
  // arr = append(arr, aliases)
  // arr = append(arr, r.FormValue("aliasName"))
  // arr = append(arr, time.Now().Format("2006-01-02"))
  const channelID = pushData[2];
  const updatedBy = pushData[6];
  // const obj = {
  //   channelID: pushData[2],
  //   channelName: pushData[3],
  //   channelDescription: pushData[4],
  //   updatedBy: pushData[6],
  //   updatedAt: pushData[7]
  // };
  const aliases = pushData[5];
  try {
    const channel = await getIDB('channel', channelID);
    console.log('channel', channel);
    let allAliases = channel.allAliases;
    for (let i = 0; i < allAliases.length; i++) {
      if (updatedBy == aliases[i][0]) {
        allAliases[i][1] = aliases[i][1];
      }
    }
    channel.allAliases = allAliases;
    channel.channelName = pushData[3];
    channel.channelDescription = pushData[4];
    channel.updatedBy = updatedBy;
    channel.updatedAt = pushData[7];
    channel.groupAliases = pushData[8];
    channel.displayStatus = 1
    upsertData(channel, 'channel', 'channelID', channelID)
      .catch((error) => {
        console.error(error);
      });


  } catch (error) {
    console.log('channel not', error);
  }
}

