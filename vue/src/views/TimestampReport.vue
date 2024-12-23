<script setup>
import { ref, onMounted, computed } from 'vue';
import TimestampDrawer from '../components/TimestampDrawer.vue';

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
  return `/timestampReport/${props.admin}/${prevNext}/${props.stamper}/`;
}
const preMonth = getRelativeMonth(thisMonth, -1);
const nextMonth = getRelativeMonth(thisMonth, 1);

let userIDs = [];
let names = [];
function makeUserIDs() {
  names = [channel.value.aliasName];
  if (Array.isArray(channel.value.groupAliases)) {
    for (const d of channel.value.groupAliases) {
      if (props.admin == d[0]) {
        for (const d2 of d[2]) {
          names.push(d2);
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
const channel = ref('');
async function fetchChannel() {
  try {
    const data = await getIDB('channel', localStorage.channelID);
    channel.value = data;
    makeUserIDs();
    const aliasName = data.aliasName;
    const allAliases = data.allAliases;
    if (props.admin == '_') {
      fetchStamperList();
      if (props.stamper != '_') {
        fetchTimestamp(props.stamper);
      }
    } else {
      fetchTimestamp(aliasName);
    }
  } catch (error) {
    console.log('channel error', error);
    channel.value = null;
  }
}

const stampers = ref('');
async function fetchStamperList() {
  try {
    const data = await getIDBs('timestamp', 'channelIDIndex', localStorage.channelID);
    stampers.value = [...new Set(data.map(item => item.aliasName))];
    console.log('stampers', stampers.value);
  } catch (error) {
    console.log('stampers error', error);
    stampers.value = null;
  }
}
const selectedStamper = ref("");
const placeholder = ref("選択");
if (props.stamper == '_') {
  selectedStamper.value = "";
} else {
  selectedStamper.value = props.stamper;
}

function onStamperChange() {
  if (selectedStamper.value) {
    const stamper = selectedStamper.value;
    // If a stamper is selected, navigate to the new URI
    // router.push(`/timestampReport/_/_/${selectedStamper.value}/`);
    location.href =  `/timestampReport/${props.admin}/${thisMonth}/${stamper}/`;
  }
}

const timestamps = ref('');
let latestEntryThisMonth;
let thisMonthEntries;
async function fetchTimestamp(aliasName) {
  try {
    timestamps.value = await getIDBbyMulti('timestamp', ['channelID', 'aliasName'], 
      [localStorage.channelID, aliasName], 60, 0, 'desc');
    thisMonthEntries = timestamps.value.filter(entry => {
      const entryMonth = entry.timeIn.slice(0, 7);
      return entryMonth === thisMonth;
    });

    latestEntryThisMonth = thisMonthEntries.reduce((max, obj) =>
      obj.timeIn > max.timeIn ? obj : max, thisMonthEntries[0]);
    daysInMonth.value = generateDaysInMonth();
  } catch (error) {
    console.log('getIDBbyMulti timestamp no record:', error);
    timestamps.value = null;
  }
}

const daysInMonth = ref('');
function generateDaysInMonth() {
  const days = [];
  const daysCount = new Date(year, month, 0).getDate();
  for (let day = 1; day <= daysCount; day++) {
    const dateStr = `${String(year)}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
    const timestamp = thisMonthEntries.find(ts => ts.timeIn.startsWith(dateStr));
    days.push({
      day: day,
      timestampID: timestamp ? timestamp.timestampID : null,
      timeIn: timestamp ? timestamp.timeIn : null,
      timeOut: timestamp ? timestamp.timeOut : null,
      breaks: timestamp ? timestamp.breaks : null,
      stampStatus: timestamp ? timestamp.stampStatus : null,
      error: { timeIn: false, timeOut: false, breaks: false }
    });
  }
  return days;
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
    console.log("Errors found, not posting data.");
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

function parseTime(date, timeStr) {
  return new Date(`${date}T${timeStr}`);
}

function formatTime(dateTime) {
  return dateTime ? timeFormat('hh:mm', dateTime) : '--:--';
}

function formatBreaksTotal(breaks) {
  if (!breaks || breaks.length === 0) return '✏️';

  const totalMinutes = breaks.reduce((total, [start, end]) => {
    const breakDuration = new Date(end) - new Date(start);
    return total + Math.floor(breakDuration / (1000 * 60)); // Convert milliseconds to minutes
  }, 0);

  const hours = Math.floor(totalMinutes / 60);
  const minutes = totalMinutes % 60;
  return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`;
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
      console.log(reason);
    })
}
let ticketURI;
function submitReport() {
  if (!confirm("実行▶️")) {
    return;
  }
  const fd = new FormData();
  fd.append('userIDs', JSON.stringify(userIDs));
  const aliasName = channel.value.aliasName;
  fd.append('aliasName', aliasName);
  fd.append('channelID', localStorage.channelID);

  const ticket = {
    status: 2,
    title: aliasName + thisMonth + '勤務表',
    assignee: props.admin,
    approver1: props.admin,
    contents: thisMonthEntries,  // 既に配列であると仮定
    contentsType: 2,
    accessNames: names
  };
  fd.append('ticket', JSON.stringify(ticket));
  const request = new Request('/TicketAdd/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .then(response => {
      if (response.ok) {
        return response.json();
      } else {
        console.log(`submitReport: ${response.status} - ${response.statusText}`);
      }
    })
    .then(data => {
      ticketURI = data.TicketID;
      delReportedTimestamps();
    })
    .catch(reason => {
      console.log('submitReport:', reason);
    });
}

function delReportedTimestamps() {
  const fd = new FormData();
  fd.append('channelID', localStorage.channelID);
  fd.append('aliasName', channel.value.aliasName);
  fd.append('userIDs', JSON.stringify(userIDs));
  fd.append('contents', JSON.stringify([latestEntryThisMonth.timeIn, 'deletePrevious']));
  fd.append('pushTitle', 'timestampRevert');
  const request = new Request('/ContentsPush/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .then(response => {
      if (response.ok) {
        location.href = `/ticket/${ticketURI}/`;
      } else {
        console.log(`delReportedTimestams: ${response.status} - ${response.statusText}`);
      }
    })
    .catch(reason => {
      console.log('delReportedTimestams:', reason);
    });
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

onMounted(() => {
  fetchChannel();
});

const showBreaksModal = ref(false);
const breakTimes = ref([]);
const currentDay = ref(null);
const currentDayIndex = ref(null);

function openBreaksModal(day, index) {
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

</script>


<template>
<TimestampDrawer />

<div id="content">

<div>
  <h2> <a :href="preMonth">&lt;&lt;</a> {{month}} <a :href="nextMonth">&gt;&gt;</a> </h2>
  <template v-if="stampers">
    <select v-model="selectedStamper" @change="onStamperChange">
      <option value="">{{ placeholder }}</option>
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
            {{ formatBreaksTotal(day.breaks) }}
          </td>
        </tr>
      </tbody>
    </table>
    <button @click="manualEdit">修正</button>
    <button @click="submitReport">提出</button>
    <button @click="copyToClipboard">コピー</button>
    
  </div>
</div>


<div v-if="showBreaksModal" class="modal-overlay">
  <div class="modal">
    <h3>{{ timeFormat('MM/DD', currentDay.timeIn) }}</h3>
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

