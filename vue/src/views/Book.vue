<script setup>
import { ref, onMounted } from 'vue';

const props = defineProps({
  id: '',
})

const channel = ref('');



// const bookPattern = ref('');
async function fetchChannel() {
  try {
    const data = await getIDB('channel', localStorage.channelID);
    channel.value = data;
  } catch (error) {
    console.log('error', error);
    channel.value = null;
  }
}
const bookPattern = ref('');
let initialStaffs = [];
async function fetchBookPattern() {
  try {
    bookPattern.value = await getIDB('bookPattern', props.id);
    initialStaffs = JSON.parse(JSON.stringify(extractStaffs(bookPattern.value)));
    console.log(bookPattern.value);
    bookPattern.value.times.forEach((time) => {
      if (!time.manualStaffs) {
        time.manualStaffs = [["", time.limitStart, time.limitEnd]];
      } else {
        time.manualStaffs.unshift(["", time.limitStart, time.limitEnd]);
      }
    });
  } catch (error) {
    console.log('error', error);
    bookPattern.value = null;
  }
}

function extractStaffs(bookPattern) {
  const staffs = [];
  bookPattern.times.forEach((time, timeIndex) => {
    if (time.staffs !== undefined) {
      time.staffs.forEach((staff) => {
        const role = staff[1];
        const staffNames = staff[2];
        staffs.push({
          start: time.start,
          role: role,
          name: staffNames
        });
      });        
    }
  });
  return staffs
}

onMounted(() => {
  fetchChannel();
  fetchBookPattern();
});

const bookPatternID = ref(props.id);

// const registrations = ref(bookPattern.value.times.map(() => 0));

function editOK(index, i1, okStaffs) {
  let revert = false;
  if (bookPattern.value.times[index].staffs[i1][3] !== undefined) {
    revert = true;
  }
  if (!okStaffs) {
    bookPattern.value.times[index].staffs[i1][2] = [channel.value.aliasName];
    bookPattern.value.times[index].staffs[i1][3] = 1;
    return
  }
  const staffIndex = okStaffs.indexOf(channel.value.aliasName);
  if (staffIndex !== -1) {
    okStaffs.splice(staffIndex, 1);
    bookPattern.value.times[index].staffs[i1][3] = -1;
  } else {
    okStaffs.push(channel.value.aliasName);
    bookPattern.value.times[index].staffs[i1][3] = 1;
  }
  if (revert) {
    bookPattern.value.times[index].staffs[i1].splice(3, 1);
  }
  return
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
  if (iamAdmin) {
    const updatedStaffs = extractStaffs(bookPattern.value);
    console.log(initialStaffs);
    console.log(updatedStaffs);
    const changedData = compareStaffs(initialStaffs, updatedStaffs);
    fd.append('contents', JSON.stringify(changedData));
  } else {
    const okStaffs = processOkStaffs();
    console.log(okStaffs);
    if (okStaffs.length < 1) {
      return
    }
    fd.append('contents', JSON.stringify(okStaffs));
  }
  // console.log( removeEmptyElements(bookPattern.value) );
  fd.append('pushTitle', 'bookPattern');
  const request = new Request('/ContentsPush/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      console.log(reason);
    })
  // if () {
    
  // }
}

function compareStaffs(initialStaffs, updatedStaffs) {
  const changedRecords = [];
  updatedStaffs.forEach((updatedRecord, index) => {
    const initialRecord = initialStaffs[index];
    const isDifferent = (JSON.stringify(updatedRecord.name)
      !== JSON.stringify(initialRecord.name));
    if (isDifferent) {
      changedRecords.push([
          updatedRecord.start,
          updatedRecord.role,
          updatedRecord.name
        ]);
    }
  });
  return ['changeStaffs', bookPattern.value.bookPatternID, changedRecords];
}

const processOkStaffs = () => {
  // bookPattern.value.times[index].staffs[i1][2]
  const okStaffs = [];
  bookPattern.value.times.forEach((time) => {
    time.staffs.forEach((staff) => {
      if (staff[3] !== undefined) {
        okStaffs.push([time.date, time.start, channel.value.aliasName, staff[1], staff[3]]);
      }
    });
  });
  if (okStaffs.length > 0) {
    return ['okStaff', bookPattern.value.bookPatternID, okStaffs];
  } else {
    return [];
  }
};

const IamAdmin = () => {
  let iamAdmin = false;
  if (Array.isArray(channel.value.groupAliases)) {
    for (const d of channel.value.groupAliases) {
      console.log('channel.value.groupAlias', d);
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
const moveStaffToTop = (staffName, okStaffs) => {
  const index = okStaffs.indexOf(staffName);
  if (index > -1) {
    okStaffs.splice(index, 1);
    okStaffs.unshift(staffName);
  }
};

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

      <p v-for="(staff, i1) in timeSlot.staffs" > {{ staff[1] ? staff[1] : 'スタッフ' + i1 }} : 
        <button
          v-for="(staffName, i2) in staff[2]"
          :style="{
            color: staffName === channel.aliasName ? 'black' : 'white', 
            backgroundColor: i2 < staff[0] ? 'blue' : 'silver'
          }"
          @click="IamAdmin() && moveStaffToTop(staffName, staff[2])"
          >
          {{ staffName }}
        </button>
        <br>
        <button 
          @click="editOK(index, i1, staff[2])" 
          :style="{ backgroundColor: staff[2] && staff[2].includes(channel.aliasName) ? 'blue' : 'silver' }">
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
