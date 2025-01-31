<script setup>
import { ref, onMounted, computed } from 'vue';
import SelectPeople from '@/components/SelectPeople.vue';
import { userIDsByName } from '@/my/channelFunc';
const props = defineProps({
  id: String,
  text: String,
  dates: String,
});
const channelID = localStorage.getItem("channelID");

document.title = 'カレンダー';

function parseDates(dates) {
  console.log('dates', dates);
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
  todo: props.text ?? ''
});

const googleURL = computed(() => {
  const text = encodeURIComponent(calendar.value.todo);
  const start = timeFormat('YYYYMMDDThhmmss', calendar.value.timeStart);
  const end = timeFormat('YYYYMMDDThhmmss', calendar.value.timeEnd);
  return `https://www.google.com/calendar/render?action=TEMPLATE&text=${text}&dates=${start}/${end}`;
});

let channel;
let aliases;
let groups;
onMounted(async () => {
  channel = await getIDB('channel', channelID);
  aliases = await getIDBs('alias', 'channelIDIndex', channelID, 10000);
  groups = await getIDBs('group', 'channelIDIndex', channelID, 10000);
  if (props.id) {
    calendar.value = await getIDB('calendar', props.id);
    currentDate.value = timeFormat('YYYY-MM-DD', calendar.value.timeStart);
  }
});

const submit = async () => {
  if (!confirm("実行▶️")) {
    return;
  }
  calendar.value.calendarID ||= generateRandomCode(8);
  calendar.value.channelID = channelID;
  calendar.value.aliasName = channel.myname;
  const fd = new FormData();
  fd.append('userIDs', JSON.stringify(userIDsByName(aliases, [channel.myname])));
  // fd.append('userIDs', JSON.stringify([...new Set(userIDs)]));
  fd.append('channelID', channelID);
  fd.append('updatedBy', channel.myname);
  fd.append('contents', JSON.stringify(calendar.value));
  fd.append('pushTitle', 'calendar');
  await sendRequest('/ContentsPush/', fd);
  window.close();
};

const delCalendar = async () => {
  if (!confirm("削除🗑")) {
    return;
  }
  calendar.value.delete = 1;
  const fd = new FormData();
  fd.append('userIDs', JSON.stringify(userIDsByName(aliases, [channel.myname])));
  fd.append('channelID', channelID);
  fd.append('updatedBy', channel.myname);
  fd.append('contents', JSON.stringify(calendar.value));
  fd.append('pushTitle', 'calendar');
  await sendRequest('/ContentsPush/', fd);
  window.close();
};

</script>

<template>
  <div class="modal">
    <div class="modal-content">
      <input type="datetime-local" v-model="calendar.timeStart" />
      <span> ~ </span>
      <input type="datetime-local" v-model="calendar.timeEnd" />
      <textarea type="text" placeholder="Todo" v-model="calendar.todo" required></textarea>
      <SelectPeople v-if="groups"
        :aliases="aliases"
        :groups="groups"
        :placeholder="'参加ユーザー'"
        v-model="calendar.aliasNames"
        />
      <button @click="submit" :disabled="!calendar.todo">▶️</button>
      <button @click="delCalendar" :disabled="!calendar.calendarID">🗑</button>
      <button><a onclick="window.close();">x</a></button>
    </div>
  </div>
  <textarea class="google-url">{{googleURL}}</textarea>
</template>

<style scoped>

button {
  width: 80%;
  height: 30px;
  margin: 8px;
}

.modal {
  display: flex;
  justify-content: center;
}

.modal-content {
  background-color: white;
  padding: 20px;
  border-radius: 5px;
  max-width: 400px;
  width: 100%;
}

.modal-content textarea {
  width: 100%;
}


.google-url {
  width: 80%;
  height: 50px;
  margin: 8px;
  padding: 8px;
}

</style>
