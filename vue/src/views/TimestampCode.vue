<script setup>
import { ref, onMounted, watchEffect } from 'vue';

import QRCode from 'qrcode';

import Advertisement from '@/components/Advertisement.vue';
import DrawerTimestamp from '@/components/DrawerTimestamp.vue';
import SelectGroup from '@/components/SelectGroup.vue';

import { pushReceive } from '@/pushReceive/pushReceive.js';


document.title = 'タイムスタンプ設定'

const stampCodes = ref([]);
let currents;
const channel = ref('');
const groups = ref([]);
const aliases = ref([]);
const options = ref([]);
const selectedGroup = ref(null);

function selectOption(option) {
  selectedGroup.value = option;
}

function qrLink (stampCode) {
  const name = selectedGroup.value.groupName;
  return `${window.location.origin}/timestamp/${name}/${stampCode.code}/`;
}

async function generateQR(stampCode) {
  const name = selectedGroup.value.groupName;
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
        timeFormat('YYYY-MM-DD', stampCode.planDate),
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

function tF(a, b = null){ return timeFormat(a, b) }
const errorMessage = ref('')
onMounted(async () => {
  channel.value = await getIDB('channel', localStorage.getItem('channelID'));
  groups.value = await getIDBs('group', 'channelIDIndex', localStorage.getItem('channelID'), 10000);
  console.log('groups.value', groups.value)
  aliases.value = await getIDBs('alias', 'channelIDIndex', localStorage.getItem('channelID'), 10000);
  stampCodes.value = await getIDBs('timestampCode', 'channelIDIndex', localStorage.getItem('channelID'));
  currents = JSON.parse(JSON.stringify(stampCodes.value));
  // selectedGroup.value = stampCodes.value.find(sc => sc.adminGroup);
  const stampCode = stampCodes.value.find(sc => sc.adminGroup);
  // console.log('stampCode', stampCode.adminGroup);
  // selectedGroup.value = groups.value.find(group => group.groupName === stampCode.adminGroup);
  const foundGroup = groups.value.find(group => group.groupName === stampCode?.adminGroup);

  if (foundGroup) {
      selectedGroup.value = foundGroup;
  } else if (groups.value.length > 0) {
      selectedGroup.value = groups.value[0];
  } else {
      selectedGroup.value = null; // グループが空だった場合の安全処理
  }

  console.log('selectedGroup.value', selectedGroup.value);
});

function createEmptyStampCode() {
  return {
    code: generateRandomCode(4),
    planDate: timeFormat('YYYY-MM-DDT12:00'),
    until: 12,
  };
}

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

async function handleSubmit(event) {
  const recordsToPost = [];
  for (let stampCode of stampCodes.value) {
    stampCode.channelID = localStorage.channelID;
    stampCode.adminGroup = selectedGroup.value.groupName;
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
  fd.append('channelID', localStorage.getItem('channelID'));
  fd.append('updatedBy', channel.value.myname);
  fd.append('pushTitle', 'timestampCode');
  const pushNames = groups.value
    .filter(group => group.groupName === selectedGroup.value.groupName)
    .flatMap(group => group.aliasNames || [])
  fd.append('pushNames', JSON.stringify(pushNames))
  fd.append('contents', JSON.stringify(stampCodes));
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ContentsPush/', fd);
  if (!res.csrf) errorMessage.value = res
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
}


</script>


<template>
<div id="drawer_column"><DrawerTimestamp /></div>
<div id="content">
  <h2 class="sp_head">タイムスタンプQR登録</h2>
  <div v-if="errorMessage"> 
    <div class="errorMessage">{{errorMessage}}<br>データ取得に失敗しました。</div>
    <a href="/setting/"> データ設定ページ </a><br>
    <a href="/sign/"> サインインページ </a>
  </div>
  <div v-if="isEditMode">
    <form @submit.prevent="handleSubmit">
      <div v-for="(stampCode, index) in stampCodes" :key="index">
        <div>
          <span>予定時刻:</span>
          <input v-model="stampCode.planDate" type="datetime-local" required />
          <span> ~ </span>
          <input v-model="stampCode.until" type="number" required class="until" />
          <span>時間後まで</span>
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
        <SelectGroup v-if="selectedGroup" :groups="groups" v-model="selectedGroup"/>
      </div>
      <br>
      <div class="button"><button type="submit"> ▶️ </button></div>
    </form>
    <br>
    <div v-if="selectedGroup">
      <a :href="'/timestampReport/' + selectedGroup.groupName + `/?month=${tF('YYYY-MM')}`">
      {{ '/timestampReport/' + selectedGroup.groupName + `/?month=${tF('YYYY-MM')}` }} </a>
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
<div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>
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

