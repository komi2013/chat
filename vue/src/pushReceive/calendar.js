import { useCalendarsStore } from '../stores/calendars.js';

export async function calendar(pushData) {
  const calendarsStore = useCalendarsStore();

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
  const calendar = pushData[4];

  if (channelID == calendar.channelID && aliasName == calendar.aliasName) {
    upsertIDB(calendar, 'calendar', 'calendarID', calendar.calendarID)
      .catch((error) => {
        console.error(error);
      });
    calendarsStore.upsert(calendar);
  }

  // obj = {
  //   ca: generateRandomCode(8),
  //   title: title,
  //   channelID: channelID,
  //   aliasName: aliasName,
  //   contents: contents
  // };

  // for (let stampCode of stampCodes) {
  //   if (stampCode.isDel) {
  //     deleteIDB('calendar', 'calendarID', calendar.calendarID);
  //   }
  //   upsertIDB(calendar, 'calendar', 'calendarID', calendar.calendarID)
  //     .catch((error) => {
  //       console.error(error);
  //     });
  // }
}

