<script setup>
import { ref, onMounted, computed } from 'vue';

import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue'
import SelectPeople from '@/components/SelectPeople.vue';

import { userIDsByName } from '@/my/channelFunc';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  id: String,
  text: String,
  dates: String,
});
const channelID = localStorage.getItem("channelID");

document.title = 'カレンダー';

function parseDates(dates) {
  if (!dates) return null;
  const parts = dates.split('/');
  const parseDateTime = (dtStr) => {
    const match = dtStr.match(/^(\d{4})(\d{2})(\d{2})T(\d{2})(\d{2})(\d{2})$/);
    if (!match) return null;
    const [, year, month, day, hour, minute, second] = match.map(Number);
    return new Date(year, month - 1, day, hour, minute, second);
  };
  const start = parseDateTime(parts[0]);
  const end = parts[1] ? parseDateTime(parts[1]) : new Date(start.getTime() + 30 * 60 * 1000);
  return { start, end };
}

const { start, end } = parseDates(props.dates) || {
  start: new Date(),
  end: new Date(new Date().getTime() + 30 * 60 * 1000)
};

const currentDate = ref(timeFormat('YYYY-MM-DD', start));

const calendar = ref({
  timeStart: timeFormat('YYYY-MM-DDThh:mm', start),
  timeEnd: timeFormat('YYYY-MM-DDThh:mm', end),
  todo: props.text ?? '',
  aliasNames: [],
  repeatOption: 'none'
});

const googleURL = computed(() => {
  const text = encodeURIComponent(calendar.value.todo);
  const start = timeFormat('YYYYMMDDThhmmss', calendar.value.timeStart);
  const end = timeFormat('YYYYMMDDThhmmss', calendar.value.timeEnd);
  return `https://www.google.com/calendar/render?action=TEMPLATE&text=${text}&dates=${start}/${end}`;
});

const channel = ref(null);
const groups = ref([]);
const aliases = ref([]);
const fetched = ref(false);
onMounted(async () => {
  channel.value = await getIDB('channel', channelID);
  groups.value = await getIDBs('group', 'channelIDIndex', channelID, 10000);
  aliases.value = await getIDBs('alias', 'channelIDIndex', channelID, 10000);
  if (props.id) {
    calendar.value = await getIDB('calendar', props.id);
    currentDate.value = timeFormat('YYYY-MM-DD', calendar.value.timeStart);
  }
  calendar.value.aliasNames = [channel.value.myname];
  fetched.value = true;
});


const submit = async () => {
  let events = generateRepeatedEvents(calendar.value);
  console.log(events);
  if (!confirm("実行▶️")) {
    return;
  }
  for (const event of events) {
    event.calendarID ||= generateRandomCode(8);
    event.channelID = channelID;
    event.aliasName = channel.value.myname;
    const fd = new FormData();
    fd.append('userIDs', JSON.stringify(userIDsByName(aliases.value, event.aliasNames)));
    fd.append('channelID', channelID);
    fd.append('updatedBy', channel.value.myname);
    fd.append('contents', JSON.stringify(event));
    fd.append('pushTitle', 'calendar');
    fd.append('csrf', localStorage.getItem('csrf'));
    const res = await sendRequest('/ContentsPush/', fd);
    if (res.csrf) {
      localStorage.setItem('csrf', res.csrf);
    }
    res.pushContents.forEach(content => {
      pushReceive(content);
    });
  }
  window.close();
};

const delCalendar = async () => {
  if (!confirm("削除🗑")) {
    return;
  }
  calendar.value.delete = 1;
  const fd = new FormData();
  fd.append('userIDs', JSON.stringify(userIDsByName(aliases.value, calendar.value.aliasNames)));
  fd.append('channelID', channelID);
  fd.append('updatedBy', channel.value.myname);
  fd.append('contents', JSON.stringify(calendar.value));
  fd.append('pushTitle', 'calendar');
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ContentsPush/', fd);
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
  window.close();
};

const generateRepeatedEvents = (calendar) => {
  const { timeStart, timeEnd, repeatOption } = calendar;
  if (!timeStart || !timeEnd) return [];
  let start = new Date(timeStart);
  let end = new Date(timeEnd);
  let events = [];
  if (!repeatOption || repeatOption === "none") {
    events.push({ ...calendar, timeStart: timeFormat('YYYY-MM-DDThh:mm', start), timeEnd: timeFormat('YYYY-MM-DDThh:mm', end) });
  } else if (repeatOption === "weekly") {
    for (let i = 0; i < 26; i++) {
      let nextStart = new Date(start);
      let nextEnd = new Date(end);
      nextStart.setDate(nextStart.getDate() + 7 * i);
      nextEnd.setDate(nextEnd.getDate() + 7 * i);
      events.push({ ...calendar, timeStart: timeFormat('YYYY-MM-DDThh:mm', nextStart), timeEnd: timeFormat('YYYY-MM-DDThh:mm', nextEnd) });
    }
  } else if (repeatOption === "monthly") {
    let originalDay = start.getDate();
    for (let i = 1; i <= 6; i++) {
      let nextStart = new Date(start);
      let nextEnd = new Date(end);
      nextStart.setMonth(nextStart.getMonth() + i);
      nextEnd.setMonth(nextEnd.getMonth() + i);
      let lastDayOfNextMonth = new Date(nextStart.getFullYear(), nextStart.getMonth() + 1, 0).getDate();
      nextStart.setDate(Math.min(originalDay, lastDayOfNextMonth));
      nextEnd.setDate(Math.min(originalDay, lastDayOfNextMonth));
      events.push({ ...calendar, timeStart: timeFormat('YYYY-MM-DDThh:mm', nextStart), timeEnd: timeFormat('YYYY-MM-DDThh:mm', nextEnd) });
    }
  }
  return events;
};

</script>

<template>
<Drawer />
<div id="content">
  <div class="modal-content">
    <input type="datetime-local" v-model="calendar.timeStart" />
    <span> ~ </span>
    <input type="datetime-local" v-model="calendar.timeEnd" />
    <textarea type="text" placeholder="Todo" v-model="calendar.todo" required></textarea>
    <button @click="submit" :disabled="!calendar.todo">▶️</button>
    <button @click="delCalendar" :disabled="!calendar.calendarID">🗑</button>
    <button><a onclick="window.close();">x</a></button>
    <SelectPeople v-if="fetched"
      :aliases="aliases"
      :groups="groups"
      :placeholder="'参加ユーザー'"
      v-model="calendar.aliasNames"
      />
    <div>
      <label>繰り返しオプション:</label>
      <select v-model="calendar.repeatOption">
        <option value="none">繰り返しなし</option>
        <option value="weekly">毎週同じ曜日</option>
        <option value="monthly">毎月同じ日</option>
      </select>
    </div>
  </div>
  <textarea class="google-url">{{googleURL}}</textarea>
</div>
<div id="ad_right"> <br><br><br> <Advertisement /> <br><br><br> <Advertisement /> <br><br><br> <Advertisement /> </div>

</template>

<style scoped>

button {
  width: 80%;
  height: 30px;
  margin: 8px;
}

.modal-content {
  background-color: white;
  padding: 20px;
  border-radius: 5px;
  max-width: 400px;
  width: 100%;
}

textarea {
  width: 90%;
  height: 50px;
  margin: 2px;
  padding: 2px;
}

</style>
