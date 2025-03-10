<script setup>
import { ref, onMounted } from 'vue';

import DrawerTimestamp from '@/components/DrawerTimestamp.vue';
import { userIDsByName, userIDsByGroups } from '@/my/channelFunc';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  id: '',
})
console.log(props.id);

// async function fetchChannel() {
//   try {
//     const data = await getIDB('channel', localStorage.channelID);
//     channel.value = data;
//     for (let i = 0; i < data.groupAliases.length; i++) {
//       groups.value.push([data.groupAliases[i][0], data.groupAliases[i][1]]);
//     }
//     console.log('groups', groups);
//   } catch (error) {
//     console.log('error', error);
//     channel.value = null;
//   }
// }

// async function fetchBookPattern() {
//   try {
//     bookPattern.value = await getIDB('bookPattern', props.id);
//     selectedGroup.value = [bookPattern.value.adminGroup];
//   } catch (error) {
//     console.log('error', error);
//   }
// }

// async function findBookPattern() {
//   const fd = new FormData();
//   fd.append('bookPatternID', props.id);
//   const data = await sendRequest('/BookPatternGet/', fd);
//   if (data) {
//     bookPattern.value = data;
//     selectedGroup.value = [bookPattern.value.adminGroup];
//   }
// }


const bookPattern = ref({
  _id: "", // 本来は ObjectId 形式
  admin_group: "",
  join_names: [],
  book_title: "",
  asks: [],
  ask_choices: [],
  facilities: [
    { facility_count: 0, facility_name: "" }
  ],
  times: [
    {
      date: "",
      limit_start: "",
      limit_end: "",
      shift_staffs: [
        { alias_name: "", shift_start: "", shift_end: "", skills: [], seq: 0 }
      ],
      books: [
        { book_start: "", book_end: "", answers: ["", "", ""], service_id: 0 }
      ]
    }
  ],
  services: [
    { id: 1, service_name: "", need_skill: "", price: 0, prepaid_price: 0, spend_minute: 0 }
  ]
});

// 施設を追加する関数
const addFacility = () => {
  bookPattern.value.facilities.push({ facility_count: 0, facility_name: "" });
};

// スタッフを追加する関数
const addStaff = (timeIndex) => {
  bookPattern.value.times[timeIndex].shift_staffs.push({ alias_name: "", shift_start: "", shift_end: "", skills: [], seq: 1 });
};



const publicWindow = ref(false); // チェックボックスの初期値

const channel = ref(null);
const groups = ref([]);
const aliases = ref([]);
const selectedGroup = ref('');
onMounted(async () => {
  channel.value = await getIDB('channel', localStorage.getItem('channelID'));
  groups.value = await getIDBs('group', 'channelIDIndex', localStorage.getItem('channelID'), 10000);
  aliases.value = await getIDBs('alias', 'channelIDIndex', localStorage.getItem('channelID'), 10000);
  if (props.id.length > 4) {
    // findBookPattern();
    const fd = new FormData();
    fd.append('bookPatternID', props.id);
    fd.append('channelID', localStorage.getItem('channelID'));
    fd.append('aliasName', channel.value.myname);
    fd.append('csrf', localStorage.getItem('csrf'));
    const res = await sendRequest('/BookPatternGet/', fd);
    console.log('res', res);
    res.csrf && localStorage.setItem('csrf', res.csrf);
    res.pushContents.forEach(content => {
      pushReceive(content);
    });
    if (res.bookPattern) {
      bookPattern.value = res.bookPattern;
      selectedGroup.value = [bookPattern.value.adminGroup];
    }
  } else if (props.id.length > 0) {
    // fetchBookPattern();
    bookPattern.value = await getIDB('bookPattern', props.id);
    selectedGroup.value = [bookPattern.value.adminGroup];
  }
});

function selectGroup(group) {
  selectedGroup.value = group;
}

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


// const publicWindow = ref(false);
const submit = async () => {
  bookPattern.value.adminGroup = selectedGroup.value.groupName;
  console.log('bookPattern', bookPattern.value);
  if (!confirm("実行▶️")) {
    return;
  }
  const fd = new FormData();
  // let userIDs = [];
  // let names = [channel.value.aliasName];
  // if (Array.isArray(channel.value.groupAliases)) {
  //   for (const d of channel.value.groupAliases) {
  //     if (selectedGroup.value[0] == d[0]) {
  //       for (const d2 of d[2]) {
  //         names.push(d2);
  //       }
  //     }
  //   }
  // }
  // for (const d of channel.value.allAliases) {
  //   if (names.includes(d[0])) {
  //     userIDs.push(d[2]);
  //   }
  // }
  const userIDs = [...new Set([
    ...userIDsByGroups(aliases.value, groups.value, selectedGroup.value.groupName),
    ...userIDsByName(aliases.value, [channel.value.myname])
  ])];
  fd.append('userIDs', JSON.stringify([...new Set(userIDs)]));
  fd.append('channelID', localStorage.getItem('channelID'));
  fd.append('updatedBy', channel.value.myname);
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

  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest(uri, fd);
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
  location.href = '/bookpattern/' + bookPattern.value.bookPatternID;
  // const data = await sendRequest(uri, fd);
  // if (data) {
  //   location.href = '/bookpattern/' + data[0];
  // }
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

    <div>
      <h2>予約設定ページ</h2>

      <!-- 予約タイトル -->
      <label>予約タイトル:</label>
      <input v-model="bookPattern.book_title" type="text" placeholder="予約タイトルを入力" />

      <hr />

      <!-- 施設情報 -->
      <h3>施設情報</h3>
      <div v-for="(facility, index) in bookPattern.facilities" :key="index" class="inline-form">
        <label>施設数:</label>
        <input v-model="facility.facility_count" type="number" placeholder="施設・道具の数" />
        
        <label>施設名:</label>
        <input v-model="facility.facility_name" type="text" placeholder="施設・道具の名称" />
      </div>
      <button @click="addFacility">施設を追加</button>

      <hr />

      <!-- タイムスケジュール -->
      <h3>予約枠</h3>
      <div v-for="(time, timeIndex) in bookPattern.times" :key="timeIndex" class="time-form">
        <label>日付:</label>
        <input v-model="time.date" type="date" />

        <label>予約受付時間:</label>
        <input v-model="time.limit_start" type="time" /> ~ <input v-model="time.limit_end" type="time" />

        <h4>スタッフシフト</h4>
        <div v-for="(staff, staffIndex) in time.shift_staffs" :key="staffIndex" class="inline-form">
          <label>スタッフ名:</label>
          <input v-model="staff.alias_name" type="text" placeholder="スタッフ名" />

          <label>シフト時間:</label>
          <input v-model="staff.shift_start" type="time" /> ~ <input v-model="staff.shift_end" type="time" />

          <label>スキル:</label>
          <input v-model="staff.skills" type="text" placeholder="スキル（カンマ区切り）" />
        </div>
        <button @click="addStaff(timeIndex)">スタッフを追加</button>
        <button @click="removeTime(timeIndex)">この日付を削除</button>
      </div>

      <hr />

      <!-- サービス情報 -->
      <h3>サービス</h3>
      <div v-for="service in bookPattern.services" :key="service.id" class="inline-form">
        <label>サービス名:</label>
        <input v-model="service.service_name" type="text" />

        <label>価格:</label>
        <input v-model="service.price" type="number" />

        <label>前払い価格:</label>
        <input v-model="service.prepaid_price" type="number" />

        <label>必要スキル:</label>
        <input v-model="service.need_skill" type="text" />

        <label>必要設備:</label>
        <input v-model="service.need_facility" type="text" />

        <label>所要時間（分）:</label>
        <input v-model="service.spend_minute" type="number" />
      </div>
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
    <SelectGroup v-if="selectedGroup" :groups="groups" v-model="selectedGroup"/>
    <SelectGroup v-if="groups && !selectedGroup" :groups="groups" v-model="selectedGroup"/>
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
