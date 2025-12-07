<script setup>
import { ref, onMounted } from 'vue';

import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue';
import PeopleImg from '@/components/PeopleImg.vue';

import { getRandomEmoji, getRandomColor } from '@/my/emoji';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  id: String,
  code: String
})

localStorage.setItem('channelID', props.id)
document.title = 'プロフィール編集'

const channel = ref(null);
const alias = ref(null);
let aliases;
const fetched = ref(false);
const isEditable = ref(false);
const aliasName = ref(null);
const aliasImg = ref(',' + getRandomEmoji() + ',' + getRandomColor());
let sameUserAliases;
let joinGroups;
const errorMessage = ref('')
onMounted(async () => {
  if (!props.code) {
    channel.value = await getIDB('channel', props.id);
    aliases = await getIDBs('alias', 'channelIDIndex', props.id, 10000);
    aliasName.value = channel.value.myname;
    isEditable.value = true;
    alias.value = aliases.find(
      (item) => item.aliasName === aliasName.value
    );
    if (alias.value) {
      aliasImg.value = alias.value.aliasImg;
      sameUserAliases = aliases.filter(
        item => item.userID === alias.value.userID && item.aliasID !== alias.value.aliasID
      );
      // console.log('sameUserAliases', sameUserAliases);
    }
    const groups = await getIDBs('group', 'channelIDIndex', props.id, 10000);
    joinGroups = groups.filter(group => group.aliasNames.includes(aliasName.value));
  } else if (props.code) {
    isEditable.value = true;
  }
  if (localStorage.getItem('csrf')) {
    if (localStorage.getItem('TO')) {
      localStorage.removeItem("TO")
    }
    fetched.value = true
  } else {
    localStorage.setItem('TO', window.location.pathname + window.location.search)
  }
})

async function aliasEdit() {
  if (!confirm("実行▶️")) {
    return;
  }
  if (props.code) {
    join();
  } else {
    editAlias();
  }
}

async function editAlias() {
  const fd = new FormData();
  fd.append('channelID', props.id);
  fd.append('updatedBy', channel.value.myname);
  fd.append('pushTitle', 'alias');
  fd.append('aliasNames', JSON.stringify(aliases.map(d => d.aliasName)))
  const contents = [alias.value.userID, channel.value.myname, alias.value.aliasBio];
  fd.append('contents', JSON.stringify(contents));
  // fd.append('imgPaths', JSON.stringify([aliasImg.value]));
  fd.append('imgPath', aliasImg.value);
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ChannelEdit/', fd);
  if (!res.csrf) errorMessage.value = res
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
}

async function join () {
  const fd = new FormData()
  fd.append('channelID', props.id)
  fd.append('code', props.code)
  fd.append('myname', aliasName.value)
  fd.append('myimg', aliasImg.value)
  fd.append('csrf', localStorage.getItem('csrf'))
  const res = await sendRequest('/ChannelJoin/', fd)
  if (!res.csrf) {
    errorMessage.value = res
    return
  }
  localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  location.href = '/profile/' + props.id + '/'
}

// async function switchAlias (aliasName) {
//   const newChannel = JSON.parse(JSON.stringify(channel.value));
//   newChannel.myname = aliasName;
//   console.log(newChannel);
//   upsertIDB(newChannel, 'channel', 'channelID', props.id);
//   location.href = '/profile/' + props.id + '/';
// }


</script>

<template>
<div id="drawer_column"><Drawer v-if="aliases" :aliases="aliases" :channel="channel" /></div>
<div id="content">
  <div v-if="fetched">
    <div v-if="errorMessage"> <div class="errorMessage">{{errorMessage}}</div> </div>
    <div class="sp_head">
      <div v-if="channel"><a :href="'/channel/' + id + '/'">{{channel.channelName}}</a></div>
        <!--     <span>
          <a :href="'/profile/' + id + '/'"> ⬅ </a>
        </span> -->
      <div v-if="!channel">チャネル内ニックネーム設定</div>
    </div>
    <br>
    <div class="join">
      <div class="icon-name">
        <template v-if="!isEditable">
          <img v-if="aliasImg && aliasImg.charAt(0) != ','" 
            :src="aliasImg" class="new-alias-img">
          <span v-if="aliasImg && aliasImg.charAt(0) == ','"
            class="new-alias-img" 
            :style="'background-color:' + aliasImg.split(',')[2] ">
              <span>{{aliasImg.split(',')[1]}}</span>
          </span>
        </template>

        <input v-if="props.code" type="text" v-model="aliasName" placeholder="このチャネルのニックネーム" class="aliasName">
        <span v-if="!props.code" class="aliasName">{{aliasName}}</span>
      </div>
      <PeopleImg v-if="isEditable" v-model="aliasImg" />

      <div v-if="!isEditable && alias" class="display-mode">
        <p>{{ alias.aliasBio }}</p>
      </div>
      <textarea 
        v-if="isEditable && alias"
        class="edit-mode" 
        v-model="alias.aliasBio" 
        placeholder="自己紹介を入力してください">
      </textarea>
      <button v-if="isEditable" @click="aliasEdit">▶️</button>
    </div>
    <h3>参加グループ一覧</h3>
    <div v-for="(d) in joinGroups" >
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
      <!--     <h3>マイニックネーム一覧</h3>
          <div v-for="(d) in sameUserAliases" >
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
            <div v-if="isEditable" @click="switchAlias(d.aliasName)" class="switch-alias">🔀</div>
          </div>
       -->
  </div>
  <div v-if="!fetched"> 
    <a href="/sign/"> サインインページ </a>でサインインしてください
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

button {
  padding: 8px;
  margin: 10px;
  width: 200px;
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

.edit-mode {
  width: 94%;
  min-height: 100px;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 5px;
  font-size: 14px;
  font-family: Arial, sans-serif;
}

.people-list {
  padding: 5px;
  display: inline-flex;
  width: 80%;
}

.switch-alias {
  display: inline-flex;
  margin-left: 10px;
}
</style>

