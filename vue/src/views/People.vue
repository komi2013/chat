<script setup>
import { ref, onMounted } from 'vue';

import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue';
import PeopleImg from '@/components/PeopleImg.vue';

import { getRandomEmoji, getRandomColor } from '@/my/emoji';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  id: String,
  name: String
})

localStorage.setItem('channelID', props.id)
document.title = 'ニックネーム一覧'

const channel = ref(null);
const person = ref(null);
const groups = ref([]);
const aliases = ref([]);
const fetched = ref(false);
// const isEditable = ref(false);
// const image = ref(',' + getRandomEmoji() + ',' + getRandomColor());
const myAliases = ref([]);
let joinGroups;
const isGroup = ref(false);
onMounted(async () => {
  channel.value = await getIDB('channel', props.id);
  groups.value = await getIDBs('group', 'channelIDIndex', props.id, 10000);
  aliases.value = await getIDBs('alias', 'channelIDIndex', props.id, 10000);
  const group = groups.value.find((item) => item.groupName === props.name);
  if (group) {
    isGroup.value = true;
    person.value = {
      name: group.groupName,
      image: group.groupImg || "",
      id: group.groupID || null,
      bio: ""
    };
    myAliases.value = aliases.value
      .filter(item => group.aliasNames.includes(item.aliasName));
  } else {
    person.value = aliases.value.find((item) => item.aliasName === props.name);
    person.value = {
      name: person.value.aliasName || "",
      image: person.value.aliasImg || "",
      id: person.value.aliasID || null,
      bio: person.value.bio || "",
      userID: person.value.userID || "",
    };
    myAliases.value = aliases.value
      .filter(item => item.userID === person.value.userID && item.aliasID !== person.value.aliasID);
    console.log('sameUserAliases', myAliases.value);
    joinGroups = groups.value.filter(group => group.aliasNames.includes(channel.value.myname));
  }
  fetched.value = true;
});


async function directMessage() {
  if (channel.value.myname == props.name) {
    location.href = `/thread/${channel.value.channelID}/@${props.name}/`; 
  } else {
    const sortedNames = [channel.value.myname, props.name].sort();
    const sortedURI = `${sortedNames[0]}@${sortedNames[1]}`;
    location.href = `/thread/${channel.value.channelID}/${sortedURI}/`;    
  }
}

</script>

<template>
<div id="drawer_column"><Drawer v-if="aliases" :aliases="aliases" :channel="channel" /></div>
<div id="content">
  <div v-if="fetched">
    <div class="sp_head">
      <a v-if="channel" :href="'/channel/' + id + '/'">{{channel.channelName}}</a>
<!--       <span v-if="!isGroup && person.name === channel.myname">
        <a :href="'/profile/' + id + '/'"> ✏️ </a>
      </span> -->
      <span v-if="isGroup">
        <a :href="'/group/' + id + '/'"> ⬅️ </a>
      </span>
    </div>
    <div class="join">
      <div class="icon-name">
          <img v-if="person.image && person.image.charAt(0) != ','" 
            :src="person.image" class="new-alias-img">
          <span v-if="person.image && person.image.charAt(0) == ','"
            class="new-alias-img" 
            :style="'background-color:' + person.image.split(',')[2] ">
              <span>{{person.image.split(',')[1]}}</span>
          </span>
        <span class="aliasName">{{person.name}}</span>
      </div>
      <div class="display-mode">
        <p>{{ person.bio }}</p>
      </div>
    </div>

    <h3>ニックネーム一覧</h3>
    <div v-for="(d) in myAliases" >
      <div class="people-list">
        <a :href="'/people/' + id + '/' + d.aliasName + '/' ">
          <img v-if="d.aliasImg && d.aliasImg.charAt(0) != ','" 
            :src="d.aliasImg" class="people-img">
          <span v-if="d.aliasImg && d.aliasImg.charAt(0) == ','"
            class="people-img" 
            :style="'background-color:' + d.aliasImg.split(',')[2] ">
              <span>{{d.aliasImg.split(',')[1]}}</span>
          </span>
          <span> {{d.aliasName}} </span>
        </a>
      </div>
    </div>

    <h3 v-if="!isGroup">参加グループ一覧</h3>
    <div v-if="!isGroup" v-for="(d) in joinGroups" >
      <div class="people-list">
        <a :href="'/people/' + id + '/' + d.groupName + '/' ">
          <img v-if="d.groupImg && d.groupImg.charAt(0) != ','" 
            :src="d.groupImg" class="people-img">
          <span v-if="d.groupImg && d.groupImg.charAt(0) == ','"
            class="people-img" 
            :style="'background-color:' + d.groupImg.split(',')[2] ">
              <span>{{d.groupImg.split(',')[1]}}</span>
          </span>
          <span> {{d.groupName}} </span>
        </a>
      </div>
    </div>
    <div v-if="channel && channel.myname !== props.name" class="dmMessage">
      <button @click="directMessage"> DMメッセージ送信 </button>
    </div>
  </div>
  <div v-if="!fetched"> 
    <div class="errorMessage">データ取得に失敗しました。</div>
    <a href="/setting/"> データ設定ページ </a><br>
    <a href="/sign/"> サインインページ </a>
  </div>
</div>

<div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>

</template>

<style>

.join {
  width: 100%;
  text-align: center;
}

.icon-name {
  display: inline-flex;
  width: 94%;
  align-items: center;
}

.dmMessage {
  width: 100%;
  text-align: center;
}

.dmMessage button {
  padding: 8px;
  margin: 10px;
  width: 80%;
  font-size: 16px;
  background-color: #007bff;
  color: #fff;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

button:hover {
  background-color: #0056b3;
}

.display-mode {
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 5px;
  text-align: left;
  white-space: pre-wrap;
}

.people-list {
  padding: 5px;
  display: inline-flex;
  width: 80%;
}

</style>

