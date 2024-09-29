import { getIDB, upsertData } from '../my/indexDB.js';

export async function ticketAdd(pushData) {
  const pushID = pushData[0];
  const fd = new FormData();
  fd.append('pushID', pushID);
  const request = new Request('/PushResponse/', {
    method: 'POST',
    body: fd,
  });
  fetch(request);
  const ID = pushData[2];
  const title = pushData[3];
  const ticket {
    ticketID: pushData[2],
    title: pushData[3]
  }
  upsertData(ticket, 'ticketAdd', 'ticketID', ticket.ticketID)
    .catch((error) => {
      console.error(error);
    });

// thisMonth need to connect latestEntry
// latest timeIn need to connect ticketID

}

