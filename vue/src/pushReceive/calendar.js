import { useCalendarsStore } from '../stores/calendars.js';

export async function calendar(pushData) {
  const calendarsStore = useCalendarsStore();

  // const pushID = pushData[0];
  // const fd = new FormData();
  // fd.append('pushID', pushID);
  // const request = new Request('/PushResponse/', {
  //   method: 'POST',
  //   body: fd,
  // });
  // fetch(request);
  const channelID = pushData[2];
  const aliasName = pushData[3];
  const calendar = pushData[4];
  if (calendar.delete) {
    deleteIDB('calendar', 'calendarID', calendar.calendarID);
    calendarsStore.delete(calendar.calendarID);
  } else if (channelID == calendar.channelID && aliasName == calendar.aliasName) {
    upsertIDB(calendar, 'calendar', 'calendarID', calendar.calendarID);
    calendar.title = calendar.todo ? Array.from(calendar.todo).slice(0, 10).join('') : ''; 
    calendarsStore.upsert(calendar);
  }
}

