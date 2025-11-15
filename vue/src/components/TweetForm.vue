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
      <button v-if="!alreadyJoinFlag" class="emoji" @click="asAnonymous" :class="{ 'selected': anonymous }">🎭</button>
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

import { emojiRanges, isEmojiInRange, getRandomEmoji, getRandomColor } from '@/my/emoji';
import { htmlToMarkdown, markdownToHtml, removeMark } from '@/my/markdown.js';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  message: Object,
  threadHead: Object,
  backID: String,
  newTweet: Boolean,
  alreadyJoinFlag: Boolean,
});

const message = props.message;
let task = ref(false);
const messageID = message.messageID
const parentID = message.parentID
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
      const formats = quill.getFormat(index, url.length);
      if (!formats.link) {
        quill.formatText(index, url.length, 'link', url);
      }
    }
  });
}


onMounted(async () => {
  if ( !props.alreadyJoinFlag ) {
    removeAnonymous()
  }
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
  const imgOK = await getImagePath(messageID) // make imgPath
  if (!imgOK) return
  clicked = true
  const messageData = htmlToMarkdown(quill.root.innerHTML.replace(/\uFEFF/g, ''));
  const addMentions = (names) => {
    for (const aliasName of names) {
      mentionNames.push(aliasName)
    }
  }
  let mentionNames = [];
  for (const name of props.threadHead.nicknames) {
    if (messageData.includes(`＠＠${name}・＠＠`)) {
      addMentions([name])
    }
  }
  const fd = new FormData()
  fd.append('parentID', parentID ?? '')
  fd.append('messageTxt', messageData)
  fd.append('backID', props.backID ?? '')
  fd.append('csrf', localStorage.getItem('csrf'))
  fd.append('imgPath', imgPath)
  fd.append('anonymous', anonymous.value ?? '')
  fd.append('anonymousImg', anonymousImg.value ?? '')
  const res = await sendRequest('/TweetPost/', fd)
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
  if (res.error) {
    errorMessage.value = res.error 
    return
  }
  if (props.newTweet) {
    const backIDURL = props.backID ? '?backID=' + props.backID : ''
    location.href = `/tweet/${res.date}/${res.parentID}.html${backIDURL}`
  } else {
    location.href = ''
  }
}

let imgPath = ''
async function getImagePath(messageID) {
  const fileInput = document.getElementById('fileInput_' + messageID);
  if (!fileInput || fileInput.files.length === 0) {
    return true
  }
  if (fileInput.files.length > 1) {
    alert('アップロードできるのは画像1枚のみです');
    return false
  }
  const file = fileInput.files[0];
  if (!file.type.startsWith('image/')) {
    alert('画像ファイルのみアップロード可能です');
    return false
  }
  imgPath = await new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      const img = new Image();
      img.onload = () => {
        const maxSize = 250;
        let width = img.width;
        let height = img.height;
        if (width > height) {
          if (width > maxSize) {
            height = Math.round(height * (maxSize / width));
            width = maxSize;
          }
        } else {
          if (height > maxSize) {
            width = Math.round(width * (maxSize / height));
            height = maxSize;
          }
        }
        const canvas = document.createElement('canvas');
        canvas.width = width;
        canvas.height = height;
        const ctx = canvas.getContext('2d');
        ctx.drawImage(img, 0, 0, width, height);
        const dataURL = canvas.toDataURL('image/jpeg', 0.9);
        resolve(dataURL);
      };
      img.onerror = reject;
      img.src = e.target.result;
    };
    reader.onerror = reject;
    reader.readAsDataURL(file);
  });

  return true
}

const backIDURL = props.backID ? '?backID=' + props.backID : ''
const anonymous = ref(localStorage.getItem('anonymous'))
const anonymousImg = ref(localStorage.getItem('anonymousImg'))

function generateRandomName() {
  const firstNames = ['Fox', 'Wolf', 'Star', 'Light', 'River', 'Aqua', 'Comet', 'Nova', 'Ember', 'Blue', 'Yellow']
  const lastNames = ['子', '丑', '寅', '卯', '辰', '巳', '午', '未', '申', '酉', '戌', '亥']
  const first = firstNames[Math.floor(Math.random() * firstNames.length)]
  const last = lastNames[Math.floor(Math.random() * lastNames.length)]
  const num = Math.floor(Math.random() * 999)
  return `${first}${num}${last}`
}

function getAliasImg() {
  return "," + getRandomEmoji() + "," + getRandomColor()
}

function asAnonymous() {
  if (anonymous.value) {
    removeAnonymous()
  } else {
    anonymous.value = generateRandomName()
    localStorage.setItem('anonymous', anonymous.value)
    anonymousImg.value = getAliasImg()
    localStorage.setItem('anonymousImg', anonymousImg.value)
  }
}

function removeAnonymous() {
  anonymous.value = ''
  localStorage.removeItem('anonymous')
  anonymousImg.value = ''
  localStorage.removeItem('anonymousImg')
}
</script>

<style>

.editLeft {
  display: inline-block;
  width: 99%;
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
