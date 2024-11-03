export async function shiftStaffEdit(pushData) {
  const pushID = pushData[0];
  const fd = new FormData();
  fd.append('pushID', pushID);
  const request = new Request('/PushResponse/', {
    method: 'POST',
    body: fd,
  });
  fetch(request);

  pushData[4].forEach((shiftStaff) => {
    if (shiftStaff.delete) {
      deleteIDB('shiftStaff', 'shiftStaffID', shiftStaff.shiftStaffID);
    } else {
      upsertIDB(shiftStaff, 'shiftStaff', 'shiftStaffID', shiftStaff.shiftStaffID)
        .catch((error) => {
          console.error(error);
        });      
      }
  });

// thisMonth need to connect latestEntry
// latest timeIn need to connect ticketID

}