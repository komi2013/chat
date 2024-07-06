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
// const groupAlias = ref('');
const groupAliases = ref('');

async function fetchChannel() {
  try {
    const data = await getIDB('channel', props.id);
    channel.value = data;
    groupAliases.value = data.groupAliases;

  } catch (error) {
    channel.value = null;
  }
}

const filteredGroupAliases = computed(() => {
  if (props.groupAliasName && groupAliases.value) {
    return groupAliases.value.filter(groupAlias => groupAlias[0] === props.groupAliasName);
  } else {
    return groupAliases.value;
  }
});

function isEditable(groupAlias) {
  if (groupAlias[3]) {
    return true;
  }
  return groupAlias[2].includes(channel.value.aliasName);
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
    console.log(groupAliases.value);
  }
  memberInput.value[i] = '';
  suggestions.value[i] = [];
}

function removeName(i2, i) {
  groupAliases.value[i][2].splice(i2, 1);
}

const newGroupName = ref('');
const newGroupImage = ref('');
const memberInput = ref([]);
const newGroupMembers = ref([]);
const suggestions = ref([]);

function updateSuggestions(i) {
  console.log('tako', memberInput.value[i]);
  const input = memberInput.value[i].toLowerCase();
  suggestions.value[i] = channel.value.allAliases.filter(alias =>
    alias[0].toLowerCase().includes(input)
  );
}


// function addMember(suggestion) {
//   if (!newGroupMembers.value.includes(suggestion[0])) {
//     newGroupMembers.value.push(suggestion[0]);
//   }
//   memberInput.value = '';
//   suggestions.value = [];
// }

// function removeMember(index) {
//   newGroupMembers.value.splice(index, 1);
// }

function newGroup() {
  groupAliases.value.unshift(['', '', [], true]);
  console.log('push', groupAliases);
}

function editGroup() {
  console.log('ga', groupAliases);
  // newGroupName.value = groupAlias[0];
  // newGroupMembers.value = [...groupAlias[2]];
}

// function saveGroupAlias() {
//   if (newGroupName.value && newGroupMembers.value.length) {
//     // Update the existing group alias or add a new one
//     const existingGroup = groupAliases.value.find(group => group[0] === newGroupName.value);
//     if (existingGroup) {
//       existingGroup[2] = [...newGroupMembers.value];
//     } else {
//       groupAliases.value.push([newGroupName.value, '/path/to/new/image.jpg', [...newGroupMembers.value]]);
//     }
//     // Reset inputs
//     newGroupName.value = '';
//     newGroupMembers.value = [];
//   } else {
//     alert('Please fill all the fields and add at least one member.');
//   }
// }

// Add new group to channel.groupAliases
// function addGroup() {
//   if (newGroupName.value && newGroupImage.value && newGroupMembers.value.length) {
//     // channel.groupAliases.push([newGroupName.value, newGroupImage.value, [...newGroupMembers.value]]);
//     // Reset inputs
//     newGroupName.value = '';
//     newGroupImage.value = '';
//     newGroupMembers.value = [];
//   } else {
//     alert('Please fill all the fields and add at least one member.');
//   }
// }

const attach = () => {
  const fileInput = document.getElementById('fileInput');
  if (fileInput) {
    fileInput.click();
  }
  fileInput.addEventListener('change', handleFileInputChange);
}
const fileInfo = ref({});
fileInfo.value = '🖼️';
const handleFileInputChange = (event) => {
  const files = event.target.files;
  const newFileInfo = document.createElement('div');
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    const fileContainer = document.createElement('div');
    if (file.type.startsWith('image/')) {
      const image = document.createElement('img');
      image.src = URL.createObjectURL(file);
      image.style.maxWidth = '50px';
      image.style.maxHeight = '50px';
      fileContainer.appendChild(image);
    } else {
      const fileName = document.createTextNode(file.name);
      fileContainer.appendChild(fileName);
    }
    newFileInfo.appendChild(fileContainer);
  }
  fileInfo.value = newFileInfo.outerHTML;
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

<button @click="newGroup"> + </button>

<!-- <table>
  <tr>
    <td @click="attach" style="text-align: center;">
      <div v-html="fileInfo"></div>
      <input type="file" style="position: fixed; left: -300px;" multiple id="fileInput">
    </td>
    <td colspan="2">
      <input placeholder="グループエイリアス" v-model="newGroupName" />
    </td>
  </tr>
  <tr>
    <td colspan="3">
      <input v-model="memberInput" @input="updateSuggestions" />
      <table v-if="suggestions.length && memberInput" class="dropdown">
        <tr v-for="(suggestion, index) in suggestions" @click="addMember(suggestion)">
          <td style="width: 50px;"><img :src="suggestion[1]"></td>
          <td>{{ suggestion[0] }}</td>
        </tr>
      </table>
    </td>
  </tr>
  <tr v-for="(member, index) in newGroupMembers" :key="index">
    <td><button @click="removeMember(index)">🗑</button></td>
    <td><img :src="getImagePath(member)"></td>
    <td>{{ member }}</td>
  </tr>
  <tr>
    <td colspan="3" style="text-align: center;">
      <button style="width: 80%;" @click="saveGroupAlias">Save</button>
    </td>
  </tr>
</table>
 -->
<table>
  <template v-for="(groupAlias, i) in filteredGroupAliases">
    <tr>
      <td>
        <img v-if="groupAlias[1]" :src="groupAlias[1]">
        <span v-if="!groupAlias[1]">🖼️</span>
      </td>
      <td colspan="2">
        <input v-if="isEditable(groupAlias)" v-model="groupAlias[0]" />
        <div v-if="!isEditable(groupAlias)">
          {{ groupAlias[0] }}
        </div>
      </td>
    </tr>
    <tr v-if="isEditable(groupAlias)" >
      <td colspan="3">
        <input v-model="memberInput[i]" @input="updateSuggestions(i)" />
        <table v-if="suggestions[i] && memberInput[i]" class="dropdown">
          <tr v-for="(suggestion, index) in suggestions[i]" @click="addName(suggestion, index)">
            <td><img :src="suggestion[1]"></td>
            <td>{{ suggestion[0] }}</td>
          </tr>
        </table>
      </td>
    </tr>
    <tr v-for="(member, i2) in groupAlias[2]">
      <td>
        <button v-if="isEditable(groupAlias)"
          @click="removeName(i2, i)">🗑</button>
      </td>
      <td>
        <img :src="getImagePath(member)">
      </td>
      <td>{{ member }}</td>
    </tr>
  </template>
</table>

<div style="text-align: center; width: 100%;">
  <button style="width: 80%;" @click="editGroup">▶️</button>
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

.member {
  padding-left: 50px;
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

