<script setup>
import { ref, computed } from 'vue';
import DrawerColumn from '../components/DrawerColumn.vue'
import EventModal from '../components/EventModal.vue';

// 時間のリストを生成（0時〜23時）
const hours = ref(Array.from({ length: 24 }, (_, i) => i));

// サンプルスケジュールデータ
const schedules = ref([
  {
    id: 1,
    timeStart: '2024-10-07T09:30',
    timeEnd: '2024-10-07T10:30',
    title: 'Meeting',
    todo: '',
    scheduleType: 1,
  },
  {
    id: 2,
    timeStart: '2024-10-07T14:00',
    timeEnd: '2024-10-07T15:00',
    title: 'Lunch with team Lunch with team 小松清次郎',
    todo: '',
    scheduleType: 2,
  },
  {
    id: 3,
    timeStart: '2024-10-08T16:00',
    timeEnd: '2024-10-08T18:00',
    title: 'Project Work',
    todo: '',
    scheduleType: 3,
  },
]);

// 今週の開始日を取得（月曜日から）
const getStartOfWeek = (date) => {
  const start = new Date(date);
  const day = start.getDay();
  const diff = (day === 0 ? 6 : day - 1); // 日曜日を最終日に設定
  start.setDate(start.getDate() - diff);
  start.setHours(0, 0, 0, 0); // 時間をリセット
  return start;
};

// 今週の月曜日から日曜日までの日付を取得
const weekDates = computed(() => {
  const startOfWeek = getStartOfWeek(new Date());
  return Array.from({ length: 7 }, (_, i) => {
    const newDate = new Date(startOfWeek);
    newDate.setDate(startOfWeek.getDate() + i);
    return newDate;
  });
});

// 指定された日と時間帯に該当するイベントを取得
const getEventsForDayAndHour = (day, hour) => {
  return schedules.value.filter((d) => {
    const start = new Date(d.timeStart);
    const end = new Date(d.timeEnd);
    // ローカル時間に基づいてイベントの時間を比較する
    return (
      start.getFullYear() === day.getFullYear() &&
      start.getMonth() === day.getMonth() &&
      start.getDate() === day.getDate() &&
      start.getHours() <= hour && // ローカル時間の時間を使用
      start.getHours() >= hour // ローカル時間の時間を使用
    );
  });
};

// 日付を DD フォーマットに変換
const formatDate = (date) => {
  return timeFormat('DD', date);
};

// 曜日をフォーマット
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

const timeStartString = '2024-10-07T09:20';
const timeStartDate = new Date(timeStartString);

console.log(timeStartDate); 


const showModal = ref(false);
const selectedEvent = ref({
  timeStart: '',
  timeEnd: '',
  title: '',
  todo: '',
  scheduleType: 0,
});

// モーダルを開く関数（新規追加）
const openModal = (day, hour) => {
  showModal.value = true;

  // クリックされた時間をもとにデフォルト値を設定
  console.log(day, hour);
  const start = new Date(day);
  start.setHours(hour, 0);
  const end = new Date(start);
  end.setMinutes(start.getMinutes() + 30);

  selectedEvent.value.timeStart = timeFormat('YYYY-MM-DDThh:mm', start);
  selectedEvent.value.timeEnd = timeFormat('YYYY-MM-DDThh:mm', end);
  selectedEvent.value.title = '';
  selectedEvent.value.todo = '';
  selectedEvent.value.scheduleType = 0;
};

// モーダルを開く関数（編集）
const openModalForEdit = (event) => {
  showModal.value = true;
  selectedEvent.value = { ...event };
};

// モーダルを閉じる
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
        <div v-for="hour in hours" :key="hour" class="day-slot" @click="openModal(day, hour)">
          <div
            v-for="d in getEventsForDayAndHour(day, hour)"
            :key="d.id"
            :class="['event', `type-${d.scheduleType}`]"
            :style="decideHeightTop(d)"
          >
            {{ d.title }}
          </div>
        </div>
      </div>
    </div>
    <!-- モーダルを表示 -->
    <EventModal
      v-if="showModal"
      :time="selectedEvent"
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
  /*border-left: 1px solid #ccc;*/
  position: relative;
}

.day-slot {
  border-bottom: 1px solid #ccc;
  position: relative;
  height: 59px;
}

.event {
  position: absolute;
  /*width: 90%;*/
  /*left: 5%;*/
  /*padding: 5px;*/
  border-radius: 4px;
  color: #fff;
  font-size: 0.8rem;
  word-break: break-word;
}

.type-1 {
  background-color: #ff7979;
}

.type-2 {
  background-color: #badc58;
}

.type-3 {
  background-color: #f9ca24;
}

.type-4 {
  background-color: #9b59b6;
}

</style>
