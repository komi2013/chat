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

// const generateKey = (shift) => `${shift.shiftStart}_${shift.role}`;

// const generateSchedules = (del = false) => {
//   schedules.value = [];
//   if (del) {
//     return
//   } else {
//     serviceID.value;
//     bookPattern.value.workStaffs.forEach((work) => {
//       const newSchedule = {
//         timeStart: `${work.workStart}`,
//         timeEnd: `${work.workEnd}`
//       };
//       schedules.value.push(newSchedule);
//     });    
//   }
// };

const generateSchedules = () => {
  if (!bookPattern.value) return;

  // `serviceID.value` が設定されていない場合 → 元のロジック
  // serviceID.value = 2;
  if (!serviceID.value) {
    bookPattern.value.workStaffs.forEach((work) => {
      const newSchedule = {
        timeStart: work.workStart,
        timeEnd: work.workEnd,
        aliasName: work.aliasName
      };
      schedules.value.push(newSchedule);
    });
    return;
  }

  // `serviceID.value` に該当するサービスを取得
  const selectedService = bookPattern.value.services.find(service => service.serviceID === serviceID.value);
  if (!selectedService) return;

  // `needSkill` を取得（無い場合は制限なし）
  const requiredSkill = selectedService.needSkill || null;

  // `needSkill` を持つスタッフ (`staffSkills` から取得)
  let skilledStaffAliases = new Set();
  if (requiredSkill) {
    skilledStaffAliases = new Set(
      bookPattern.value.staffSkills
        .filter(staff => staff.skills.includes(requiredSkill))
        .map(staff => staff.aliasName)
    );
  }

  // `workStaffs` の `aliasName` が `skilledStaffAliases` に含まれる場合のみ `schedules.value` に追加
  bookPattern.value.workStaffs.forEach((work) => {
    if (!requiredSkill || skilledStaffAliases.has(work.aliasName)) {
      const newSchedule = {
        timeStart: work.workStart,
        timeEnd: work.workEnd,
        aliasName: work.aliasName,
        serviceID: serviceID.value
      };
      schedules.value.push(newSchedule);
    }
  });
};


const getEventsForDayAndHour = (day, hour) => {
  if (!bookPattern.value) return [];

  // `bookStart` の日付と時間を `YYYY-MM-DD-HH` の形式で取得（重複なし）
  const bookedHours = new Set(
    bookPattern.value.books.map(book => {
      const bookDate = new Date(book.bookStart);
      return `${bookDate.toISOString().split('T')[0]}-${bookDate.getHours()}`;
    })
  );

  // console.log('bookedHours', bookedHours); // 確認用ログ

  return schedules.value
    .filter((event) => {
      const start = new Date(event.timeStart);
      const end = new Date(event.timeEnd);

      // 対象の日付と時間を `YYYY-MM-DD-HH` 形式に変換
      const eventKey = `${start.toISOString().split('T')[0]}-${hour}`;

      // 指定した `hour` に該当するイベントを抽出
      const isWithinHour =
        start.getFullYear() === day.getFullYear() &&
        start.getMonth() === day.getMonth() &&
        start.getDate() === day.getDate() &&
        (
          (start.getHours() === hour) ||
          (start.getHours() < hour && end.getHours() > hour)
        );

      if (!isWithinHour) return false;

      // 予約のある `hour`（かつ同じ日）のみ、予約数チェックを行う
      if (bookedHours.has(eventKey)) {
        let bookCount = bookPattern.value.books.reduce((count, book) => {
          const bookStart = new Date(book.bookStart);
          const bookEnd = new Date(book.bookEnd);

          const isOverlapping =
            (start >= bookStart && start < bookEnd) ||
            (end > bookStart && end <= bookEnd) ||
            (start <= bookStart && end >= bookEnd);

          return isOverlapping ? count + 1 : count;
        }, 0);

        // 同じ時間帯のスケジュール数を取得
        const sameHourCount = schedules.value.filter(e => {
          const eStart = new Date(e.timeStart);
          const eEnd = new Date(e.timeEnd);
          return (
            eStart.getFullYear() === day.getFullYear() &&
            eStart.getMonth() === day.getMonth() &&
            eStart.getDate() === day.getDate() &&
            (
              (eStart.getHours() === hour) ||
              (eStart.getHours() < hour && eEnd.getHours() > hour)
            )
          );
        }).length;

        // 施設のカウントを考慮する
        let facilityLimitExceeded = false;

        // 予約されているサービスの `needFacility` を取得
        bookPattern.value.books.forEach(book => {
          const service = bookPattern.value.services.find(s => s.serviceID === book.serviceID);
          if (service?.needFacility) {
            // `facilityName` が `needFacility` に一致する `facilityCount` を取得
            const facility = bookPattern.value.facilities.find(f => f.facilityName === service.needFacility);
            if (facility) {
              // 同じ `needFacility` を持つ予約の数をカウント
              const bookedFacilityCount = bookPattern.value.books.filter(b => {
                const bService = bookPattern.value.services.find(s => s.serviceID === b.serviceID);
                return bService?.needFacility === service.needFacility;
              }).length;

              // 予約数が施設のカウントを超えたら制限
              if (bookedFacilityCount >= facility.facilityCount) {
                facilityLimitExceeded = true;
              }
            }
          }
        });

        // 予約数が `同じ時間帯のスケジュール数` 以上、または `施設制限` を超えていたら除外
        if (bookCount >= sameHourCount || facilityLimitExceeded) return false;
      }

      return true;
    })
    .map((event, index, array) => ({ ...event, index, total: array.length }));
};


const decideHeightTop = (schedule) => {
  const start = new Date(schedule.timeStart);
  const end = new Date(schedule.timeEnd);
  const minuteHeight = 1;
  const top = start.getMinutes() * minuteHeight;
  let height = (end - start) / (1000 * 60) * minuteHeight;
  // console.log('height', height);
  if (height > 60) {
    height = 60;
  }
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
          <td v-for="(day, index) in monthDates" :key="'day-' + index" class="day-slot" >
            <div v-for="d in getEventsForDayAndHour(day, hour)" 
               :key="'event-' + d.id" 
               class="event"
               :style="decideHeightTop(d)"
               @click="openModal(day, hour)"
               >
               予約可
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
