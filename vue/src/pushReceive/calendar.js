import { useCalendarsStore } from '@/stores/calendars.js';

export async function calendar(pd) {
  const calendarsStore = useCalendarsStore();
  const channelID = pd[2];
  const aliasName = pd[3];
  const calendar = pd[4];
  const pre = await getIDB('calendar', calendar.calendarID);
  if (calendar.delete) {
    deleteIDB('calendar', 'calendarID', calendar.calendarID);
    calendarsStore.delete(calendar.calendarID);
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

