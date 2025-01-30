<script setup>
import { ref, onMounted, computed, nextTick } from 'vue';

import NoticePopup from '../components/NoticePopup.vue';
import SelectPeople from '../components/SelectPeople.vue';
import { useCalendarsStore } from '../stores/calendars.js';

const props = defineProps({
  date: String, // 文字列形式の日付（例: '2025-02-01'）
});
const channelID = localStorage.getItem("channelID");

const hours = ref(Array.from({ length: 24 }, (_, i) => i));
const calendarsStore = useCalendarsStore();

const schedules = computed(() => {
  return calendarsStore.calendars;
});

const getStartOfWeek = (date) => {
  const start = new Date(date);
  const day = start.getDay();
  const diff = (day === 0 ? 6 : day - 1);
  start.setDate(start.getDate() - diff);
  start.setHours(0, 0, 0, 0);
  return start;
};

const today = props.date || timeFormat('YYYY-MM-DD');
function calculateDate(days) {
  const date = new Date(today);
  date.setDate(date.getDate() + days);
  return date.toISOString().split('T')[0]; // 'YYYY-MM-DD'
}

const nextWeek = calculateDate(7);
const preWeek = calculateDate(-7);

const weekDates = getWeekDates();
function getWeekDates() {
  const startOfWeek = getStartOfWeek(new Date(today));
  return Array.from({ length: 7 }, (_, i) => {
    const newDate = new Date(startOfWeek);
    newDate.setDate(startOfWeek.getDate() + i);
    return newDate;
  });
}

const getEventsForDayAndHour = (day, hour) => {
  return schedules.value.filter((event) => {
    const start = new Date(event.timeStart);
    const end = new Date(event.timeEnd);
    return (
      start.getFullYear() === day.getFullYear() &&
      start.getMonth() === day.getMonth() &&
      start.getDate() === day.getDate() &&
      start.getHours() <= hour &&
      end.getHours() >= hour
    );
  });
};

const formatDate = (date) => {
  return timeFormat('DD', date);
};

const formatDay = (date) => {
  const daysOfWeek = ['日', '月', '火', '水', '木', '金', '土'];
  return daysOfWeek[date.getDay()];
};

const decideHeightTop = (schedule, index) => {
  const start = new Date(schedule.timeStart);
  const end = new Date(schedule.timeEnd);
  const startMinutes = start.getHours() * 60 + start.getMinutes();
  const endMinutes = end.getHours() * 60 + end.getMinutes();
  const minuteHeight = 1;
  const top = start.getMinutes() * minuteHeight;
  const height = (endMinutes - startMinutes) * minuteHeight;
  const opacity = (index === 1) ? 1 : Math.random() * 0.8 + 0.1;
  const zindex = opacity * 10;
  let backgroundColor;
  switch (schedule.nameCount) {
    case 1:
      backgroundColor = "rgba(255, 0, 0, 0.25)"; // 赤の25%
      break;
    case 2:
      backgroundColor = "rgba(255, 255, 0, 0.25)"; // 黄の25%
      break;
    case 3:
      backgroundColor = "rgba(128, 0, 128, 0.25)"; // 紫の25%
      break;
    default:
      backgroundColor = "rgba(0, 0, 255, 0.25)"; // 青の25%
  }
  return {
    height: `${height}px`,
    top: `${top}px`,
    opacity: `${opacity}`,
    "z-index": `${zindex}`,
    "background-color": `${backgroundColor}`,
  };
};

let channel;
let aliases;
let groups;
onMounted(async () => {
  channel = await getIDB('channel', channelID);
  aliases = await getIDBs('alias', 'channelIDIndex', channelID, 10000);
  groups = await getIDBs('group', 'channelIDIndex', channelID, 10000);
  const calendarData = await getAllIDBs('calendar');
  console.log(calendarData);
  calendarData.forEach((d) => {
    d.title = d.todo ? Array.from(d.todo).slice(0, 10).join('') : ''; 
    calendarsStore.upsert(d);
  });
  const container = document.getElementById('calendar-container-move');
  if (container) {
    setupPCEvents(container);
    setupTouchEvents(container);
    container.scrollTo({ top: 600 });
  }
});

let startX = 0;
function setupPCEvents(container) {
  container.addEventListener('dragstart', (event) => {
    startX = event.clientX;
  });

  container.addEventListener('dragend', (event) => {
    const endX = event.clientX;
    handleDragOrSwipe(endX);
  });
}

function setupTouchEvents(container) {
  container.addEventListener('touchstart', (event) => {
    startX = event.touches[0].clientX;
  });

  container.addEventListener('touchend', (event) => {
    const endX = event.changedTouches[0].clientX;
    handleDragOrSwipe(endX);
  });
}

function handleDragOrSwipe(endX) {
  const deltaX = endX - startX;
  if (Math.abs(deltaX) > 50) {
    if (deltaX > 0) {
      location.href = '/calendar/' + preWeek + '/';
    } else {
      location.href = '/calendar/' + nextWeek + '/';
    }
  }
}

const newSchedule = (day, hour) => {
  const start = new Date(day);
  start.setHours(hour, 0);
  const end = new Date(start);
  end.setMinutes(start.getMinutes() + 30);
  const timeStart = timeFormat('YYYYMMDDThhmmss', start);
  location.href = `/calendarEdit/?dates=${timeStart}/`;
};

const handleSelectedItemsChange = (change) => {
  const { diff, item } = change;
  if (diff === 1) {
    console.log("Item added:", item);
    let userIDs = [];
    // for (const d of channel.allAliases) {
    //   if (d[0] === item.name) {
    //     userIDs.push(d[2]);
    //   }
    // }
    const fd = new FormData();
    fd.append('userIDs', JSON.stringify([...new Set(userIDs)]));
    fd.append('channelID', localStorage.channelID);
    fd.append('aliasName', channel.aliasName);
    fd.append('targetStore', 'calendar');
    console.log(today);
    const param = {
      date: today
    }
    fd.append('param', JSON.stringify(param));
    // fd.append('pushTitle', 'pushSelect');
    // const request = new Request('/StoreSelect/', {
    //   method: 'POST',
    //   body: fd,
    // });
    // fetch(request)
    //   .catch((reason)=>{
    //     console.error(reason);
    //   })
    sendRequest('/StoreSelect/', fd);
  } else if (diff === -1) {
    console.log("Item removed:", item);
  }
};

</script>

<template>
<div id="drawer_column">
  <label for="drawer_check" class="pc_disp_none for_drawer">≡</label>
  <input id="drawer_check" type="checkbox" class="pulling pc_disp_none">
  <table id="drawer">
    <tr><td><a href="/" > 🏠 ホーム </a></td></tr>
    <tr>
      <td>
        <SelectPeople v-if="channel"
          @update:selectedItems="handleSelectedItemsChange"
          :channel="channel"
          />
      </td>
    </tr>
    <tr><td><a href="/sign/" > 🔒 ログイン </a></td></tr>
  </table>
</div>
<div id="content">
  <div class="calendar-container" id="calendar-container-move" draggable="true">
    <div class="header">
      <div class="time-slot" style="align-items: center;">≡</div>
      <div v-for="(day, index) in weekDates" :key="index" class="day-header">
        <div>{{ formatDate(day) }}</div>
        <div>{{ formatDay(day) }}</div>
      </div>
    </div>
    <div class="calendar-grid">
      <div>
        <div v-for="hour in hours" :key="hour" class="time-slot">
          <span>{{ hour }}</span>
        </div>
      </div>
      <div class="day-column" v-for="(day, index) in weekDates" :key="index">
        <div v-for="hour in hours" :key="hour" class="day-slot" @click="newSchedule(day, hour)">
          <div
            v-for="d, in getEventsForDayAndHour(day, hour)"
            :key="d.id"
            :style="decideHeightTop(d, getEventsForDayAndHour(day, hour).length)"
            class="event"
          >
            <template v-if="d.nameCount">
              xxx
            </template>
            <template v-if="!d.nameCount">
              <a :href="`/calendarEdit/${d.calendarID}/`"> {{ d.title }} </a>
            </template>
          </div>
        </div>
      </div>
    </div>
  </div>
  <NoticePopup />
</div>
</template>

<style scoped>
.calendar-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  max-width: 1000px;
  margin: 0 auto;
  box-sizing: border-box;
  position: relative;
  overflow-x: auto; /* 水平スクロールを許可 */
  overflow-y: visible; /* 垂直方向のスクロールを許可 */
  height: 100vh; /* ビューポート全体の高さ */
}

.header {
  display: grid;
  grid-template-columns: 20px repeat(7, 1fr); /* 左の時間軸の幅 */
  background-color: #f0f0f0;
  border-bottom: 1px solid #ccc;
  position: sticky; /* 上部に固定 */
  top: 0;
  z-index: 5;
}

.calendar-grid {
  display: grid;
  grid-template-columns: 20px repeat(7, 1fr);
  grid-auto-rows: 60px; /* 各行の高さ */
}

.time-slot {
  height: 59px;
  display: flex;
  justify-content: center;
  /*align-items: center;*/
  font-size: 10px;
  border-bottom: 1px solid #f0f0f0;
}

.time-slot span {
  padding-top: 2px;
}

.day-header {
  text-align: center;
  border-left: 1px solid #ccc;
  padding: 5px 0;
}


.day-column {
  /*border-left: 1px solid #ccc;*/
  position: relative;
}

.day-slot {
  border-bottom: 1px solid #ccc;
  border-left: 1px solid #ccc;
  position: relative;
  height: 59px;
}

.event {
  position: absolute;
  border-radius: 4px;
  font-size: 0.8rem;
  word-break: break-word;
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
    top : 63px;
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


ul {
  list-style-type: none;
  padding: 0;
}

li {
  cursor: pointer;
  display: flex;
  align-items: center;
  margin-bottom: 5px;
}

.selected-item {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

img {
  margin-right: 10px;
}

</style>
