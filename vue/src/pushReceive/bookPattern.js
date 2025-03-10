export async function bookPattern(pushData) {
  const pushID = pushData[0];
  const fd = new FormData();
  fd.append('pushID', pushID);
  const request = new Request('/PushResponse/', {
    method: 'POST',
    body: fd,
  });
  fetch(request);
  const channelID = pushData[2];
  if (!pushData[4][0]) {
    const bookPattern = pushData[4];
    upsertIDB(bookPattern, 'bookPattern', 'bookPatternID', bookPattern.bookPatternID)
      .catch((error) => {
        console.error(error);
      });
  } else if (pushData[4][0] == 'okStaff') {
    console.log('okay');
    console.log(pushData[4]);
    const bookPattern = await getIDB('bookPattern', pushData[4][1]);
    updateOkStaffs(bookPattern, pushData[4][2]);
  } else if (pushData[4][0] == 'changeStaffs') {
    const bookPattern = await getIDB('bookPattern', pushData[4][1]);
    updateStaffs(bookPattern, pushData[4][2]);
  }
}

const updateOkStaffs = (bookPattern, pushStaffs) => {
  const times = bookPattern.times;
  pushStaffs.forEach(([date, start, joinName, bpRole, diff]) => {
    times.forEach(time => {
      if (time.date === date && time.start === start) {
        time.staffs.forEach(staff => {
          let [staffCount, role, staffList] = staff;
          if (bpRole == role) {
            if (diff > 0) {
              if (staffList === undefined) {
                staffList = [joinName];
              } else {
                staffList.push(joinName);
              }
            } else {
              if (staffList !== undefined) {
                const staffIndex = staffList.indexOf(joinName);
                if (staffIndex !== -1) {
                  staffList.splice(staffIndex, 1);
                }
              }
            }
            if (staffList !== undefined) {
              staff[2] = staffList;
            }
          }
        });
      }
    });
  });
  upsertIDB(bookPattern, 'bookPattern', 'bookPatternID', bookPattern.bookPatternID);
};

function updateStaffs(bookPattern, updateData) {
  updateData.forEach(([start, role, newStaffs]) => {
    bookPattern.times.forEach(time => {
      if (time.start === start) {
        time.staffs.forEach((staff, index) => {
          if (role === null || staff[1] === role) {
            time.staffs[index][2] = newStaffs;
          }
        });
      }
    });
  });
  upsertIDB(bookPattern, 'bookPattern', 'bookPatternID', bookPattern.bookPatternID);
}
