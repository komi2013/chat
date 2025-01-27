<script setup>
import { ref, computed, onBeforeMount, onMounted } from 'vue'

import Quill from 'quill';
import "quill/dist/quill.snow.css";

import DrawerThread from '@/components/DrawerThread.vue';
import SelectPeople from '@/components/SelectPeople.vue';
import SelectAlias from '@/components/SelectAlias.vue';

import { htmlToMarkdown, markdownToHtml } from '@/my/markdown.js';
import { userIDsByName } from '@/my/channelFunc';

const props = defineProps({
  parent_id: ''
})
const channelID = localStorage.getItem("channelID");
const channel = ref(null);
const aliases = ref([]);
const groups = ref([]);
const threadHead = ref(null);

// onBeforeMount(async () => {
//   channel.value = await getIDB('channel', channelID);
//   aliases.value = await getIDBs('alias', 'channelIDIndex', channelID, 10000);
//   groups.value = await getIDBs('group', 'channelIDIndex', channelID, 10000);
//   threadHead.value = await getIDB('threadHead', props.parent_id);
//   console.log(threadHead.value);
// });

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

onMounted(async () => {
  channel.value = await getIDB('channel', channelID);
  aliases.value = await getIDBs('alias', 'channelIDIndex', channelID, 10000);
  groups.value = await getIDBs('group', 'channelIDIndex', channelID, 10000);
  threadHead.value = await getIDB('threadHead', props.parent_id);
  console.log(threadHead.value);
  await initQuill();
});

const postThreadHead = () => {
  if (!confirm("実行▶️")) {
    return;
  }
  const description = htmlToMarkdown(quill.value.root.innerHTML.replace(/\uFEFF/g, ''));
  console.log(description);
  const fd = new FormData();
  fd.append('channelID', channelID);
  fd.append('updatedBy', channel.value.myname);
  fd.append('userIDs', JSON.stringify(userIDsByName(aliases.value, threadHead.value.aliasNames)));
  fd.append('pushTitle', 'threadHead');
  const contents = [
  	threadHead.value.parentID.replace(channelID, ""),
    threadHead.value.title,
    description,
    threadHead.value.aliasNames,
    threadHead.value.adminNames,
    threadHead.value.broadcastFlag
  ];
  fd.append('contents', JSON.stringify(contents));
  sendRequest('/ContentsPush/', fd);
}

</script>

<template>
<DrawerThread />
<div id="content"><br><br>
  <input v-if="threadHead" type="text" class="inputText" v-model="threadHead.title"/>

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

  <div v-if="threadHead">
    <input type="checkbox" id="broadcastFlag" v-model="threadHead.broadcastFlag" />
    <label for="broadcastFlag">ブロードキャスト：管理者以外投稿できません</label>    
  </div>

  <SelectPeople v-if="threadHead"
    :channel="channel"
    :aliases="aliases"
    :groups="groups"
    :placeholder="'参加ユーザー'"
    v-model="threadHead.aliasNames"
    class="choosePeople"
    />

  <SelectAlias v-if="threadHead"
    :channel="channel"
    :aliases="aliases"
    :groups="groups"
    :editable="false"
    :placeholder="'管理ユーザー'"
    v-model="threadHead.adminNames"
    class="choosePeople"
    />

  <button @click="postThreadHead" class="postButton">➡️</button>
</div>
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

.choosePeople {
	padding: 6px;
}

</style>

