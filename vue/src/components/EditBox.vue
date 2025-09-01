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
      <button class="emoji" @click="msgUpsert(messageID, false)">▶️</button>
    </div>
    <div :id="'edit_' + messageID"
      v-html="editTxt[messageID]"
      >
    </div>
  </div>
  <div class="files" v-html="fileInfo[messageID]"></div>
  <input type="file" style="position: fixed; left: -300px;" multiple :id="'fileInput_' + messageID">

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
import { userIDsByName } from '@/my/channelFunc';
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
  // console.log('editable', editable)
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

let dm = false;
let clicked = false;
const msgUpsert = async (messageID, delMessage) => {
  if (!messageID && quill.root.innerHTML == '<p><br></p>') {
    return;
  }
  if (clicked) {
    return;
  }
  clicked = true;
  const messageData = delMessage ? '' : htmlToMarkdown(quill.root.innerHTML.replace(/\uFEFF/g, ''));
  const pushTitle = messageID ? 'threadEdit' : 'thread';
  const thisMsgID = messageID ? messageID : base62Encode(Math.floor(Date.now())) + generateRandomCode(1);
  const myAlias = props.aliases.find(alias => alias.aliasName === localStorage.getItem('myname'));
  let alreadyNames = props.threadHead.newThread ? [] : props.threadHead.aliasNames
  dm = props.threadHead.parentID.includes('@');
  if (dm) {
    const splitNames = props.threadHead.parentID.split('@')
    const matchedGroup = props.groups.find(group => splitNames.includes(group.groupName))
    const dmAliasNames = matchedGroup?.aliasNames || []
    const dmNames = splitNames.filter(name => name !== matchedGroup?.groupName)
    alreadyNames = [...new Set([...alreadyNames, ...dmNames, ...dmAliasNames])]
  }
  const alreadyUserIDs = new Set(userIDsByName(props.aliases, alreadyNames))
  console.log('alreadyUserIDs', alreadyUserIDs)
  let yets = [];
  let newNames = []
  if (Array.isArray(props.groups) && !dm) {
    for (const d of props.groups) {
      const atName = `＠＠${d.groupName}・＠＠`;
      if (messageData.includes(atName)) {
        for (const d2 of d.aliasNames) {
          newNames.push(d2);
          if (task.value) {
            yets.push({
              aliasName: d2,
              emoji: '☑️'
            });       
          }
        }
      }
    }
  }
  for (const d of props.aliases) {
    const atName = `＠＠${d.aliasName}・＠＠`;
    if (messageData.includes(atName) && !dm) {
      if (!dm) {
        newNames.push(d.aliasName);        
      }
      if (task.value) {
        yets.push({
          aliasName: d.aliasName,
          emoji: '☑️'
        });       
      }
    }
  }
  const newUserIDs = userIDsByName(props.aliases, newNames);
  // const alreadyUserIDsSet = new Set(alreadyUserIDs)
  const totalNames = [...new Set([...newNames, ...alreadyNames, localStorage.getItem('myname')])]
  const uniqueNewIDs = newUserIDs.filter(id => !alreadyUserIDs.has(id));
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
  const userIDs = JSON.stringify(userIDsByName(props.aliases, totalNames))
  fd.set('userIDs', userIDs)
  let editThreadHead = props.threadHead
  editThreadHead.joinNames = [...(editThreadHead.joinNames ?? [localStorage.getItem('myname')]), ...newNames]
  editThreadHead.aliasNames = totalNames
  console.log('editThreadHead.newReply', editThreadHead.newReply)
  if (uniqueNewIDs.length > 0 || editThreadHead.newReply) {
    delete editThreadHead.newReply
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
  console.log('props.threadHead.newThread, userIDs, totalNames', props.threadHead.newThread, userIDs, totalNames)
  if (props.threadHead.newThread) {
    editThreadHead.messageTxt = messageData
    editThreadHead.aliasImg = myAlias.aliasImg
    editThreadHead.title = getSubstring(removeMark(messageData), 0, 30)
    editThreadHead.adminNames = [localStorage.getItem('myname')]
    delete editThreadHead.newThread
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
