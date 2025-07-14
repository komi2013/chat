export async function ticket(pushData) {
  const pushID = pushData[0];
  const fd = new FormData();
  fd.append('pushID', pushID);
  const request = new Request('/PushResponse/', {
    method: 'POST',
    body: fd,
  });
  fetch(request);
  const ticket = pushData[4];

  if (ticket.channelID === pushData[2] && ticket.aliasName === pushData[3]) {
    upsertIDB(ticket, 'ticket', 'ticketID', ticket.ticketID);
  }

// thisMonth need to connect latestEntry
// latest timeIn need to connect ticketID

  // arr = append(arr, channelID)
  // arr = append(arr, aliasName)
  // arr = append(arr, contents)

}

