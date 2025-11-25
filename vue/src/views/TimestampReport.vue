<script setup>
import { ref, onMounted, computed } from 'vue';

import Advertisement from '@/components/Advertisement.vue';
import DrawerTimestamp from '@/components/DrawerTimestamp.vue';
import NoticePopup from '@/components/NoticePopup.vue';
import SelectGroup from '@/components/SelectGroup.vue';

import { pushReceive } from '@/pushReceive/pushReceive.js';

import { useNoticesStore } from '@/stores/notices.js';

const props = defineProps({
  admin: String,
  month: String,
  stamper: String
});

document.title = '勤怠表'

function tF(a, b = null){ return timeFormat(a, b) }

const thisMonth = props.month ? props.month : timeFormat('YYYY-MM');
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
  return prevNext;
}
const preMonth = getRelativeMonth(thisMonth, -1);
const nextMonth = getRelativeMonth(thisMonth, 1);

function formatBreaksTotal(breaks, status) {
  if ((!breaks || breaks.length === 0) && status === 2) return '';
  if (!breaks || !myReport.value) return '';
  if (!breaks || breaks.length === 0) return '✏️';

  const totalMinutes = breaks.reduce((total, breakTime) => {
    if (!breakTime.start || !breakTime.end) return total; // 片方がない場合はカウントしない
    
    const breakDuration = new Date(breakTime.end) - new Date(breakTime.start);
    return total + Math.floor(breakDuration / (1000 * 60)); // ミリ秒を分に変換
  }, 0);

  const hours = Math.floor(totalMinutes / 60);
  const minutes = totalMinutes % 60;
  console.log(totalMinutes)
  return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`;
}

function formatTime(dateTime) {
  return dateTime ? timeFormat('hh:mm', dateTime) : '--:--';
}

const channel = ref(null);
const groups = ref([]);
const aliases = ref([]);
const selectedGroup = ref(null);
let targetName;
const iamAdmin = ref(false);
const myReport = ref(false)
const stampers = ref('');
const daysInMonth = ref('');
const timestamps = ref('');
let thisMonthEntries;
const approveds = ref(null);
let pushNames = []
const fetched = ref(false)
const errorMessage = ref('')
onMounted(async () => {
  channel.value = await getIDB('channel', localStorage.getItem('channelID'));
  groups.value = await getIDBs('group', 'channelIDIndex', localStorage.getItem('channelID'), 10000);
  aliases.value = await getIDBs('alias', 'channelIDIndex', localStorage.getItem('channelID'), 10000);
  const matchingGroup = groups.value.find(group => group.groupName === props.admin);
  if (matchingGroup && matchingGroup.aliasNames.includes(channel.value.myname)) {
    iamAdmin.value = true
  }
  targetName = props.stamper ? props.stamper : channel.value.myname
  if (targetName == channel.value.myname) {
    myReport.value = true
  }
  if (iamAdmin.value) {
    const data = await getIDBs('timestamp', 'channelIDIndex', localStorage.getItem('channelID'), 10000);
    stampers.value = [...new Set(data.map(item => item.aliasName))];
  }
  timestamps.value = await getIDBbyMulti('timestamp', ['channelID', 'aliasName'], 
    [localStorage.getItem('channelID'), targetName], 60, 0, 'desc');
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
  const nextApprover = localStorage.getItem('nextApprover');
  if (nextApprover) {
    selectedGroup.value = groups.value.find(group => group.groupName === nextApprover) || null;
  }
  pushNames = [...new Set([
    ...groups.value.filter(group => group.groupName === props.admin).flatMap(group => group.aliasNames || []),
    ...[channel.value.myname]
  ])]
  fetched.value = true
});

function generateDaysInMonth() {
  const days = [];
  const daysCount = iamAdmin.value ? new Date(year, month, 0).getDate() + 31 : new Date(year, month, 0).getDate();
  for (let day = 0; day < daysCount; day++) {
    const baseDate = new Date(year, month - 1, 1);
    baseDate.setDate(baseDate.getDate() + day);
    const dateStr = `${baseDate.getFullYear()}-${String(baseDate.getMonth() + 1).padStart(2, '0')}-${String(baseDate.getDate()).padStart(2, '0')}`;
    const timestamp = thisMonthEntries.find(ts => ts.timeIn.startsWith(dateStr));
    const dayData = {
      day: baseDate.getDate(),
      timestampID: timestamp ? timestamp.timestampID : null,
      timeIn: timestamp ? timestamp.timeIn : null,
      timeOut: timestamp ? timestamp.timeOut : null,
      breaks: timestamp ? timestamp.breaks : null,
      stampStatus: timestamp ? timestamp.stampStatus : null,
      approveds: timestamp ? timestamp.approveds : null,
      submit: timestamp ? timestamp.submit : null,
      error: { timeIn: false, timeOut: false, breaks: false }
    };
    // iamAdmin の場合、timestamp があるときだけ push
    if (myReport.value || timestamp) {
      days.push(dayData);
    }
  }
  return days;
}

const selectedStamper = ref( props.stamper ? props.stamper : "");
function onStamperChange() {
  if (selectedStamper.value) {
    location.href =  `/timestampReport/${props.admin}/?month=${thisMonth}&stamper=${selectedStamper.value}`;
  }
}

const allSubmit = ref(false);
function toggleAllSubmit() {
  allSubmit.value = !allSubmit.value;
  daysInMonth.value.forEach(day => {
    if (day.timestampID) {
      day.submit = allSubmit.value;
    }
  });
}

const breakTimes = ref([]);
function addBreak() {
  breakTimes.value.push({ start: '', end: '' });
}

function removeBreak(index) {
  breakTimes.value.splice(index, 1);
}

function saveBreaks() {
  const timeInDate = currentDay.value.timeIn.split('T')[0];
  currentDay.value.breaks = breakTimes.value.map(b => ({
    start: b.start ? `${timeInDate}T${b.start}` : null,
    end: b.end ? `${timeInDate}T${b.end}` : null
  }));

  daysInMonth.value[currentDayIndex.value].stampStatus = 1;
  console.log(daysInMonth.value)
  closeBreaksModal();
}

const currentDay = ref(null);
const currentDayIndex = ref(null);
const showBreaksModal = ref(false);
function openBreaksModal(day, index) {
  if (day.stampStatus == 2) return;
  currentDay.value = day;
  currentDayIndex.value = index;
  if (day.breaks && Array.isArray(day.breaks)) {
    breakTimes.value = day.breaks.map(b => ({
      start: b.start ? b.start.split('T')[1].substring(0, 5) : '',
      end: b.end ? b.end.split('T')[1].substring(0, 5) : ''
    }));
  } else {
    breakTimes.value = [];
  }
  showBreaksModal.value = true;
}

function closeBreaksModal() {
  showBreaksModal.value = false;
}

async function approve() {
  if (!confirm("実行▶️")) {
    return;
  }
  const fd = new FormData();
  localStorage.setItem('nextApprover', selectedGroup.value.groupName); // just shortcut already selected
  // const nextApproverIDs = userIDsByGroups(aliases.value, groups.value, selectedGroup.value.groupName);
  const approverNames = groups.value
    .filter(group => group.groupName === selectedGroup.value.groupName)
    .flatMap(group => group.aliasNames || [])
  fd.append('pushNames', JSON.stringify([...new Set([...pushNames, ...approverNames])]));
  fd.append('channelID', localStorage.getItem('channelID'));
  fd.append('updatedBy', channel.value.myname);
  fd.append('contents', JSON.stringify([1, thisMonthEntries, props.stamper, selectedGroup.value.groupName]));
  fd.append('pushTitle', 'timestampReport');
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ContentsPush/', fd);
  if (!res.csrf) {
    errorMessage.value = res
    return
  }
  localStorage.setItem('csrf', res.csrf);
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  if (res.error) { errorMessage.value = res.error }
  location.href = ''
}

async function deleteReport() {
  const fd = new FormData();
  fd.append('channelID', localStorage.getItem('channelID'));
  fd.append('updatedBy', channel.value.myname);
  fd.append('pushNames', JSON.stringify(pushNames));
  fd.append('contents', JSON.stringify([2, targetName]));
  fd.append('pushTitle', 'timestampReport');
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ContentsPush/', fd);
  if (!res.csrf) {
    errorMessage.value = res
    return
  }
  localStorage.setItem('csrf', res.csrf);
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  if (res.error) { errorMessage.value = res.error }
  location.href = ''
}

async function submitReport() {
  if (!confirm("実行▶️")) {
    return;
  }
  const fd = new FormData();
  fd.append('pushNames', JSON.stringify(pushNames));
  fd.append('channelID', localStorage.getItem('channelID'));
  fd.append('updatedBy', channel.value.myname);
  const timestampIDs = thisMonthEntries.map(entry => entry.timestampID);
  const minTimestampID = timestampIDs.reduce((min, current) => current < min ? current : min);
  const maxTimestampID = timestampIDs.reduce((max, current) => current > max ? current : max);
  fd.append('contents', JSON.stringify([channel.value.myname, minTimestampID, maxTimestampID]));
  fd.append('pushTitle', 'timestampReport');
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ContentsPush/', fd);
  if (!res.csrf) {
    errorMessage.value = res
    return
  }
  localStorage.setItem('csrf', res.csrf);
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  if (res.error) { errorMessage.value = res.error }
  location.href = ''
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
      cells[2].classList.add('error');
      daysInMonth.value[index].error.timeOut = true;
      hasErrors = true;
    } else {
      cells[2].classList.remove('error');
      daysInMonth.value[index].error.timeOut = false;
      if (timeOutStr !== originalTimeOut) {
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

async function manualPost(changedRecords) {
  if (!confirm("実行▶️")) {
    return;
  }
  const fd = new FormData();
  fd.append('pushNames', JSON.stringify(pushNames));
  fd.append('channelID', localStorage.getItem('channelID'));
  fd.append('updatedBy', channel.value.myname);
  fd.append('contents', JSON.stringify(changedRecords));
  fd.append('pushTitle', 'timestampReport');
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ContentsPush/', fd);
  if (!res.csrf) {
    errorMessage.value = res
    return
  }
  localStorage.setItem('csrf', res.csrf);
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  if (res.error) { errorMessage.value = res.error }
  location.href = ''
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
    const noticesStore = useNoticesStore();
    noticesStore.setNotice('コピーしました');
  }).catch(err => {
    console.error('Failed to copy text: ', err);
  });
}


</script>


<template>
<div id="drawer_column"><DrawerTimestamp /></div>

<div id="content" v-if="fetched">
  <div>
    <h2 class="sp_head">
      <a :href="`/timestampReport/${props.admin}/?month=${preMonth}&stamper=${targetName}`">&lt;&lt;</a>
      {{month}}
      <a :href="`/timestampReport/${props.admin}/?month=${nextMonth}&stamper=${targetName}`">&gt;&gt;</a>
    </h2>
    <div v-if="errorMessage"> 
      <div class="errorMessage">{{errorMessage}}<br>データ取得に失敗しました。</div>
      <a href="/setting/"> データ設定ページ </a><br>
      <a href="/sign/"> サインインページ </a>
    </div>
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
          <th v-if="myReport">
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
          <td v-if="myReport">
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

    <button v-if="myReport" @click="manualEdit">修正</button>
    <button v-if="myReport" @click="submitReport">提出</button>
    <button v-if="iamAdmin" @click="approve">承認</button>
    <button v-if="iamAdmin" @click="copyToClipboard">コピー</button>
    <button v-if="approveds" @click="deleteReport">削除</button>
  </div>
</div>

<div v-if="showBreaksModal" class="modal-overlay">
  <div class="modal">
    <h3>{{ tF('MM/DD', currentDay.timeIn) }}</h3>
    <div v-for="(breakItem, breakIndex) in breakTimes" :key="breakIndex">
      <input v-model="breakItem.start" type="time">
      <input v-model="breakItem.end" type="time">
      <button @click="removeBreak(breakIndex)"> ❌ </button>
    </div>
    <button @click="addBreak"> + </button>
    <br>
    <button @click="saveBreaks"> 保存▶️ </button>
    <button @click="closeBreaksModal"> ❌ </button>
  </div>
</div>
<div v-if="!fetched"> 
  <div class="errorMessage">データ取得に失敗しました。</div>
  <a href="/setting/"> データ設定ページ </a><br>
  <a href="/sign/"> サインインページ </a>
</div>
<div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>
<NoticePopup />
</template>

<style scoped>

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

