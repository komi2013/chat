export async function reception(pd) {
  // const calendarsStore = useCalendarsStore();
  const channelID = pd[2];
  const aliasName = pd[3];
  const rData = pd[4];
  console.log('rData', rData);
  const pre = await getIDB('reception', rData.receptionID);
  if (rData.delete) {
    deleteIDB('reception', 'receptionID', rData.receptionID);
    // calendarsStore.delete(calendar.calendarID);
  } else {
    upsertIDB(rData, 'reception', 'receptionID', rData.receptionID);
    // calendar.title = calendar.todo ? Array.from(calendar.todo).slice(0, 10).join('') : ''; 
    // calendarsStore.upsert(calendar);
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

