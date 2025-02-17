import { useCalendarsStore } from '@/stores/calendars.js';
import { useNoticesStore } from '@/stores/notices.js';

let calendarNameCount = 1;
export async function storeShare(pd) {
  // const pushID = pd[0];
  const channelID = pd[2];
  const aliasName = pd[3];
  const targetStore = pd[4][0];
  const save = pd[4][1];
  const contents = pd[4][2];
  if (targetStore == 'calendar') {
    const calendarsStore = useCalendarsStore();
    contents.forEach((d) => {
      d.nameCount = calendarNameCount;
      d.title = aliasName;
      calendarsStore.upsert(d);
    });
    const noticesStore = useNoticesStore();
    noticesStore.setNotice(aliasName + 'のデータを取得');
    calendarNameCount = calendarNameCount + 1;
  }
  if (save == 1) {
    contents.forEach((d) => {
      const dataCopy = { ...d };
      delete dataCopy.ID;
      upsertIDB(dataCopy, targetStore, targetStore, d.ID);
    });
  }
}

