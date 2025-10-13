import { advertisement } from './advertisement.js';
import { alias } from './alias.js';
import { answer } from './answer.js';
import { bookmark } from './bookmark.js';
import { bookPattern } from './bookPattern.js';
import { calendar } from './calendar.js';
import { channelEdit } from './channelEdit.js';
// import { channelJoin } from './channelJoin.js';
import { chunk } from './chunk.js';
import { emoji } from './emoji.js';
import { entryForm } from './entryForm.js';
import { group } from './group.js';
import { reception } from './reception.js';
import { receptionOrder } from './receptionOrder.js';
import { shiftStaffEdit } from './shiftStaffEdit.js';
import { storeSelect } from './storeSelect.js';
import { storeShare } from './storeShare.js';
import { thread } from './thread.js';
import { threadEdit } from './threadEdit.js';
import { threadHead } from './threadHead.js';
import { ticket } from './ticket.js';
import { timestamp } from './timestamp.js';
import { timestampCode } from './timestampCode.js';
import { timestampReport } from './timestampReport.js';
import { timestampRevert } from './timestampRevert.js';
import { topEdit } from './topEdit.js';

export async function pushReceive(notificationData, direct = false) {
  // console.log('before data', notificationData)
  // return
  // console.log('after data', notificationData)
  const pd = JSON.parse(notificationData)
  pd.directPush = direct
  const dupli = {
    pushDuplicationID: pd[1] + pd[2] + pd[3] + pd[0],
    pushID: pd[0],
    pushTitle: pd[1],
    channelID: pd[2],
    updatedBy: pd[3],
    updatedAt: timeFormat(),
    preContents: pd[4]
  }
  const pre = await getIDB('pushDuplication', dupli.pushDuplicationID)
  if (pre) {
    deleteIDB('pushDuplication', 'pushDuplicationID', pre.pushDuplicationID)
    return
  }
  upsertIDB(dupli, 'pushDuplication', 'pushDuplicationID', dupli.pushDuplicationID)
  const actions = {
    advertisement: advertisement,
    alias: alias,
    answer: answer,
    bookmark: bookmark,
    bookPattern: bookPattern,
    calendar: calendar,
    // channelAdd: channelAdd,
    channelEdit: channelEdit,
    // channelJoin: channelJoin,
    chunk: chunk,
    emoji: emoji,
    entryForm: entryForm,
    group: group,
    reception: reception,
    receptionOrder: receptionOrder,
    shiftStaffEdit: shiftStaffEdit,
    storeSelect: storeSelect,
    storeShare: storeShare,
    schedule: shiftStaffEdit,
    thread: thread,
    threadEdit: threadEdit,
    threadHead: threadHead,
    ticket: ticket,
    timestamp: timestamp,
    timestampCode: timestampCode,
    timestampReport: timestampReport,
    timestampRevert: timestampRevert,
    topEdit: topEdit,
  };

  const action = actions[pd[1]];
  
  if (action) {
    action(pd)  // 対応する関数を呼び出す
  } else {
    console.log('Unknown action.')
  }
}

