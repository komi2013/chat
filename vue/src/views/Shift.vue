<script setup>
import { ref, onMounted } from 'vue';

const props = defineProps({
  id: '',
})

const channel = ref('');
async function fetchChannel() {
  try {
    channel.value = await getIDB('channel', localStorage.channelID);
  } catch (error) {
    console.log('error', error);
    channel.value = null;
  }
}
const bookPattern = ref('');
const availableSkills = ref([]);
async function fetchBookPattern() {
  try {
    bookPattern.value = await getIDB('bookPattern', props.id);
    console.log(bookPattern.value);
    if (bookPattern.value.parentID) {
      findParent(bookPattern.value.parentID);
    } else {
      fetchShiftStaff();
    }
  } catch (error) {
    console.log('error', error);
    bookPattern.value = null;
  }
}

const shiftStaff = ref([]);
let initialStaffs = [];
async function fetchShiftStaff() {
  try {
    shiftStaff.value = await getIDBs('shiftStaff', 'bookPatternIDIndex', props.id);
    initialStaffs = JSON.parse(JSON.stringify(shiftStaff.value));
  } catch (error) {
    console.log('error', error);
    // shiftStaff.value = [];
  }
}

async function findParent(parentID) {
  const fd = new FormData();
  // fd.append('channelID', localStorage.channelID);
  // fd.append('aliasName', channel.value.aliasName);
  fd.append('bookPatternID', parentID);
  const res = await sendRequest('/BookPatternGet/', fd);
  res.times.forEach((time) => {
    if (time.shiftStaff) {
      time.shiftStaff.forEach((staff) => {
        shiftStaff.value.push({
          shiftStaffID: props.id + time.date.replace(/-/g, "") + staff.shiftStart.replace(":", "") + staff.aliasName,
          aliasName: staff.aliasName,
          shiftStart: `${time.date}T${staff.shiftStart}`, // 日付と時間を結合
          shiftEnd: `${time.date}T${staff.shiftEnd}`, // 日付と時間を結合
          skills: staff.skills,
          seq: staff.seq,
        });
      });
    }
  });
  console.log('/BookPatternGet/', res);
  initialStaffs = JSON.parse(JSON.stringify(shiftStaff.value));
}

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

onMounted(() => {
  fetchChannel();
  fetchBookPattern();
});

const moveStaffToTop = (timeSlot, name) => {
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

function onOffOK(timeSlot) {
  const { date, start, end } = timeSlot;
  const shiftStaffID = props.id + date.replace(/-/g, "") + start.replace(":", "") + channel.value.aliasName;
  console.log(shiftStaff.value, shiftStaffID);
  const existingIndex = shiftStaff.value.findIndex(staff => staff.shiftStaffID === shiftStaffID);
  if (existingIndex !== -1) {
    shiftStaff.value.splice(existingIndex, 1);
  } else {
    const count = shiftStaff.value.filter(staff => staff.shiftStart === date + "T" + start).length + 1;
    const staff = {
      shiftStaffID: shiftStaffID,
      bookPatternID: props.id,
      aliasName: channel.value.aliasName,
      shiftStart: date + 'T' + start,
      shiftEnd: date + 'T' + end,
      skills: availableSkills.value,
      seq: count
    }
    shiftStaff.value.push(staff);    
  }
}

function submitOK() {
  const changedData = compareAndUpdateStaffs(initialStaffs, shiftStaff.value);
  if (changedData.length < 1) {
    return
  }
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
  fd.append('shiftStaffs', JSON.stringify(changedData));
  if (bookPattern.value.parentID) {
    const request = new Request('/ShiftStaffEdit/', {
      method: 'POST',
      body: fd,
    });
    fetch(request)
      .catch((reason)=>{
        console.log(reason);
      })
  } else {
    fd.append('pushTitle', 'shiftStaffEdit');
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
  console.log('initialStaffs, updatedStaffs', initialStaffs, updatedStaffs);
  const changedRecords = [];
  updatedStaffs.forEach((updatedRecord) => {
    const initialRecord = initialStaffs.find(
      (initial) =>
        initial.aliasName === updatedRecord.aliasName &&
        initial.shiftStart === updatedRecord.shiftStart
    );
    if (!initialRecord) {
      changedRecords.push({
        ...updatedRecord,
      });
    } else if (initialRecord.seq !== updatedRecord.seq) {
      changedRecords.push({
        ...updatedRecord,
      });
    }
  });
  initialStaffs.forEach((initialRecord) => {

    const existsInUpdated = (updatedStaffs || []).some(
      (updatedRecord) =>
        updatedRecord.aliasName === initialRecord.aliasName &&
        updatedRecord.shiftStart === initialRecord.shiftStart
    );
    console.log('existsInUpdated', existsInUpdated);
    if (!existsInUpdated) {
      changedRecords.push({
        ...initialRecord,
        shiftStaffID: initialRecord.shiftStaffID,
        delete: true,
      });
    }
  });
  console.log('changedRecords', changedRecords);
  const submitData = changedRecords.map(staff => {
    const { shiftStaffID, bookPatternID, ...rest } = staff;
    return {
      ...rest,
      bookPatternID: bookPattern.value.parentID,
      shiftStaffID: shiftStaffID
    };
  });
  return submitData;
}

</script>

<template>
  <div class="book-pattern-page">
    <h1>{{ bookPattern.bookTitle }}</h1>

  <div>
    <h3>スキルを選択:</h3>
    <div v-for="(skill, index) in bookPattern.skills" :key="index">
      <label>
        <input
          type="checkbox"
          :value="skill"
          v-model="availableSkills"
        />
        {{ skill }}
      </label>
    </div>

    <p>選択されたスキル: {{ availableSkills }}</p>
  </div>

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
          @click="onOffOK(timeSlot)" 
          :style="{ backgroundColor: myNameInclude(timeSlot) ? 'blue' : 'silver' }">
          ◯
        </button>        
      </p>
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
