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

export function userIDsByName(channel, names) {
  let userIDs = [];
  for (const d of channel.allAliases) {
    if (names.includes(d[0])) {
      userIDs.push(d[2]);
    }
  }
  return [...new Set(userIDs)];
}