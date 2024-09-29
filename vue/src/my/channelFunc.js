export function takeUserIDs(channel, name = null) {
  let userIDs = [];
  let names = [channel.aliasName];
  if (name) {
    names.push(name);
  }
  if (Array.isArray(channel.groupAliases)) {
    for (const d of channel.groupAliases) {
      if (name) {
        if (name== d[0]) {
          for (const d2 of d[2]) {
            names.push(d2);
          }
        }
      }
    }
  }
  for (const d of channel.allAliases) {
    if (names.includes(d[0])) {
      userIDs.push(d[2]);
    }
  }
  return userIDs
}