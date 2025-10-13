import { useNoticesStore } from '@/stores/notices.js';
import { pushReceive } from '@/pushReceive/pushReceive.js';

let channel;
let aliases;
let channelID;
let aliasName;
let code;
let action;
let now;
let timestampID;
let timestampCode;
let planDate;
let timestamps;
let timestampRecord;
let until;

export async function timestamp(pd) {
  const pushID = pd[0];
  channelID = pd[2];
  // updatedBy = pd[3];
  code = pd[4][0];
  action = pd[4][1];
  now = pd[4][2];
  aliasName = pd[4][3];
  channel = await getIDB('channel', channelID);
  aliases = await getIDBs('alias', 'channelIDIndex', localStorage.getItem('channelID'), 10000);
  timestampCode = await getIDB('timestampCode', code);
  if (timestampCode) {
    planDate = timestampCode.planDate;
    let date = new Date(planDate);
    date.setHours(date.getHours() + timestampCode.until);
    until = timeFormat('YYYY-MM-DDThh:mm', date.toISOString());
  }
  timestamps = await getIDBbyMulti('timestamp', ['channelID', 'aliasName'], 
    [channelID, aliasName], 30, 0, 'desc');
  const latestEntry = timestamps.reduce((max, obj) => 
    obj.timestampID > max.timestampID ? obj : max, timestamps[0]);
  const pre = latestEntry;
  if (pre) timestampRecord = JSON.parse(JSON.stringify(pre));
  if (!planDate || (now >= planDate && now <= until)) {
    stampTime();
  } else {
    revertTimestamp();
  }
  const logID = pd[1] + pd[2] + pd[3] + pd[0];
  const log = {
    logID: logID,
    pushID: pd[0],
    pushTitle: pd[1],
    channelID: pd[2],
    updatedBy: pd[3],
    updatedAt: timeFormat(),
    preContents: pre
  }
  if (pre) log.preContents = pre;
  upsertIDB(log, 'log', 'log', log.logID);
}

async function stampTime() {
    switch (action) {
      case 'startWork':
        timestampRecord = {
          timestampID: now + aliasName,
          channelID: channelID,
          aliasName: aliasName,
          timeIn: now
        };
        break;
      case 'endWork':
        timestampRecord.timeOut = now;
        break;
      case 'startBreak':
        if (timestampRecord.breaks) {
          timestampRecord.breaks.push({start: now});
        } else {
          timestampRecord.breaks = [{start: now}];
        }
        break;
      case 'endBreak':
        const lastLine = timestampRecord.breaks.length - 1;
        let lastBreak = timestampRecord.breaks[lastLine];
        lastBreak.end = now;
        timestampRecord.breaks[lastLine] = lastBreak;
        break;
      case 'del_startWork':
        console.log(now + aliasName, aliasName);
        const res = await deleteIDB('timestamp', 'timestampID', now + aliasName);
        console.log(res);
        break;
      case 'del_endWork':
        delete timestampRecord.timeOut;
        break;
      case 'del_startBreak':
        if (timestampRecord.breaks && timestampRecord.breaks.length > 0) {
          timestampRecord.breaks.pop();
        }
        break;
      case 'del_endBreak':
        if (timestampRecord.breaks && timestampRecord.breaks.length > 0) {
          const lastIndex = timestampRecord.breaks.length - 1;
          delete timestampRecord.breaks[lastIndex].end;
        }
        break;
    }
    if (code == 'revert' && aliasName === channel.myname) {
      const noticesStore = useNoticesStore();
      noticesStore.setNotice('打刻できませんでした');
    }
    if (action != 'del_startWork') {
      await upsertIDB(timestampRecord, 'timestamp', 'timestampID', timestampRecord.timestampID);
    }
}

async function revertTimestamp() {
  const fd = new FormData();
  fd.append('pushNames', JSON.stringify([aliasName]))
  fd.append('channelID', channelID);
  fd.append('updatedBy', channel.myname);
  fd.append('contents', JSON.stringify(['revert', 'del_' + action, now, aliasName]));
  fd.append('pushTitle', 'timestamp');
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ContentsJustPush/', fd);
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
}
