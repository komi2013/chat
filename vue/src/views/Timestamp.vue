<script setup>
import { ref, onMounted, computed } from 'vue';
import QRCode from 'qrcode';

import Advertisement from '@/components/Advertisement.vue';
import DrawerTimestamp from '@/components/DrawerTimestamp.vue';
import NoticePopup from '@/components/NoticePopup.vue';

import { userIDsByName, userIDsByGroups } from '@/my/channelFunc';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  adminName: String,
  code: String
})

document.title = 'タイムスタンプ'

const channel = ref('');
const groups = ref([]);
const aliases = ref([]);
const timestamps = ref('');
// let timestampID;

function tF(a, b = null){ return timeFormat(a, b) }

const latestTimestamp = computed(() => {
  if (!timestamps.value) return null;
  if (timestamps.value.length === 0) return null;
  return timestamps.value.sort((a, b) => new Date(b.timeIn) - new Date(a.timeIn))[0];
});

function formatDateTime(dateTime) {
  if (dateTime == null) {
    return '';
  }
  return timeFormat('MM/DD hh:mm', dateTime);
}

onMounted(async () => {
  channel.value = await getIDB('channel', localStorage.getItem('channelID'));
  groups.value = await getIDBs('group', 'channelIDIndex', localStorage.getItem('channelID'), 10000);
  aliases.value = await getIDBs('alias', 'channelIDIndex', localStorage.getItem('channelID'), 10000);

  timestamps.value = await getIDBbyMulti('timestamp', ['channelID', 'aliasName'], 
    [localStorage.getItem('channelID'), channel.value.myname], 30, 0, 'desc');
  // if () {

  // }
  // const latestEntry = timestamps.value.reduce((max, obj) => 
  //   obj.timestampID > max.timestampID ? obj : max, timestamps.value[0]);
  // timestampID = latestEntry.timestampID;
  // console.log(timestamps.value);
});

async function stamp(action) {
  if (!confirm("実行▶️")) {
    return;
  }
  const fd = new FormData();
  fd.append('channelID', localStorage.getItem('channelID'));
  fd.append('updatedBy', channel.value.myname);
  fd.append('pushTitle', 'timestamp');
  const userIDs = [...new Set([
    ...userIDsByGroups(aliases.value, groups.value, props.adminName),
    ...userIDsByName(aliases.value, [channel.value.myname])
  ])];

  fd.append('userIDs', JSON.stringify(userIDs));
  const now = timeFormat('YYYY-MM-DDThh:mm');
  fd.append('contents', JSON.stringify([props.code, action, now, channel.value.myname]));
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ContentsPush/', fd);
  if (!res.csrf) errorMessage.value = res
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
}


</script>


<template>
<DrawerTimestamp />

<div id="content">

<h2 class="sp_head">タイムスタンプ</h2>
<div style="width: 100%; text-align: center;"><Advertisement /></div>
<br>
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
            {{ formatDateTime(breakPeriod.start) }} - {{ formatDateTime(breakPeriod.end) }}
          </li>
        </ul>
      </div>
    </div>
    <div v-else>
      <p>No timestamp data available</p>
    </div>
  </div>

  <div>
    <a :href="'/timestampReport/' + adminName + `/?month=${tF('YYYY-MM')}` ">
    {{ '/timestampReport/' + adminName + `/?month=${tF('YYYY-MM')}` }} </a>
  </div>

</div>
<div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>
<NoticePopup />
</template>

<style scoped>

.button {
  text-align: center;
  width: 100%;
}

.button button {
  width: 80%;
  height: 30px;
}
</style>

