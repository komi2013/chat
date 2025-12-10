<template>
  <div v-show="editable">
    <div class="editLeft" :id="'toolbar_' + messageID">
      <button class="ql-bold"></button>
      <button class="ql-strike"></button>
      <button class="ql-blockquote"></button>
      <button class="ql-code-block"></button>
      <button class="ql-link"></button>
      <select class="ql-color">
        <option value="red">Red</option>
        <option value=""></option>
      </select>
      <button class="emoji" @click="attach(messageID)">🌄</button>
      <button class="emoji" v-if="messageID" @click="msgUpsert(messageID, true)">🗑</button>
      <button class="emoji" @click="tasking" :class="{ 'selected': task }">🔖</button>
      <button class="emoji" @click="asGroup = true" :class="{ 'selected': selectedGroup.groupID }">👥</button>
      <button class="emoji" v-if="showPhone" @click="callVisible = true"> ☎️ </button>
      <button class="emoji" @click="msgUpsert(messageID, false)">▶️</button>
    </div>
    <div :id="'edit_' + messageID"
      v-html="editTxt[messageID]"
      >
    </div>
  </div>
  <div class="files" v-html="fileInfo[messageID]"></div>
  <input type="file" style="position: fixed; left: -300px;" multiple :id="'fileInput_' + messageID">

  <!-- Call Modal -->
  <div v-if="callVisible" class="call-modal">
    <div class="call-dialog">
      <h3>SkyWay ビデオ通話</h3>
      <video ref="localVideo" autoplay playsinline muted></video>
      <video ref="remoteVideo" autoplay playsinline></video>
      <button @click="startSkyway">通話開始</button>
      <button @click="endSkyway">終了</button>
      状態: {{ callState }}
    </div>
  </div>

  <EditOptionModal :show="asGroup" @close="asGroup = false">
    <SelectGroup
      :groups="myGroups"
      v-model="selectedGroup"
      @update:modelValue="handleSelection"
    />
  </EditOptionModal>
</template>

<script setup>
import { ref, defineProps, onMounted } from 'vue';
import { useMessagesStore } from '@/stores/messages.js';

import Quill from 'quill';
import "quill-mention";
import "quill/dist/quill.snow.css";

import EditOptionModal from '@/components/EditOptionModal.vue';
import SelectGroup from '@/components/SelectGroup.vue';

import { htmlToMarkdown, markdownToHtml, removeMark } from '@/my/markdown.js';
import { startSkywayCall } from '@/my/skyway.js';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  channel: Object,
  aliases: Array,
  groups: Array,
  message: Object,
  threadHead: Object,
});

const message = props.message;
// console.log('message', message)
let task = ref(false);
const messageID = props.message.messageID;
const messagesStore = useMessagesStore();

let editTxt = ref({});
editTxt.value[messageID] = markdownToHtml(props.message.messageTxt, props.channel, []);

let quill
async function initQuill() {
  quill = new Quill('#edit_' + messageID, {
    modules: {
      toolbar: '#toolbar_' + messageID,
      mention: {
        allowedChars: /^[A-Za-z\sÅÄÖåäö]*$/,
        mentionDenotationChars: ["@"],
        source: function(searchTerm, renderList, mentionChar) {
          let values;
          if (mentionChar === "@") {
            let aliasForMention = props.aliases
            if (Array.isArray(props.groups)) {
              aliasForMention = props.aliases.concat(
                props.groups.map(group => ({
                  aliasID: group.groupID,
                  aliasName: group.groupName,
                  aliasImg: group.groupImg,
                }))
              );
            }

            values = aliasForMention.map((alias, index) => {
              return {
                id: alias.aliasID,
                value: alias.aliasName,
                icon: alias.aliasImg
              };
            });
          }

          if (searchTerm.length === 0) {
            renderList(values, searchTerm);
          } else {
            const matches = values.filter(item => item.value.toLowerCase().includes(searchTerm.toLowerCase()));
            renderList(matches, searchTerm);
          }
        },
        renderItem: function(item) {
          const mentionWithImage = document.createElement("div");
          if (item.icon.charAt(0) == ',') {
            const arr = item.icon.split(',');
            mentionWithImage.innerHTML = 
              `<span class="min-icon" style="background-color:${arr[2]}"><span>${arr[1]}</span></span>${item.value}`;
          } else {
            mentionWithImage.innerHTML = `<img src="${item.icon}" class="min-icon">${item.value}`;
          }
          return mentionWithImage;
        },
        onOpen: function() {
          const quillMentionList = document.getElementById('quill-mention-list');
          const rect = quillMentionList.getBoundingClientRect();
          if (rect.left > 150 && rect.left < 300) {
            quillMentionList.style.left = (- 1 * rect.left) + 'px';
          }
        }
      }
    },
    theme: 'snow'
  })
}
let editable = ref(false)
onMounted(async () => {
  editable.value = !props.threadHead.broadcastFlag || (props.threadHead.broadcastFlag && props.threadHead.adminNames.includes(localStorage.getItem('myname')))
  if (editable.value) {
    await initQuill()
  }
})

const tasking = () => {
  task.value = !task.value;
};

const attach = () => {
  const fileInput = document.getElementById('fileInput_' + messageID);
  if (fileInput) {
    fileInput.click();
  }
  fileInput.addEventListener('change', handleFileInputChange);
}

const fileInfo = ref({});
const handleFileInputChange = (event) => {
  const files = event.target.files;
  const newFileInfo = document.createElement('div');
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    const fileContainer = document.createElement('div');
    if (file.type.startsWith('image/')) {
      const image = document.createElement('img');
      image.src = URL.createObjectURL(file);
      image.style.maxWidth = '50px';
      image.style.maxHeight = '50px';
      fileContainer.appendChild(image);
    } else {
      const fileName = document.createTextNode(file.name);
      fileContainer.appendChild(fileName);
    }
    newFileInfo.appendChild(fileContainer);
  }
  fileInfo.value[messageID] = newFileInfo.outerHTML;
}

const asGroup = ref(false);
const emptyGroup = {groupID:'', groupName:'グループなし', groupImg:''}
const selectedGroup = ref(emptyGroup)
const myGroups = ref(props.groups.filter(group => group.aliasNames.includes(localStorage.getItem('myname'))))
myGroups.value.unshift(emptyGroup)
function handleSelection(group) {
  selectedGroup.value = group;
  asGroup.value = false;
}

let clicked = false;
const msgUpsert = async (messageID, delMessage) => {
  if (!messageID && quill.root.innerHTML == '<p><br></p>') {
    return;
  }
  if (clicked) return
  clicked = true;
  const messageData = delMessage ? '' : htmlToMarkdown(quill.root.innerHTML.replace(/\uFEFF/g, ''));
  const pushTitle = messageID ? 'threadEdit' : 'thread';
  const thisMsgID = messageID ? messageID : base62Encode(Math.floor(Date.now())) + generateRandomCode(1);
  const myAlias = props.aliases.find(alias => alias.aliasName === localStorage.getItem('myname'));
  let alreadyNames = props.threadHead.newThread ? [] : props.threadHead.aliasNames
  const dm = props.threadHead.parentID.includes('@')
  let toInquiryUser = false
  if (dm) {
    const splitNames = props.threadHead.parentID.split('@')
    if (splitNames[0] == "") toInquiryUser = true
    const matchedGroup = props.groups.find(group => splitNames.includes(group.groupName))
    const dmAliasNames = matchedGroup?.aliasNames || []
    const dmNames = splitNames.filter(name => name !== matchedGroup?.groupName)
    alreadyNames = [...new Set([...alreadyNames, ...dmNames, ...dmAliasNames])]
  }
  // const alreadyUserIDs = new Set(userIDsByName(props.aliases, alreadyNames))
  // console.log('alreadyUserIDs', alreadyUserIDs)
  let yets = [];
  let mentionNames = [];
  if (!dm) {
    const addMentions = (names) => {
      for (const aliasName of names) {
        mentionNames.push(aliasName);
        if (task.value) {
          yets.push({ aliasName, emoji: '☑️' });
        }
      }
    };
    if (Array.isArray(props.groups)) {
      for (const g of props.groups) {
        if (messageData.includes(`＠＠${g.groupName}・＠＠`)) {
          addMentions(g.aliasNames);
        }
      }
    }
    for (const a of props.aliases) {
      if (messageData.includes(`＠＠${a.aliasName}・＠＠`)) {
        addMentions([a.aliasName]);
      }
    }
  }

  // const newUserIDs = userIDsByName(props.aliases, newNames);
  // const alreadyUserIDsSet = new Set(alreadyUserIDs)
  const totalNames = [...new Set([...mentionNames, ...alreadyNames, localStorage.getItem('myname')])]
  // const uniqueNewIDs = newUserIDs.filter(id => !alreadyUserIDs.has(id));
  const fileInput = document.getElementById('fileInput_' + messageID);
  if (fileInput && fileInput.files.length > 10) {
    alert('too many files');
    return;
  }
  const fd = new FormData();
  if (fileInput && fileInput.files.length > 0) {
    for (const file of fileInput.files) {
      fd.append('files[]', file);
    }
  }

  fd.append('channelID', localStorage.getItem('channelID'));
  fd.append('updatedBy', localStorage.getItem('myname'));
  // fd.append('userIDs', JSON.stringify([...new Set([...newUserIDs, ...alreadyUserIDs])]));
  // const userIDs = JSON.stringify([...new Set([...newUserIDs, ...alreadyUserIDs])])
  // const userIDs = JSON.stringify(userIDsByName(props.aliases, totalNames))
  fd.set('pushNames', JSON.stringify(totalNames))
  let editThreadHead = props.threadHead
  editThreadHead.aliasNames = totalNames
  if (mentionNames.length > 0 || editThreadHead.newReply) {
    if (editThreadHead.newReply) {
      delete editThreadHead.newReply
      editThreadHead.newThread = true
    }
    fd.set('contents', JSON.stringify(editThreadHead));
    fd.set('pushTitle', 'threadHead');
    fd.set('csrf', localStorage.getItem("csrf"));
    const res = await sendRequest('/ContentsPush/', fd);
    res.csrf && localStorage.setItem('csrf', res.csrf);
    if (Array.isArray(res.pushContents)) {
      for (const content of res.pushContents) {
        await pushReceive(content)
      }
    }
  }
  // console.log('props.threadHead.newThread, userIDs, totalNames', props.threadHead.newThread, userIDs, totalNames)
  if (editThreadHead.newThread) {
    editThreadHead.messageTxt = messageData
    editThreadHead.aliasImg = myAlias.aliasImg
    editThreadHead.title = getSubstring(removeMark(messageData), 0, 30)
    editThreadHead.adminNames = [localStorage.getItem('myname')]
    // delete editThreadHead.newThread
    fd.set('pushTitle', 'threadHead')
    fd.set('contents', JSON.stringify(editThreadHead))
  } else {
    const contents = [
      props.message.parentID,
      thisMsgID,
      messageData,
      selectedGroup.value.groupImg ? selectedGroup.value.groupImg : myAlias.aliasImg,
      totalNames,
      props.threadHead.backID ?? '',
      yets,
      selectedGroup.value.groupName === 'グループなし' ? '' : selectedGroup.value.groupName,
      ...(messageID ? [Math.floor(Date.now() / 1000)] : [])
    ]
    fd.set('contents', JSON.stringify(contents))
    fd.set('pushTitle', pushTitle)
  }
  fd.set('csrf', localStorage.getItem("csrf"));
  const uri = toInquiryUser ? '/ReceptionThreadCustomer/' : '/ContentsPush/'
  const res = await sendRequest('/ContentsPush/', fd);
  res.csrf && localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  quill.root.innerHTML = '';
  fileInfo.value = [];
  task.value = false;
  asGroup.value = false;
  clicked = false;
  if (props.threadHead.newThread) { location.href = '' }
}

const SKYWAY_API_KEY = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiJhOWVkMGMxZi0xMzFmLTQzNGEtYjgzOS04ZmE0ZmE0ZmMyMGMiLCJpYXQiOjE3NjUzNDkxODAsImV4cCI6MTc5Njg4NTE4MCwic2NvcGUiOnsiYXBwIjp7ImlkIjoiYzBiMmE0MTktOWMzYy00MDhmLWFiMzItNjcxNDA3ZDZlM2FkIiwidHVybiI6dHJ1ZSwiYWN0aW9ucyI6WyJyZWFkIl0sImNoYW5uZWxzIjpbeyJpZCI6IioiLCJuYW1lIjoiKiIsImFjdGlvbnMiOlsid3JpdGUiXSwibWVtYmVycyI6W3siaWQiOiIqIiwibmFtZSI6IioiLCJhY3Rpb25zIjpbIndyaXRlIl0sInB1YmxpY2F0aW9uIjp7ImFjdGlvbnMiOlsid3JpdGUiXX0sInN1YnNjcmlwdGlvbiI6eyJhY3Rpb25zIjpbIndyaXRlIl19fV0sInNmdUJvdHMiOlt7ImFjdGlvbnMiOlsid3JpdGUiXSwiZm9yd2FyZGluZ3MiOlt7ImFjdGlvbnMiOlsid3JpdGUiXX1dfV19XX19fQ.J_zpQtJuKsvRtz8ss0xh8J6cma651jFV-Yiqpou_CjI';

console.log("STEP1 token(raw) =", SKYWAY_API_KEY);

const callVisible = ref(false);
const callState = ref('idle');

const localVideo = ref(null);
const remoteVideo = ref(null);

let session = null;
const showPhone = ref(true)
// showPhone.value = true

const startSkyway = async () => {
  callState.value = 'connecting';

  const token = SKYWAY_API_KEY; // ← token に名前変更したほうが良い
console.log("STEP2 token(before startSkywayCall) =", token);

  session = await startSkywayCall({
    token,                                // ← 修正
    roomName: generateRoomName(),
    localVideoEl: localVideo.value,
    remoteVideoEl: remoteVideo.value,
    onStatus: (status) => callState.value = status
  });
};

const endSkyway = () => {
  if (session) {
    session.end();
    session = null;
  }
  callState.value = 'ended';
  callVisible.value = false;
};

function generateRoomName() {
  // DM の相手と自分の名前からルーム名生成（順番固定）
  const me = localStorage.getItem('myname');
  const peer = determineRemoteName();
  return ['dm', me, peer].sort().join('_');
}

function determineRemoteName(){
  // threadHead.parentID の構成から相手の名前を取り出す（あなたのアプリの命名規則に合わせて調整してください）
  // current user:
  const me = localStorage.getItem('myname');
  const splitNames = props.threadHead.parentID.split('@').filter(Boolean);
  const groupNames = (Array.isArray(props.groups) ? props.groups.map(g => g.groupName) : []);
  const nonGroup = splitNames.filter(n => !groupNames.includes(n));
  // 非自分の名前を返す
  if (nonGroup.length === 2) {
    return nonGroup.find(n => n !== me);
  }
  return nonGroup.length === 1 ? nonGroup[0] : '';
}

</script>

<style>

.editLeft {
  display: inline-block;
  width: 99%;
}

/*.editRight {
  display: inline-block;
  width: 30%;
}
*/
.editText .ql-container.ql-snow {
  border: 1px solid #d1d5db;
  border-bottom-width: 0;
}
.editText .ql-editor {
  padding: 4px 0px;
}

.files {
  border-top: none;
  border-right: 1px solid #d1d5db;
  border-bottom: 1px solid #d1d5db;
  border-left: 1px solid #d1d5db;
}
.ql-snow.ql-toolbar {
  padding: 8px 0px;
}
.ql-snow.ql-toolbar .emoji {
  font-size: 12px;
  padding-top: 0px;
}
.ql-snow.ql-toolbar .selected {
  background-color: #92a7b54a;
  border-radius: 5px;
}
.ql-mention-list-container {
  background-color: white;
  bottom: 0px;
}

.ql-mention-list {
  display: flex;
  flex-direction: column;
  position: absolute;
  left: -30px;
  width: 300px;
  bottom: 0px;
}

.ql-mention-list-item {
  display: flex;
  align-items: center;
  background-color: white;
}

.ql-mention-list-item-text {
  margin-left: 8px;
}

.ql-mention-list-item-image {
  width: 24px;
  height: 24px;
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

.mention {
  background-color: #a7cad63d;
  color: blue;
}

</style>
