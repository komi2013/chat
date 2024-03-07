<template>
  <QuillEditor :toolbar="'#my-toolbar_' + messageID" v-model:content="editorContent[messageID]" contentType="html" >
    <template #toolbar>
      <div class="editLeft" :id="'my-toolbar_' + messageID">
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
      <div class="editRight ql-toolbar ql-snow">
        <button v-if="messageID" @click="cancelEdit(message, messageID)">⬅</button>
        <button @click="msgUpsert(messageID)">▶️</button>
      </div>
    </template>
  </QuillEditor>
</template>

<script setup>
import { ref, defineProps } from 'vue';
import { useMessagesStore } from '../stores/messages.js';

import { QuillEditor } from '@vueup/vue-quill';
import '@vueup/vue-quill/dist/vue-quill.snow.css';

import { htmlToMarkdown, markdownToHtml } from '../my/markdown.js';

const props = defineProps({
  message: Object,
  channel: Object,
  parent_id: String
});

const messageID = props.message.messageID;

const messagesStore = useMessagesStore();

let editorContent = ref({});
editorContent.value[messageID] = markdownToHtml(props.message.messageTxt);
const cancelEdit = (message, messageID) => {
  message.editFlg = false;
  messagesStore.update(message, message.messageID);
};

const msgUpsert = (messageId) => {
  // 'strong', 's', 'blockquote', 'pre', 'a', 'span' 'p',    
  // console.log(editorContent.value[messageId]);
  // console.log(htmlToMarkdown(editorContent.value[messageId]));

  const messageData = htmlToMarkdown(editorContent.value[messageId]);

  const uri = messageId ? (props.parent_id ? '/ThreadEdit/' : '/MessageEdit/') : (props.parent_id ? '/ThreadPost/' : '/MessagePost/');
  const fd = new FormData();
  fd.append('parentID', props.parent_id);
  fd.append('channelID', props.channel.channelID);
  fd.append('messageID', messageId);
  fd.append('messageTxt', messageData);
  fd.append('aliasName', props.channel.aliasName);
  const request = new Request(uri, {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      alert(reason)
    })
}

</script>

<style>
.markdown {
  padding: 4px;
}

.editLeft {
  text-align: left;
  display: inline-block;
  width: 76%;
}

.editRight {
  display: inline-block;
  width: 24%;
}

.selected {
  border: 1px solid #3498db;
}

.chgble {
  text-align: left;
  outline: 1px solid blue;
  width: 96%;
  display: inline-block;
}
.quote {
  padding: 3px;
  margin: 0 0 0 5px;
  border-left: 3px solid #ccc;
}
.editText .ql-container.ql-snow {
  border: 1px solid #d1d5db;
}
.editText .ql-editor {
  padding: 4px 0px;
}
</style>
