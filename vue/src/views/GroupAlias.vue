<script setup>
import { ref, computed, onBeforeMount } from 'vue';
import DrawerColumn from '../components/DrawerColumn.vue';
import EditBox from '../components/EditBox.vue';
import Messages from '../components/Messages.vue';
import { useThreadHeadsStore } from '../stores/threadHeads.js';
import { useChannelsStore } from '../stores/channels.js';
import { get_formated_time } from '../my/get_formated_time.js';
import { generateRandomCode } from '../my/strings.js';
import { getIDB, getIDBs, upsertData } from '../my/indexDB.js';
import { isEmojiOpen, selectedMessageId, openEmoji, closeEmoji, selectEmoji, calcEmoji, emojiPath } from '../my/emoji.js';
import { isOtherOpen, otherMessageId, openOther, closeOther, selectOther, activeEdit } from '../my/other.js';

const props = defineProps({
  id: '',
  groupAliasName: ''
})

const channel = ref('');
const groupAliases = ref('');
const groupAliasName = ref(props.groupAliasName);
const groupAliasNames = ref([]);
const postable = ref(false);

async function fetchChannel() {
  try {
    const data = await getIDB('channel', props.id);
    channel.value = data;
    groupAliases.value = data.groupAliases;
    for (let i = 0; i < data.groupAliases.length; i++) {
      groupAliasNames.value.push(data.groupAliases[i][0]);
    }
  } catch (error) {
    channel.value = null;
  }
}

function isEditable(groupAlias) {
  if (groupAliasName.value) {
    return false;
  }
  if (groupAlias[3]) {
    return true;
  }
  if (groupAlias[2].includes(channel.value.aliasName)) {
    return true;
  }
  return false;
}

function updateGroupAliasName(event, groupAlias) {
  groupAlias[0] = event.target.innerText;
}

function getImagePath(member) {
  const alias = channel.value.allAliases.find(alias => alias[0] === member);
  return alias ? alias[1] : '';
}

function addName(suggestion, i) {
  if (!groupAliases.value[i][2].includes(suggestion[0])) {
    groupAliases.value[i][2].push(suggestion[0]);
  }
  memberInput.value[i] = '';
  suggestions.value[i] = [];
}

function removeName(i2, i) {
  groupAliases.value[i][2].splice(i2, 1);
}

const memberInput = ref([]);
const suggestions = ref([]);

function updateSuggestions(i) {
  const input = memberInput.value[i].toLowerCase();
  suggestions.value[i] = channel.value.allAliases.filter(alias =>
    alias[0].toLowerCase().includes(input)
  );
}

function newGroup() {
  groupAliases.value.unshift(['', '', [], true]);
}

function removeGroup(i) {
  groupAliases.value.splice(i, 1);
}

function editGroup() {
  console.log('ga', groupAliases);
  if (!confirm("▶️")) {
    return;
  }
  const fd = new FormData();
  fd.append('channelID', channel.value.channelID);
  fd.append('channelName', channel.value.channelName);
  fd.append('channelDescription', channel.value.channelDescription);
  let aliases = [];
  let userIDS = [];
  channel.value.allAliases.forEach(item => {
    aliases.push([item[0], item[1]]);
    userIDS.push(item[2]);
  });
  fd.append('aliases', JSON.stringify(aliases));
  fd.append('userIDs', JSON.stringify(userIDS));
  fd.append('aliasName', channel.value.aliasName);
  fd.append('groupAliases', JSON.stringify(channel.value.groupAliases));
  const request = new Request('/ChannelEdit/', {
    method: 'POST',
    body: fd,
  })
  fetch(request)
    .then((response) => response.json())
    .then((json)=>{
      // when status not 1
    })
    .catch((reason)=>{
      console.log(reason)
    })
}

const fileInfo = ref({});
fileInfo.value = '🖼️';

const attach = (i, editable) => {
  if (!editable) {
    return;
  }
  const fileInput = document.getElementById('fileInput_' + i);
  if (fileInput) {
    fileInput.click();
  }
  fileInput.addEventListener('change', (event) => handleFileInputChange(event, i));
}

const handleFileInputChange = (event, i) => {
  const file = event.target.files[0];
  const newFileInfo = document.createElement('div');
  const fileContainer = document.createElement('div');

  if (file.type.startsWith('image/')) {
    const reader = new FileReader();
    reader.onload = (e) => {
      const base64Image = e.target.result;
      resizeImage(base64Image, 50, 50, (resizedBase64Image) => {
        groupAliases.value[i][1] = resizedBase64Image;

        const image = document.createElement('img');
        image.src = resizedBase64Image;
      });
    };
    reader.readAsDataURL(file);
  } else {
    return;
  }
}

const resizeImage = (base64Str, maxWidth, maxHeight, callback) => {
  const img = new Image();
  img.onload = () => {
    const canvas = document.createElement('canvas');
    let width = img.width;
    let height = img.height;

    // 幅と高さの比率を計算して、サイズを変更
    if (width > height) {
      if (width > maxWidth) {
        height *= maxWidth / width;
        width = maxWidth;
      }
    } else {
      if (height > maxHeight) {
        width *= maxHeight / height;
        height = maxHeight;
      }
    }

    canvas.width = width;
    canvas.height = height;

    const ctx = canvas.getContext('2d');
    ctx.drawImage(img, 0, 0, width, height);

    const resizedBase64 = canvas.toDataURL('image/png');
    callback(resizedBase64);
  }
  img.src = base64Str;
}


onBeforeMount(async () => {
  await fetchChannel();
});


</script>

<template>
<DrawerColumn />
<div id="content">
  <div class="headTitle">
    <div v-if="channel">
      <a :href="'/channelInfo/' + channel.channelID"> {{ channel.channelName }} </a>
    </div>
    <div style="line-height: 50px;">
      &nbsp;
    </div>
  </div>

<div class="editButton">
  <button v-if="!groupAliasName" @click="newGroup"> + </button>
  <a v-if="groupAliasName" :href="'/groupAlias/' + props.id + '/'"><button> ✏️ </button></a>
</div>

<table>
  <template v-for="(groupAlias, i) in groupAliases">
    <template v-if="!groupAliasName || (groupAliasName === groupAliasNames[i])">
    <tr>
      <td @click="attach(i, isEditable(groupAlias))">
        <img v-if="groupAlias[1]" :src="groupAlias[1]">
        <span v-if="!groupAlias[1]">🖼️</span>
        <input type="file" style="position: fixed; left: -300px;" :id="'fileInput_' + i">
      </td>
      <td>
        <input v-if="isEditable(groupAlias)" v-model="groupAlias[0]" />
        <div v-if="!isEditable(groupAlias)">
          {{ groupAlias[0] }}
        </div>
      </td>
      <td> <a v-if="!groupAliasName" :href="'/groupAlias/' + props.id + '/' + groupAlias[0] + '/' "> ⏭️ </a> </td>
      <td> <a v-if="isEditable(groupAlias)" @click="removeGroup(i)"> 🗑 </a> </td>
    </tr>
    <tr v-if="isEditable(groupAlias)" >
      <td colspan="3">
        <input v-model="memberInput[i]" @input="updateSuggestions(i)" />
        <table v-if="suggestions[i] && memberInput[i]" class="dropdown">
          <tr v-for="suggestion in suggestions[i]" @click="addName(suggestion, i)">
            <td><img :src="suggestion[1]"></td>
            <td>{{ suggestion[0] }}</td>
          </tr>
        </table>
      </td>
    </tr>
    <tr v-for="(member, i2) in groupAlias[2]">
      <td>
        <button v-if="isEditable(groupAlias)"
          @click="removeName(i2, i)"> 🗑 </button>
      </td>
      <td colspan="2">
        <img :src="getImagePath(member)">
        {{ member }}
      </td>
    </tr>
    <tr><td colspan="3"><hr></td></tr>
    </template>
  </template>
</table>

<div v-if="!groupAliasName" class="editButton">
  <button @click="editGroup">▶️</button>
</div>

</div>
</template>

<style>

#content {
  height: 100%;
  overflow-y: auto;
}

img {
  max-width: 50px;
  max-height: 50px;
}

hr {
  border: none;             /* デフォルトの境界線を取り除く */
  border-top: 1px solid black; /* 上部に1pxの境界線を追加 */
  margin: 20px 0;
}

.member {
  padding-left: 50px;
}

.editButton {
  text-align: center;
  width: 100%;
}

.editButton button {
  width: 80%;
  height: 30px;
}


.dropdown {
  padding: 10px;
  border: 1px solid #ccc;
  position: absolute;
  background: white;
  z-index: 1000;
  max-width: 200px;
  overflow-y: auto;
}

.dropdown ul {
  list-style-type: none;
  padding: 0;
  margin: 0;
}

.dropdown li {
  padding: 8px;
  cursor: pointer;
}

.dropdown li:hover {
  background-color: #f0f0f0;
}

@media screen and (min-width : 701px) { 
  .headTitle {
    margin-left: 50px;
    display: flex;
  }
}

@media screen and (max-width : 700px) {
  .headTitle {
    margin-left: 50px;
    display: flex;
  }
  .headTitle div {
    display: table-cell;
  }
}
</style>

