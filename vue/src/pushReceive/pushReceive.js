import { bookmark } from './bookmark.js';
import { bookPattern } from './bookPattern.js';
import { calendar } from './calendar.js';
import { channelAdd } from './channelAdd.js';
import { channelEdit } from './channelEdit.js';
import { channelJoin } from './channelJoin.js';
import { chunk } from './chunk.js';
import { emoji } from './emoji.js';
import { groupAliasEdit } from './groupAliasEdit.js';
import { receptionOrder } from './receptionOrder.js';
import { rookie } from './rookie.js';
import { shiftStaffEdit } from './shiftStaffEdit.js';
import { storeSelect } from './storeSelect.js';
import { thread } from './thread.js';
import { threadEdit } from './threadEdit.js';
import { ticket } from './ticket.js';
import { timestamp } from './timestamp.js';
import { timestampCode } from './timestampCode.js';
import { timestampReport } from './timestampReport.js';
import { timestampRevert } from './timestampRevert.js';

export function pushReceive(notificationData) {
  const data = JSON.parse(notificationData);

  const actions = {
    bookmark: bookmark,
    bookPattern: bookPattern,
    calendar: calendar,
    channelAdd: channelAdd,
    channelEdit: channelEdit,
    channelJoin: channelJoin,
    chunk: chunk,
    emoji: emoji,
    groupAliasEdit: groupAliasEdit,
    receptionOrder: receptionOrder,
    rookie: rookie,
    shiftStaffEdit: shiftStaffEdit,
    storeSelect: storeSelect,
    schedule: shiftStaffEdit,
    thread: thread,
    threadEdit: threadEdit,
    ticket: ticket,
    timestamp: timestamp,
    timestampCode: timestampCode,
    timestampReport: timestampReport,
    timestampRevert: timestampRevert,
  };

  const action = actions[data[1]];
  
  if (action) {
    action(data);  // 対応する関数を呼び出す
  } else {
    console.log('Unknown action.');
  }
}

