export function takeUserIDs(channel, groupName = null) {
  let userIDs = [];
  let groupNames = [channel.aliasName];
  if (groupName) {
    groupNames.push(groupName);
  }
  if (Array.isArray(channel.groupAliases)) {
    for (const d of channel.groupAliases) {
      if (groupName) {
        if (groupName == d[0]) {
          for (const d2 of d[2]) {
            groupNames.push(d2);
          }
        }
      }
    }
  }
  for (const d of channel.allAliases) {
    if (groupNames.includes(d[0])) {
      userIDs.push(d[2]);
    }
  }
  return [...new Set(userIDs)];
}

// export async function fetchChannel(channelId) {
//   const result = {
//     channel: null,
//     groupAliases: [],
//     groupAliasNames: [],
//   };

//   try {
//     const data = await getIDB('channel', channelId);
//     result.channel = data;
//     result.groupAliases = data.groupAliases || [];
//     result.groupAliasNames = result.groupAliases.map(alias => alias[0]);
//   } catch (error) {
//     console.error('fetchChannelData:', error);
//   }

//   return result;
// }

export async function fetchChannel(channelId) {
  try {
    return await getIDB('channel', channelId);
  } catch (error) {
    console.error('fetchChannel:', error);
    return null;
  }
}

export async function fetchAliases(channelId) {
  try {
    return await getIDBs('alias', 'channelIDIndex', channelId);
  } catch (error) {
    console.error('fetchAliases:', error);
    return [];
  }
}

export async function fetchGroups(channelId) {
  try {
    return await getIDBs('group', 'channelIDIndex', channelId);
  } catch (error) {
    console.error('fetchGroups:', error);
    return [];
  }
}

// export function userIDsAll(channel, aliases) {
//   let userIDs = [];
//   for (const d of aliases) {
//     if (names.includes(d[0])) {
//       userIDs.push(d[2]);
//     }
//   }
//   return [...new Set(userIDs)];
// }

export function userIDsByName(channel, aliases, names = null) {
  let userIDs = [];
  for (const d of aliases) {
    if (names.includes(d.aliasName)) {
      userIDs.push(d.userID);
    }
  }
  return [...new Set(userIDs)];
}

export function userIDsByGroups(channel, aliases, groups, groupNames) {
  let userIDs = [];
  // let groupNames = [channel.aliasName];
  // if (groupName) {
  //   groupNames.push(groupName);
  // }
  if (Array.isArray(channel.groupAliases)) {
    for (const d of channel.groupAliases) {
      if (groupName == d[0]) {
        for (const d2 of d[2]) {
          groupNames.push(d2);
        }
      }
    }
  }
  for (const d of channel.allAliases) {
    if (groupNames.includes(d[0])) {
      userIDs.push(d[2]);
    }
  }
  return [...new Set(userIDs)];
}
