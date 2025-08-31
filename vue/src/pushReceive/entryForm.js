export async function entryForm(pd) {
  const channelID = pd[2]
  const updatedBy = pd[3]
  let editEntryForm = pd[4]
  const pre = await getIDB('entryForm', editEntryForm.entryFormID) || editEntryForm
  // let entryForm;
  // if (pre) {
  //   entryForm = JSON.parse(JSON.stringify(pre))
  // }
  upsertIDB(editEntryForm, 'entryForm', 'entryFormID', editEntryForm.entryFormID)
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

