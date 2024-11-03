<script setup>
import { ref, onMounted } from 'vue';

const props = defineProps({
  id: '',
})

const channel = ref('');
// const bookPattern = ref('');
async function fetchChannel() {
  try {
    channel.value = await getIDB('channel', localStorage.channelID);
  } catch (error) {
    console.log('error', error);
    channel.value = null;
  }
}
const bookPattern = ref('');
async function fetchBookPattern() {
  try {
    bookPattern.value = await getIDB('bookPattern', props.id);
    bookPattern.value.times.forEach((time) => {
      if (!time.manualStaffs) {
        time.manualStaffs = [["", time.limitStart, time.limitEnd]];
      } else {
        time.manualStaffs.unshift(["", time.limitStart, time.limitEnd]);
      }
    });
    if (bookPattern.value.parentID) {
      findBookParent();
    }
  } catch (error) {
    console.log('error', error);
    bookPattern.value = null;
  }
}

const shiftStaff = ref('');
let initialStaffs = [];
async function fetchShiftStaff() {
  try {
    shiftStaff.value = await getIDB('shiftStaff', props.id);
    initialStaffs = JSON.parse(JSON.stringify(shiftStaff.value));
  } catch (error) {
    console.log('error', error);
    shiftStaff.value = [];
  }
}

// function extractStaffs(bookPattern) {
//   const staffs = [];
//   bookPattern.times.forEach((time, timeIndex) => {
//     if (time.staffs !== undefined) {
//       time.staffs.forEach((staff) => {
//         const role = staff[1];
//         const staffNames = staff[2];
//         staffs.push({
//           start: time.start,
//           role: role,
//           name: staffNames
//         });
//       });
//     }
//   });
//   return staffs
// }

const IamAdmin = () => {
  let iamAdmin = true;
  if (Array.isArray(channel.value.groupAliases)) {
    for (const d of channel.value.groupAliases) {
      if (bookPattern.value.adminGroup == d[0]) {
        for (const name of d[2]) {
          if (channel.value.aliasName == name) {
            iamAdmin = true;
          }
        }
      }
    }
  }
  return iamAdmin;
}

const getShiftStaffByDate = (timeSlot) => {
  // if (!Array.isArray(shiftStaff.value)) {
  //   return []
  // }
  const { date, start } = timeSlot;
  const result = (shiftStaff.value || []).filter(shiftStaff => {
      const shiftDate = shiftStaff.shiftStart.slice(0, 10);
      const shiftTime = shiftStaff.shiftStart.slice(11);
      return (
        shiftDate === date &&
        shiftTime === start
      );
    })
    .sort((a, b) => a.seq - b.seq);
  return result;
};

const myNameInclude = (timeSlot) => {
  // if (!shiftStaff.value) {
  //   return
  // }
  const { date, start } = timeSlot;
  return (shiftStaff.value || []).some(shiftStaff => {
    const shiftDate = shiftStaff.shiftStart.slice(0, 10);
    const shiftTime = shiftStaff.shiftStart.slice(11);

    return (
      shiftDate === date &&
      shiftTime === start &&
      shiftStaff.aliasName === channel.value.aliasName
    );
  });
};

// manual data
    // const shiftStaffs = ref([
    // {
    //   shiftStaffID: props.id + "202411011000" + "staffA",
    //   bookPatternID: props.id,
    //   aliasName: "staffA",
    //   shiftStart: "2024-11-01T10:00",
    //   shiftEnd: "2024-11-01T15:00",
    //   role: "stylist",
    //   seq: 2
    // },
    // {
    //   shiftStaffID: props.id + "202411011000" + "staffB",
    //   bookPatternID: props.id,
    //   aliasName: "staffB",
    //   shiftStart: "2024-11-01T10:00",
    //   shiftEnd: "2024-11-01T15:00",
    //   role: "stylist",
    //   seq: 1
    // }
    // ]);


onMounted(() => {
  fetchChannel();
  fetchBookPattern();
  fetchShiftStaff();
});

let parent;
async function findBookParent() {
  const fd = new FormData();
  fd.append('channelID', localStorage.channelID);
  fd.append('aliasName', channel.value.aliasName);
  fd.append('windowID', localStorage.channelID + bookPattern.value.parentID);
  const request = new Request('/WindowGet/', {
    method: 'POST',
    body: fd,
  });
  try {
    const response = await fetch(request);
    if (response.ok) {
      parent = await response.json();
    } else {
      console.error('Failed to findBookParent data', response.status);
    }
  } catch (error) {
    console.error('Error findBookParent data:', error);
  }
}

const bookPatternID = ref(props.id);

const moveStaffToTop = (timeSlot, name) => {
  // console.log(shiftStaff.value);
  // if (!Array.isArray(shiftStaff.value)) {
  //   return
  // }
  const { date, start } = timeSlot;
  const shiftStart = date + 'T' + start;
  const staffs = (shiftStaff.value || []).filter(staff => staff.shiftStart === shiftStart);
  staffs.sort((a, b) => {
    if (a.aliasName === name) return -1;
    if (b.aliasName === name) return 1;
    return a.seq - b.seq;
  });
  staffs.forEach((staff, index) => {
    staff.seq = index + 1;
  });
};

function editOK(timeSlot, role, seq) {
  // if (!shiftStaff.value) {
  //   return
  // }
  const { date, start, end } = timeSlot;
  const shiftStaffID = props.id + date.replace(/-/g, "") + start.replace(":", "") + channel.value.aliasName;
  const existingIndex = shiftStaff.value.findIndex(staff => staff.shiftStaffID === shiftStaffID);
  if (existingIndex !== -1) {
    shiftStaff.value.splice(existingIndex, 1); // 同じIDが存在すれば削除
  } else {
    const staff = {
      shiftStaffID: shiftStaffID,
      bookPatternID: props.id,
      aliasName: channel.value.aliasName,
      shiftStart: date + "T" + start,
      shiftEnd: date + "T" + end,
      role: role,
      seq: seq
    }
    shiftStaff.value.push(staff);    
  }
}

function submitOK() {
  if (!confirm("実行▶️")) {
    return;
  }
  const fd = new FormData();
  let userIDs = [];
  let names = [channel.value.aliasName];
  let iamAdmin = false;
  if (Array.isArray(channel.value.groupAliases)) {
    for (const d of channel.value.groupAliases) {
      if (bookPattern.value.adminGroup == d[0]) {
        for (const d2 of d[2]) {
          names.push(d2);
          if (channel.value.aliasName == d2) {
            iamAdmin = true;
          }
        }
      }
    }
  }
  if (Array.isArray(bookPattern.value.joinNames)) {
    for (const d of bookPattern.value.joinNames) {
      names.push(d);
    }
  }
  for (const d of channel.value.allAliases) {
    if (names.includes(d[0])) {
      userIDs.push(d[2]);
    }
  }
  fd.append('userIDs', JSON.stringify([...new Set(userIDs)]));
  fd.append('channelID', localStorage.channelID);
  fd.append('aliasName', channel.value.aliasName);
  const changedData = compareAndUpdateStaffs(initialStaffs, shiftStaff.value);
  console.log(changedData.length);
  if (changedData.length < 1) {
    return
  }
  fd.append('contents', JSON.stringify(changedData));
    // if (iamAdmin) {
    //   // const updatedStaffs = extractStaffs(bookPattern.value);
    //   // console.log(initialStaffs);
    //   // console.log(updatedStaffs);

    //   console.log('contents', changedData);
    // } else {
    //   const okStaffs = processOkStaffs();
    //   if (okStaffs.length < 1) {
    //     return
    //   }
    //   fd.append('contents', JSON.stringify(okStaffs));
    // }

  if (bookPattern.value.parentID) {
    fd.append('windowID', localStorage.channelID + bookPattern.value.parentID);
    const request = new Request('/ShiftStaffEdit/', {
      method: 'POST',
      body: fd,
    });
    fetch(request)
      .catch((reason)=>{
        console.log(reason);
      })
  } else {
    fd.append('pushTitle', 'shiftStaff');
    const request = new Request('/ContentsPush/', {
      method: 'POST',
      body: fd,
    });
    fetch(request)
      .catch((reason)=>{
        console.log(reason);
      })


  }

}

function compareAndUpdateStaffs(initialStaffs, updatedStaffs) {
  const changedRecords = [];

  // updatedStaffs をループして、追加や変更があれば changedRecords に追加
  updatedStaffs.forEach((updatedRecord) => {
    const initialRecord = initialStaffs.find(
      (initial) =>
        initial.aliasName === updatedRecord.aliasName &&
        initial.shiftStart === updatedRecord.shiftStart
    );

    // initialRecord が存在しない場合は追加とみなす
    if (!initialRecord) {
      changedRecords.push({
        ...updatedRecord,
        // type: "added",
      });
    } else if (initialRecord.seq !== updatedRecord.seq) {
      // seq に変更がある場合は変更として記録
      changedRecords.push({
        ...updatedRecord,
        // type: "modified",
        // changedField: "seq",
        // previousSeq: initialRecord.seq,
        // newSeq: updatedRecord.seq,
      });
    }
  });

  // initialStaffs にあるが updatedStaffs にないレコードを削除フラグ付きで changedRecords に追加

  initialStaffs.forEach((initialRecord) => {
    const existsInUpdated = (updatedStaffs || []).some(
      (updatedRecord) =>
        updatedRecord.aliasName === initialRecord.aliasName &&
        updatedRecord.shiftStart === initialRecord.shiftStart
    );

    if (!existsInUpdated) {
      changedRecords.push({
        shiftStaffID: updatedRecord.shiftStaffID,
        delete: 1,
      });
    }
  });

  return changedRecords;
}

// const processOkStaffs = () => {
//   // bookPattern.value.times[index].staffs[i1][2]
//   const okStaffs = [];
//   bookPattern.value.times.forEach((time) => {
//     time.staffs.forEach((staff) => {
//       if (staff[3] !== undefined) {
//         okStaffs.push([time.date, time.start, channel.value.aliasName, staff[1], staff[3], time.end]);
//       }
//     });
//   });
//   if (okStaffs.length > 0) {
//     return ['okStaff', bookPattern.value.bookPatternID, okStaffs];
//   } else {
//     return [];
//   }
// };

// function updateOpenTimes(parent, okStaffs) {
//   const currentTime = new Date().toISOString();

//   // openTimesが存在しない場合は新しく配列を作成
//   if (!parent.openTimes) {
//     parent.openTimes = [];
//   }

//   // okStaffsのデータから必要な情報を追加または削除
//   okStaffs[2].forEach(([date, time, name, role, flag]) => {
//     const entry = [name, `${date}T${time}`, currentTime];

//     if (flag === 1) {
//       // 追加処理: flagが1の場合にopenTimesに追加
//       parent.openTimes.push(entry);
//     } else if (flag === -1) {
//       // 削除処理: flagが-1の場合に一致するエントリを削除
//       parent.openTimes = parent.openTimes.filter(
//         ([existingName, existingDateTime]) => !(existingName === name && existingDateTime === `${date}T${time}`)
//       );
//     }
//   });
// }


// const copyTimes = (startTime, endTime) => {
//   bookPattern.value.times.forEach((timeSlot, i1) => {
//     timeSlot.manualStaffs.forEach((staff, index) => {
//       if (index !== 0) {
//         staff[1] = startTime;
//         staff[2] = endTime;
//       }
//     });
//   });
// };

</script>

<template>
  <div class="book-pattern-page">
    <h1>{{ bookPattern.bookTitle }} ({{ bookPatternID }})</h1>

    <div v-for="(timeSlot, index) in bookPattern.times" :key="index" class="time-slot">
      <p>{{ timeSlot.date }} {{ timeSlot.start }} {{ timeSlot.limitStart }} - {{ timeSlot.end }} {{ timeSlot.limitEnd }}</p>

      <p v-for="(staff, i1) in timeSlot.staffs" :key="i1" > {{ staff[1] ? staff[1] : 'スタッフ' }} : 
        <button
          v-for="(d2, i2) in getShiftStaffByDate(timeSlot)" :key="i2"
          :style="{
            color: d2.aliasName === channel.aliasName ? 'black' : 'white', 
            backgroundColor: i2 < staff[0] ? 'blue' : 'silver'
          }"
          @click="IamAdmin() && moveStaffToTop(timeSlot, d2.aliasName)"
          >
          {{ d2.aliasName }}
        </button>
        <br>
        <button 
          @click="editOK(timeSlot, staff[1], i2)" 
          :style="{ backgroundColor: myNameInclude(timeSlot) ? 'blue' : 'silver' }">
          ◯
        </button>        
      </p>
<!--       <p v-for="(staff, i1) in timeSlot.manualStaffs" >
        <template v-if="staff[0] == channel.aliasName || staff[0] == ''">
          <input v-model="staff[1]" type="time" /> ~ <input v-model="staff[2]" type="time" />
          <span v-if="index === 0">
            <button @click="copyTimes(staff[1], staff[2])">コピー</button>
          </span>
        </template>
        <template v-else>
          <p> {{staff[0]}} {{staff[1]}} ~ {{staff[2]}} </p>
        </template>
      </p> -->
    </div>
    <button @click="submitOK">提出</button>
  </div>
</template>

<style scoped>
.book-pattern-page {
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
  background-color: #f9f9f9;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.time-slot {
  border: 1px solid #ccc;
  padding: 10px;
  margin-bottom: 10px;
  border-radius: 4px;
}

button {
  background-color: #007bff;
  color: white;
  border: none;
  padding: 8px 12px;
  cursor: pointer;
  border-radius: 4px;
  height: 36px;
  margin-left: 4px;
}

button:hover {
  background-color: #0056b3;
}
</style>
