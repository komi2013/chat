<script setup>
import { ref, onMounted } from 'vue';

import Advertisement from '@/components/Advertisement.vue';
import TimestampDrawer from '@/components/DrawerTimestamp.vue';
import { takeUserIDs } from '@/my/channelFunc.js';

const props = defineProps({
  ticketID: String,
});

const statusOptions = ref([
  { label: '下書き', value: 0 },
  { label: '進行中', value: 1 },
  { label: 'レビュー中', value: 2 },
  { label: '完了', value: 3 }
]);

function tF(a, b = null){ return timeFormat(a, b) }

const channel = ref(null);
const aliases = ref(null);
const groups = ref(null);
onMounted(async () => {
  channel.value = await getIDB('channel', props.channel_id);
  aliases.value = await getIDBs('alias', 'channelIDIndex', props.channel_id, 10000);
  groups.value = await getIDBs('group', 'channelIDIndex', props.channel_id, 10000);
  ticket.value = await getIDB('ticket', props.ticketID);
  await fetchTicket();
});

// async function fetchChannel() {
//   try {
//     channel = await getIDB('channel', localStorage.channelID);
//     if (props.ticketID) {
//       fetchTicket();
//     }
//     const individualAliases = channel.allAliases.map(alias => alias[0]);
//     const groupAliases = channel.groupAliases.map(group => group[0]);
//     assigneeOptions.value = [...new Set([...individualAliases, ...groupAliases])];
//   } catch (error) {
//     console.log('channel error', error);
//   }
// }

const ticket = ref({
  ticketID: localStorage.channelID + generateRandomCode(4),
  title: '',           // Default empty title
  status: 0,           // Default status (e.g., 0 might mean "Draft")
  assignee: '',        // Default assignee (empty)
  description: '',     // Default empty description
  createdBy: '',// Default createdBy value
  createdAt: new Date().toISOString(),  // Set to current date by default
  updatedAt: null,     // Default no updated date
  contentsType: 1,     // Example default contentsType, set to 1 or whatever your use case is
  contents: []         // Empty array for contents, assuming this might hold timestamp data
});

const assigneeOptions = ref('');
const selectedStatus = ref('');
const selectedAssignee = ref('');
async function fetchTicket() {
  ticket.value = await getIDB('ticket', props.ticketID);
  // console.log(ticket.value);

  // ticket.value = data.ticket;
  // ticketLogs.value = data.ticketLogs;
  selectedStatus.value = statusOptions.value.find(option => option.value === ticket.value.status)?.value || 0
  selectedAssignee.value = assigneeOptions.value.includes(ticket.value.assignee) ? ticket.value.assignee : assigneeOptions.value[0];
  if (ticket.value.contentsType == 2) {
    timestamps = ticket.value.contents;
    await fetchTimestamp(timestamps);
    const latestEntry = timestamps.reduce((max, obj) => 
      obj.timeIn > max.timeIn ? obj : max, timestamps[0]);
    const year = tF('YYYY', latestEntry.timeIn);
    const month = tF('MM', latestEntry.timeIn);
    thisMonth.value = tF('YYYY年MM月', latestEntry.timeIn);
    daysInMonth.value = generateDaysInMonth(year, month);
  }
}

// const ticketLogs = ref(null);

const newComment = ref('');
function saveChanges() {
  if (!confirm("実行▶️")) {
    return;
  }
  const fd = new FormData();
  const userIDs = takeUserIDs(channel);
  fd.append('userIDs', JSON.stringify([...new Set(userIDs)]));
  fd.append('channelID', localStorage.channelID);
  ticket.value.channelID = localStorage.channelID;
  fd.append('aliasName', channel.aliasName);
  ticket.value.aliasName = channel.aliasName;
  const newLog = {
    createdAt: new Date().toISOString(),
    createdBy: channel.aliasName,
    changedTexts: [{ title: 'コメント', description: newComment.value }]
  };
  if (ticket.value.ticketLogs) {
    ticket.value.ticketLogs.push({ ...newLog });    
  } else if (newComment.value) {
    ticket.value.ticketLogs = [newLog];
  }
  fd.append('contents', JSON.stringify(ticket.value));
  fd.append('pushTitle', 'ticket');
  const request = new Request('/ContentsPush/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      console.log(reason)
    })
}


async function fetchTimestamp(timestamps) {
  try {
    const ts = await getIDBbyMulti('timestamp', ['channelID', 'aliasName'], 
      [localStorage.channelID, ticket.value.createdBy], 60, 0, 'desc');
    const checkMonth = ts.filter(entry => {
      return entry.stampStatus >= 20;
    });
    console.log('check fraud', timestamps, checkMonth);
    compareRecords(timestamps, checkMonth);
    console.log(timestamps);
  } catch (error) {
    console.log('getIDBbyMulti timestamp no record:', error);
  }
}

function compareRecords(timestamps, checkMonth) {
  const differences = [];
  const timestampsIDs = new Set(timestamps.map(record => record.timestampID));
  const checkMonthIDs = new Set(checkMonth.map(record => record.timestampID));
  timestamps.forEach(record => {
    const matchingRecord = checkMonth.find(r => r.timestampID === record.timestampID);
    if (matchingRecord) {
      if (record.timeIn !== matchingRecord.timeIn ||
          record.timeOut !== matchingRecord.timeOut ||
          record.stampStatus !== matchingRecord.stampStatus ||
          JSON.stringify(record.breaks) !== JSON.stringify(matchingRecord.breaks)) {
          record.error = 'データの違いがあります';
      }
    } else {
      record.error = 'このレコードが存在しません';
    }
  });
  checkMonth.forEach(record => {
    if (!timestampsIDs.has(record.timestampID)) {
      // differences.push({
      //   timestampID: record.timestampID,
      //   message: 'timestampsにこのレコードが存在しません',
      //   checkMonth: record
      // });
    }
  });
}

const thisMonth = ref('');
let timestamps;
const daysInMonth = ref('');
function generateDaysInMonth(year, month) {
  const days = [];
  const daysCount = new Date(year, month, 0).getDate();
  for (let day = 1; day <= daysCount; day++) {
    const dateStr = `${year}-${month}-${String(day).padStart(2, '0')}`;
    const timestamp = timestamps.find(ts => ts.timeIn.startsWith(dateStr));
    days.push({
      day: day,
      timestampID: timestamp ? timestamp.timestampID : null,
      timeIn: timestamp ? timestamp.timeIn : null,
      timeOut: timestamp ? timestamp.timeOut : null,
      breaks: timestamp ? timestamp.breaks : null,
      stampStatus: timestamp ? timestamp.stampStatus : 0,
      error: timestamp ? timestamp.error : null
    });
  }
  return days;
}

function formatBreaksTotal(breaks) {
  if (!breaks || breaks.length === 0) return '';

  const totalMinutes = breaks.reduce((total, [start, end]) => {
    const breakDuration = new Date(end) - new Date(start);
    return total + Math.floor(breakDuration / (1000 * 60)); // Convert milliseconds to minutes
  }, 0);

  const hours = Math.floor(totalMinutes / 60);
  const minutes = totalMinutes % 60;
  return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`;
}

</script>

<template>
<TimestampDrawer />
<div id="content">
  <div class="ticket-edit-page">
    <br>
    <div class="ticket-form">
      <input v-model="ticket.title" type="text" />
      <div class="form-row">
        <label>ステータス</label>
        <select v-model="selectedStatus">
          <option v-for="option in statusOptions" :key="option.value" :value="option.value">
            {{ option.label }}
          </option>
        </select>
      </div>
      <div class="form-row">
        <label>担当者</label>
        <select v-model="selectedAssignee">
          <option v-for="assignee in assigneeOptions" :key="assignee" :value="assignee">
            {{ assignee }}
          </option>
        </select>
      </div>

      <label>文書</label>
      <textarea v-model="ticket.description"></textarea>

      <div>
        <p>作成者: {{ ticket.aliasName }}</p>
        <p>発行日: {{ tF('YYYY-MM-DD', ticket.createdAt) }}</p>
        <p v-if="ticket.updatedAt">更新日: {{ tF('YYYY-MM-DD', ticket.updatedAt) }}</p>
      </div>
      <span v-if="ticket.contentsType == 2">{{ thisMonth }}</span>
      <table v-if="ticket.contentsType == 2" class="timestamp-table">
        <thead>
          <tr>
            <th>日</th>
            <th>開始</th>
            <th>終了</th>
            <th>休憩</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(day, index) in daysInMonth" :key="day" :class="['status' + day.stampStatus, { 'error': day.error }]">
            <td>{{ day.day }}</td>
            <td>
              {{ day.timeIn ? tF('hh:mm', day.timeIn) : '--:--' }}
              <p v-if="day.error"> {{day.error}} </p>
            </td>
            <td>
              {{ day.timeOut ? tF('hh:mm', day.timeOut) : '--:--' }}
            </td>
            <td>
              {{ formatBreaksTotal(day.breaks) }}
            </td>
          </tr>
        </tbody>
      </table>
      <p>
        閲覧者: <span v-for="name in ticket.accessNames"> {{name}} </span>
      </p>
      <div>
        <p v-for="d in ticket.ticketLogs">
          <span>{{tF('YYYY-MM-DD hh:mm:ss', d.createdAt)}}</span>
          &nbsp;
          <span>{{d.createdBy}}</span>
          <p v-for="d2 in d.changedTexts">
            <span v-if="d2.title">{{d2.title}}:</span>
            <span v-if="d2.description">{{d2.description}}</span>
          </p>
          <div><hr></div>
        </p>
      </div>
      <div v-if="props.ticketID">
        <label>コメント</label>
        <textarea v-model="newComment"></textarea>
      </div>
      <button @click="saveChanges">変更を保存</button>
    </div>
  </div>
</div>
<div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>

</template>

<style scoped>
.ticket-edit-page {
  margin: 0 auto;
  padding: 20px;
  background-color: #f9f9f9;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.ticket-form {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.form-row {
  display: flex;
  gap: 10px;
}

.error {
  color: red;
}

.status1 {
  background-color: #e6c7f0;
}

hr {
  border: none;
  border-top: 1px solid black;
  margin: 20px 0;
}

label {
  font-weight: bold;
  padding-top: 10px;
}

input, select, textarea {
  padding: 8px;
  font-size: 14px;
  border-radius: 4px;
  border: 1px solid #ccc;
  /*width: 100%;*/
}

textarea {
  resize: vertical;
  height: 100px;
  width: 96%;
}

.timestamps-page {
  width: 100%;
  max-width: 600px;
  margin: 0 auto;
  background-color: #f9f9f9;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

table {
  width: 100%;
  border-collapse: collapse;
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
  padding: 10px 15px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

button:hover {
  background-color: #0056b3;
}
</style>
