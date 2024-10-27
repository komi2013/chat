export async function ticketEdit(pushData) {
  const pushID = pushData[0];
  const fd = new FormData();
  fd.append('pushID', pushID);
  const request = new Request('/PushResponse/', {
    method: 'POST',
    body: fd,
  });
  fetch(request);
  const ticket = {
    ticketID: pushData[2],
    title: pushData[3]
  }
  upsertIDB(ticket, 'ticket', 'ticketID', ticket.ticketID)
    .catch((error) => {
      console.error(error);
    });

// thisMonth need to connect latestEntry
// latest timeIn need to connect ticketID

}

