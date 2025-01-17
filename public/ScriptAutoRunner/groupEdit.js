setTimeout(() => {
  
  // const chunks = getIDBs('chunk', 'chunkPassIndex', 'Az', 1000);
  // console.log(chunks);

  // const channelID = 'hQKP';
  // const aliasName = 'mik1';
  // const groupName = 'group1';
  // const groupImg = '/me.jpg';
  // const aliasNames = ['ivan13'];
  // const editLog = {
  //   updatedBy: aliasName,
  //   updatedAt: timeFormat()
  // }
  // const group = {
  //   groupID: channelID + groupName,
  //   channelID: channelID,
  //   groupName: groupName,
  //   groupImg: groupImg,
  //   aliasNames: aliasNames,
  //   editLogs: [editLog]
  // }
  // console.log(group);
  // upsertIDB(group, 'group', 'group', group.groupID)
  //   .catch((error) => {
  //     console.error(error);
  //   });

  const aliasName = 'sei1';
  const alias = {
    aliasID: "hQKP" + aliasName,
    channelID: "hQKP",
    aliasName: aliasName,
    aliasImg: "/me.jpg",
    userID: "seijiro"
  }
  upsertIDB(alias, 'alias', 'aliasID', alias.aliasID)
    .catch((error) => {
      console.error(error);
    });


}, "1000");
