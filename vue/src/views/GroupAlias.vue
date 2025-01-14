<script setup>
import { ref, computed, onBeforeMount } from 'vue';
import DrawerColumn from '../components/DrawerColumn.vue';
import SelectAlias from '@/components/SelectAlias.vue';

import { useThreadHeadsStore } from '../stores/threadHeads.js';
import { useChannelsStore } from '../stores/channels.js';
import { isEmojiOpen, selectedMessageId, openEmoji, closeEmoji, selectEmoji, calcEmoji, emojiPath } from '../my/emoji.js';
import { fetchChannel, fetchAliases, fetchGroups, userIDsByName } from '@/my/channelFunc';
// import { isOtherOpen, otherMessageId, openOther, closeOther, selectOther, activeEdit } from '../my/toggle.js';

const props = defineProps({
  id: '',
  groupAliasName: ''
})

const channel = ref(null);
const aliases = ref([]);
const groups = ref([]);
const groupAliasName = ref(props.groupAliasName);

function isEditable(groupAlias) {
  // const mynames = [channel.value.aliasName, ''];
  // if (Array.isArray(channel.value.groupAliases)) {
  console.log(groupAlias.aliasNames);
  // if (!groupAlias.aliasNames) {
  //   return true;
  // }
  // const names = groupAlias.aliasNames;
  if ( groupAlias.aliasNames.includes(channel.value.aliasName) ) {
    return true;
  }
  if (groupAliasName.value) {
    return false;
  }
  return false;
}

// function updateGroupAliasName(event, groupAlias) {
//   groupAlias[0] = event.target.innerText;
// }

// function getImagePath(aliasName) {
//   const alias = aliases.value.find(alias => alias.aliasName === aliasName);
//   return alias.aliasImg ? alias.aliasImg : '';
// }

// function removeName(groupIndex, aliasName) {
//   const group = groups.value[groupIndex];
//   if (!group) {
//     console.warn(`Group at index ${groupIndex} not found.`);
//     return;
//   }
//   const aliasIndex = group.aliasNames.findIndex(
//     (name) => name === aliasName
//   );
//   group.aliasNames.splice(aliasIndex, 1);
//   console.log(groups.value);
// }

function newGroup() {
  const group = {
    groupID: '',
    channelID: props.id,
    groupName: '',
    groupImg: '',
    aliasNames: [channel.value.aliasName]
  }
  groups.value.push(group);
}

function removeGroup(i) {
  groups.value.splice(i, 1);
}

function editGroup(group) {
  console.log('group', group);
  if (!confirm("▶️")) {
    return;
  }
  const fd = new FormData();
  fd.append('channelID', channel.value.channelID);
  const userIDs = userIDsByName(channel, aliases);
  fd.append('userIDs', JSON.stringify(userIDs));
  fd.append('aliasName', channel.value.aliasName);
  fd.append('contents', JSON.stringify(group));
  const request = new Request('/ContentsPush/', {
    method: 'POST',
    body: fd,
  });
  fetch(request);
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
        groups.value[i].groupImg = resizedBase64Image;

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

function getAliasesByNames(aliasNames) {
  return aliasNames
    .map((name) => {
      const match = aliases.value.find((entry) => entry.aliasName === name);
      if (match) {
        return {
          id: match.aliasID,
          name: match.aliasName,
          image: match.aliasImg,
        };
      }
      return null;
    })
    .filter(Boolean); // null 値を除外
}

const fetched = ref(false);
onBeforeMount(async () => {
  channel.value = await fetchChannel(props.id);
  aliases.value = await fetchAliases(props.id);
  groups.value = await fetchGroups(props.id);
  fetched.value = true;
  // fetchAlias();
});

const updateSelectedAlias = (change) => {
  console.log('change', change);
}

</script>

<template>
<DrawerColumn />
<div id="content" v-if="fetched">
  <div class="headTitle">
    <div>
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
  <template v-for="(groupAlias, i) in groups">
    <template v-if="!groupAliasName || (groupAliasName === groupAlias.groupName)">
      <tr>
        <td @click="attach(i, isEditable(groupAlias))">
          <img v-if="groupAlias.groupImg" :src="groupAlias.groupImg">
          <span v-else>🖼️</span>
          <input type="file" style="position: fixed; left: -300px;" :id="'fileInput_' + i">
        </td>
        <td>
          <input v-if="isEditable(groupAlias)" v-model="groupAlias.groupName" />
          <div v-if="!isEditable(groupAlias)">
            {{ groupAlias.groupName }}
          </div>
        </td>
        <td> <a v-if="!groupAliasName" :href="'/groupAlias/' + props.id + '/' + groupAlias.groupName + '/' "> ⏭️ </a> </td>
        <td> <a v-if="isEditable(groupAlias)" @click="removeGroup(i)"> 🗑 </a> </td>
      </tr>
      <tr>
        <td colspan="3">
          <SelectAlias v-model="groupAlias.aliasNames" :aliases="aliases" :editable="isEditable(groupAlias)" />
        </td>
      </tr>
      <div v-if="isEditable(groupAlias)" class="editButton">
        <button @click="editGroup(groupAlias)">▶️</button>
      </div>
      <tr><td colspan="3"><hr></td></tr>
    </template>
  </template>
</table>

</div>
<div v-if="!fetched"> Loading... or Something Went </div>
</template>

<style>

/*#content {
  height: 100%;
  overflow-y: auto;
}
*/
img {
  max-width: 50px;
  max-height: 50px;
}

hr {
  border: none;             /* デフォルトの境界線を取り除く */
  border-top: 1px solid black; /* 上部に1pxの境界線を追加 */
  margin: 20px 0;
}

.editButton {
  text-align: center;
  width: 100%;
}

.editButton button {
  width: 80%;
  height: 30px;
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

