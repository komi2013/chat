<script setup>
import { ref, onMounted, computed } from 'vue';

import QRCode from 'qrcode';

import TimestampDrawer from '../components/TimestampDrawer.vue';
import { get_formated_time } from '../my/get_formated_time.js';
import { getIDB, getAllIDBs, deleteData, getIDBs, getByMulti } from '../my/indexDB.js';
import { generateRandomCode } from '../my/strings.js';

const props = defineProps({
  name: '',
  code: ''
})
const channel = ref('');
async function fetchChannel() {
  try {
    const data = await getIDB('channel', localStorage.channelID);
    channel.value = data;
    fetchTimestamp();
    // console.log(channel.value);
    const aliasName = data.aliasName;
    const allAliases = data.allAliases;
    for (let i = 0; i < allAliases.length; i++) {
      if (allAliases[i][0] === aliasName) {

      }
    }
  } catch (error) {
    console.log('error', error);
    channel.value = null;
  }
}
const timestamps = ref('');
let timestampID;
async function fetchTimestamp() {
  try {
    timestamps.value = await getByMulti('timestamp', ['channelID', 'aliasName'], 
      [localStorage.channelID, channel.value.aliasName], 30, 0, 'desc');
    const latestEntry = timestamps.value.reduce((max, obj) => 
      obj.timestampID > max.timestampID ? obj : max, timestamps.value[0]);
    timestampID = latestEntry.timestampID;
    console.log(timestamps.value);
  } catch (error) {
    console.log('error', error);
    timestamps.value = null;
  }
}

function stamp(action) {
  if (!confirm("実行▶️")) {
    return;
  }
  const fd = new FormData();
  let userIDs = [];
  let names = [channel.value.aliasName];
  names.push(props.name);
  if (Array.isArray(channel.value.groupAliases)) {
    for (const d of channel.value.groupAliases) {
      if (props.name == d[0]) {
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
  const now = get_formated_time('YYYY-MM-DDThh:mm');
  fd.append('userIDs', JSON.stringify([...new Set(userIDs)]));
  fd.append('channelID', localStorage.channelID);
  fd.append('aliasName', channel.value.aliasName);
  fd.append('contents', JSON.stringify([
    props.code, action, now, props.name]));
  fd.append('pushTitle', 'timestamp');
  const request = new Request('/ContentsPush/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      console.log(reason)
    })
}

const latestTimestamp = computed(() => {
  if (!timestamps.value) return null;
  if (timestamps.value.length === 0) return null;
  return timestamps.value.sort((a, b) => new Date(b.timeIn) - new Date(a.timeIn))[0];
});

function formatDateTime(dateTime) {
  if (dateTime == null) {
    return '';
  }
  return get_formated_time('MM/DD hh:mm', dateTime);
}

onMounted(() => {
  fetchChannel();
});
</script>


<template>
<TimestampDrawer />

<div id="content">

<h2>タイムスタンプ</h2>
<div class="button"><button @click="stamp('startWork')">▶️勤務スタート</button></div>
<br>
<div class="button"><button @click="stamp('endWork')">⏹️勤務終了</button></div>
<br>
<div class="button"><button @click="stamp('startBreak')">☕▶️休憩スタート</button></div>
<br>
<div class="button"><button @click="stamp('endBreak')">☕⏹️休憩終了</button></div>

  <div class="timestamp-container">
    <div v-if="latestTimestamp">
      <div>
        <h4>勤務:</h4>
        <ul>
          <li>
            {{ formatDateTime(latestTimestamp.timeIn) }} - {{ formatDateTime(latestTimestamp.timeOut) }}
          </li>
        </ul>
      </div>
      <div>
        <h4>☕休憩:</h4>
        <ul>
          <li v-for="(breakPeriod, index) in latestTimestamp.breaks" :key="index">
            {{ formatDateTime(breakPeriod[0]) }} - {{ formatDateTime(breakPeriod[1]) }}
          </li>
        </ul>
      </div>
    </div>
    <div v-else>
      <p>No timestamp data available</p>
    </div>
  </div>

</div>
</template>

<style scoped>
h2 {
  position: relative;
  left: 50px;
  top: -10px;
}

.button {
  text-align: center;
  width: 100%;
}

.button button {
  width: 80%;
  height: 30px;
}
</style>

