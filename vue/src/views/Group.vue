<script setup>
import { ref, computed, onBeforeMount } from 'vue';

import Advertisement from '@/components/Advertisement.vue';
import Drawer from '../components/Drawer.vue';
import SelectAlias from '@/components/SelectAlias.vue';
import PeopleImg from '@/components/PeopleImg.vue';

import { useThreadHeadsStore } from '@/stores/threadHeads.js';
import { useChannelsStore } from '@/stores/channels.js';

import { isEmojiOpen, selectedMessageId, openEmoji, closeEmoji, selectEmoji, calcEmoji, emojiPath } from '@/my/emoji.js';
import { userIDsByName } from '@/my/channelFunc';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  id: '',
  groupName: ''
})

localStorage.setItem('channelID', props.id);

const channel = ref(null)
const groups = ref([])
const aliases = ref([])
const fetched = ref(false)
let groupLockUntilDate
const today = new Date()
const errorMessage = ref('')
onBeforeMount(async () => {
  channel.value = await getIDB('channel', props.id);
  groups.value = await getIDBs('group', 'channelIDIndex', props.id, 10000);
  aliases.value = await getIDBs('alias', 'channelIDIndex', props.id, 10000);
  groupLockUntilDate = new Date(channel.value.groupLockUntilDate);
  groups.value.forEach(group => {
    if (group.aliasNames.includes(channel.value.myname) && today > groupLockUntilDate) {
    // if (group.aliasNames.includes(channel.value.myname)) {
      group.editable = true;
    } else {
      group.editable = false;
    }
  });

  fetched.value = true;
});

// const groupName = ref(props.groupName);

function newGroup() {
  const group = {
    groupID: '',
    channelID: props.id,
    groupName: '',
    groupImg: '',
    aliasNames: [channel.value.myname],
    newOne: true ,
    editable: today < groupLockUntilDate ? false : true
  }
  groups.value.push(group);
}

async function removeGroup(group) {
  if (!confirm("▶️実行")) {
    return
  }
  const fd = new FormData()
  fd.append('channelID', channel.value.channelID)
  fd.append('userIDs', JSON.stringify(userIDsByName(aliases.value)))
  fd.append('updatedBy', channel.value.myname)
  fd.append('pushTitle', 'group')
  const contents = [
    group.groupName, ''
  ]
  fd.append('contents', JSON.stringify(contents))
  fd.append('csrf', localStorage.getItem('csrf'))
  const res = await sendRequest('/ContentsPush/', fd)
  if (!res.csrf) errorMessage.value = res
  res.csrf && localStorage.setItem('csrf', res.csrf)
  res.pushContents.forEach(content => {
    pushReceive(content)
  })
}

async function editGroup(group) {
  if (!confirm("▶️実行")) {
    return;
  }
  const fd = new FormData();
  fd.append('channelID', channel.value.channelID);
  fd.append('userIDs', JSON.stringify(aliases.value.map(d => d.userID)));
  fd.append('updatedBy', channel.value.myname);
  fd.append('pushTitle', 'group');
  const contents = [
    group.groupName, group.aliasNames
  ]
  fd.append('contents', JSON.stringify(contents));
  fd.append('imgPath', group.groupImg);
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ContentsPush/', fd);
  if (!res.csrf) errorMessage.value = res
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
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

</script>

<template>
<Drawer />
<div id="content" v-if="fetched">
  <div class="headTitle">
    <div>
      <a :href="'/channel/' + channel.channelID"> {{ channel.channelName }} </a>
    </div>
  </div>
  <div v-if="errorMessage"> 
    <div class="errorMessage">{{errorMessage}}<br>データ取得に失敗しました。</div>
    <a href="/setting/"> データ設定ページ </a><br>
    <a href="/sign/"> サインインページ </a>
  </div>
  <table>
    <template v-for="(groupAlias, i) in groups">
      <tr><td colspan="3">
        <input v-if="groupAlias.newOne" type="text" v-model="groupAlias.groupName" placeholder="グループ名" class="group-name">
        <PeopleImg v-if="groupAlias.editable" v-model="groupAlias.groupImg" />
        <template v-if="!groupAlias.editable">
          <a :href="'/people/' + props.id + '/' + groupAlias.groupName + '/' ">
            <img v-if="groupAlias.groupImg && groupAlias.groupImg.charAt(0) != ','" 
              :src="groupAlias.groupImg" class="people-img">
            <span v-if="groupAlias.groupImg && groupAlias.groupImg.charAt(0) == ','"
              class="people-img" 
              :style="'background-color:' + groupAlias.groupImg.split(',')[2] ">
                <span>{{groupAlias.groupImg.split(',')[1]}}</span>
            </span>
          </a>
        </template>
        <a :href="'/people/' + props.id + '/' + groupAlias.groupName + '/' ">
          <span v-if="!groupAlias.newOne">{{groupAlias.groupName}}</span>
        </a>
      </td></tr>
      <tr>
        <td colspan="3" class="height">
          <SelectAlias v-model="groupAlias.aliasNames" :aliases="aliases" :editable="groupAlias.editable" />
        </td>
      </tr>
      <tr>
        <td class="center">&nbsp;
          <!-- <button @click="removeGroup(groupAlias)"> ✉️ </button> -->
          <!-- <a v-if="!groupName" :href="'/group/' + props.id + '/' + groupAlias.groupName + '/' "> ⏭️ </a> -->
        </td>

        <td v-if="groupAlias.editable" class="center">
          <button @click="removeGroup(groupAlias)"> 🗑 </button>
        </td>
        <td v-if="groupAlias.editable" class="center">
          <button @click="editGroup(groupAlias)">▶️</button>
        </td>
      </tr>
      <div >
      </div>
      <tr><td colspan="3"><hr></td></tr>
    </template>
  </table>

  <div class="center">
    <button @click="newGroup"> + </button>
  </div>

</div>
<div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>

</template>

<style>

.headTitle div {
  width: 90%;
  display: inline-block;
  height: 50px
}

.headTitle span {
  line-height: 50px;
  width: 50px;
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

