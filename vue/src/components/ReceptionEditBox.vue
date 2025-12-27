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
      <button class="emoji" @click="msgUpsert(messageID)">▶️</button>
    </div>
    <div :id="'edit_' + messageID"
      v-html="editTxt[messageID]"
      >
    </div>
  </div>
  <div class="files" v-html="fileInfo[messageID]"></div>
  <input type="file" style="position: fixed; left: -300px;" multiple :id="'fileInput_' + messageID">
  <div v-if="errorMessage" class="errorMessage">{{ errorMessage }}</div>
</template>

<script setup>
import { ref, defineProps, onMounted } from 'vue';
import { useMessagesStore } from '@/stores/messages.js';

import Quill from 'quill';
import "quill/dist/quill.snow.css";

import { htmlToMarkdown, markdownToHtml, removeMark } from '@/my/markdown.js';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  channel: {
    type: Object,
    default: () => ({
      channelID: '',
      myname: 'お客様',
      channelName: '問い合わせ対応'
    })
  },
  message: {
    type: Object,
    default: () => ({
      messageTxt: '',
      messageID: '',
      parentID: ''
    })
  },
  threadHead: {
    type: Object,
    default: () => ({
      receptionID: '',
      parentID: '',
      title: '新規問い合わせ'
    })
  }
})

const errorMessage = ref('')
const messagesStore = useMessagesStore();
const messageID = props.message.messageID
const editTxt = ref({});
const fileInfo = ref({});
const quill = ref({});
const editable = ref(true);

onMounted(async () => {
  if (messageID) {
    const message = messagesStore.messages.find(m => m.messageID === messageID);
    if (message) {
      editTxt.value[messageID] = markdownToHtml(message.messageTxt, props.channel);
    }
  }
  await initQuill();
});

async function initQuill() {
  const toolbar = document.getElementById('toolbar_' + messageID);
  if (!toolbar) return;
  
  quill.value[messageID] = new Quill('#edit_' + messageID, {
    modules: {
      toolbar: '#toolbar_' + messageID,
    },
    theme: 'snow'
  });
  
  if (editTxt.value[messageID]) {
    quill.value[messageID].root.innerHTML = editTxt.value[messageID];
  }
}

function attach(messageID) {
  const fileInput = document.getElementById('fileInput_' + messageID);
  fileInput.click();
  fileInput.onchange = (e) => {
    const files = Array.from(e.target.files);
    files.forEach(file => {
      uploadFile(file, messageID);
    });
  };
}

async function uploadFile(file, messageID) {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('csrf', localStorage.getItem('csrf'));
  
  const response = await fetch('/upload/', {
    method: 'POST',
    body: formData
  });
  const result = await response.json();
  
  if (result.filelink) {
    const currentContent = quill.value[messageID].root.innerHTML;
    quill.value[messageID].root.innerHTML = currentContent + `＊f＊${result.filelink}・＊f＊ `;
  }
}

let clicked = false
async function msgUpsert(messageID) {
  if (quill.value[messageID].root.innerHTML == '<p><br></p>') return
  if (clicked) return
  clicked = true
  const content = quill.value[messageID].root.innerHTML.replace(/\uFEFF/g, '');
  const messageTxt = htmlToMarkdown(content);
  const thisMsgID = messageID ? messageID : base62Encode(Math.floor(Date.now())) + generateRandomCode(1);
  const fd = new FormData()
  let editThreadHead = props.threadHead
  console.log('editThreadHead', editThreadHead)
  fd.append('channelID', props.channel.channelID)
  fd.append('messageID', thisMsgID);
  fd.append('parentID', props.message.parentID);
  fd.append('messageTxt', messageTxt);
  fd.append('csrf', localStorage.getItem('csrf'))
  if (editThreadHead.newThread || editThreadHead.newReply) {
    editThreadHead.messageTxt = messageTxt
    editThreadHead.newThread = true
    delete editThreadHead.newReply
    fd.set('threadHead', JSON.stringify(editThreadHead))
  }
  const res = await sendRequest('/ReceptionThreadCustomer/', fd);
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
  if (res.error) { errorMessage.value = res.error }
  quill.value[messageID].root.innerHTML = ''
  fileInfo.value = []
  clicked = false
  if (props.threadHead.newThread) { location.href = '' }
}

</script>

<style>
.editLeft {
  display: flex;
  gap: 5px;
  margin-bottom: 10px;
}

.editLeft button {
  padding: 5px 10px;
  border: 1px solid #ccc;
  background: white;
  cursor: pointer;
  border-radius: 3px;
}

.editLeft button:hover {
  background: #f0f0f0;
}

.editLeft button.selected {
  background: #007bff;
  color: white;
}

.ql-snow.ql-toolbar {
  padding: 8px 0px;
}

.ql-snow.ql-toolbar .attachment {
  font-size: 12px;
  padding-top: 0px;
}

.files {
  margin-top: 10px;
}

.emoji {
  font-size: 16px;
}

/* エディターのスタイル */
#edit_new {
  min-height: 100px;
  border: 1px solid #ccc;
  border-radius: 4px;
}

</style>
