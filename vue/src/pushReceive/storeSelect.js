import { userIDsByName } from '@/my/channelFunc.js';

export async function storeSelect(pd) {  
  const pushID = pd[0];
  const channelID = pd[2];
  const aliasName = pd[3];
  const targetStore = pd[4];
  const param = pd[5];
  const channel = await getIDB('channel', channelID);
  const aliases = await getIDBs('alias', 'channelIDIndex', channelID, 10000);
  const userIDs = userIDsByName(aliases, [aliasName]);
  const calendar = await getAllIDBs('calendar');
  const startDate = new Date(param.date);
  const endDate = new Date(param.date);
  startDate.setDate(startDate.getDate() - 1);
  endDate.setDate(endDate.getDate() + 13);
  const contents = calendar
    .filter(event => {
      const eventDate = new Date(event.timeStart);
      return eventDate >= startDate && eventDate <= endDate;
    })
    .map(event => {
      return {
        calendarID: event.calendarID,
        timeStart: event.timeStart,
        timeEnd: event.timeEnd
      };
    });
  const fd = new FormData();
  fd.append('pushID', pushID);
  fd.append('userIDs', JSON.stringify(userIDs));
  fd.append('channelID', channelID);
  fd.append('aliasName', channel.myname);
  fd.append('contents', JSON.stringify(contents));
  fd.append('targetStore', targetStore);
  fd.append('csrf', localStorage.getItem("csrf"));
  const res = await sendRequest('/StorePush/', fd);
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents && res.pushContents.forEach(content => {
    pushReceive(content);
  });
  const logID = pd[1] + pd[2] + pd[3] + pd[4] + pd[0];
  const log = {
    logID: logID,
    pushID: pd[0],
    pushTitle: pd[1],
    channelID: pd[2],
    updatedBy: pd[3],
    updatedAt: timeFormat(),
    preContents: pre
  }
  upsertIDB(log, 'log', 'log', log.logID);
}

