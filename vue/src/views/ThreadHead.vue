<script setup>
import { ref, onMounted } from 'vue'

import Quill from 'quill';
import "quill/dist/quill.snow.css";

import Advertisement from '@/components/Advertisement.vue';
import DrawerThread from '@/components/DrawerThread.vue';
import SelectPeople from '@/components/SelectPeople.vue';
import SelectAlias from '@/components/SelectAlias.vue';

import { htmlToMarkdown, markdownToHtml } from '@/my/markdown.js';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  parent_id: String
})

document.title = 'スレッド設定'

const channelID = localStorage.getItem("channelID");
const channel = ref(null);
const aliases = ref([]);
const groups = ref([]);
const threadHead = ref(null);

let adminEditable = false;
const quill = ref(null);
async function initQuill() {
  quill.value = await new Quill('#description', {
    modules: {
      toolbar: '#toolbar',
    },
    theme: 'snow'
  });
  quill.value.root.innerHTML = markdownToHtml(threadHead.value.description, channel.value);
}
const fetched = ref(false)
const errorMessage = ref('')
onMounted(async () => {
  channel.value = await getIDB('channel', channelID);
  aliases.value = await getIDBs('alias', 'channelIDIndex', channelID, 10000);
  groups.value = await getIDBs('group', 'channelIDIndex', channelID, 10000);
  threadHead.value = await getIDB('threadHead', props.parent_id);
  adminEditable = threadHead.value.adminNames.includes(channel.value.myname);
  // const originalAdminNames = JSON.parse(JSON.stringify(threadHead.value.adminNames))
  // const originalAliasNames = JSON.parse(JSON.stringify(threadHead.value.aliasNames))
  // const originalBroadcastFlag = JSON.parse(JSON.stringify(threadHead.value.broadcastFlag))
  await initQuill();
  fetched.value = true
});

const postThreadHead = async () => {
  if (!confirm("実行▶️")) {
    return;
  }
  const description = htmlToMarkdown(quill.value.root.innerHTML.replace(/\uFEFF/g, ''));
  const fd = new FormData();
  fd.append('channelID', channelID);
  fd.append('updatedBy', channel.value.myname);
  fd.append('pushNames', JSON.stringify(threadHead.value.aliasNames))
  fd.append('pushTitle', 'threadHead');
  fd.append('csrf', localStorage.getItem('csrf'));
  let editThreadHead = threadHead.value;
  editThreadHead.description = description;
  fd.append('contents', JSON.stringify(editThreadHead));
  const res = await sendRequest('/ContentsPush/', fd);
  if (!res.csrf) errorMessage.value = res
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
}

function backTo() {
  const backID = props.parent_id;
  const secondPart = backID.replace(channelID, '');
  location.href = '/thread/' + channelID + '/' + secondPart + '/';
}

</script>

<template>
<DrawerThread />
<div id="content">
  <div v-if="fetched" class="sp_head">
    <div>
      <span v-if="threadHead && threadHead.parentID.includes('@')" >{{threadHead.title}}</span>
    </div>
    <span>
      <a @click="backTo"> ⬅ </a>
    </span>
  </div>
  <div v-if="errorMessage"> 
    <div class="errorMessage">{{errorMessage}}<br>データ取得に失敗しました。</div>
    <a href="/setting/"> データ設定ページ </a><br>
    <a href="/sign/"> サインインページ </a>
  </div>
  <br>
  <input v-if="threadHead && !threadHead.parentID.includes('@')" type="text" class="inputText" v-model="threadHead.title"/>

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
  <div id="description" ></div>

  <div v-if="fetched">
    <input type="checkbox" id="muteStatus"
      :checked="threadHead.displayStatus === 3"
      @change="threadHead.displayStatus = $event.target.checked ? 3 : null"
    />
    <label for="muteStatus">ミュート</label>
  </div>

  <div v-if="fetched">
    <input type="checkbox" id="broadcastFlag" v-model="threadHead.broadcastFlag" :disabled="!adminEditable" />
    <label for="broadcastFlag">ブロードキャスト：管理者以外投稿できません</label>    
  </div>

  <SelectPeople v-if="fetched"
    :channel="channel"
    :aliases="aliases"
    :groups="groups"
    :placeholder="'参加ユーザー'"
    v-model="threadHead.aliasNames"
    class="choosePeople"
    />

  <SelectAlias v-if="fetched"
    :channel="channel"
    :aliases="aliases"
    :groups="groups"
    :editable="adminEditable"
    :placeholder="'管理ユーザー'"
    v-model="threadHead.adminNames"
    class="choosePeople"
    />

  <button @click="postThreadHead" class="postButton">➡️</button>
</div>
<div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>

</template>

<style>

.headTitle div {
  width: 74%;
  display: inline-block;
  margin-left: 60px;
}

.headTitle span {
  line-height: 50px;
  width: 50px;
}

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

.choosePeople {
	padding: 6px;
}

</style>

