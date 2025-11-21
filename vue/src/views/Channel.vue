<script setup>
import { ref, onMounted } from 'vue'
import QRCode from 'qrcode';
import Quill from 'quill';
import "quill/dist/quill.snow.css";

import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue'
import PeopleImg from '@/components/PeopleImg.vue'
import SelectAlias from '@/components/SelectAlias.vue'

import { markdownToHtml, htmlToMarkdown } from '@/my/markdown.js';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  id: String,
})

props.id && localStorage.setItem('channelID', props.id);

const channel = ref({
  channelDescription: '',
  channelName: ''
});
const myname = ref(null);
const myimg = ref(null);
let userID
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
const errorMessage = ref('')
const iamAdmin = ref(false)
const iamGuest = ref(false)
const adminNames = ref([])
let initAdminNames = []
onMounted(async () => {
  channels.value = await getAllIDBs('channel');
  groups.value = await getIDBs('group', 'channelIDIndex', props.id, 10000);
  aliases.value = await getIDBs('alias', 'channelIDIndex', props.id, 10000);
  const threadHeadsAll = await getIDBs('threadHead', 'channelIDIndex', props.id, 1000);
  threadHeads.value = threadHeadsAll.filter(d => !d.backID);

  if (props.id) {
    channel.value = channels.value.find(d => d.channelID === props.id);
    document.title = channel.value.channelName
    localStorage.setItem('channelID', props.id)
    localStorage.setItem('myname', channel.value.myname)
  } else {
    document.title = '組織・チャネル設定'
  }
  myname.value = channel.value.myname
  // console.log('myname', myname.value)
  userID = aliases.value.find((d) => d.aliasName === myname.value)?.userID
  // console.log('userID', userID)
  adminNames.value = aliases.value
    .filter(alias => alias.accessRight === "admin")
    .map(alias => alias.aliasName)
  initAdminNames = [...adminNames.value]
  iamAdmin.value = adminNames.value.includes(channel.value.myname)
  aliases.value = aliases.value.map((alias) => ({
    ...alias,
    removable: alias.userID === userID || iamAdmin.value,
  }))
  // console.log('aliases', aliases.value)
  iamGuest.value = aliases.value.some(
    alias => alias.accessRight === "guest" && alias.aliasName === channel.value.myname
  )
  await initQuill()
})

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

const invitationCode = ref('');
const invitationQR = ref('');
const guest = ref(false)
// const untilDays = ref(1);
const generateInvitation = ref('')
async function channelEdit () {
  const fd = new FormData()
  fd.append('channelID', props.id)
  fd.append('updatedBy', channel.value.myname)
  const pushNames = aliases.value
      .filter(d => d.accessRight !== 'guest' && d.accessRight !== 'inquirer')
      .map(d => d.aliasName)
  fd.append('pushNames', JSON.stringify(pushNames))
  fd.append('channelName', channel.value.channelName)
  fd.append('channelDescription', htmlToMarkdown(quill.value.root.innerHTML.replace(/\uFEFF/g, '')))
  fd.append('csrf', localStorage.getItem('csrf'))
  const res = await sendRequest('/ChannelEdit/', fd)
  if (!res.csrf) {
    errorMessage.value = res
    return
  }
  if (res.error) errorMessage.value = res.error
  res.csrf && localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
}

async function channelAdd () {
  if (!myname.value || !channel.value.channelName) {
    alert("入力してください")
    return
  }
  if (!confirm("実行▶️")) return
  const fd = new FormData()
  fd.append('channelName', channel.value.channelName)
  fd.append('channelDescription', htmlToMarkdown(quill.value.root.innerHTML.replace(/\uFEFF/g, '')))
  fd.append('myname', myname.value);
  fd.append('myimg', myimg.value);
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ChannelAdd/', fd);
  if (!res.csrf) errorMessage.value = res
  if (res.error) errorMessage.value = res.error
  res.csrf && localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  if (res.channelID) location.href = `/channel/${res.channelID}/`
}

const invite = async () => {
  if (!confirm("実行")) return
  let fd = new FormData();
  fd.append('channelID', props.id);
  fd.append('updatedBy', channel.value.myname);
  if (generateInvitation.value) fd.append('generateInvitation', 1)
  if (guest.value) fd.append('guest', 1)
  // fd.append('untilDate', untilDate)
  fd.append('csrf', localStorage.getItem('csrf'))
  const res = await sendRequest('/ChannelEdit/', fd)
  if (!res.csrf) errorMessage.value = res
  invitationCode.value = `${window.location.origin}/profile/${props.id}/?code=${res.invitationCode}`
  invitationQR.value = await QRCode.toDataURL(invitationCode.value)
  // untilDays.value = res.untilDate
  res.csrf && localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
}

async function removeChannel () {
  if (!confirm("実行")) return
  const fd = new FormData()
  fd.append('channelID', props.id)
  fd.append('updatedBy', channel.value.myname)
  fd.append('channelDelete', '1')
  fd.append('csrf', localStorage.getItem('csrf'))
  const res = await sendRequest('/ChannelDelete/', fd)
  if (!res.csrf) {
    errorMessage.value = res
    return
  }
  if (res.error) errorMessage.value = res.error
  res.csrf && localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
}

const removeName = (alias) => {
  alias.deleteFlag = true;
};

const removeNames = async () => {
  const deleteAliases = aliases.value
    .filter(d => d.deleteFlag)
    .map(d => d.aliasName)
  if (!confirm("実行")) return
  const fd = new FormData();
  fd.append('channelID', props.id);
  fd.append('updatedBy', channel.value.myname);
  fd.append('pushNames', JSON.stringify(aliases.value.map(d => d.aliasName)));
  fd.append('deleteAliases', JSON.stringify(deleteAliases));
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ChannelEdit/', fd);
  if (!res.csrf) errorMessage.value = res
  res.csrf && localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  // location.href = ''
}

async function adminEdit() {
  if (!confirm("実行")) return
  const diffAdmins = getAdminDiffData()
  if (diffAdmins.length == 0) return
  const fd = new FormData()
  for (const alias of diffAdmins) {
    fd.set('channelID', props.id)
    fd.set('updatedBy', channel.value.myname)
    const pushNames = aliases.value
        .filter(d => d.accessRight !== 'guest' && d.accessRight !== 'inquirer')
        .map(d => d.aliasName)
    fd.append('pushNames', JSON.stringify(pushNames))
    // fd.set('pushTitle', 'alias')
    // const contents = [alias.userID, alias.aliasName, alias.aliasBio, alias.accessRight]
    fd.set('admin', JSON.stringify(alias))
    fd.set('imgPath', alias.aliasImg);
    fd.set('csrf', localStorage.getItem('csrf'))
    const res = await sendRequest('/ChannelEdit/', fd)
    if (!res.csrf) {
      errorMessage.value = res
      return
    }
    if (res.error) errorMessage.value = res.error
    res.csrf && localStorage.setItem('csrf', res.csrf)
    if (Array.isArray(res.pushContents)) {
      for (const content of res.pushContents) {
        await pushReceive(content)
      }
    }
  }
  // location.href = ''
}

function getAdminDiffData() {
  const before = initAdminNames
  const after = adminNames.value
  const addedNames   = after.filter(name => !before.includes(name))
  const removedNames = before.filter(name => !after.includes(name))
  const addedData = aliases.value
    .filter(a => addedNames.includes(a.aliasName))
    .map(a => ({ ...a, accessRight: "admin" }))
  const removedData = aliases.value
    .filter(a => removedNames.includes(a.aliasName))
    .map(a => ({ ...a, accessRight: null }))
  return [...addedData, ...removedData]
}

</script>

<template>
<div id="drawer_column"><Drawer v-if="aliases" :aliases="aliases" :channel="channel" /></div>
<div id="content">
  <br><br>
  <div v-if="errorMessage"> 
    <div class="errorMessage">{{errorMessage}}</div>
  </div>
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
  <div id="description" ></div>
  <div v-if="id"> このチャネルのニックネーム: {{channel.myname}} </div><br>
  <template v-if="!id">
    <input type="text" v-model="myname" placeholder="このチャネルのニックネーム" class="inputText">
    <PeopleImg v-model="myimg" />
  </template>

  <button @click="channelAdd" class="postButton" >チャネル登録</button><br>
  <button @click="channelPost" class="postButton" :disabled="!channel.channelName || !myname || iamGuest">設定変更</button><br>

  <div v-if="id && !iamGuest" class="invitation">
    <button @click="invite" class="postButton"> <span>招待URL</span> </button>
    <div>
      <label>
        <input type="radio" value="" v-model="generateInvitation" />参照
      </label>
      <label>
        <input type="radio" value="generate" v-model="generateInvitation" />生成
      </label>
      <span>&nbsp;&nbsp;</span>
      <input type="checkbox" id="guest" v-model="guest" />
      <label for="guest">ゲスト</label>
    </div>
<!--     <div v-if="generateInvitation" class="optionRight">
      <input type="number" id="until" v-model="untilDays" />
      <label for="until">日まで有効</label>      
    </div> -->
    <div> <a :href="invitationCode"> {{invitationCode}} </a> </div>
    <div> <img :src="invitationQR"></div>    
  </div>
  <div v-if="id"> <a :href="'/group/' + id + '/'">
    👪 グループアカウント作成・編集 
  </a> </div>

  <h3>ユーザー一覧</h3>
  <div v-for="d in aliases" class="aliases" :class="{ 'deleted': d.deleteFlag }">
      <a :href="d.aliasName === channel.myname
          ? `/profile/${id}/`
          : `/people/${id}/${d.aliasName}/`">
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
  <button @click="removeNames" class="postButton">ユーザー削除</button>
  <button @click="removeChannel" class="postButton">このチャネルの削除</button>

  <template v-if="id">
    <h3>スレッド一覧</h3>
    <ul v-if="threadHeads">
      <li v-for="d in threadHeads">
        <a :href="'/thread/' + d.channelID + '/' + d.parentID + '/'">
          {{d.title}}
        </a>
      </li>
      <li><a :href="newThreadURL"><button> + 新規 </button></a></li>
    </ul>
  </template>

  <div v-if="iamAdmin">
    <SelectAlias
      :channel="channel"
      :aliases="aliases.filter(a => !a.accessRight || a.accessRight === 'admin')"
      :groups="groups"
      :editable="iamAdmin"
      :placeholder="'管理ユーザー'"
      v-model="adminNames"
      class="choosePeople"
      />
    <button @click="adminEdit" class="postButton">管理者登録設定</button>
  </div>

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

