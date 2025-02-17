import { userIDsByName } from '@/my/channelFunc.js';

export async function storeSelect(pd) {  

  // const channelID = pushData[2];
  // const aliasName = pushData[3];
  // const messageID = pushData[4][0];
  // const emojiValue = pushData[4][1];
  // const parentID = pushData[4][2];

  const pushID = pd[0];
  const channelID = pd[2];
  const aliasName = pd[3];
  const targetStore = pd[4][0];
  const param = pd[4][1];
  const channel = await getIDB('channel', channelID);
  const aliases = await getIDBs('alias', 'channelIDIndex', channelID, 10000);
  const userIDs = userIDsByName(aliases, [aliasName]);
  const calendar = await getAllIDBs('calendar');
  const startDate = new Date(param.date);
  const endDate = new Date(param.date);
  startDate.setDate(startDate.getDate() - 1);
  endDate.setDate(endDate.getDate() + 13);
  const calendarData = calendar
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


  // const fd = new FormData();
  // fd.append('userIDs', JSON.stringify(userIDsByName(aliases.value, [addName])));
  // fd.append('channelID', channelID);
  // fd.append('updatedBy', channel.value.myname);
  // const param = { date: today };
  // const contents = ['calendar', calendarData];
  // fd.append('contents', JSON.stringify(contents));
  // fd.append('pushTitle', 'storeSelect');
  // fd.append('csrf', localStorage.getItem("csrf"));
  // const res = await sendRequest('/ContentsJustPush/', fd);
  // res.csrf && localStorage.setItem('csrf', res.csrf);
  // res.pushContents.forEach(content => {
  //   pushReceive(content);
  // });

  const fd = new FormData();
  // fd.append('pushID', pushID);
  fd.append('userIDs', JSON.stringify(userIDs));
  fd.append('channelID', channelID);
  fd.append('updatedBy', channel.myname);
  const contents = ['calendar', 0, calendarData];
  fd.append('contents', JSON.stringify(contents));
  fd.append('pushTitle', 'storeShare');
  fd.append('csrf', localStorage.getItem("csrf"));
  const res = await sendRequest('/ContentsJustPush/', fd);
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
    updatedAt: timeFormat()
  }
  upsertIDB(log, 'log', 'log', log.logID);
}

