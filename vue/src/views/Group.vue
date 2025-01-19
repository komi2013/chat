<script setup>
import { ref, computed, onBeforeMount } from 'vue';
import DrawerColumn from '../components/DrawerColumn.vue';
import SelectAlias from '@/components/SelectAlias.vue';
import PeopleImg from '@/components/PeopleImg.vue';

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
const fetched = ref(false);
onBeforeMount(async () => {
  channel.value = await fetchChannel(props.id);
  aliases.value = await fetchAliases(props.id);
  groups.value = await fetchGroups(props.id);
  fetched.value = true;
});

const groupName = ref(props.groupName);
// const isEditable = ref(true);

function isEditable(groupAlias) {
  if ( groupAlias.aliasNames.includes(channel.value.myname) ) {
    return true;
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
    aliasNames: [channel.value.myname],
    newOne: true 
  }
  groups.value.push(group);
}

function removeGroup(group) {
  if (!confirm("▶️実行")) {
    return;
  }
  const fd = new FormData();
  fd.append('channelID', channel.value.channelID);
  fd.append('userIDs', JSON.stringify(userIDsByName(aliases.value)));
  fd.append('updatedBy', channel.value.myname);
  fd.append('pushTitle', 'group');
  const contents = [
    group.groupName, ''
  ]
  fd.append('contents', JSON.stringify(contents));
  sendRequest('/ContentsPush/', fd);
}

function editGroup(group) {
  if (!confirm("▶️実行")) {
    return;
  }
  const fd = new FormData();
  fd.append('channelID', channel.value.channelID);
  fd.append('userIDs', JSON.stringify(userIDsByName(aliases.value)));
  fd.append('updatedBy', channel.value.myname);
  fd.append('pushTitle', 'group');
  const contents = [
    group.groupName, group.aliasNames
  ]
  fd.append('contents', JSON.stringify(contents));
  fd.append('imgPath', group.groupImg);
  sendRequest('/ContentsPush/', fd);
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

const updateSelectedAlias = (change) => {
  console.log('change', change);
}

</script>

<template>
<DrawerColumn />
<div id="content" v-if="fetched">
  <div class="headTitle">
    <div>
      <a :href="'/channel/' + channel.channelID"> {{ channel.channelName }} </a>
    </div>
    <div style="line-height: 50px;">
      &nbsp;
    </div>
  </div>

<table>
  <template v-for="(groupAlias, i) in groups">
    <template v-if="!groupName || (groupName === groupAlias.groupName)">
      <tr><td colspan="3">
        <input v-if="groupAlias.newOne" type="text" v-model="groupAlias.groupName" placeholder="グループ名" class="group-name">
        <span v-if="!groupAlias.newOne">{{groupAlias.groupName}}</span>
      </td></tr>
      <tr>
        <td colspan="3">
          <PeopleImg v-if="isEditable(groupAlias)" v-model="groupAlias.groupImg" />
          <template v-if="!isEditable(groupAlias)">
            <img v-if="groupAlias.groupImg && groupAlias.groupImg.charAt(0) != ','" 
              :src="groupAlias.groupImg" class="people-img">
            <span v-if="groupAlias.groupImg && groupAlias.groupImg.charAt(0) == ','"
              class="people-img" 
              :style="'background-color:' + groupAlias.groupImg.split(',')[2] ">
                <span>{{groupAlias.groupImg.split(',')[1]}}</span>
            </span>
          </template>
        </td>
      </tr>
      <tr>
        <td colspan="3" class="height">
          <SelectAlias v-model="groupAlias.aliasNames" :aliases="aliases" :editable="isEditable(groupAlias)" />
        </td>
      </tr>
      <tr>
        <td class="center">
          <a v-if="!groupName" :href="'/group/' + props.id + '/' + groupAlias.groupName + '/' "> ⏭️ </a>
        </td>

        <td v-if="isEditable(groupAlias)" class="center">
          <button v-if="isEditable(groupAlias)" @click="removeGroup(groupAlias)"> 🗑 </button>
        </td>
        <td v-if="isEditable(groupAlias)" class="center">
          <button @click="editGroup(groupAlias)">▶️</button>
        </td>
      </tr>
      <div >
      </div>
      <tr><td colspan="3"><hr></td></tr>
    </template>
  </template>
</table>

<div class="center">
  <button v-if="!groupName" @click="newGroup"> + </button>
  <a v-if="groupName" :href="'/group/' + props.id + '/'"><button> ✏️ </button></a>
</div>

</div>
<div v-if="!fetched"> Loading... or Something Went </div>
</template>

<style>

img {
  max-width: 50px;
  max-height: 50px;
}

hr {
  border: none;
  border-top: 1px solid black;
  margin: 20px 0;
}

.center {
  text-align: center;
}

.center button {
  width: 80%;
  height: 30px;
}

.height {
  height: 60px;
}

.group-name {
  margin: 10px;
  padding: 10px;
}

@media screen and (min-width : 701px) { 
  .headTitle {
    margin-left: 6px;
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

