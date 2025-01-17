<script setup>
import { ref, computed, onBeforeMount, onMounted } from 'vue'
import QRCode from 'qrcode';
import Quill from 'quill';
import "quill/dist/quill.snow.css";

import DrawerColumn from '../components/DrawerColumn.vue'
import PeopleImg from '../components/PeopleImg.vue'
import { markdownToHtml, htmlToMarkdown } from '../my/markdown.js';
import { fetchChannel, fetchAliases, fetchGroups, userIDsByName } from '@/my/channelFunc';

const props = defineProps({
  id: '',
})

const channel = ref({
  channelDescription: '',
  channelName: ''
});
const myname = ref([]);
const myimg = ref([]);
const channels = ref(null);
const aliases = ref([]);
async function fetchAllChannel() {
  try {
    channels.value = await getAllIDBs('channel');
  } catch (error) {
    console.warn('all channel:', error);
  }
}

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
  let userIDs = userIDsByName(aliases.value);
  fd.append('userIDs', JSON.stringify(userIDsByName(aliases.value)));
  const contents = [
    channel.value.channelName,
    htmlToMarkdown(quill.value.root.innerHTML.replace(/\uFEFF/g, ''))
  ];
  fd.append('contents', JSON.stringify(contents));
  sendRequest('/ContentsPush/', fd);
}

async function channelAdd () {
  const fd = new FormData();
  fd.append('channelName', channel.value.channelName);
  fd.append('channelDescription', htmlToMarkdown(quill.value.root.innerHTML.replace(/\uFEFF/g, '')));
  fd.append('myname', myname.value);
  fd.append('myimg', myimg.value);
  const res = await sendRequest('/ChannelAdd/', fd);
  // location.href = res.channelID;
}

const invitationCode = ref('');
const invitationQR = ref('');
const mention = ref(true);
const invite = async () => {
  console.log(aliases.value);
  if (!confirm("実行▶️")) {
    return;
  }
  const fd = new FormData();
  fd.append('channelID', props.id);
  fd.append('updatedBy', channel.value.myname);
  fd.append('channelName', channel.value.channelName);
  fd.append('channelDescription', htmlToMarkdown(quill.value.root.innerHTML.replace(/\uFEFF/g, '')));
  fd.append('userIDs', JSON.stringify(userIDsByName(aliases.value)));
  fd.append('aliasNames', JSON.stringify(aliases.value.map(d => d.aliasName)));
  if (!mention.value) {
    fd.append('noRightMention', 1);
  }
  const res = await sendRequest('/ChannelInvite/', fd);
  invitationCode.value = `${window.location.origin}/profile/${props.id}/?code=${res[0]}`;
  invitationQR.value = await QRCode.toDataURL(invitationCode.value);
};

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

onMounted(async () => {
  await fetchAllChannel();
  aliases.value = await fetchAliases(props.id);
  // groups.value = await fetchGroups(props.id);
  if (props.id) {
    channel.value = channels.value.find(d => d.channelID === props.id);
  }
  initQuill();
});

</script>

<template>
<DrawerColumn />

<div id="content">
<br><br>

<div>
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
  <template v-if="!props.id">
    <input type="text" v-model="myname" placeholder="このチャネルのニックネーム" class="inputText">
    <PeopleImg v-model="myimg" />
  </template>

  <button @click="channelPost" class="postButton">▶️</button><br>
  <div v-if="props.id">
    <button @click="invite" class="postButton"> <span>✉️</span> <span>招待URL</span> </button>
    <div class="optionRight">
      <input type="checkbox" id="mention" v-model="mention" />
      <label for="mention">メンション権限</label>
    </div>
    <div> <a :href="invitationCode"> {{invitationCode}} </a> </div>
    <div> <img :src="invitationQR"></div>    
  </div>
  <div v-if="props.id"> <a :href="'/groupAlias/' + props.id + '/'">
    グループアカウント作成・編集 
  </a> </div>
</div>

<h3>全てのチャネルチーム一覧</h3>
<ul v-if="channels">
  <li v-for="d in channels" :key="d.channelID">
    <a :href="'/channelInfo/' + d.channelID + '/'">{{ d.channelName }}</a>
  </li>
  <li><a href="/channelInfo/">新規チャネル作成</a></li>
</ul>

</div>
<!-- <div v-if="!fetched"> <br><br> Loading... or Something Went </div> -->
</template>

<style>

.ql-snow.ql-toolbar {
  padding: 8px 0px;
}
.ql-snow.ql-toolbar .attachment {
  font-size: 12px;
  padding-top: 0px;
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
  background-color: #007bff;
  color: #fff;
  cursor: pointer;
  transition: background-color 0.3s;
}

.postButton:hover {
  background-color: #0056b3;
}

.optionRight {
  padding: 4px;
}
</style>

