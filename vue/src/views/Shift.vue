<script setup>
import { ref, onMounted, computed, watch } from 'vue';

import { userIDsByName } from '@/my/channelFunc';

const props = defineProps({
  id: String,
});

const channelID = localStorage.getItem("channelID");
function tF(a, b = null){ return timeFormat(a, b) }
const hours = ref(Array.from({ length: 24 }, (_, i) => i));
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

const generateKey = (shift) => `${shift.shiftStart}_${shift.role}`;

const generateSchedules = () => {
  console.log('times', bookPattern.value.times);
  bookPattern.value.shifts.forEach((shift, staffIndex) => {
    const newSchedule = {
      key: generateKey(shift),
      aliasNames: [...shift.aliasNames],
      timeStart: `${shift.shiftStart}`,
      timeEnd: `${shift.shiftEnd}`,
      role: shift.role,
      open: shift.open,
      start: shift.shiftStart,
      end: shift.shiftEnd,
      fix: shift.fix
    };
    schedules.value.push(newSchedule);
  });
};

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
  const width = 118 / schedule.total;
  const left = schedule.index * width;
  const colors = ["rgba(255, 0, 0, 0.25)", "rgba(255, 255, 0, 0.25)", "rgba(128, 0, 128, 0.25)", "rgba(0, 0, 255, 0.25)"];
  const backgroundColor = colors[schedule.nameCount - 1] || colors[3];
  return {
    height: `${height -6}px`,
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
const bookPattern = ref(null);
const availableSkills = ref([]);
const iamAdmin = ref(false);
const shiftStaffs = ref([]);
// let initialStaffs = [];
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
});

function jump(days) {
  const currentDate = new Date(today);
  currentDate.setDate(currentDate.getDate() + days);
  const formattedDate = currentDate.toISOString().split('T')[0];
  location.href = `/calendar/${formattedDate}/`;
}

const moveStaffToTop = (list, aliasName) => {
  const index = list.indexOf(aliasName);
  if (index > 0) {
    list.unshift(list.splice(index, 1)[0]);
  }
};

const onOffOK = (list, name) => {
  const index = list.indexOf(name);
  if (index === -1) {
    list.push(name);
  } else {
    list.splice(index, 1);
  }
};

const toggleFix = (scheduleKey) => {
  const schedule = schedules.value.find(s => s.key === scheduleKey);
  if (schedule) {
    schedule.fix = !schedule.fix; // fixの状態を切り替え
    console.log(`Updated fix for ${scheduleKey}:`, schedule.fix);
  }
};

const submitShift = async () => {
  const updatedShifts = bookPattern.value.shifts
    .filter(shift => {
      const key = generateKey(shift);
      const schedule = schedules.value.find(s => s.key === key);
      return schedule &&
        (
          JSON.stringify(schedule.aliasNames) !== JSON.stringify(shift.aliasNames)
          || (schedule.fix && !shift.fix)
        );
    })
    .map(shift => {
      const key = generateKey(shift);
      const schedule = schedules.value.find(s => s.key === key);
      return {
        aliasNames: schedule.aliasNames,
        shiftStart: schedule.start,
        shiftEnd: schedule.end,
        role: schedule.role,
        fix: schedule.fix,
        open: schedule.open
      };
    });

  if (updatedShifts.length > 0) {
    const fd = new FormData();
    fd.append('bookPatternID', props.id);
    fd.append('channelID', localStorage.getItem('channelID'));
    fd.append('aliasName', channel.value.myname);
    fd.append('csrf', localStorage.getItem('csrf'));
    fd.append('updatedShifts', JSON.stringify(updatedShifts));
    // fd.append('confirmed', confirmed);
    fd.append('availableSkills', JSON.stringify(availableSkills.value));
    const res = await sendRequest('/BookPatternShift/', fd);
    res.csrf && localStorage.setItem('csrf', res.csrf);
    res.pushContents.forEach(content => {
      pushReceive(content);
    });

  }
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
          <h3>スキルを選択:</h3>
          <div v-for="(skill, index) in bookPattern.skills" :key="index">
            <label>
              <input
                type="checkbox"
                :value="skill"
                v-model="availableSkills"
              />
              {{ skill }}
            </label>
          </div>
          
          <p>選択されたスキル: {{ availableSkills }}</p>
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
              :class="{'sunday': day.getDay() === 0}" >
            <div v-for="d in getEventsForDayAndHour(day, hour)" 
               :key="'event-' + d.id" 
               class="event"
               :style="decideHeightTop(d)" >
               {{ d.role }} : <br>
              <button
                v-for="(aliasName, openIndex) in d.aliasNames"
                :style="{
                  color: aliasName === channel.myname ? 'black' : 'white', 
                  backgroundColor: openIndex < d.open ? 'blue' : 'silver'
                }"
                @click="iamAdmin && moveStaffToTop(d.aliasNames, aliasName)"

                >
                {{ aliasName }}
              </button>
              <br>
              <button 
                @click="onOffOK(d.aliasNames, channel.myname)" 
                :style="{ backgroundColor: d.aliasNames.includes(channel.myname) ? 'blue' : 'silver' }">
                登録
              </button>
              <button 
                @click="toggleFix(d.key)"
                :style="{ backgroundColor: d.fix ? 'blue' : 'silver' }">
                確定
              </button>
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
