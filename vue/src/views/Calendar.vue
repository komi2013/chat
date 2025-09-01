<script setup>
import { ref, onMounted, computed, watch } from 'vue';

import Advertisement from '@/components/Advertisement.vue';
import NoticePopup from '@/components/NoticePopup.vue';
import SelectPeople from '@/components/SelectPeople.vue';
import { useCalendarsStore } from '@/stores/calendars.js';
// import { useNoticesStore } from '@/stores/notices.js';

import { userIDsByName } from '@/my/channelFunc';

const props = defineProps({
  date: String,
});

const channelID = localStorage.getItem("channelID");
function tF(a, b = null){ return timeFormat(a, b) }
const hours = ref(Array.from({ length: 24 }, (_, i) => i));

const calendarsStore = useCalendarsStore();
const schedules = computed(() => calendarsStore.calendars);

const today = props.date || timeFormat('YYYY-MM-DD');
document.title = today
const getNext30Days = () => {
  return Array.from({ length: 30 }, (_, i) => {
    const date = new Date(today);
    date.setDate(date.getDate() + i);
    return date;
  });
};
const monthDates = getNext30Days();
const getEventsForDayAndHour = (day, hour) => {
  const events = schedules.value.filter((event) => {
    const start = new Date(event.timeStart);

    return (
      start.getFullYear() === day.getFullYear() &&
      start.getMonth() === day.getMonth() &&
      start.getDate() === day.getDate() &&
      start.getHours() === hour
    );
  });
  return events.map((event, index) => ({ ...event, index, total: events.length }));
};

const decideHeightTop = (schedule) => {
  const start = new Date(schedule.timeStart);
  const end = new Date(schedule.timeEnd);
  const minuteHeight = 1;
  const top = start.getMinutes() * minuteHeight;
  const height = (end - start) / (1000 * 60) * minuteHeight;
  console.log('height', height);
  const opacity = schedule.index === 0 ? 1 : 0.5;
  const zindex = (schedule.index + 1) * 4;
  const width = 120 / schedule.total;
  const left = schedule.index * width;
  const colors = ["rgba(255, 0, 0, 0.25)", "rgba(255, 255, 0, 0.25)", "rgba(128, 0, 128, 0.25)", "rgba(0, 0, 255, 0.25)"];
  const backgroundColor = colors[schedule.nameCount - 1] || colors[3];
  return {
    height: `${height}px`,
    top: `${top}px`,
    opacity: `${opacity}`,
    "z-index": `${zindex}`,
    "background-color": backgroundColor,
    width: `${width}px`,
    left: `${left}px`,
  };
};

const channel = ref(null);
const groups = ref([]);
const aliases = ref([]);
onMounted(async() => {
  channel.value = await getIDB('channel', channelID);
  groups.value = await getIDBs('group', 'channelIDIndex', channelID, 10000);
  aliases.value = await getIDBs('alias', 'channelIDIndex', channelID, 10000);
  const calendarData = await getAllIDBs('calendar');
  calendarData.forEach((d) => {
    d.title = d.todo ? Array.from(d.todo).slice(0, 10).join('') : ''; 
    calendarsStore.upsert(d);
  });

  document.documentElement.scrollTo({ top: 600, behavior: "smooth" });
  document.body.scrollTo({ top: 600, behavior: "smooth" });

});

const newSchedule = (day, hour) => {
  const start = new Date(day);
  start.setHours(hour, 0);
  const timeStart = timeFormat('YYYYMMDDThhmmss', start);
  window.open(`/calendarEdit/?dates=${timeStart}`);
  // location.href = `/calendarEdit/?dates=${timeStart}`;
};

const openCalendar = (calendarID) => {
  window.open(`/calendarEdit/${calendarID}/`);
};

const searchUsers = ref([])
const errorMessage = ref('')
watch(searchUsers, async (newNames, oldNames) => {
  console.log('検索ユーザーが変更されました:', newNames, oldNames);
  const oldSet = new Set(oldNames || []);
  const newSet = new Set(newNames || []);
  const addName = [...newSet].find(name => !oldSet.has(name));
  const subName = [...oldSet].find(name => !newSet.has(name));

    // fd.append('contents', JSON.stringify(calendar.value));
    // fd.append('targetStore', 'calendar');
    // console.log(today);
    // fd.append('param', JSON.stringify(param));

  if (addName) {
    const fd = new FormData()
    fd.append('userIDs', JSON.stringify(userIDsByName(aliases.value, [addName])))
    fd.append('channelID', channelID)
    fd.append('updatedBy', localStorage.getItem('myname'))
    const param = { date: today }
    const contents = ['calendar', param]
    fd.append('contents', JSON.stringify(contents))
    fd.append('pushTitle', 'storeSelect')
    fd.append('csrf', localStorage.getItem("csrf"))
    const res = await sendRequest('/ContentsJustPush/', fd)
    if (!res.csrf) errorMessage.value = res
    res.csrf && localStorage.setItem('csrf', res.csrf)
    res.pushContents.forEach(content => {
      pushReceive(content)
    });
  } else if (subName) {
    const delSchedules = schedules.value.filter(schedule => schedule.aliasName === subName);
    delSchedules.forEach((d) => {
      calendarsStore.delete(d.channelID)
    })
  }
})

function jump(days) {
  const currentDate = new Date(today);
  currentDate.setDate(currentDate.getDate() + days);
  const formattedDate = currentDate.toISOString().split('T')[0];
  location.href = `/calendar/${formattedDate}/`;
}

</script>

<template>
  <div id="drawer_column">
    <label for="drawer_check" class="pc_disp_none for_drawer"></label>
    <input id="drawer_check" type="checkbox" class="pulling pc_disp_none">
    <table id="drawer">
      <tr><td><a href="/" > 🏠 ホーム </a></td></tr>
      <tr><td style="text-align: center;"> <Advertisement /> </td></tr>
      <tr>
        <td>
          <SelectPeople v-if="aliases"
            :groups="groups"
            :aliases="aliases"
            :placeholder="'検索ユーザー'"
            v-model="searchUsers"
            />
        </td>
      </tr>
      <tr><td><a href="/sign/" > サインイン </a></td></tr>
    </table>
  </div>
  <div id="content">
    <div v-if="errorMessage"> 
      <div class="errorMessage">{{errorMessage}}<br>データ取得に失敗しました。</div>
      <a href="/setting/"> データ設定ページ </a><br>
      <a href="/sign/"> サインインページ </a>
    </div>
    <table id="calendar-container-move">
      <thead>
        <tr class="header">
          <th>≡</th>
          <th v-for="(day, index) in monthDates" :key="'day-header-' + index">
            <span>{{ tF('WWW', day) }}</span>
            <span>{{ tF('MM/DD', day) }}</span>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="hour in hours">
          <td class="back-column" @click="jump(-30)">
            <span v-if="hour % 6 == 5">
              ⬅                
            </span>
          </td>
          <td v-for="(day, index) in monthDates" :key="'day-' + index" 
              class="day-slot"
              :class="{'sunday': day.getDay() === 0}"
              @click="newSchedule(day, hour)">
            <div v-for="d in getEventsForDayAndHour(day, hour)" 
               :key="'event-' + d.id" 
               class="event"
               :style="decideHeightTop(d)"
               @click.stop="openCalendar(d.calendarID)"
               >
               {{ d.title }}
            </div>
            <span v-if="index % 3 === 0 && getEventsForDayAndHour(day, hour).length === 0"
                  class="time-hour">
              {{ hour }}
            </span>
          </td>
          <td class="back-column" @click="jump(30)">
            <span v-if="hour % 6 == 5">
              ➡               
            </span>
          </td>
        </tr>
      </tbody>
      <thead>
        <tr class="header">
          <th>≡</th>
          <th v-for="(day, index) in monthDates" :key="'day-header-' + index">
            <span>{{ tF('DD', day) }}</span>
            <span>{{ tF('WWW', day) }}</span>
          </th>
        </tr>
      </thead>
    </table>
  </div>
  <NoticePopup />
</template>

<style scoped>

.header {
  background-color: #f0f0f0;
  height: 30px;
}

.back-column {
  background-color: #fafafa;
  padding: 0 2px;
}

.day-slot {
  height: 57px;
  min-width: 120px;
  border: 1px solid #ccc;
  position: relative;
}

.event {
  position: absolute;
  top: 0px;
  border-radius: 4px;
  font-size: 0.6rem;
  word-break: break-word;
  background-color: rgba(0, 0, 255, 0.25);
  padding: 2px;
  z-index: 5
}

.time-hour {
  display: inline-flex;
  height: 100%;
  align-items: start;
  font-size: 0.5rem;
}

.sunday {
  background-color: rgba(255, 0, 0, 0.1);
}

#drawer td {
  background-color: #EEEEEE;
}

@media screen and (min-width : 701px) {
  #drawer {
    margin-top : -1px;
    background-color: white;
  }
}

@media screen and (max-width : 700px) {
  #drawer {
    width: 80%;
    overflow: scroll;
    position: absolute;
    z-index: 30;
    margin: 0;
    background-color: white;
    left: -100%;
    top : 33px;
    float: left;
  }
  .pulling {
    position: absolute;
    top: 0px;
    height: 59px;
    width: 50px;
    opacity: 0;
    z-index: 30;
  }
  .pulling:checked ~ #drawer{
    left: 0px;
  }
  .for_drawer {
    position: absolute;
    font-size: 40px;
    top: -10px;
    width: 50px;
    text-align: center;
  }
}

</style>
