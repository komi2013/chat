<script setup>
import { ref, onMounted, watchEffect } from 'vue';

import QRCode from 'qrcode';

import TimestampDrawer from '../components/TimestampDrawer.vue';
import { get_formated_time } from '../my/get_formated_time.js';
import { getIDB, getIDBs, getAllIDBs, deleteData } from '../my/indexDB.js';
import { generateRandomCode } from '../my/strings.js';

const stampCodes = ref([]);
let currents;
function createEmptyStampCode() {
  return {
    code: generateRandomCode(4),
    planDate: get_formated_time('YYYY-MM-DDT12:00'),
    until: 12,
  };
}

async function fetchChannel() {
  try {
    const data = await getIDB('channel', localStorage.channelID);
    channel.value = data;
    const aliasName = data.aliasName;
    const allAliases = data.allAliases;
    for (let i = 0; i < allAliases.length; i++) {
      if (allAliases[i][0] === aliasName) {
        options.value.push([aliasName, allAliases[i][1]]);
        break;
      }
    }
    for (let i = 0; i < data.groupAliases.length; i++) {
      options.value.push([data.groupAliases[i][0], data.groupAliases[i][1]]);
    }
  } catch (error) {
    console.log('error', error);
    channel.value = null;
  }
}

const selectedOption = ref(null);
async function fetchCodes() {
  try {
    stampCodes.value = await getIDBs('timestampCode', 'channelIDIndex', localStorage.channelID);
    currents = JSON.parse(JSON.stringify(stampCodes.value));
    const stampCode = stampCodes.value.find(sc => sc.name);
    for (let d of options.value) {
      if (stampCode && stampCode.name && d[0] == stampCode.name) {
        selectedOption.value = d;
      }      
    }
  } catch (error) {
    stampCodes.value = [createEmptyStampCode()];
  }
}

async function handleSubmit(event) {
  const recordsToPost = [];
  for (let stampCode of stampCodes.value) {
    stampCode.channelID = localStorage.channelID;
    stampCode.name = selectedOption.value[0];
    const dataToStore = JSON.parse(JSON.stringify(stampCode));

    if (currents) {
      const current = currents.find(current => current.code === stampCode.code);
      if (current) {
        if (
          current.planDate !== stampCode.planDate ||
          current.until !== stampCode.until
        ) {
          recordsToPost.push({ ...dataToStore });
        }
      } else {
        recordsToPost.push({ ...dataToStore });
      }
    } else {
      recordsToPost.push({ ...dataToStore });
    }
  }
  if (currents) {
    for (let current of currents) {
      const existsInNewList = stampCodes.value.some(stampCode => stampCode.code === current.code);
      if (!existsInNewList) {
        recordsToPost.push({ ...current, isDel: true });
      }
    }
  }
  if (recordsToPost.length > 0) {
    await postData(recordsToPost);
  }
}

async function postData(stampCodes) {
  if (!confirm("実行▶️")) {
    return;
  }
  const fd = new FormData();
  let userIDs = [];
  let names = [channel.value.aliasName];
  if (Array.isArray(channel.value.groupAliases)) {
    for (const d of channel.value.groupAliases) {
      if (selectedOption.value[0] == d[0]) {
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
  fd.append('contents', JSON.stringify(stampCodes));
  fd.append('pushTitle', 'timestampCode');
  const request = new Request('/ContentsPush/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      console.log(reason);
    })
}

// Function to clear the form
function clearForm() {
  stampCodes.value = [createEmptyStampCode()];
}

function addStampCode() {
  stampCodes.value.push(createEmptyStampCode());
}

function removeStampCode(index) {
  stampCodes.value.splice(index, 1);
  if (stampCodes.value.length === 0) {
    addStampCode();
  }
}

const channel = ref('');
const options = ref([]);


function selectOption(option) {
  selectedOption.value = option;
}

function qrLink (stampCode) {
  const name = selectedOption.value[0];
  return `${window.location.origin}/timestamp/${name}/${stampCode.code}/`;
}

async function generateQR(stampCode) {
  const name = selectedOption.value[0];
  const url = `${window.location.origin}/timestamp/${name}/${stampCode.code}/`;
  const qr = await QRCode.toDataURL(url);
  return qr;
}
const qrlist = ref([]);
async function generateQRCodes() {
  for (let i = 0; i < stampCodes.value.length; i++) {
    const stampCode = stampCodes.value[i];
    const qrCode = await generateQR(stampCode);
    qrlist.value[i] = [
        qrCode, 
        get_formated_time('YYYY-MM-DD', stampCode.planDate),
        stampCode
      ];
  }
}

const isEditMode = ref(true);
function toggleMode() {
  isEditMode.value = !isEditMode.value;
}

watchEffect(() => {
  if (!isEditMode.value) {
    generateQRCodes();
  }
});


onMounted(() => {
  fetchChannel();
  fetchCodes();
});
</script>


<template>
<TimestampDrawer />
<div id="content">
  <div v-if="isEditMode">
    <h2>タイムスタンプQR登録</h2>
    <form @submit.prevent="handleSubmit">
      <div v-for="(stampCode, index) in stampCodes" :key="index">
        <div>
          <label>ランダムコード:</label>
          <input v-model="stampCode.code" type="text" required />
        </div>
        <div>
          <label>予定時刻:</label>
          <input v-model="stampCode.planDate" type="datetime-local" required />
        </div>
        <div>
          <span> ~ </span>
          <input v-model="stampCode.until" type="number" required class="until" />
          <span>時間後まで有効</span>
        </div>
        <div class="button">
          <button type="button" @click="removeStampCode(index)"> 🗑 </button>
        </div>
        <div><hr></div>
      </div>
      <div class="button">
        <button type="button" @click="addStampCode"> ＋ </button>
      </div>
      <br>
      <div class="dropdown-menu">
        <div 
          v-for="option in options" 
          :key="option[0]" 
          class="dropdown-item" 
          :class="{ 'selected': option === selectedOption }"
          @click="selectOption(option)"
        >
          <img :src="option[1]" alt="Option Image" class="option-image" />
          {{ option[0] }}
        </div>
      </div>
      <br>
      <div class="button"><button type="submit"> ▶️ </button></div>
    </form>
    <br>
    <div v-if="selectedOption">
      <a :href="'/timestampReport/' + selectedOption[0] + '/_/'">
      {{ '/timestampReport/' + selectedOption[0] + '/_/' }} </a>
    </div>
  </div>
  <br>
  <div class="toggle-icons button">
    <button @click="toggleMode">{{ isEditMode ? '🖨️' : '✏️' }}</button>
  </div>
  <div v-if="!isEditMode" v-for="(qr, index) in qrlist" class="qrcode">
    <img :src="qr[0]"><br>
    <a :href="qrLink(qr[2])">{{qr[1]}}</a>
  </div>

</div>
</template>

<style scoped>
#content {
  position: relative;
  left: 10px;
}
hr {
  border: none;
  border-top: 1px solid black;
  margin: 20px 0;
}
h2 {
  position: relative;
  left: 50px;
  top: -10px;
}

.until {
  width: 40px;
  margin-bottom: 10px;
}

.button {
  text-align: center;
  width: 100%;
}
.button button {
  width: 80%;
  height: 30px;
}
.dropdown-menu {
  border: 1px solid #ccc;
  border-radius: 4px;
  background-color: white;
}

.dropdown-item {
  padding: 10px;
  display: flex;
  align-items: center;
  cursor: pointer;
  border-bottom: 1px solid #ccc;
}

.dropdown-item:last-child {
  border-bottom: none;
}

.dropdown-item:hover,
.dropdown-item.selected {
  background-color: #f0f0f0;
}

.option-image {
  width: 20px;
  height: 20px;
  margin-right: 10px;
}

.selected-image-display {
  width: 40px;
  height: 40px;
  margin-right: 10px;
}

.dropdown-item.selected img {
  filter: hue-rotate(90deg); /* Example effect to change the color of the image */
}

.dropdown-item.selected {
  color: blue; /* Change this to whatever color you want for selected text */
}

.selected-display {
  margin-top: 20px;
  display: flex;
  align-items: center;
}

.qrcode {
  display: inline-block;
  /*text-align: center;*/
}
.qrcode a {
  position: relative;
  top: -20px;
  left: 30px;
}
</style>

