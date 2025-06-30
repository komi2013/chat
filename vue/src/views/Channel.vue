<script setup>
import { ref, computed, onBeforeMount, onMounted } from 'vue'
import QRCode from 'qrcode';
import Quill from 'quill';
import "quill/dist/quill.snow.css";

import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue'
import PeopleImg from '@/components/PeopleImg.vue'

import { markdownToHtml, htmlToMarkdown } from '@/my/markdown.js';
import { userIDsByName } from '@/my/channelFunc';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  id: '',
})

props.id && localStorage.setItem('channelID', props.id);

const channel = ref({
  channelDescription: '',
  channelName: ''
});
const myname = ref(null);
const myimg = ref(null);
const userID = ref(null);
const channels = ref([]);
const aliases = ref([]);
const groups = ref([]);
const threadHeads = ref([]);

const quill = ref(null);
function initQuill() {
  quill.value = new Quill('#description', {
    modules: {
      toolbar: '#toolbar',
    },
    theme: 'snow'
  });
  quill.value.root.innerHTML = markdownToHtml(channel.value.channelDescription, channel.value);
}

const newThreadURL = '/thread/' + props.id + '/' + generateRandomCode(3) + '/';

onMounted(async () => {
  channels.value = await getAllIDBs('channel');
  groups.value = await getIDBs('group', 'channelIDIndex', props.id, 10000);
  aliases.value = await getIDBs('alias', 'channelIDIndex', props.id, 10000);
  const threadHeadsAll = await getIDBs('threadHead', 'channelIDIndex', props.id, 1000);
  threadHeads.value = threadHeadsAll.filter(d => !d.backID);

  if (props.id) {
    channel.value = channels.value.find(d => d.channelID === props.id);
  }
  myname.value = channel.value.myname;
  userID.value = aliases.value.find((d) => d.aliasName === myname.value)?.userID;
  // const isAdmin = channel.value.adminNames.includes(myname.value);
  const isAdmin = false;
  aliases.value = aliases.value.map((alias) => ({
    ...alias,
    removable: alias.userID === userID.value || isAdmin,
  }));
  await initQuill();
});

const channelPost = async () => {
  if (!confirm("実行▶️")) {
    return;
  }
  if (props.id) { // edit
    channelEdit();
  } else {
    channelAdd();
  }
}

async function channelEdit () {
  const fd = new FormData();
  fd.append('channelID', props.id);
  fd.append('updatedBy', channel.value.myname);
  fd.append('pushTitle', 'channelEdit');
  fd.append('userIDs', JSON.stringify(aliases.value.map(d => d.userID)));
  const contents = [
    channel.value.channelName,
    htmlToMarkdown(quill.value.root.innerHTML.replace(/\uFEFF/g, ''))
  ];
  fd.append('contents', JSON.stringify(contents));
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ContentsPush/', fd);
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
}

async function channelAdd () {
  const fd = new FormData();
  fd.append('channelName', channel.value.channelName);
  fd.append('channelDescription', htmlToMarkdown(quill.value.root.innerHTML.replace(/\uFEFF/g, '')));
  fd.append('myname', myname.value);
  fd.append('myimg', myimg.value);
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ChannelAdd/', fd);
  res.csrf && localStorage.setItem('csrf', res.csrf);
  location.href = res.channelID;
}

const invitationCode = ref('');
const invitationQR = ref('');
const mention = ref(true);
const untilDays = ref(1);
const invite = async () => {
  if (!confirm("実行▶️")) {
    return;
  }
  const until = new Date();
  until.setDate(until.getDate() + untilDays.value);
  const untilDate = timeFormat('YYYY-MM-DD', until);
  let fd = new FormData();
  fd.append('channelID', props.id);
  fd.append('updatedBy', channel.value.myname);
  fd.append('channelName', channel.value.channelName);
  fd.append('channelDescription', htmlToMarkdown(quill.value.root.innerHTML.replace(/\uFEFF/g, '')));
  fd.append('userIDs', JSON.stringify(aliases.value.map(d => d.userID)));
  fd.append('aliasNames', JSON.stringify(aliases.value.map(d => d.aliasName)));
  fd.append('aliases', JSON.stringify(aliases.value));
  fd.append('groups', JSON.stringify(groups.value));
  if (!mention.value) {
    fd.append('noRightMention', 1);
  }
  fd.append('untilDate', untilDate);
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ChannelInvite/', fd);
  invitationCode.value = `${window.location.origin}/profile/${props.id}/?code=${res.invitationCode}`;
  invitationQR.value = await QRCode.toDataURL(invitationCode.value);
  res.csrf && localStorage.setItem('csrf', res.csrf);
  if (res) {
    fd = new FormData();
    fd.append('channelID', props.id);
    fd.append('updatedBy', channel.value.myname);
    fd.append('pushTitle', 'channelEdit');
    fd.append('userIDs', JSON.stringify(aliases.value.map(d => d.userID)));
    const contents = [
      channel.value.channelName,
      htmlToMarkdown(quill.value.root.innerHTML.replace(/\uFEFF/g, '')),
      untilDate
    ];
    fd.append('contents', JSON.stringify(contents));
    fd.append('csrf', localStorage.getItem('csrf'));
    const res = await sendRequest('/ContentsPush/', fd);
    res.csrf && localStorage.setItem('csrf', res.csrf);
    res.pushContents.forEach(content => {
      pushReceive(content);
    });
  }
};

const removeName = (alias) => {
  alias.deleteFlag = true;
};

const removeNames = async () => {
  console.log(aliases.value);
  const deleteAliases = aliases.value.filter(d => d.deleteFlag);
  console.log('deleteAliases', deleteAliases);
  if (!confirm("実行▶️")) {
    return;
  }
  const fd = new FormData();
  fd.append('channelID', props.id);
  fd.append('updatedBy', channel.value.myname);
  fd.append('userIDs', JSON.stringify(aliases.value.map(d => d.userID)));
  // fd.append('deleteUserIDs', JSON.stringify(deleteAliases.map(d => d.userID)));
  // const contents = [
  //   channel.value.channelName,
  //   htmlToMarkdown(quill.value.root.innerHTML.replace(/\uFEFF/g, ''))
  // ];
  fd.append('deleteAliases', JSON.stringify(deleteAliases));
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ChannelDelete/', fd);
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
};

</script>

<template>
<Drawer />

<div id="content">
  <br><br>
  <input type="text" v-model="channel.channelName" placeholder="グループ名" class="inputText">
  <div class="editLeft" id="toolbar">
    <button class="ql-bold"></button>
    <button class="ql-strike"></button>
    <button class="ql-blockquote"></button>
    <button class="ql-code-block"></button>
    <button class="ql-link"></button>
    <select class="ql-color">
      <option value="red">Red</option>
      <option value=""></option>
    </select>
  </div>
  <div id="description" ></div><br>
  <template v-if="!id">
    <input type="text" v-model="myname" placeholder="このチャネルのニックネーム" class="inputText">
    <PeopleImg v-model="myimg" />
  </template>

  <button @click="channelPost" class="postButton" :disabled="!channel.channelName || !myname">▶️</button><br>

  <div v-if="id"> このチャネルのニックネーム: {{channel.myname}} </div>

  <div v-if="id" class="invitation">
    <button @click="invite" class="postButton"> <span>✉️</span> <span>招待URL</span> </button>
    <div class="optionRight">
      <input type="checkbox" id="mention" v-model="mention" />
      <label for="mention">メンション権限</label>
      <span>&nbsp;&nbsp;</span>
      <input type="number" id="until" v-model="untilDays" />
      <label for="until">日まで有効</label>      
    </div>
    <div> <a :href="invitationCode"> {{invitationCode}} </a> </div>
    <div> <img :src="invitationQR"></div>    
  </div>
  <div v-if="id"> <a :href="'/group/' + id + '/'">
    👪 グループアカウント作成・編集 
  </a> </div>

  <h3>ユーザー一覧</h3>
  <div v-for="d in aliases" class="aliases" :class="{ 'deleted': d.deleteFlag }">
    <a :href="'/people/' + id + '/' + d.aliasName + '/'">
      <img v-if="d.aliasImg && d.aliasImg.charAt(0) != ','" 
        :src="d.aliasImg" class="min-icon">
      <span v-if="d.aliasImg && d.aliasImg.charAt(0) == ','"
        :style="'background-color:' + d.aliasImg.split(',')[2] "
        class="min-icon">
          <span>{{d.aliasImg.split(',')[1]}}</span>
      </span>
      <span>{{ d.aliasName }}</span>
    </a>
    <button v-if="d.removable" @click="removeName(d)" >x</button>
  </div>
  <button @click="removeNames" class="postButton">▶️</button>

  <template v-if="id">
    <h3>スレッド一覧</h3>
    <ul v-if="threadHeads">
      <li v-for="d in threadHeads">
        <a :href="'/thread/' + d.parentID.slice(0, 4) + '/' + d.parentID.slice(4) + '/'">
          {{d.title}}
        </a>
      </li>
      <li><a :href="newThreadURL"><button> + 新規 </button></a></li>
    </ul>
  </template>

  <h3>チャネル一覧</h3>
  <ul v-if="channels">
    <li v-for="d in channels" :key="d.channelID">
      <a :href="'/channel/' + d.channelID + '/'">{{ d.channelName }}</a>
    </li>
    <li><a href="/channel/"><button> + 新規 </button></a></li>
  </ul>

</div>
<div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>

</template>

<style>

.ql-snow.ql-toolbar {
  padding: 8px 0px;
}
.ql-snow.ql-toolbar .attachment {
  font-size: 12px;
  padding-top: 0px;
}

#description {
  height: auto;
}

.inputText {
  width: 100%;
  padding: 10px;
  margin-bottom: 10px;
  border: 1px solid #ccc;
  border-radius: 5px;
  box-sizing: border-box;
}

.postButton {
  display: block;
  width: 100%;
  padding: 10px;
  border: none;
  border-radius: 5px;
}

.optionRight {
  padding: 4px;
}

.aliases {
  padding: 4px;
  display: inline-flex;
}

.deleted {
  background-color: gray;
  opacity: 0.4;
}

.min-icon {
  width: 26px;
  max-width: 26px;
  height: 26px;
  max-height: 26px;
  border-radius: 4px;
  display: inline-flex;
  vertical-align: middle;
  justify-content: center;
  align-items: center;
}

.invitation {
  padding: 10px;
  margin: 10px;
  border: 1px solid #ccc;
}

.invitation input[type=number] {
  width: 32px;
}

</style>

