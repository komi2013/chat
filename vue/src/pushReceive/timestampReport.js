export async function timestampReport(pushData) {
  const pushID = pushData[0];
  const channelID = pushData[2];
  const aliasName = pushData[3];
  console.log('data', pushData[4]);
  if (pushData[4][0] === 1) { // approve
    const timestamps = pushData[4][1];
    timestamps.forEach(record => {
      if (!record.approveds) {
          record.approveds = [aliasName];
      } else if (!record.approveds.includes(aliasName)) {
          record.approveds.push(aliasName);
      }
      upsertIDB(record, 'timestamp', 'timestampID', record.timestampID);
    });
  } else if (pushData[4][0] === 2) {
    const timestamps = await getIDBbyMulti('timestamp', ['channelID', 'aliasName'], 
      [channelID, pushData[4][1]], 60, 0, 'asc');
    timestamps.forEach(timestamp => {
      if (timestamp.approveds && timestamp.approveds.length > 0) {
        deleteIDB('timestamp', 'timestampID', timestamp.timestampID);
      }
    });
  } else if (typeof pushData[4][0] === "string") { // submit stampStatus 2
    const timestamps = await getIDBbyMulti('timestamp', ['channelID', 'aliasName'], 
      [channelID, pushData[4][0]], 60, 0, 'desc');
    const minTimestampID = pushData[4][1];
    const maxTimestampID = pushData[4][2];
    const updatedTimestamps = timestamps.map(entry => {
      if (entry.timestampID && entry.timestampID >= minTimestampID && entry.timestampID <= maxTimestampID) {
        const { approveds, ...rest } = entry;
        return { ...rest, stampStatus: 2 };
      }
      return entry;
    });
    for (let d of updatedTimestamps) {
      upsertIDB(d, 'timestamp', 'timestampID', d.timestampID);
    }
  } else { // manual edit stampStatus 1
    const timestamps = pushData[4];
    for (let d of timestamps) {
      d.channelID = channelID;
      d.aliasName = aliasName;
      d.stampStatus = 1;
      console.log('d', d);
      if (d.timestampID == null) {
        d.timestampID = `${d.timeIn}${d.aliasName}`;
      }
      if (d.approveds) {
        delete d.approveds;
      }
      upsertIDB(d, 'timestamp', 'timestampID', d.timestampID);
    }
  }
}

