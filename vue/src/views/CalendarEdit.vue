<script setup>
import { ref } from 'vue';
import SelectPeople from '../components/SelectPeople.vue';

const props = defineProps({
  id: String,
  start: String,
});

document.title = 'カレンダー';
const channel = ref(null);
const person = ref(null);
fetchChannel();
async function fetchChannel() {
  try {
    channel.value = await getIDB('channel', localStorage.channelID);
    if (channel.value.allAliases && Array.isArray(channel.value.allAliases)) {
      const aliasName = channel.value.aliasName;
      const aliases = channel.value.allAliases.find((alias) => alias[0] === aliasName);
      console.log(aliases);
      if (aliases) {
        person.value = {
          name: aliases[0],
          image: aliases[1] !== "null" ? aliases[1] : null,
        };
      }
    }
    if (props.id != '_') {
      fetchCalendar();
    }
  } catch (error) {
    console.error('channel error', error);
  }
}

const start = new Date(props.start || new Date());
const end = new Date(start.getTime() + 30 * 60 * 1000);
const currentDate = ref(timeFormat('YYYY-MM-DD', start));

const calendar = ref({
  timeStart: timeFormat('YYYY-MM-DDThh:mm', start),
  timeEnd: timeFormat('YYYY-MM-DDThh:mm', end),
  title: '',
  todo: ''
});

async function fetchCalendar() {
  try {
    calendar.value = await getIDB('calendar', props.id);
    currentDate.value = timeFormat('YYYY-MM-DD', calendar.value.timeStart);
  } catch (error) {
    console.error('calendar error', error);
  }
}

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
  if (!confirm("実行▶️")) {
    return;
  }
  let userIDs = [];
  for (const d of channel.value.allAliases) {
    if (channel.value.aliasName === d[0]) {
      userIDs.push(d[2]);
    }
    for (const dd of joinNames) {
      if (dd === d[0]) {
        userIDs.push(d[2]);
      }
    }
  }
  calendar.value.calendarID ||= generateRandomCode(8);
  calendar.value.channelID = channel.value.channelID;
  calendar.value.aliasName = channel.value.aliasName;
  const fd = new FormData();
  fd.append('userIDs', JSON.stringify([...new Set(userIDs)]));
  fd.append('channelID', localStorage.channelID);
  fd.append('aliasName', channel.value.aliasName);
  fd.append('contents', JSON.stringify(calendar.value));
  fd.append('pushTitle', 'calendar');
  const request = new Request('/ContentsPush/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      console.error(reason);
    })
  location.href = '/calendar/' + currentDate.value + '/';
};
</script>

<template>
  <div class="modal">
    <div class="modal-content">
      <input type="datetime-local" v-model="calendar.timeStart" />
      <span> ~ </span>
      <input type="datetime-local" v-model="calendar.timeEnd" />
      <input type="text" placeholder="タイトル" v-model="calendar.title" class="title" />
      <textarea type="text" placeholder="Todo" v-model="calendar.todo"></textarea>
      <SelectPeople v-if="channel && person"
        @update:selectedItems="handleSelectedItemsChange"
        :channel="channel"
        :person="person"
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

.title {
  width: 100%;
  margin-bottom: 10px;
}

.modal-content textarea {
  width: 100%;
}

</style>
