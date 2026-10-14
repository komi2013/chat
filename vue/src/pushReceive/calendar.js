import { useCalendarsStore } from '@/stores/calendars.js';

export async function calendar(pd) {
  const calendarsStore = useCalendarsStore();
  const channelID = pd[2];
  const aliasName = pd[3];
  const contents = pd[4];
  if (Array.isArray(contents) && contents[0] === 'calendar' && Array.isArray(contents[2])) {
    contents[2].forEach((event) => {
      if (!event?.read) return;
      event.title = aliasName;
      calendarsStore.upsert(event);
    });
    return;
  }

  const calendar = contents;
  const pre = await getIDB('calendar', calendar.calendarID);
  if (calendar.delete) {
    deleteIDB('calendar', 'calendarID', calendar.calendarID);
    calendarsStore.delete(calendar.calendarID);
  } else if (calendar.read) {
    calendar.title = calendar.todo ? Array.from(calendar.todo).slice(0, 10).join('') : '';
    calendarsStore.upsert(calendar);
  } else if (channelID == calendar.channelID && aliasName == calendar.aliasName) {
    upsertIDB(calendar, 'calendar', 'calendarID', calendar.calendarID);
    calendar.title = calendar.todo ? Array.from(calendar.todo).slice(0, 10).join('') : ''; 
    calendarsStore.upsert(calendar);
  }
  if (pre) {
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
    upsertIDB(log, 'log', 'log', log.logID);
  }
}

