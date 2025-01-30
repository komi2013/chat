<script setup>
import { ref, onMounted } from 'vue';
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

// const start = new Date(props.start || new Date());
// const end = new Date(start.getTime() + 30 * 60 * 1000);
const currentDate = ref(timeFormat('YYYY-MM-DD', start));

const calendar = ref({
  timeStart: timeFormat('YYYY-MM-DDThh:mm', start),
  timeEnd: timeFormat('YYYY-MM-DDThh:mm', end),
  todo: props.text ?? ''
});

// async function fetchCalendar() {
//   try {
//     calendar.value = await getIDB('calendar', props.id);
//     currentDate.value = timeFormat('YYYY-MM-DD', calendar.value.timeStart);
//   } catch (error) {
//     console.error('calendar error', error);
//   }
// }

let channel;
let aliases;
let groups;
onMounted(async () => {
  channel = await getIDB('channel', channelID);
  aliases = await getIDBs('alias', 'channelIDIndex', channelID, 10000);
  groups = await getIDBs('group', 'channelIDIndex', channelID, 10000);
  // const myAlias = aliases.find((d) => d.aliasName === channel.myname);
  // person.value = {
  //   name: myAlias.aliasName,
  //   image: myAlias.aliasImg,
  // };
  if (props.id != '_') {
    calendar.value = await getIDB('calendar', props.id);
    currentDate.value = timeFormat('YYYY-MM-DD', calendar.value.timeStart);
  }
});


let joinNames = [];
const handleSelectedItemsChange = (change) => {
  const { diff, item } = change;
  if (diff === 1) {
    console.log("Item added:", item);
    joinNames.push(item.name);
  } else if (diff === -1) {
    console.log("Item removed:", item);
    const index = joinNames.indexOf(item.name);
    if (index !== -1) {
      joinNames.splice(index, 1);
    }
  }
};

const submit = () => {
  console.log(JSON.stringify(userIDsByName(aliases, [channel.myname])));
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
  sendRequest('/ContentsPush/', fd);
  // location.href = '/calendar/' + currentDate.value + '/';
};
</script>

<template>
  <div class="modal">
    <div class="modal-content">
      <input type="datetime-local" v-model="calendar.timeStart" />
      <span> ~ </span>
      <input type="datetime-local" v-model="calendar.timeEnd" />
      <textarea type="text" placeholder="Todo" v-model="calendar.todo"></textarea>
      <SelectPeople v-if="groups"
        @update:selectedItems="handleSelectedItemsChange"
        :channel="channel"
        :aliases="aliases"
        :groups="groups"
        :placeholder="'参加ユーザー'"
        v-model="calendar.aliasNames"

        />
      <button @click="submit">投稿</button>
      <button><a :href="`/calendar/${currentDate}/`">閉じる</a></button>
    </div>
  </div>
</template>

<style scoped>
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

/*.title {
  width: 100%;
  margin-bottom: 10px;
}
*/
.modal-content textarea {
  width: 100%;
}

</style>
