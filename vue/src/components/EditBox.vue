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
        📎
      </button>
    </div>
    <div class="editRight ql-toolbar ql-snow">
      <button v-if="messageID" @click="msgUpsert(messageID, true)">🗑</button>
      <button @click="task" :class="{ 'task': message.task }">🔖</button>
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
  message: Object,
  channel: Object,
  threadHead: Object
});

const message = props.message;
const messageID = props.message.messageID;
const messagesStore = useMessagesStore();

let editTxt = ref({});
editTxt.value[messageID] = markdownToHtml(props.message.messageTxt, props.channel);
// console.log('props.message.messageTxt' , props.message.messageTxt);
// console.log('editTxt.value[messageID]' , editTxt.value[messageID]);
const task = () => {
  message.task = !message.task;
  console.log(message);
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

const msgUpsert = (messageID, delMessage) => {
  const messageData = delMessage ? '' : htmlToMarkdown(quill.root.innerHTML);
  const threadFlg = props.message.parentID;
  const uri = messageID ? (threadFlg ? '/ThreadEdit/' : '/MessageEdit/') : (threadFlg ? '/ThreadPost/' : '/MessagePost/');
  const fd = new FormData();
  // fd.append('parentID', props.message.parentID);
  fd.append('channelID', props.channel.channelID);
  fd.append('messageID', messageID);
  fd.append('messageTxt', messageData);
  fd.append('aliasName', props.channel.aliasName);
  if (props.threadHead && props.threadHead.backID) {
    fd.append('backID', props.threadHead.backID);
  }
  if (message.task) {
    fd.append('task', 1);
  }
  const fileInput = document.getElementById('fileInput_' + messageID);
  if (fileInput && fileInput.files.length > 0) {
    for (const file of fileInput.files) {
      fd.append('files[]', file);
    }
  }

  const request = new Request(uri, {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      alert(reason)
    })
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
            values = props.channel.aliasArray.map((alias, index) => {
              return {
                id: index + 1,
                value: alias[0],
                imageUrl: alias[1]
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
  text-align: left;
  display: inline-block;
  width: 76%;
}

.editRight {
  display: inline-block;
  width: 24%;
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
