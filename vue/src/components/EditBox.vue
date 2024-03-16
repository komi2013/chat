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
    </div>
    <div class="editRight ql-toolbar ql-snow">
      <button v-if="messageID" @click="cancelEdit(message, messageID)">⬅</button>
      <button @click="msgUpsert(messageID)">▶️</button>
    </div>
    <div :id="'edit_' + messageID"
      v-html="editTxt[messageID]"
      ref="editor">
    </div>
  </div>
</template>

<script setup>
import { ref, defineProps, onMounted } from 'vue';
import { useMessagesStore } from '../stores/messages.js';

import { htmlToMarkdown, markdownToHtml } from '../my/markdown.js';

import Quill from 'quill';
import "quill-mention";
import "quill/dist/quill.snow.css";
// import { MentionBlot } from '../my/MentionBlot.js';

const props = defineProps({
  message: Object,
  channel: Object,
  threadHead: Object
});

const messageID = props.message.messageID;
const messagesStore = useMessagesStore();

let editTxt = ref({});
editTxt.value[messageID] = markdownToHtml(props.message.messageTxt);
// console.log(editTxt.value[messageID]);
const cancelEdit = (message, messageID) => {
  message.editFlg = false;
  messagesStore.update(message, message.messageID);
};

const msgUpsert = (messageID) => {
  const messageData = htmlToMarkdown(editTxt.value[messageID]);
  const threadFlg = props.message.parentID;
  const uri = messageID ? (threadFlg ? '/ThreadEdit/' : '/MessageEdit/') : (threadFlg ? '/ThreadPost/' : '/MessagePost/');
  const fd = new FormData();
  fd.append('parentID', props.message.parentID);
  fd.append('channelID', props.channel.channelID);
  fd.append('messageID', messageID);
  fd.append('messageTxt', messageData);
  fd.append('aliasName', props.channel.aliasName);
  if (props.threadHead && props.threadHead.backID) {
    fd.append('backID', props.threadHead.backID);
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

let aliasArray = [
  ['sei1', '/me.jpg'],
  ['ivan1', '/ivan.png'],
  ['group2', '/group.png']
];

let editor = ref(null);

// Quill.register(MentionBlot);

const MentionBlot = Quill.import("blots/mention");

class StyledMentionBlot extends MentionBlot {
  static render(data) {
    const element = document.createElement('span');
    element.innerText = data.value;
    element.style.color = data.color;
    return element;
  }
}
StyledMentionBlot.blotName = "styled-mention";

Quill.register(StyledMentionBlot);


let quill;
onMounted(() => {
  quill = new Quill('#edit_' + messageID, {
    modules: {
      toolbar: '#toolbar_' + messageID,

      mention: {
        allowedChars: /^[A-Za-z\sÅÄÖåäö]*$/,
        mentionDenotationChars: ["@"],
        blotName: 'styled-mention',
        source: function(searchTerm, renderList, mentionChar) {
          let values;

          if (mentionChar === "@") {
            values = [
              { id: 1, value: "komatsu", imageUrl: "/me.jpg" },
              { id: 2, value: "ivan", imageUrl: "/ivan.png" },
              { id: 3, value: "seijiroseijiroseijiroseijiroseijiroseijiroseijiroseijiroseijiro", imageUrl: "/me.jpg" },
              // Add more users as needed
            ];
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
          if (rect.left > 150) {
            quillMentionList.style.left = (- 1 * rect.left) + 'px';
          }
        }
        // onSelect: function(item, insertItem) {
        //   // console.log('insert HTML', item, insertItem );
        //   insertItem({id:'123',value:'My Mention'},true, {blotName: "Inline"})
        //   // return '<a>test</a>';
        // }
      }
    },
    theme: 'snow'
  });
});

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

.ql-mention-list-container {
  background-color: white;
  bottom: 0px;
}

.ql-mention-list {
  display: flex;
  flex-direction: column-reverse;
  position: absolute;
  left: -30px;
  width: 300px;
  bottom: 0px;
}

.ql-mention-list-item {
  display: flex; /* メンションリストアイテムをフレックスボックスとして配置 */
  align-items: center; /* 垂直方向に中央揃え */
  background-color: white;
}

/* メンションリストアイテムのテキストスタイルを変更する */
.ql-mention-list-item-text {
  margin-left: 8px; /* テキストと画像の間隔を設定 */
}

/* メンションリストアイテムの画像スタイルを変更する */
.ql-mention-list-item-image {
  width: 24px; /* 画像の幅を設定 */
  height: 24px; /* 画像の高さを設定 */
}

.mention {
  background-color: #a7cad63d;
  color: blue;
}
</style>
