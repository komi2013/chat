<script setup>
import { ref, onMounted } from 'vue';

import BookModal from '@/components/BookModal.vue';
import { userIDsByName } from '@/my/channelFunc';

const props = defineProps({
  id: String,
  date: String,
  serviceID: Number
});

const channelID = localStorage.getItem("channelID");
function tF(a, b = null){ return timeFormat(a, b) }
const hours = ref(Array.from({ length: 24 }, (_, i) => i));
const serviceID = ref(props.serviceID || null);
// const calendarsStore = useCalendarsStore();
// const schedules = computed(() => calendarsStore.calendars);

// const schedules = computed(() => {
//   console.log("schedules recalculated!", calendarsStore.calendars);
//   return calendarsStore.calendars;
// });

const today = props.date || timeFormat('YYYY-MM-DD');
const getNext30Days = () => {
  return Array.from({ length: 30 }, (_, i) => {
    const date = new Date(today);
    date.setDate(date.getDate() + i);
    return date;
  });
};
const monthDates = getNext30Days();


async function findBookPattern() {
  const fd = new FormData();
  fd.append('bookPatternID', props.id);
  fd.append('channelID', localStorage.getItem('channelID'));
  fd.append('aliasName', channel.value.myname);
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/BookPatternGet/', fd);
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
  return res.bookPattern;
}

const generateSchedules = () => {
  if (!bookPattern.value) return;
  schedules.value = []; // 既存のデータをリセット

  const today = props.date || timeFormat("YYYY-MM-DD");
  const monthDates = Array.from({ length: 30 }, (_, i) => {
    const date = new Date(today);
    date.setDate(date.getDate() + i);
    return date.toISOString().split("T")[0]; // "YYYY-MM-DD" の形式
  });

  let tempSchedules = []; // 各 `workStaffs` の処理結果を一時保存

  monthDates.forEach(date => {
    // `books` のフラグをリセット（毎回 `workStaffs` のループごとにリセットしない）
    bookPattern.value.books.forEach(book => (book.splitted = false));

    bookPattern.value.workStaffs.forEach(work => {
      if (!isSameDate(work.workStart, date)) return;

      let start = new Date(work.workStart);
      let end = new Date(work.workEnd);

      console.log(`【workStaffsループ】${formatDateTime(start)} ~ ${formatDateTime(end)}`);

      let splitBlocks = splitWorkTime(start, end, bookPattern.value.books);

      splitBlocks.forEach(({ start, end }) => {
        tempSchedules.push({
          timeStart: start,
          timeEnd: end,
        });
      });
    });
  });

  // **最終的に分割された `schedule` を統合**
  console.log('tempSchedules', tempSchedules);
  schedules.value = mergeSchedules(tempSchedules);

  console.log("【最終 schedules】", JSON.stringify(schedules.value, null, 2));
};

const splitWorkTime = (workStart, workEnd, books) => {
  console.log(`【splitWorkTime】処理開始: ${formatDateTime(workStart)} ~ ${formatDateTime(workEnd)}`);

  let blocks = [{ start: new Date(workStart), end: new Date(workEnd) }];
  let remainingBooks = books
    .filter(book => isOverlapping(workStart, workEnd, book.bookStart, book.bookEnd) && !book.splitted)
    .sort((a, b) => new Date(a.bookStart) - new Date(b.bookStart));

  remainingBooks.forEach(book => {
    let bookStart = new Date(book.bookStart);
    let bookEnd = new Date(book.bookEnd);

    console.log(`  【予約判定】${formatDateTime(bookStart)} ~ ${formatDateTime(bookEnd)}`);

    let newBlocks = [];

    blocks.forEach(({ start, end }) => {
      console.log('start:', formatDateTime(start), 'end:', formatDateTime(end));
      if (bookStart > start && bookStart < end) {
        console.log(`  → 分割1: ${formatDateTime(start)} ~ ${formatDateTime(bookStart)}`);
        newBlocks.push({ start: new Date(start), end: new Date(bookStart) });
        book.splitted = true;
      }

      console.log(formatDateTime(bookEnd), '>', formatDateTime(start), '&&', formatDateTime(bookEnd), '<', formatDateTime(end));
      // 17:30 ~ 18:30 book

      if (bookStart > start && bookEnd < end) {
        console.log(`  → 分割2: ${formatDateTime(bookEnd)} ~ ${formatDateTime(end)}`);
        newBlocks.push({ start: new Date(bookEnd), end: new Date(end) });
      }
    });

    blocks = newBlocks.length > 0 ? newBlocks : blocks;
  });

  return blocks;
};

const mergeSchedules = (schedules) => {
  if (schedules.length === 0) return [];

  schedules.sort((a, b) => new Date(a.timeStart) - new Date(b.timeStart));

  let mergedSchedules = [schedules[0]];

  for (let i = 1; i < schedules.length; i++) {
    let lastSchedule = mergedSchedules[mergedSchedules.length - 1];
    let currentSchedule = schedules[i];

    if (new Date(lastSchedule.timeEnd) >= new Date(currentSchedule.timeStart)) {
      lastSchedule.timeEnd = new Date(Math.max(new Date(lastSchedule.timeEnd).getTime(), new Date(currentSchedule.timeEnd).getTime()));
    } else {
      mergedSchedules.push(currentSchedule);
    }
  }

  return mergedSchedules.map(schedule => ({
    timeStart: formatDateTime(schedule.timeStart),
    timeEnd: formatDateTime(schedule.timeEnd),
  }));
};

// **関数はそのまま**
const isSameDate = (workStart, date) => workStart.split("T")[0] === date;
const isOverlapping = (startA, endA, startB, endB) => !(new Date(startA) >= new Date(endB) || new Date(endA) <= new Date(startB));
const formatDateTime = (date) => date.getFullYear() + "-" + String(date.getMonth() + 1).padStart(2, "0") + "-" + String(date.getDate()).padStart(2, "0") + "T" + String(date.getHours()).padStart(2, "0") + ":" + String(date.getMinutes()).padStart(2, "0");


const getEventsForDayAndHour = (day, hour) => {
  const events = schedules.value.filter((event) => {
    const start = new Date(event.timeStart);
    const end = new Date(event.timeEnd);
    if ( hour == 17 && day.getDate() == 18) {
      console.log('getEventsForDayAndHour', hour, start.getHours(), '<=', hour, '&&', end.getHours(), '>=', hour);
    }
    return (
      start.getFullYear() === day.getFullYear() &&
      start.getMonth() === day.getMonth() &&
      start.getDate() === day.getDate() &&
      start.getHours() <= hour &&
      end.getHours() >= hour
    );
  });
  return events.map((event, index) => ({ ...event, index, total: events.length }));
};


const decideHeightTop = (schedule, hour) => {
  const start = new Date(schedule.timeStart);
  const end = new Date(schedule.timeEnd);
  const minuteHeight = 1;
  let top = start.getMinutes() * minuteHeight;
  if (hour > start.getHours()) {
    top = 0;
  }
  let height;
  // let height = (end - start) / (1000 * 60) * minuteHeight;
  // if ( (hour == 17 || hour == 18 || hour == 19) && start.getDate()) {
  //   console.log('decideHeightTop', hour, height, start, end);
  // }
  if (end.getHours() > hour) {
    height = 60;
  } else {
    height = end.getMinutes();
  }
  // height = height - top;
  // const opacity = schedule.index === 0 ? 1 : 0.5;

  const opacity = 1;
  const zindex = 2;
  const width = 118;
  const left = schedule.index * width;
  // const colors = ["rgba(255, 0, 0, 0.25)", "rgba(255, 255, 0, 0.25)", "rgba(128, 0, 128, 0.25)", "rgba(0, 0, 255, 0.25)"];
  // const backgroundColor = colors[schedule.nameCount - 1] || colors[3];
  return {
    height: `${height -6}px`,
    top: `${top}px`,
    opacity: `${opacity}`,
    "z-index": `${zindex}`,
    "background-color": 'white',
    width: `${width}px`,
    // left: `${left}px`,
  };
};

const channel = ref(null);
const groups = ref([]);
const aliases = ref([]);
const bookPattern = ref(null);
const availableSkills = ref([]);
const iamAdmin = ref(false);
const schedules = ref([]);
onMounted(async() => {
  channel.value = await getIDB('channel', channelID);
  groups.value = await getIDBs('group', 'channelIDIndex', channelID, 10000);
  aliases.value = await getIDBs('alias', 'channelIDIndex', channelID, 10000);
  if (props.id && props.id.length > 5) {
    bookPattern.value = await findBookPattern();
  } else {
    bookPattern.value = await getIDB('bookPattern', props.id);
  }
  const matchingGroup = groups.value.find(group => group.groupName === bookPattern.value.adminGroup);
  if (matchingGroup && matchingGroup.aliasNames.includes(channel.value.myname)) {
    iamAdmin.value = true;
  }
  const matchedStaff = bookPattern.value.staffSkills.find(
    (staff) => staff.aliasName === channel.value.myname
  );
  availableSkills.value = matchedStaff ? [...matchedStaff.skills] : [];
  generateSchedules();
  console.log(schedules.value);
  window.scrollTo({top: 600, behavior: "smooth"});
  document.title = bookPattern.value.bookTitle;
});

function jump(days) {
  const currentDate = new Date(today);
  currentDate.setDate(currentDate.getDate() + days);
  const formattedDate = currentDate.toISOString().split('T')[0];
  location.href = `/calendar/${formattedDate}/`;
}

const selectedServiceId = ref(null);
function handleServiceChange(serviceId) {
  const selected = bookPattern.value.services.find(service => service.id === serviceId);
  console.log('selected', selected);
  generateSchedules(true);
  // generateOpenTimes(selected);
}

const showModal = ref(false);
const selectedEvent = ref({
  timeStart: '',
  timeEnd: '',
});

const openModal = (day, hour) => {
  console.log('hour', hour);
  showModal.value = true;
  const start = new Date(day);
  start.setHours(hour, 0);
  const end = new Date(start);
  end.setMinutes(start.getMinutes() + 60);

  selectedEvent.value.timeStart = timeFormat('YYYY-MM-DDThh:mm', start);
  selectedEvent.value.timeEnd = timeFormat('YYYY-MM-DDThh:mm', end);
};

const closeModal = () => {
  showModal.value = false;
};


</script>

<template>
  <div id="drawer_column">
    <label for="drawer_check" class="pc_disp_none for_drawer"></label>
    <input id="drawer_check" type="checkbox" class="pulling pc_disp_none">
    <table id="drawer">
      <tr><td><a href="/" > 🏠 ホーム </a></td></tr>
      <tr v-if="bookPattern">
        <td>
          <select id="menu-select" class="menu" v-model="selectedServiceId" @change="handleServiceChange(selectedServiceId)">
            <option disabled value="">メニュー</option>
            <option v-for="service in bookPattern.services" :key="service.id" :value="service.id">
              {{ service.serviceName }} - {{ service.price }}円
            </option>
          </select>
        </td>
      </tr>
      <tr>
        <td>
          <button @click="submitShift">提出</button>
          <!-- <button @click="submitShift(1)">確定</button> -->
        </td>
      </tr>
      <tr><td><a href="/sign/" > 🔒 ログイン </a></td></tr>
    </table>
  </div>
  <div id="content">
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
               :style="decideHeightTop(d, hour)"
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
          <th></th>
          <th v-for="(day, index) in monthDates" :key="'day-header-' + index">
            <span>{{ tF('DD', day) }}</span>
            <span>{{ tF('WWW', day) }}</span>
          </th>
        </tr>
      </thead>
    </table>
    <BookModal
      v-if="showModal"
      :time="selectedEvent"
      :bookPattern="bookPattern"
      :serviceID="selectedServiceId"
      :myname="channel.myname"
      @close="closeModal"
      @submit="submitEvent"
    />
  </div>
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
  background-color: gray;
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

.event button {
  padding: 4px;
  margin: 4px;
  border-radius: 4px;
}

.time-hour {
  display: inline-flex;
  height: 100%;
  align-items: start;
  font-size: 0.5rem;
}

/*.sunday {
  background-color: rgba(255, 0, 0, 0.1);
}

*/
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
