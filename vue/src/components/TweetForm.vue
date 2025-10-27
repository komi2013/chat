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
      <button class="emoji" @click="attach(messageID)">🌄</button>
      <button class="emoji" v-if="messageID" @click="msgUpsert(messageID)">🗑</button>
      <button class="emoji" @click="msgUpsert(messageID, false)">▶️</button>
    </div>
    <div :id="'edit_' + messageID"
      v-html="editTxt[messageID]"
      >
    </div>
  </div>
  <div v-if="errorMessage">
    <p style="color:red;">{{ errorMessage }}</p>
  </div>
  <div class="files" v-html="fileInfo[messageID]"></div>
  <input type="file" style="position: fixed; left: -300px;" multiple :id="'fileInput_' + messageID">
  <p style="text-align:right; font-size:12px; color:gray;">
    {{ remainingChars }} / 280
  </p>
</template>

<script setup>
import { ref, defineProps, onMounted } from 'vue';
import { useMessagesStore } from '@/stores/messages.js';

import Quill from 'quill';
import "quill-mention";
import "quill/dist/quill.snow.css";

import EditOptionModal from '@/components/EditOptionModal.vue';

import { htmlToMarkdown, markdownToHtml, removeMark } from '@/my/markdown.js';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  message: Object,
  threadHead: Object,
  backID: String,
});

const message = props.message;
// console.log('message', message)
let task = ref(false);
const messageID = message.messageID
const parentID = message.parentID
// const messagesStore = useMessagesStore();

let editTxt = ref({});
editTxt.value[messageID] = markdownToHtml(props.message.messageTxt)

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
            let aliasForMention = props.threadHead.nicknames
            values = aliasForMention.map((name) => {
              return {
                id: name,
                value: name,
                icon: ''
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
          return item.value
          // const mentionWithImage = document.createElement("div");
          // if (item.icon.charAt(0) == ',') {
          //   const arr = item.icon.split(',');
          //   mentionWithImage.innerHTML = 
          //     `<span class="min-icon" style="background-color:${arr[2]}"><span>${arr[1]}</span></span>${item.value}`;
          // } else {
          //   mentionWithImage.innerHTML = `<img src="${item.icon}" class="min-icon">${item.value}`;
          // }
          // return mentionWithImage;
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
  handleTextLimit()
  handleAutoLink()
}

const remainingChars = ref(280)
function handleTextLimit() {
  quill.on('text-change', () => {
    const text = quill.getText().trimEnd()
    if (text.length > 280) {
      // 制限を超えた部分を削除（過剰入力を防ぐ）
      quill.deleteText(280, text.length)
    }
    remainingChars.value = 280 - quill.getLength() + 1
  })
}

function handleAutoLink() {
  quill.on('text-change', (delta, oldDelta, source) => {
    if (source !== 'user') return;
    const text = quill.getText();
    const urlRegex = /(https?:\/\/[^\s]+)/g;
    let match;
    while ((match = urlRegex.exec(text)) !== null) {
      const url = match[0];
      const index = match.index;
      // すでにリンクが設定されていない場合のみリンク化
      const formats = quill.getFormat(index, url.length);
      if (!formats.link) {
        quill.formatText(index, url.length, 'link', url);
      }
    }
  });
}


onMounted(async () => {
  await initQuill()
})

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

let clicked = false
const errorMessage = ref('')
const msgUpsert = async (messageID) => {
  if (quill.root.innerHTML == '<p><br></p>') return
  if (clicked) return
  clicked = true
  const messageData = htmlToMarkdown(quill.root.innerHTML.replace(/\uFEFF/g, ''));
  let mentionNames = [];
  for (const name of props.threadHead.nicknames) {
    if (messageData.includes(`＠＠${name}・＠＠`)) {
      addMentions([name])
    }
  }
  const addMentions = (names) => {
    for (const aliasName of names) {
      mentionNames.push(aliasName)
    }
  }
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
  fd.append('parentID', parentID ?? '')
  fd.append('messageTxt', messageData)
  fd.append('backID', props.backID ?? '')
  fd.append('csrf', localStorage.getItem('csrf'))
  const res = await sendRequest('/TweetPost/', fd)
  res.csrf && localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  quill.root.innerHTML = ''
  fileInfo.value = []
  clicked = false;
  if (res.error) {
    errorMessage.value = res.error
    return
  }
  const backIDURL = props.backID ? '?backID=' + props.backID : ''
  // console.log(`/tweet/${res.date}/${res.parentID}.html${backIDURL}`)
  location.href = `/tweet/${res.date}/${res.parentID}.html${backIDURL}`

  // if (props.threadHead.newThread) {
  //   location.href = ''
  // }
}


  const backIDURL = props.backID ? '?backID=' + props.backID : ''
  
  console.log(`/tweet/.html${backIDURL}`)
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

.ql-editor a {
  color: #007bff !important; /* 任意のリンク色に変更 */
}

.ql-editor a:hover {
  color: #0056b3 !important;
}
</style>
