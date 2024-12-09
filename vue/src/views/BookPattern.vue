<script setup>
import { ref, onMounted } from 'vue';
import TimestampDrawer from '../components/TimestampDrawer.vue';
// import { getIDB, getIDBs, getAllIDBs, deleteData } from '../my/indexDB.js';
import { generateRandomCode } from '../my/strings.js';

const props = defineProps({
  id: '',
})
console.log(props.id);
const channel = ref('');
const groups = ref([]);
const selectedGroup = ref('');
function selectGroup(group) {
  selectedGroup.value = group;
}

async function fetchChannel() {
  try {
    const data = await getIDB('channel', localStorage.channelID);
    channel.value = data;
    for (let i = 0; i < data.groupAliases.length; i++) {
      groups.value.push([data.groupAliases[i][0], data.groupAliases[i][1]]);
    }
    console.log('groups', groups);
  } catch (error) {
    console.log('error', error);
    channel.value = null;
  }
}

const bookPattern = ref('');
async function fetchBookPattern() {
  try {
    bookPattern.value = await getIDB('bookPattern', props.id);
    selectedGroup.value = [bookPattern.value.adminGroup];
  } catch (error) {
    console.log('error', error);
  }
}

async function findBookPattern() {
  const fd = new FormData();
  fd.append('bookPatternID', props.id);
  const data = await sendRequest('/BookPatternGet/', fd);
  if (data) {
    bookPattern.value = data;
    selectedGroup.value = [bookPattern.value.adminGroup];
  }
}

onMounted(() => {
  fetchChannel();
  if (props.id.length > 4) {
    findBookPattern();
  } else if (props.id.length > 0) {
    fetchBookPattern();
  }
});

const currentMonth = ref(new Date().getMonth());
const currentYear = ref(new Date().getFullYear());

const selectedDates = ref([]);

const addTime = () => {
  const lastTime = bookPattern.value.times[bookPattern.value.times.length - 1];
  const newTime = {
    ...lastTime,
  };
  console.log('newTime', newTime);
  newTime.start = lastTime.end
  newTime.end = calculateNewEndTime(lastTime.start, lastTime.end);
  bookPattern.value.times.push(newTime);
};

const calculateNewEndTime = (start, end) => {
  const [startHours, startMinutes] = start.split(':').map(Number);
  const [endHours, endMinutes] = end.split(':').map(Number);
  const startTotalMinutes = startHours * 60 + startMinutes;
  const endTotalMinutes = endHours * 60 + endMinutes;
  let difference = endTotalMinutes - startTotalMinutes;
  if (difference < 0) {
    difference += 24 * 60;
  }
  const newEndTotalMinutes = endTotalMinutes + difference;
  const adjustedMinutes = newEndTotalMinutes % (24 * 60);
  const newHours = Math.floor(adjustedMinutes / 60);
  const newMinutes = adjustedMinutes % 60;
  return `${newHours.toString().padStart(2, '0')}:${newMinutes.toString().padStart(2, '0')}`;
};

const removeTime = (index) => {
  if (bookPattern.value.times && bookPattern.value.times.length > 0) {
    bookPattern.value.times.splice(index, 1);
  }
};

const applyFormToSelectedDates = () => {
  selectedDates.value.forEach((selectedDate) => {
    const existingTime = bookPattern.value.times.find(time => time.date === selectedDate);

    if (!existingTime) {
      const lastTime = bookPattern.value.times[bookPattern.value.times.length - 1];
      const newTime = {
        ...lastTime,
        date: selectedDate,
      };

      bookPattern.value.times.push(newTime);
    }
  });
};

const getDaysInMonth = (month, year) => {
  const date = new Date(year, month, 1);
  const days = [];
  while (date.getMonth() === month) {
    days.push(new Date(date).toISOString().split('T')[0]);
    date.setDate(date.getDate() + 1);
  }
  return days;
};

const toggleDateSelection = (date) => {
  const index = selectedDates.value.indexOf(date);
  if (index === -1) {
    selectedDates.value.push(date);
  } else {
    selectedDates.value.splice(index, 1);
  }
};


const publicWindow = ref(false);
const submit = async () => {
  bookPattern.value.adminGroup = selectedGroup.value[0];
  if (!confirm("実行▶️")) {
    return;
  }
  const fd = new FormData();
  let userIDs = [];
  let names = [channel.value.aliasName];
  if (Array.isArray(channel.value.groupAliases)) {
    for (const d of channel.value.groupAliases) {
      if (selectedGroup.value[0] == d[0]) {
        for (const d2 of d[2]) {
          names.push(d2);
        }
      }
    }
  }
  for (const d of channel.value.allAliases) {
    if (names.includes(d[0])) {
      userIDs.push(d[2]);
    }
  }
  fd.append('userIDs', JSON.stringify([...new Set(userIDs)]));
  fd.append('channelID', localStorage.channelID);
  fd.append('aliasName', channel.value.aliasName);
  fd.append('contents', JSON.stringify(removeEmptyElements(bookPattern.value)));
  fd.append('pushTitle', 'bookPattern');
  let uri = '/ContentsPush/';
  if (publicWindow.value) {
    if (bookPattern.value.bookPatternID) {
      uri = '/BookPatternEdit/';
    } else {
      uri = '/BookPatternAdd/';
    }
    fd.delete('pushTitle');
    fd.append('contentsTitle', 'bookPattern');
  } else {
    if (!bookPattern.value.bookPatternID) {
      bookPattern.value.bookPatternID = generateRandomCode(4);
    }
  }
  const data = await sendRequest(uri, fd);
  if (data) {
    location.href = '/bookpattern/' + data[0];
  }
};

function removeEmptyElements(obj) {
  if (typeof obj === 'object' && obj !== null) {
    const newObj = Array.isArray(obj) ? [] : {};
    Object.keys(obj).forEach(key => {
      const value = obj[key];
      const cleanedValue = removeEmptyElements(value);
      if (
        cleanedValue !== null &&
        cleanedValue !== "" &&
        !(Array.isArray(cleanedValue) && (cleanedValue.length === 0 || cleanedValue.every(item => item === null))) &&
        !(typeof cleanedValue === 'object' && !Array.isArray(cleanedValue) && Object.keys(cleanedValue).length === 0)
      ) {
        newObj[key] = cleanedValue;
      }
    });
    return newObj;
  }
  return obj;
}

</script>

<template>
  <br><br>
  <div>
    <template v-if="bookPattern.bookTitle">
      <input v-model="bookPattern.bookTitle" type="text" placeholder="予約設定" style="width: 96%;" />      
    </template>
    <template v-if="!bookPattern.bookTitle">
      <span>予約設定ページ</span>
    </template>

    <div v-if="bookPattern.facilities" class="inline-form">
      <input v-model="bookPattern.facilities[0]" type="number" placeholder="施設・道具の数" />
      <input v-model="bookPattern.facilities[1]" type="text" placeholder="施設・道具の名称" />
    </div>

    <div v-if="bookPattern.maxFacility" class="inline-form">

      <label>最大収容人数:</label>
      <input v-model="bookPattern.maxFacility" type="number" placeholder="最大収容人数" />

    </div>

    <div v-if="bookPattern.parentID" class="inline-form">
      <label>ペアレントID:</label>
      <input v-model="bookPattern.parentID" type="text" placeholder="ペアレントID" />
    </div>

    <div class="inline-form">
      <label>公開:</label>
      <input v-model="publicWindow" type="checkbox" />
    </div>

    <hr />

    <div v-for="time in bookPattern.times" class="time-form">
      <div class="inline-form">
        <label>Book Title:</label>
        <input v-model="time.bookTitle" type="text" />
      </div>

      <div class="inline-form">
        <label>Date:</label>
        <input v-model="time.date" type="date" />
      </div>

      <div v-if="time.limitStart" class="inline-form">
        <label>開始時間 ~ 終了時間:</label>
        <input v-model="time.limitStart" type="time" /> ~ <input v-model="time.limitEnd" type="time" />
      </div>

      <div v-if="time.start" class="inline-form">
        <label>開始時間 ~ 終了時間:</label>
        <input v-model="time.start" type="time" /> ~ <input v-model="time.end" type="time" />
      </div>

      <!-- スタッフ情報の入力 -->
      <div v-if="time.staffs" v-for="(staff, staffIndex) in time.staffs" :key="staffIndex" class="inline-form">
        <label>Required Number:</label>
        <input v-model="staff[0]" type="number" placeholder="必要人数" />

        <label>Staff Role:</label>
        <input v-model="staff[1]" type="text" placeholder="スタッフの役割" />

<!--         <label>Staff Members:</label>
        <div v-for="(member, memberIndex) in staff[2] || []" :key="memberIndex">
          <input v-model="staff[2][memberIndex]" type="text" placeholder="スタッフ名" />
        </div> -->
      </div>

      <button @click="removeTime(index)">この日付を削除</button>
      <hr />

    </div>


    <button @click="addTime">日付を追加</button>

    <hr />

    <!-- カレンダー表示 -->
    <h2>カレンダー</h2>
    <div class="calendar">
      <div
        v-for="day in getDaysInMonth(currentMonth, currentYear)"
        :key="day"
        :class="['calendar-day', { selected: selectedDates.includes(day) }]"
        @click="toggleDateSelection(day)"
      >
        {{ day.split('-')[2] }} <!-- 日付のみ表示 -->
      </div>
    </div>

    <button @click="applyFormToSelectedDates">選択した日にちに適用</button>
    <div style="margin-left: 10px;">
      <span>質問内容</span>
      <div v-if="bookPattern.asks" v-for="(ask, askIndex) in bookPattern.asks" class="inline-form">
        <input v-model="bookPattern.asks[askIndex]" type="text" />
      </div>

      <span>単一選択の質問</span>
      <div v-if="bookPattern.askChoices" v-for="(choice, choiceIndex) in bookPattern.askChoices" :key="choiceIndex">
        <input v-model="bookPattern.askChoices[choiceIndex][0]" type="text" />
        <template v-for="(c, ci) in bookPattern.askChoices[choiceIndex]">
          <input v-if="ci > 0" v-model="bookPattern.askChoices[choiceIndex][ci]" type="text" /><br>
        </template>
        <br>
      </div>

      <span>複数選択の質問</span>
      <div v-if="bookPattern.askMultiChoices" v-for="(multiChoice, index) in bookPattern.askMultiChoices" :key="index" >
        <input v-model="bookPattern.askMultiChoices[index][0]" type="text" />
        <template v-for="(c, ci) in bookPattern.askMultiChoices[index]">
          <input v-if="ci > 0" v-model="bookPattern.askMultiChoices[index][ci]" type="text" /><br>
        </template>
        <br>
      </div>
    </div>
    <div class="dropdown-menu">
      <div 
        v-for="group in groups"
        class="dropdown-item" 
        :class="{ 'selected': group[0] === selectedGroup[0] }"
        @click="selectGroup(group)"
      >
        <img :src="group[1]" class="option-image" />
        {{ group[0] }}
      </div>
    </div>
    <div style="text-align: center;">
      <button @click="submit" style="width: 80%;">送信</button>
    </div>
  </div>
</template>

<style scoped>
.time-form {
  margin-bottom: 20px;
}

.inline-form {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

.inline-form label {
  width: 120px;
}

.selected {
  background-color: #f0f0f0;
}

button {
  padding: 10px 15px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

button:hover {
  background-color: #0056b3;
}


/* カレンダーのスタイル */
.calendar {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 5px;
  margin: 20px 0;
}

.calendar-day {
  padding: 10px;
  background-color: #f0f0f0;
  text-align: center;
  cursor: pointer;
}

.calendar-day.selected {
  background-color: #90ee90; /* 選択された日にちの色 */
}
</style>
