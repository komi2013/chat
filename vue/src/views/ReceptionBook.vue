<script setup>
import { ref, onMounted } from 'vue';

import Advertisement from '@/components/Advertisement.vue';
import BookModal from '@/components/BookModal.vue';
import { userIDsByName } from '@/my/channelFunc';

const props = defineProps({
  id: String,
  date: String,
  menuID: Number
});

document.title = '受付予約'

const channelID = localStorage.getItem("channelID");
function tF(a, b = null){ return timeFormat(a, b) }
const hours = ref(Array.from({ length: 24 }, (_, i) => i));
const menuID = ref(props.menuID || null);
const today = props.date || timeFormat('YYYY-MM-DD');
const getNext30Days = () => {
  return Array.from({ length: 30 }, (_, i) => {
    const date = new Date(today);
    date.setDate(date.getDate() + i);
    return date;
  });
};
const monthDates = getNext30Days();

async function findReception() {
  const fd = new FormData();
  fd.append('receptionID', localStorage.getItem('channelID'));
  fd.append('channelID', localStorage.getItem('channelID'));
  fd.append('aliasName', channel.value.myname);
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ReceptionGet/', fd);
  if (!res.csrf) errorMessage.value = res
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
  return res.reception;
}

const generateSchedules = (menuID) => {
  if (!reception.value) return;
  schedules.value = [];
  const today = props.date || timeFormat("YYYY-MM-DD");
  const monthDates = Array.from({ length: 30 }, (_, i) => {
    const date = new Date(today);
    date.setDate(date.getDate() + i);
    return date.toISOString().split("T")[0];
  });
  let selectedService = null;
  let needSkill = null;
  let needFacility = null;
  if (menuID) {
    selectedService = reception.value.menus.find(service => service.menuID === menuID);
    if (!selectedService) return;

    needSkill = selectedService.needSkill;
    needFacility = selectedService.needFacility;
  }
  let tempSchedules = [];
  let unavailableTimes = [];
  monthDates.forEach(date => {
    reception.value.books.forEach(book => (book.splitted = false));
    let workStaffTime = false;
    reception.value.workStaffs.forEach(work => {
      if (!isSameDate(work.workStart, date)) return;
      let start = new Date(work.workStart);
      let end = new Date(work.workEnd);
      if (needSkill) {
        const validStaffs = getValidWorkStaffs(work, needSkill);
        // console.log('validStaffs', validStaffs);
        if (!validStaffs) return; // 該当するスタッフがいない場合はスキップ
      }

      let splitBlocks = splitWorkTime(start, end, reception.value.books);
      splitBlocks.forEach(({ start, end }) => {
        tempSchedules.push({
          timeStart: start,
          timeEnd: end,
        });
      });

      workStaffTime = true;
    });
    if (needFacility) {
      let newUnavailableTimes = makeUnavailableTime(date, needFacility);
      let uniqueTimes = new Set(unavailableTimes.map(time => JSON.stringify(time)));
      newUnavailableTimes.forEach(time => uniqueTimes.add(JSON.stringify(time)));
      unavailableTimes = Array.from(uniqueTimes).map(time => JSON.parse(time));
    }
  });

  let mergedSchedules = mergeSchedules(tempSchedules);
  tempSchedules = [];
  mergedSchedules.forEach(schedule => {
    let splitBlocks = splitWorkTime(schedule.timeStart, schedule.timeEnd, unavailableTimes);
    splitBlocks.forEach(({ start, end }) => {
      // console.log('splitBlocks.forEach', start, end);
      tempSchedules.push({
        timeStart: start,
        timeEnd: end,
      });
    });
  });
  mergedSchedules = mergeSchedules(tempSchedules);
  schedules.value = mergedSchedules;
};

const getValidWorkStaffs = (work, needSkill) => {
  return reception.value.staffSkills.some(staff => 
    staff.skills.includes(needSkill) && staff.aliasName === work.aliasName
  );
};

const needsFacility = (book, needFacility) => {
  const service = reception.value.menus.find(service => service.menuID === book.menuID);
  return service?.needFacility === needFacility;
};

const splitWorkTime = (workStart, workEnd, books) => {
  // console.log(`【splitWorkTime】処理開始: ${formatDateTime(workStart)} ~ ${formatDateTime(workEnd)}`);

  let blocks = [{ start: new Date(workStart), end: new Date(workEnd) }];
  // console.log('対象範囲:', workStart, '~', workEnd, '適用する books:', books);

  let remainingBooks = books
    .filter(book => isOverlapping(workStart, workEnd, book.bookStart, book.bookEnd))
    .sort((a, b) => new Date(a.bookStart) - new Date(b.bookStart));

  // console.log('対象の予約 (remainingBooks):', remainingBooks);

  remainingBooks.forEach(book => {
    let bookStart = new Date(book.bookStart);
    let bookEnd = new Date(book.bookEnd);

    // console.log(`  【予約判定】${formatDateTime(bookStart)} ~ ${formatDateTime(bookEnd)}`);

    let newBlocks = [];

    blocks.forEach(({ start, end }) => {
      if (bookStart <= end && bookEnd >= start) {
        if (bookStart > start) {
          // console.log(`  → 分割1: ${formatDateTime(start)} ~ ${formatDateTime(bookStart)}`);
          newBlocks.push({ start: new Date(start), end: new Date(bookStart) });
        }

        if (bookEnd < end) {
          // console.log(`  → 分割2: ${formatDateTime(bookEnd)} ~ ${formatDateTime(end)}`);
          newBlocks.push({ start: new Date(bookEnd), end: new Date(end) });
        }
      } else {
        newBlocks.push({ start, end });
      }
    });

    blocks = newBlocks.length > 0 ? newBlocks : blocks;
  });

  // console.log('【分割後のブロック】', blocks);
  return blocks;
};

const makeUnavailableTime = (date, needFacility) => {
  let unavailableTimes = [];

  if (!needFacility) return unavailableTimes;

  const relatedFacility = reception.value.facilities.find(facility => facility.facilityName === needFacility);
  if (!relatedFacility) return unavailableTimes;

  const facilityCount = relatedFacility.facilityCount;

  let facilityBooks = reception.value.books
    .filter(book => needsFacility(book, needFacility))
    .sort((a, b) => new Date(a.bookStart) - new Date(b.bookStart));

  let mergedBookings = [];
  let currentStart = null;
  let currentEnd = null;

  facilityBooks.forEach(book => {
    let bookStart = new Date(book.bookStart);
    let bookEnd = new Date(book.bookEnd);

    // **初回または前の予約と時間が空いている場合、新しいブロックを作成**
    if (!currentStart || bookStart > currentEnd) {
      if (currentStart) {
        mergedBookings.push({ bookStart: currentStart, bookEnd: currentEnd });
      }
      currentStart = bookStart;
      currentEnd = bookEnd;
    } else {
      // **時間が重なっている場合、終了時間を更新**
      currentEnd = new Date(Math.max(currentEnd.getTime(), bookEnd.getTime()));
    }
  });

  // **最後のブロックを追加**
  if (currentStart) {
    mergedBookings.push({ bookStart: currentStart, bookEnd: currentEnd });
  }

  // **フォーマットを統一**
  return mergedBookings.map(({ bookStart, bookEnd }) => ({
    bookStart: formatDateTime(bookStart),
    bookEnd: formatDateTime(bookEnd),
  }));
};

const mergeSchedules = (schedules) => {
  if (schedules.length === 0) return [];

  schedules = schedules.filter(schedule => 
    new Date(schedule.timeStart).getTime() !== new Date(schedule.timeEnd).getTime()
  );

  // **時間順にソート**
  schedules.sort((a, b) => new Date(a.timeStart) - new Date(b.timeStart));

  let mergedSchedules = [];

  // console.log('before schedules', JSON.stringify(schedules));

  for (let i = 0; i < schedules.length; i++) {
    let currentSchedule = Object.assign({}, schedules[i]); // **オブジェクトをコピー**
    
    if (mergedSchedules.length === 0) {
      mergedSchedules.push(currentSchedule);
      continue;
    }

    let lastSchedule = mergedSchedules[mergedSchedules.length - 1];

    // **時間帯が完全に一致するデータをマージしない**
    if (
      new Date(lastSchedule.timeStart).getTime() === new Date(currentSchedule.timeStart).getTime() &&
      new Date(lastSchedule.timeEnd).getTime() === new Date(currentSchedule.timeEnd).getTime()
    ) {
      continue;
    }

    // **時間が重なっている場合のみ `timeEnd` を正しく更新**
    if (new Date(lastSchedule.timeEnd) >= new Date(currentSchedule.timeStart)) {
      lastSchedule.timeEnd = new Date(Math.max(new Date(lastSchedule.timeEnd).getTime(), new Date(currentSchedule.timeEnd).getTime()));
    } else {
      mergedSchedules.push(currentSchedule);
    }
  }

  // console.log('after schedules', JSON.stringify(schedules));

  return mergedSchedules.map(schedule => ({
    timeStart: formatDateTime(schedule.timeStart),
    timeEnd: formatDateTime(schedule.timeEnd),
  }));
};

// **関数はそのまま**
const isSameDate = (workStart, date) => workStart.split("T")[0] === date;
const isOverlapping = (startA, endA, startB, endB) => {
  return !(new Date(startA) > new Date(endB) || new Date(endA) < new Date(startB));
};

const formatDateTime = (date) => {
  if (!date) return "1000-01-01T00:00:00"; // **デフォルト値**

  try {
    const d = date instanceof Date ? date : new Date(date);

    if (isNaN(d.getTime())) {
      // console.warn("Invalid date detected:", date);
      return "1000-01-01T00:00:00"; // **無効な日付もデフォルト値を返す**
    }

    return d.getFullYear() + "-" +
      String(d.getMonth() + 1).padStart(2, "0") + "-" +
      String(d.getDate()).padStart(2, "0") + "T" +
      String(d.getHours()).padStart(2, "0") + ":" +
      String(d.getMinutes()).padStart(2, "0");
  } catch (error) {
    console.error("formatDateTime error:", error, "date value:", date);
    return "1000-01-01T00:00:00"; // **エラー時もデフォルト値を返す**
  }
};


const getEventsForDayAndHour = (day, hour) => {
  const events = schedules.value.filter((event) => {
    const start = new Date(event.timeStart);
    const end = new Date(event.timeEnd);
    // if ( day.getDate() == 22 && hour == 16) {
    //   console.log('getEventsForDayAndHour', hour, start.getHours(), '<=', hour, '&&', end.getMinutes(), '>=', hour);
    //   console.log('start, end', start, end);
    // }
    return (
      start.getFullYear() === day.getFullYear() &&
      start.getMonth() === day.getMonth() &&
      start.getDate() === day.getDate() &&
      start.getHours() <= hour &&
      (end.getHours() * 60 + end.getMinutes() > hour * 60)
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
  // if (start.getDate() && hour == 16) {
  //   console.log('decideHeightTop', hour, height, start, end);
  // }
  if (end.getHours() > hour) {
    height = 60;
  } else {
    height = end.getMinutes();
  }
  const opacity = 1;
  const zindex = 2;
  const width = 118;
  const left = schedule.index * width;
  return {
    height: `${height -6}px`,
    top: `${top}px`,
    opacity: `${opacity}`,
    "z-index": `${zindex}`,
    "background-color": 'white',
    width: `${width}px`,
  };
};

const channel = ref(null);
const groups = ref([]);
const aliases = ref([]);
const reception = ref(null);
const availableSkills = ref([]);
const iamAdmin = ref(false);
const schedules = ref([]);
const errorMessage = ref('')
onMounted(async() => {
  channel.value = await getIDB('channel', channelID);
  groups.value = await getIDBs('group', 'channelIDIndex', channelID, 10000);
  aliases.value = await getIDBs('alias', 'channelIDIndex', channelID, 10000);
  if (props.id && props.id.length > 3) {
    reception.value = await findReception();
  } else {
    reception.value = await getIDB('reception', props.id);
  }

  if (reception.value.adminNames && reception.value.adminNames.includes(channel.value.myname)) {
    iamAdmin.value = true;
  }
  const matchedStaff = reception.value.staffSkills.find(
    (staff) => staff.aliasName === channel.value.myname
  );
  availableSkills.value = matchedStaff ? [...matchedStaff.skills] : [];
  generateSchedules();
  window.scrollTo({top: 600, behavior: "smooth"});
  document.title = reception.value.receptionTitle;
});

function jump(days) {
  const currentDate = new Date(today);
  currentDate.setDate(currentDate.getDate() + days);
  const formattedDate = currentDate.toISOString().split('T')[0];
  location.href = `/calendar/${formattedDate}/`;
}

const selectedServiceId = ref(null);
function handleServiceChange(menuID) {
  // console.log('selected', reception.value.menus);
  const selected = reception.value.menus.find(service => service.menuID === menuID);
  generateSchedules(selected.menuID);
}

const showModal = ref(false);
const selectedEvent = ref({
  timeStart: '',
  timeEnd: '',
});

const openModal = (day, hour) => {
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
      <tr><td style="text-align: center;"> <Advertisement /> </td></tr>
      <tr v-if="reception">
        <td>
          <select id="menu-select" class="menu" v-model="selectedServiceId" @change="handleServiceChange(selectedServiceId)">
            <option disabled value="">メニュー</option>
            <option v-for="service in reception.menus" :key="service.menuID" :value="service.menuID">
              {{ service.menuName }} - {{ service.price }}円
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
      <tr><td><a href="/sign/" > サインイン </a></td></tr>
      <tr><td style="text-align: center;"> <Advertisement /> </td></tr>
      <tr><td style="text-align: center;"> <Advertisement /> </td></tr>
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
            <span v-if="index % 3 === 0"
                  class="time-hour">
              {{ hour }}
            </span>
            <div v-for="d in getEventsForDayAndHour(day, hour)" 
               :key="'event-' + d.id" 
               class="event"
               :style="decideHeightTop(d, hour)"
               @click.stop="openModal(day, hour)"
               >
               {{ d.title }} <!-- maybe number of people is better? -->
            </div>
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
      :reception="reception"
      :menuID="selectedServiceId"
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
  z-index: 5;
  position: relative;
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
