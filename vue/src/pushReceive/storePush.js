import { userIDsByName } from '../my/channelFunc.js';
import { useCalendarsStore } from '../stores/calendars.js';
import { useNoticesStore } from '../stores/notices.js';

let nameCount = 1;
export async function storePush(pushData) {
  const pushSelectID = pushData[0];
  const channelID = pushData[2];
  const aliasName = pushData[3];
  const contents = pushData[4];
  const targetStore = pushData[5];
  const calendarsStore = useCalendarsStore();
  contents.forEach((d) => {
    d.nameCount = nameCount;
    calendarsStore.upsert(d);
  });
  const noticesStore = useNoticesStore();
  noticesStore.setNotice(aliasName + 'のデータを取得');
  nameCount = nameCount + 1;
}

