<template>
  <div>
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
      <button class="attachment" @click="attach(messageID)">
        🌄
      </button>
    </div>
    <div class="editRight ql-toolbar ql-snow">
      <button v-if="messageID" @click="msgUpsert(messageID, true)">🗑</button>
      <button @click="tasking" :class="{ 'task': task }">🔖</button>
      <button @click="msgUpsert(messageID, false)">▶️</button>
    </div>
    <div :id="'edit_' + messageID"
      v-html="editTxt[messageID]"
      >
    </div>
  </div>
  <div class="files" v-html="fileInfo[messageID]"></div>
  <input type="file" style="position: fixed; left: -300px;" multiple :id="'fileInput_' + messageID">
</template>

<script setup>
import { ref, defineProps, onMounted } from 'vue';
import { useMessagesStore } from '../stores/messages.js';

import { htmlToMarkdown, markdownToHtml } from '../my/markdown.js';

import Quill from 'quill';
import "quill-mention";
import "quill/dist/quill.snow.css";

const props = defineProps({
  channel: Object,
  aliases: Array,
  groups: Array,
  message: Object,
  threadHead: Object,
});

const message = props.message;
let task = ref(false);
const messageID = props.message.messageID;
const messagesStore = useMessagesStore();

let editTxt = ref({});
editTxt.value[messageID] = markdownToHtml(props.message.messageTxt, props.channel, []);
// console.log('props.message.messageTxt' , props.message.messageTxt);
// console.log('editTxt.value[messageID]' , editTxt.value[messageID]);
const tasking = () => {
  task.value = !task.value;
};

const attach = () => {
  const fileInput = document.getElementById('fileInput_' + messageID);
  if (fileInput) {
    fileInput.click();
  }

  // File input要素にchangeイベントリスナーを追加
  // const fileInput = document.getElementById('fileInput_');
  fileInput.addEventListener('change', handleFileInputChange);

}

let dm = false;
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
let clicked = false;
const msgUpsert = (messageID, delMessage) => {
  if (!messageID && quill.root.innerHTML == '<p><br></p>') {
    return;
  }
  if (clicked) {
    return;
  }
  clicked = true;
  const messageData = delMessage ? '' : htmlToMarkdown(quill.root.innerHTML.replace(/\uFEFF/g, ''));
  // const threadFlg = props.message.parentID;
  const uri = messageID ? '/ThreadEdit/' : '/ThreadPost/';
  const fd = new FormData();
  console.log(messageData);
  fd.append('parentID', props.message.parentID);
  fd.append('channelID', props.channel.channelID);
  fd.append('messageID', messageID);
  fd.append('messageTxt', messageData);
  fd.append('postedBy', props.channel.myname);
  const alias = props.aliases.find(alias => alias.aliasName === props.channel.myname);
  console.log('いいい', alias.aliasImg);
  fd.append('aliasImg', getAliasImg(props));
  let userIDs = [];
  let names = [props.channel.aliasName];
  if (props.threadHead) {
    fd.append('backID', props.threadHead.backID);
    fd.append('type', props.threadHead.threadType);
    dm = props.threadHead.parentID.includes('@');
    if (dm) {
      names = props.threadHead.parentID.split('@');
    }
    userIDs = props.aliases
      .filter(alias => props.threadHead.aliasNames.includes(alias.aliasName))
      .map(alias => alias.userID);
  }
  userIDs.push(localStorage.getItem("userID"));
  if (Array.isArray(props.groups) && !dm) {
    for (const d of props.groups) {
      const atName = `＠＠${d[0]}・＠＠`;
      if (messageData.includes(atName)) {
        for (const d2 of d[2]) {
          names.push(d2);
        }
      }
    }
  }
  let yets = [];
  for (const d of props.aliases) {
    if (names.includes(d.aliasName)) {
      userIDs.push(d.userID);
    }
    const atName = `＠＠${d.aliasName}・＠＠`;
    if (messageData.includes(atName) && !dm) {
      userIDs.push(d.userID);
      names.push(d.aliasName);
      yets.push([d.aliasName, '/img/yet.png']);
    }
  }
  fd.append('names', JSON.stringify([...new Set(names)]));
  fd.append('userIDs', JSON.stringify([...new Set(userIDs)]));
  if (task.value) {
    fd.append('yets', JSON.stringify(yets));
  }
  const fileInput = document.getElementById('fileInput_' + messageID);
  if (fileInput && fileInput.files.length > 10) {
    alert('too many files');
    return;
  }
  if (fileInput && fileInput.files.length > 0) {
    for (const file of fileInput.files) {
      fd.append('files[]', file);
    }
  }
  if (!messageID) {
    quill.root.innerHTML = '';
    fileInfo.value = [];
  }
  const request = new Request(uri, {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .then(function(response) {
      clicked = false;
    })
}

function getAliasImg(props) {
  let aliasImg = null;
  const myname = props.channel.myname;
  // props.aliases.forEach(d => {

  // });
  // const found = array1.find((element) => element > 10);
  for (let i = 0; i < props.aliases.length; i++) {
    console.log(props.aliases[i].aliasName, '===', myname);
    if (props.aliases[i].aliasName === myname) {
      aliasImg = props.aliases[i].aliasImg;
      break;
    }
  }
  return aliasImg;
}

let quill;
onMounted(() => {
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
                  aliasImg: group.groupImg, // 必要に応じてプロパティ名を調整
                }))
              );
            }

            values = aliasForMention.map((alias, index) => {
              return {
                id: alias.aliasID,
                value: alias.aliasName,
                imageUrl: alias.aliasImg
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
          // console.log(position);
          const mentionWithImage = document.createElement("div");
          // mentionWithImage.style.left = 0;
          mentionWithImage.innerHTML = `<img src="${item.imageUrl}" class="iconMini">${item.value}`;
          return mentionWithImage;
        },
        onOpen: function() {
          const quillMentionList = document.getElementById('quill-mention-list');
          const rect = quillMentionList.getBoundingClientRect();
          console.log(rect);
          if (rect.left > 150 && rect.left < 300) {
            quillMentionList.style.left = (- 1 * rect.left) + 'px';
          }
        }
      }
    },
    theme: 'snow'
  });
});

</script>

<style>

.editLeft {
  display: inline-block;
  width: 69%;
}

.editRight {
  display: inline-block;
  width: 30%;
}

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
.ql-snow.ql-toolbar .attachment {
  font-size: 12px;
  padding-top: 0px;
}
.ql-snow.ql-toolbar .task {
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

.mention {
  background-color: #a7cad63d;
  color: blue;
}
</style>
