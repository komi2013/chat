<script setup>
import { computed, ref, onMounted } from 'vue';
import BookModal from '../components/BookModal.vue';

// book data

const props = defineProps({
  id: '',
})

const windowID = ref(props.id);

const channel = ref('');
async function fetchChannel() {
  try {
    const data = await getIDB('channel', localStorage.channelID);
    channel.value = data;
  } catch (error) {
    console.log('error', error);
    channel.value = null;
  }
}
let bookPattern = null;
async function findBookPattern() {
  const fd = new FormData();
  fd.append('windowID', props.id);
  const data = await sendRequest('/WindowGet/', fd);
  bookPattern = data;
  const hasNeedRoles = data.times.some(time => time.needRoles && time.needRoles.length > 0);
  if (hasNeedRoles) {
    findShiftStaff();
  }
}
let shiftStaff = null;
async function findShiftStaff() {
  const fd = new FormData();
  fd.append('windowID', props.id);
  const data = await sendRequest('/ShiftStaffGet/', fd);
  shiftStaff = data;
  generateOpenTimes();
  console.log('/ShiftStaffGet/', data);
}
let book = null;
async function findBook() {
  const fd = new FormData();
  fd.append('windowID', props.id);
  const data = await sendRequest('/BookGet/', fd);
  book = data;
  console.log('/BookGet/', data);
}

// pseudo data
  // const bookPattern = {
  //   adminGroup: "2kaime",
  //   joinNames: ["sei1", "asd"],
  //   needFacilities: [[2, "perm"]],
  //   maxFacility: "30",
  //   times: [{
  //     date: "2024-11-13",
  //     bookTitle: "サロンの公開用予約リンク",
  //     needFacilities: [[1, "perm"]],
  //     needRoles: [[3, "stylist"]],
  //     limitStart: "10:00",
  //     limitEnd: "20:00",
  //     askChoices: [["性別は？", "男", "女", "その他"], ["何歳ですか？", "~ 15", "16 ~ 18", "19 ~ 25", "26 ~ 35", "35 ~"]]
  //   }],
  //   menu: [{
  //     name: "パーマ",
  //     price: 10000,
  //     needRole: "perm",
  //     needFacility: "perm"
  //   }, {
  //     name: "カット",
  //     price: 3000
  //   }]
  // };
  // const book = [{
  //   bookStart: "2024/11/13 12:00:00",
  //   bookEnd: "2024/11/13 13:00:00",
  //   answers: ["小松", "35", "2"],
  //   menuID: 2,
  //   useRole: "perm",
  //   useFacility: "perm"
  // }];

  // const shiftStaff = [{
  //   shiftStaffID: "ncsW202411031500sei1",
  //   aliasName: "sei1",
  //   shiftStart: "2024-11-13T15:00",
  //   shiftEnd: "2024-11-13T20:00",
  //   roles: ["perm", "cut"],
  //   seq: 1
  // }, {
  //   shiftStaffID: "ncsW202411011000sei1",
  //   aliasName: "sei1",
  //   shiftStart: "2024-11-11T10:00",
  //   shiftEnd: "2024-11-11T15:00",
  //   roles: ["perm", "cut"],
  //   seq: 1
  // }];

// openTimesの作成
const openTimes = ref([]);
function generateOpenTimes() {
  const limitStart = bookPattern.times[0].limitStart;
  const limitEnd = bookPattern.times[0].limitEnd;

  shiftStaff.forEach(staff => {
    const staffStart = new Date(staff.shiftStart);
    const staffEnd = new Date(staff.shiftEnd);

    // Shift内の時間を分割して1時間単位で作成
    let currentTime = new Date(staff.shiftStart);
    while (currentTime < staffEnd) {
      const nextTime = new Date(currentTime);
      nextTime.setHours(currentTime.getHours() + 1);

      // 時間内チェック
      const startTimeStr = `${staff.shiftStart.slice(0, 10)}T${limitStart}`;
      const endTimeStr = `${staff.shiftStart.slice(0, 10)}T${limitEnd}`;
      const timeStart = new Date(startTimeStr);
      const timeEnd = new Date(endTimeStr);

      // 条件を満たしているかの確認
      const roleMatch = book.every(b => !b.useRole || staff.roles.includes(b.useRole));
      const facilityMatch = book.every(b => !b.useFacility || 
        bookPattern.needFacilities.some(([count, facility]) => facility === b.useFacility && count >= book.length)
      );

      if (roleMatch && facilityMatch && currentTime >= timeStart && currentTime < timeEnd) {
        openTimes.value.push({
          timeStart: currentTime.toISOString().slice(0, 16),
          timeEnd: nextTime.toISOString().slice(0, 16),
          answers: book[0].answers,
          menuID: book[0].menuID
        });
      }

      currentTime = nextTime;
    }
  });
}

onMounted(() => {
  fetchChannel();
  findBookPattern();
  findBook();
  console.log(openTimes.value);
});

const hours = ref(Array.from({ length: 24 }, (_, i) => i));

const getStartOfWeek = (date) => {
  const start = new Date(date);
  const day = start.getDay();
  const diff = (day === 0 ? 6 : day - 1);
  start.setDate(start.getDate() - diff);
  start.setHours(0, 0, 0, 0);
  return start;
};

const weekDates = computed(() => {
  const startOfWeek = getStartOfWeek(new Date());
  return Array.from({ length: 7 }, (_, i) => {
    const newDate = new Date(startOfWeek);
    newDate.setDate(startOfWeek.getDate() + i);
    return newDate;
  });
});

const getEventsForDayAndHour = (day, hour) => {
  return openTimes.value.filter((d) => {
    const start = new Date(d.timeStart);
    const end = new Date(d.timeEnd);
    return (
      start.getFullYear() === day.getFullYear() &&
      start.getMonth() === day.getMonth() &&
      start.getDate() === day.getDate() &&
      start.getHours() <= hour &&
      start.getHours() >= hour
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

const decideHeightTop = (schedule) => {
  const start = new Date(schedule.timeStart);
  const end = new Date(schedule.timeEnd);
  const startMinutes = start.getHours() * 60 + start.getMinutes();
  const endMinutes = end.getHours() * 60 + end.getMinutes();
  const minuteHeight = 1;
  const top = start.getMinutes() * minuteHeight;
  const height = (endMinutes - startMinutes) * minuteHeight;
  return {
    height: `${height}px`,
    top: `${top}px`,
  };
};

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

const openModalForEdit = (event) => {
  showModal.value = true;
  selectedEvent.value = { ...event };
};

const closeModal = () => {
  showModal.value = false;
};

</script>

<template>
  <div class="calendar-container">
    <div class="header">
      <div class="time-slot"></div>
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
        <div v-for="hour in hours" :key="hour" class="day-slot">
          <div
            v-for="d in getEventsForDayAndHour(day, hour)"
            :class="['event']"
            :style="decideHeightTop(d)"
            @click="openModal(day, hour)"
          >
            &nbsp;&nbsp;&nbsp;&nbsp;
          </div>
        </div>
      </div>
    </div>
    <!-- モーダルを表示 -->
    <BookModal
      v-if="showModal"
      :time="selectedEvent"
      :bookPattern="bookPattern"
      :windowID="windowID"
      @close="closeModal"
      @submit="submitEvent"
    />
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
  position: relative; /* 親を基準に要素の配置 */
}

.header {
  display: grid;
  grid-template-columns: 20px repeat(7, 1fr); /* 左の時間軸の幅を20pxに設定 */
  background-color: #f0f0f0;
  border-bottom: 1px solid #ccc;
  position: fixed; /* ヘッダーを固定 */
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
  width: 100%;
}

.time-slot {
  height: 59px;
  display: flex;
  justify-content: center;
  /*align-items: center;*/
  font-size: 10px;
  border-bottom: 1px solid #ccc;
  background-color: silver;
}

.day-header {
  text-align: center;
  border-left: 1px solid #ccc;
  padding: 5px 0;
}

.calendar-grid {
  padding-top: 60px;
  display: grid;
  grid-template-columns: 20px repeat(7, 1fr); /* 左の時間軸の幅を20pxに設定 */
  grid-auto-rows: 59px;
}

.day-column {
  position: relative;
}

.day-slot {
  border-bottom: 1px solid #ccc;
  position: relative;
  height: 59px;
  background-color: silver;
}

.event {
  position: absolute;
  border-radius: 4px;
  color: #fff;
  font-size: 0.8rem;
  word-break: break-word;
  background-color: white;
  z-index: 1;
  width: 60px;
}

</style>
