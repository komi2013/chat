export async function alias(pd) {
  const channelID = pd[2];
  const updatedBy = pd[3];
  const userID = pd[4][0];
  const aliasName = pd[4][1]
  const bio = pd[4][2]
  const accessRight = pd[4][3]
  // const del = pd[4][4] ?? false
  const aliasImg = pd[5]
  const numericAccessRight = ({
    admin: 1,
    guest: 2,
    inquirer: 3,
    inquiry: 3,
    delete: -1,
  })[accessRight] ?? Number(accessRight)
  const atMark = numericAccessRight === 3 ? '@' : '' // @customer nickname
  const alias = {
    // aliasID: channelID + userID,
    aliasID: channelID + atMark + aliasName,
    channelID: channelID,
    aliasName: atMark + aliasName,
    aliasImg: aliasImg,
    userID: userID,
    aliasBio: bio,
    accessRight: numericAccessRight,
  }
  // console.log('alias', alias)
  const pre = await getIDB('alias', alias.aliasID)
  if (numericAccessRight === -1) {
    deleteIDB('alias', 'aliasID', alias.aliasID);
  } else {
    upsertIDB(alias, 'alias', 'aliasID', alias.aliasID);
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
