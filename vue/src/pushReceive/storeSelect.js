import { userIDsByName } from '../my/channelFunc.js';

export async function storeSelect(pushData) {
  const pushSelectID = pushData[0];
  const channelID = pushData[2];
  const aliasName = pushData[3];
  const targetStore = pushData[4];
  const param = pushData[5];
  const channel = await getIDB('channel', channelID);
  const userIDs = userIDsByName(channel, [aliasName]);
  const calendar = await getAllIDBs('calendar');
  const startDate = new Date(param.date);
  const endDate = new Date(param.date);
  startDate.setDate(startDate.getDate() - 7);
  endDate.setDate(endDate.getDate() + 7);
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
  fd.append('pushSelectID', pushSelectID);
  fd.append('userIDs', JSON.stringify(userIDs));
  fd.append('channelID', channelID);
  fd.append('aliasName', channel.aliasName);
  fd.append('contents', JSON.stringify(contents));
  fd.append('targetStore', targetStore);
  const request = new Request('/StorePush/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      console.error(reason);
    })
}

