<script setup>
import { computed, ref, onMounted } from 'vue';
import BookModal from '../components/BookModal.vue';

// book data

const props = defineProps({
  id: '',
})

const bookPatternID = ref(props.id);

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
const services = ref('');
async function findBookPattern() {
  const fd = new FormData();
  fd.append('bookPatternID', props.id);
  const data = await sendRequest('/BookPatternGet/', fd);
  bookPattern = data;
  services.value = data.services;
  const hasNeedSkills = data.services.some(m => m.needSkill && m.needSkill.length > 0);
  if (hasNeedSkills) {
    generateOpenTimes();
    
    console.log(openTimes.value);
  }
}

// openTimesの作成
const openTimes = ref([]);
function generateOpenTimes(service = null) {
  const limitStart = bookPattern.times[0].limitStart;
  const limitEnd = bookPattern.times[0].limitEnd;
  const intervalMinutes = service?.spendMinute || 60;
  if (service) {
    openTimes.value = [];
  }
  bookPattern.times.forEach(slot => {
    let facilities = [...bookPattern.facilities];
    let staffOpenTimes = [];
    if (slot.shiftStaffs) {
      slot.shiftStaffs.forEach(staff => {
        const staffStart = new Date(`${slot.date}T${staff.shiftStart}`);
        const staffEnd = new Date(`${slot.date}T${staff.shiftEnd}`);
        let currentTime = new Date(staffStart);
        while (currentTime < staffEnd) {
          const nextTime = new Date(currentTime);
          nextTime.setMinutes(currentTime.getMinutes() + intervalMinutes);
          const existingTime = staffOpenTimes.find(
            time => time.timeStart.getTime() === currentTime.getTime()
          );
          if (!service || !service.needSkill || (staff.skills && staff.skills.includes(service.needSkill)) ) {
            if (existingTime) {
              existingTime.count += 1;
            } else {
              staffOpenTimes.push({
                timeStart: new Date(currentTime),
                timeEnd: new Date(nextTime),
                count: 1
              });
            }
          }
          currentTime = nextTime;
        }
      });
    }
    if (slot.books) {
      slot.books.forEach(booked => {
        const bookStart = new Date(`${slot.date}T${booked.bookStart}`);
        const bookEnd = new Date(`${slot.date}T${booked.bookEnd}`);
        const bookedService = bookPattern.services.find(service => service.id === booked.serviceID);
        staffOpenTimes = staffOpenTimes.filter(openTime => {
          const overlaps = openTime.timeStart < bookEnd && openTime.timeEnd > bookStart;
          const needsFacility = service && service.needFacility && bookedService?.needFacility === service?.needFacility;
          if (overlaps && needsFacility) {
            const facilityIndex = bookPattern.facilities.findIndex(
              facility => facility === service.needFacility
            );
            if (facilityIndex !== -1 && facilities[facilityIndex - 1] > 0) {
              facilities[facilityIndex - 1] -= 1;
            }
            return false;
          }
          return true;
        });
      });
    }
    const validOpenTimes = staffOpenTimes.filter(time => time.count > 0);
    openTimes.value = [...openTimes.value, ...validOpenTimes];
  });
  console.log(openTimes.value);
}


onMounted(() => {
  fetchChannel();
  findBookPattern();
  // findBook();
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
  // console.log(openTimes.value, 'openTimes.value');
  if (!openTimes.value) { return false };
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

const selectedServiceId = ref("");

function handleServiceChange(serviceId) {
  const selected = bookPattern.services.find(service => service.id === serviceId);
  generateOpenTimes(selected);
}
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

    <select id="menu-select" class="menu" v-model="selectedServiceId" @change="handleServiceChange(selectedServiceId)">
      <option disabled value="">メニュー</option>
      <option v-for="service in services" :key="service.id" :value="service.id">
        {{ service.serviceName }} - {{ service.price }}円
      </option>
    </select>

    <!-- モーダルを表示 -->
    <BookModal
      v-if="showModal"
      :time="selectedEvent"
      :bookPattern="bookPattern"
      :serviceID="selectedServiceId"
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

.menu {
  position: fixed;
  bottom: 0;
  width: 100%;
  background: white;
  font-size: 0.8rem;
  z-index: 5;
}

</style>
