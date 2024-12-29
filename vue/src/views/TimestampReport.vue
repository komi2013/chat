<script setup>
import { ref, onMounted, computed } from 'vue';
import TimestampDrawer from '../components/TimestampDrawer.vue';
import SelectGroup from '../components/SelectGroup.vue';
import { takeUserIDs } from '../my/channelFunc.js';

const props = defineProps({
  admin: String,
  month: {
    type: String,
    default: ''
  },
  stamper: String
});

const thisMonth = props.month == '_' ? timeFormat('YYYY-MM') : props.month;
const [year, month] = thisMonth.split('-').map(Number);
function getRelativeMonth(month, offset) {
  const [year, monthNum] = month.split('-').map(Number);
  let newYear = year;
  let newMonth = monthNum + offset;
  if (newMonth === 0) {
    newMonth = 12;
    newYear -= 1;
  } else if (newMonth === 13) {
    newMonth = 1;
    newYear += 1;
  }
  const prevNext = `${newYear}-${String(newMonth).padStart(2, '0')}`;
  // return `/timestampReport/${props.admin}/${prevNext}/${props.stamper}/`;
  return prevNext;
}
const preMonth = getRelativeMonth(thisMonth, -1);
const nextMonth = getRelativeMonth(thisMonth, 1);
const channel = ref('');
const groups = ref([]);
let targetName;
async function fetchChannel() {
  try {
    const data = await getIDB('channel', localStorage.channelID);
    channel.value = data;
    makeUserIDs();
    targetName = props.stamper === '_' ? data.aliasName : props.stamper;
    const allAliases = data.allAliases;
    if (iamAdmin.value) {
      fetchStamperList();
    }
    fetchTimestamp();
    for (let i = 0; i < data.groupAliases.length; i++) {
      groups.value.push([data.groupAliases[i][0], data.groupAliases[i][1]]);
    }
  } catch (error) {
    console.error('channel error', error);
    channel.value = null;
  }
}

let userIDs = [];
let names = [];
let iamAdmin = ref(false);
function makeUserIDs() {
  names = [channel.value.aliasName];
  if (Array.isArray(channel.value.groupAliases)) {
    for (const d of channel.value.groupAliases) {
      if (props.admin == d[0]) {
        for (const d2 of d[2]) {
          names.push(d2);
          if (channel.value.aliasName === d2) {
            iamAdmin.value = true;
          }
        }
      }
    }
  }
  for (const d of channel.value.allAliases) {
    if (names.includes(d[0])) {
      userIDs.push(d[2]);
    }
  }
  userIDs = [...new Set(userIDs)];
  names =  [...new Set(names)];
}

const stampers = ref('');
async function fetchStamperList() {
  try {
    const data = await getIDBs('timestamp', 'channelIDIndex', localStorage.channelID);
    stampers.value = [...new Set(data.map(item => item.aliasName))];
  } catch (error) {
    console.error('stampers error', error);
    stampers.value = null;
  }
}
const selectedStamper = ref( props.stamper === '_' ? "" : props.stamper );
function onStamperChange() {
  if (selectedStamper.value) {
    const stamper = selectedStamper.value;
    location.href =  `/timestampReport/${props.admin}/${thisMonth}/${stamper}/`;
  }
}

const timestamps = ref('');
let thisMonthEntries;
const approveds = ref(null);
async function fetchTimestamp() {
  try {
    timestamps.value = await getIDBbyMulti('timestamp', ['channelID', 'aliasName'], 
      [localStorage.channelID, targetName], 60, 0, 'desc');
    thisMonthEntries = timestamps.value.filter(entry => {
      const entryMonth = entry.timeIn.slice(0, 7);
      if (iamAdmin.value) {
        return entryMonth === thisMonth || entryMonth === nextMonth;;
      } else {
        return entryMonth === thisMonth;
      }
    });
    const uniqueApproveds = new Set();
    for (const record of thisMonthEntries) {
      if (record.approveds && Array.isArray(record.approveds)) {
        for (const approver of record.approveds) {
          uniqueApproveds.add(approver);
        }
      }
    }
    approveds.value = uniqueApproveds.size > 0 ? Array.from(uniqueApproveds) : null;
    daysInMonth.value = generateDaysInMonth();
  } catch (error) {
    console.error('getIDBbyMulti timestamp no record:', error);
    timestamps.value = null;
  }
}

const daysInMonth = ref('');
function generateDaysInMonth() {
  const days = [];
  const daysCount = iamAdmin.value ? new Date(year, month, 0).getDate() + 31 : new Date(year, month, 0).getDate();
  for (let day = 0; day <= daysCount; day++) {
    const baseDate = new Date(year, month - 1, 1);
    baseDate.setDate(baseDate.getDate() + day);
    const dateStr = `${baseDate.getFullYear()}-${String(baseDate.getMonth() + 1).padStart(2, '0')}-${String(baseDate.getDate()).padStart(2, '0')}`;
    const timestamp = thisMonthEntries.find(ts => ts.timeIn.startsWith(dateStr));
    if (iamAdmin.value) {
      if (timestamp) {
        days.push({
          day: day,
          timestampID: timestamp.timestampID,
          timeIn: timestamp.timeIn,
          timeOut: timestamp.timeOut,
          breaks: timestamp.breaks,
          stampStatus: timestamp.stampStatus,
          approveds: timestamp.approveds,
          submit: timestamp.submit,
          error: { timeIn: false, timeOut: false, breaks: false }
        });        
      }
    } else {
      days.push({
        day: day,
        timestampID: timestamp ? timestamp.timestampID : null,
        timeIn: timestamp ? timestamp.timeIn : null,
        timeOut: timestamp ? timestamp.timeOut : null,
        breaks: timestamp ? timestamp.breaks : null,
        stampStatus: timestamp ? timestamp.stampStatus : null,
        approveds: timestamp ? timestamp.approveds : null,
        submit: timestamp ? timestamp.submit : null,
        error: { timeIn: false, timeOut: false, breaks: false }
      });
    }

  }
  return days;
}

function formatBreaksTotal(breaks, status) {
  if ((!breaks || breaks.length === 0) && status === 2) return '';
  if (!breaks || iamAdmin.value) return '';
  if (!breaks || breaks.length === 0) return '✏️';
  const totalMinutes = breaks.reduce((total, [start, end]) => {
    const breakDuration = new Date(end) - new Date(start);
    return total + Math.floor(breakDuration / (1000 * 60)); // Convert milliseconds to minutes
  }, 0);

  const hours = Math.floor(totalMinutes / 60);
  const minutes = totalMinutes % 60;
  return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`;
}

function formatTime(dateTime) {
  return dateTime ? timeFormat('hh:mm', dateTime) : '--:--';
}

function tF(a, b = null){ return timeFormat(a, b) }

onMounted(() => {
  fetchChannel();
});

const allSubmit = ref(false);
function toggleAllSubmit() {
  allSubmit.value = !allSubmit.value;
  daysInMonth.value.forEach(day => {
    day.submit = allSubmit.value;
  });
}

const showBreaksModal = ref(false);
const breakTimes = ref([]);
const currentDay = ref(null);
const currentDayIndex = ref(null);

function openBreaksModal(day, index) {
  if (day.stampStatus == 2) return
  currentDay.value = day;
  currentDayIndex.value = index;
  if (day.breaks) {
    breakTimes.value = day.breaks.map(b => [
      b[0].split('T')[1].substring(0, 5),
      b[1].split('T')[1].substring(0, 5)
    ]);    
  }

  showBreaksModal.value = true;
}

function closeBreaksModal() {
  showBreaksModal.value = false;
}

function addBreak() {
  breakTimes.value.push(['', '']);
}

function removeBreak(index) {
  breakTimes.value.splice(index, 1);
}

function saveBreaks() {
  const timeInDate = currentDay.value.timeIn.split('T')[0];
  currentDay.value.breaks = breakTimes.value.map(b => [
    `${timeInDate}T${b[0]}`,
    `${timeInDate}T${b[1]}`
  ]);
  daysInMonth.value[currentDayIndex.value].stampStatus = 1;
  closeBreaksModal();
}

function formatDate(dateStr) {
  const date = new Date(dateStr);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}

function formatBreaks(breaks) {
  return breaks.map(b => `${b[0].split('T')[1].substring(0, 5)} - ${b[1].split('T')[1].substring(0, 5)}`).join(', ');
}

function manualEdit() {
  const rows = document.querySelectorAll('.timestamp-table tbody tr');
  let hasErrors = false;

  rows.forEach((row, index) => {
    const cells = row.querySelectorAll('td');
    let timeInStr = cells[1].innerText.trim();
    let timeOutStr = cells[2].innerText.trim();
    const breaksStr = cells[3].innerText.trim();

    const originalTimeIn = daysInMonth.value[index].timeIn ? formatTime(daysInMonth.value[index].timeIn) : '--:--';
    const originalTimeOut = daysInMonth.value[index].timeOut ? formatTime(daysInMonth.value[index].timeOut) : '--:--';
    const originaBrakes = formatBreaksTotal(daysInMonth.value[index].breaks);
    timeInStr = cleanseTime(timeInStr);
    if (!isValidTime(timeInStr) && timeInStr !== '--:--') {
      cells[1].classList.add('error');
      daysInMonth.value[index].error.timeIn = true;
      hasErrors = true;
    } else {
      cells[1].classList.remove('error');
      daysInMonth.value[index].error.timeIn = false;
      if (timeInStr !== originalTimeIn) {
        const dateStr = `${thisMonth}-${String(daysInMonth.value[index].day).padStart(2, '0')}`;
        daysInMonth.value[index].timeIn = timeInStr === '--:--' ? null : timeFormat('YYYY-MM-DDThh:mm', `${dateStr}T${timeInStr}`);
        daysInMonth.value[index].stampStatus = 1;
      }
    }

    timeOutStr = cleanseTime(timeOutStr);
    if (!isValidTime(timeOutStr) && timeOutStr !== '--:--') {
      cells[2].classList.add('error'); // Add error class to highlight the cell
      daysInMonth.value[index].error.timeOut = true;
      hasErrors = true;
    } else {
      cells[2].classList.remove('error'); // Remove error class if no error
      daysInMonth.value[index].error.timeOut = false;
      if (timeOutStr !== originalTimeOut) {  // Compare with original value
        const dateStr = `${thisMonth}-${String(daysInMonth.value[index].day).padStart(2, '0')}`;
        daysInMonth.value[index].timeOut = timeOutStr === '--:--' ? null : timeFormat('YYYY-MM-DDThh:mm', `${dateStr}T${timeOutStr}`);
        daysInMonth.value[index].stampStatus = 1;
      }
    }
    if (timeInStr === '--:--' && timeOutStr !== '--:--') {
      cells[1].classList.add('error');
      daysInMonth.value[index].error.timeIn = true;
      hasErrors = true;
    }
    if (timeOutStr === '--:--' && timeInStr !== '--:--') {
      cells[2].classList.add('error');
      daysInMonth.value[index].error.timeOut = true;
      hasErrors = true;
    }
    if (breaksStr !== originaBrakes) {
      daysInMonth.value[index].stampStatus = 1;
    }
  });

  if (hasErrors) {
    console.error("Errors found, not posting data.");
    return;
  }

  const changedRecords = daysInMonth.value
    .filter(day => day.stampStatus)
    .map(day => ({
      timestampID: day.timestampID,
      timeIn: day.timeIn,
      timeOut: day.timeOut,
      breaks: day.breaks
    }));
  manualPost(changedRecords);
}

function cleanseTime(timeStr) {
  if (/^\d{3,4}$/.test(timeStr)) {
    if (timeStr.length === 3) {
      timeStr = '0' + timeStr;
    }
    return timeStr.slice(0, 2) + ':' + timeStr.slice(2);
  }
  return timeStr;
}

function isValidTime(timeStr) {
  return /^([01]\d|2[0-3]):([0-5]\d)$/.test(timeStr);
}

function manualPost(changedRecords) {
  if (!confirm("実行▶️")) {
    return;
  }
  const fd = new FormData();
  fd.append('userIDs', JSON.stringify([...new Set(userIDs)]));
  fd.append('channelID', localStorage.channelID);
  fd.append('aliasName', channel.value.aliasName);
  fd.append('contents', JSON.stringify(changedRecords));
  fd.append('pushTitle', 'timestampReport');
  const request = new Request('/ContentsPush/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      console.error(reason);
    })
}
let ticketURI;
function submitReport() {
  if (!confirm("実行▶️")) {
    return;
  }
  const fd = new FormData();
  fd.append('userIDs', JSON.stringify(userIDs));
  fd.append('channelID', localStorage.channelID);
  fd.append('aliasName', channel.value.aliasName);
  const timestampIDs = thisMonthEntries.map(entry => entry.timestampID);
  const minTimestampID = timestampIDs.reduce((min, current) => current < min ? current : min);
  const maxTimestampID = timestampIDs.reduce((max, current) => current > max ? current : max);
  fd.append('contents', JSON.stringify([channel.value.aliasName, minTimestampID, maxTimestampID]));
  fd.append('pushTitle', 'timestampReport');
  const request = new Request('/ContentsPush/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      console.error(reason);
    })
  // location.href = '';
}

function copyToClipboard() {
  const rows = daysInMonth.value.map(day => [
    day.day,
    formatTime(day.timeIn),
    formatTime(day.timeOut),
    formatBreaksTotal(day.breaks)
  ]);

  const tabDelimitedString = rows.map(row => row.join('\t')).join('\n');

  navigator.clipboard.writeText(tabDelimitedString).then(() => {
    alert('Data copied to clipboard!');
  }).catch(err => {
    console.error('Failed to copy text: ', err);
  });
}

const selectedGroup = ref(localStorage.nextApprover ? [localStorage.nextApprover] : []);

function approve() {
  if (!confirm("実行▶️")) {
    return;
  }
  const fd = new FormData();
  localStorage.nextApprover = selectedGroup.value[0];
  const nextApproverIDs = takeUserIDs(channel.value, selectedGroup.value[0]);
  fd.append('userIDs', JSON.stringify([...new Set([...userIDs, ...nextApproverIDs])]));
  fd.append('channelID', localStorage.channelID);
  fd.append('aliasName', channel.value.aliasName);
  fd.append('contents', JSON.stringify([1, thisMonthEntries, props.stamper, selectedGroup.value[0]]));
  fd.append('pushTitle', 'timestampReport');
  const request = new Request('/ContentsPush/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      console.error(reason);
    })
}

function deleteReport() {
  const fd = new FormData();
  fd.append('channelID', localStorage.channelID);
  fd.append('aliasName', channel.value.aliasName);
  fd.append('userIDs', JSON.stringify(takeUserIDs(channel.value, props.admin)));
  fd.append('contents', JSON.stringify([2, targetName]));
  fd.append('pushTitle', 'timestampReport');
  const request = new Request('/ContentsPush/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      console.error(reason);
    })
}


</script>


<template>
<TimestampDrawer />

<div id="content">

<div>
  <h2>
    <a :href="`/timestampReport/${props.admin}/${preMonth}/${props.stamper}/`">&lt;&lt;</a>
    {{month}}
    <a :href="`/timestampReport/${props.admin}/${nextMonth}/${props.stamper}/`">&gt;&gt;</a>
  </h2>
  <template v-if="stampers">
    <select v-model="selectedStamper" @change="onStamperChange">
      <option value="">選択</option>
      <option v-for="stamper in stampers" :key="stamper" :value="stamper">
        {{ stamper }}
      </option>
    </select>
  </template>
</div>
  <div class="timestamps-page">
    <table class="timestamp-table">
      <thead>
        <tr>
          <th>日</th>
          <th>開始</th>
          <th>終了</th>
          <th>休憩</th>
          <th v-if="!iamAdmin">
            提出<br>
            <input type="checkbox" @change="toggleAllSubmit" />
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(day, index) in daysInMonth" :key="day" :class="'status' + day.stampStatus">
          <td>{{ day.day }}</td>
          <td :contenteditable="day.stampStatus != 2" :class="{'error': day.error.timeIn}">
            {{ day.timeIn ? formatTime(day.timeIn) : '--:--' }}
          </td>
          <td :contenteditable="day.stampStatus != 2" :class="{'error': day.error.timeOut}">
            {{ day.timeOut ? formatTime(day.timeOut) : '--:--' }}
          </td>
          <td contenteditable="true" @click="openBreaksModal(day, index)" :class="{'error': day.error.breaks}">
            {{ formatBreaksTotal(day.breaks, day.stampStatus) }}
          </td>
          <td v-if="!iamAdmin">
            <input type="checkbox" v-model="day.submit" />
          </td>
        </tr>
      </tbody>
    </table>

    <div v-if="iamAdmin" class="approval">
      <div v-if="approveds">
        <span>承認者:</span>
        <span v-for="approver in approveds" class="approver">
          {{approver}}
        </span>
      </div>
      <div>次の承認グループ</div>
      <SelectGroup :groups="groups" v-model="selectedGroup"/>
    </div>

    <button v-if="!iamAdmin" @click="manualEdit">修正</button>
    <button v-if="!iamAdmin" @click="submitReport">提出</button>
    <button v-if="iamAdmin" @click="approve">承認</button>
    <button v-if="iamAdmin" @click="copyToClipboard">コピー</button>
    <button v-if="approveds" @click="deleteReport">削除</button>
  </div>
</div>

<div v-if="showBreaksModal" class="modal-overlay">
  <div class="modal">
    <h3>{{ tF('MM/DD', currentDay.timeIn) }}</h3>
    <div v-for="(breakItem, breakIndex) in breakTimes" :key="breakIndex">
      <input v-model="breakTimes[breakIndex][0]" type="time">
      <input v-model="breakTimes[breakIndex][1]" type="time">
      <button @click="removeBreak(breakIndex)"> ❌ </button>
    </div>
    <button @click="addBreak"> + </button>
    <br>
    <button @click="saveBreaks"> 保存▶️ </button>
    <button @click="closeBreaksModal"> ❌ </button>
  </div>
</div>

</template>

<style scoped>
h2 {
  position: relative;
  left: 70px;
  top: -20px;
}
select {
  padding: 8px;
  font-size: 14px;
}

.timestamps-page {
  width: 100%;
  max-width: 600px;
  margin: 0 auto;
/*  padding: 20px;*/
  background-color: #f9f9f9;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.error {
  color: red;
}

table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
}

th, td {
  padding: 10px;
  text-align: center;
  border: 1px solid #ccc;
}

th {
  background-color: #f2f2f2;
  font-weight: bold;
}

td[contenteditable="true"]:focus {
  outline: 2px solid #00f;
}

button {
  margin-top: 20px;
  padding: 10px 20px;
  background-color: #007bff;
  color: #fff;
  border: none;
  border-radius: 5px;
  cursor: pointer;
}

button:hover {
  background-color: #0056b3;
}

.status1 {
  background-color: #e6c7f0;
}

.status2 {
  background-color: silver;
}

.approval {
  padding: 6px;
}

.approver {
  margin-left:  6px;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
}

.modal {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  max-width: 400px;
  width: 100%;
}

.modal h3 {
  margin-bottom: 20px;
}

.modal label {
  display: inline-block;
  width: 70px;
}

.modal input[type="time"] {
  margin-right: 10px;
  padding: 5px;
}

.modal button {
  margin-top: 10px;
}

</style>

