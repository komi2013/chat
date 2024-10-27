import { getIDB, upsertIDB, getByMulti } from '../my/indexDB.js';
import { get_formated_time } from '../my/get_formated_time.js';

let channel;
let channelID;
let aliasName;
let code;
let action;
let now;
let timestampID;
let timestampCode;
let planDate;
let tts;
let tt;

export async function timestamp(pushData) {
  const pushID = pushData[0];
  const fd = new FormData();
  fd.append('pushID', pushID);
  const request = new Request('/PushResponse/', {
    method: 'POST',
    body: fd,
  });
  fetch(request);
  channelID = pushData[2];
  aliasName = pushData[3];
  code = pushData[4][0];
  action = pushData[4][1];
  now = pushData[4][2];
  name = pushData[4][3];
  // timestampID = pushData[4][4];
  // channel = getIDB('channel', channelID);
  await fetchChannel();
  await fetchTimestampCode();
  await fetchTimestamp();
  let date = new Date(planDate);
  date.setHours(date.getHours() + timestampCode.until);
  const until = get_formated_time('YYYY-MM-DDThh:mm', date.toISOString());
  if (planDate && (now < planDate || until < now)) {
    revertTimestamp();
  } else {
    stampTime();
  }
}

async function fetchChannel() {
  try {
    channel = await getIDB('channel', channelID);
  } catch (error) {
    console.log('error', error);
    channel = null;
  }
}

async function fetchTimestampCode() {
  try {
    timestampCode = await getIDB('timestampCode', code);
    planDate = timestampCode.planDate;
  } catch (error) {
    console.log('error', error);
    timestampCode = null;
  }
}

async function fetchTimestamp() {
  try {
    tts = await getIDBbyMulti('timestamp', ['channelID', 'aliasName'], 
      [channelID, aliasName], 30, 0, 'desc');
    const latestEntry = tts.reduce((max, obj) => 
      obj.timestampID > max.timestampID ? obj : max, tts[0]);
    // timestampID = latestEntry.timestampID;
    tt = latestEntry;
    console.log(tt);
  } catch (error) {
    console.log('error', error);
    tts = null;
  }
}

function revertTimestamp() {
  let userIDs = [];
  let names = [aliasName];
  for (const d of channel.allAliases) {
    if (names.includes(d[0])) {
      userIDs.push(d[2]);
    }
  }
  const fd = new FormData();
  fd.append('userIDs', JSON.stringify([...new Set(userIDs)]));
  fd.append('channelID', channelID);
  fd.append('aliasName', channel.aliasName);
  fd.append('contents', JSON.stringify([code, action, now, name]));
  fd.append('pushTitle', 'timestampRevert');
  const request = new Request('/ContentsPush/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      console.log(reason)
    })
}

function stampTime() {
    // let timestamp;
    switch (action) {
      case 'startWork':
        tt = {
          timestampID: now + aliasName,
          channelID: channelID,
          aliasName: aliasName,
          timeIn: now
          // breaks: pushData[6],
          // geos: pushData[6]
          // ips:
        };
        // tt.timeIn = now;
        break;
      case 'endWork':
        // timestamp = getIDB('timestamp', timestampID);
        tt.timeOut = now;
        break;
      case 'startBreak':
        // timestamp = getIDB('timestamp', timestampID);
        if (tt.breaks) {
          tt.breaks.push([now, null]);
        } else {
          tt.breaks = [ [now, null] ];
        }
        break;
      case 'endBreak':
        const lastLine = tt.breaks.length - 1;
        let lastBreak = tt.breaks[lastLine];
        lastBreak[1] = now;
        tt.breaks[lastLine] = lastBreak;
        break;
    }
    upsertIDB(tt, 'timestamp', 'timestampID', tt.timestampID)
      .catch((error) => {
        console.error(error);
      });

}