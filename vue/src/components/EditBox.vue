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
      @input="checkMention"
      @mousedown="handleMove"
      @keydown="handleCross"
      v-html="editTxt[messageID]">
    </div>
  </div>
  <div>
    <textarea ref="textarea" style="display: none;"></textarea>
    <div ref="dropdown" class="dropdown">
      <div>
        <div v-for="(alias, i) in suggestions"
          :style="{ backgroundColor: i == selectedIndex ? 'lightblue' : '' }"
          @click="mentioning(alias)" >
          <img :src="alias[1]" class="iconMini">
          {{alias[0]}}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, defineProps, onMounted } from 'vue';
import { useMessagesStore } from '../stores/messages.js';

import { htmlToMarkdown, markdownToHtml } from '../my/markdown.js';

import Quill from 'quill';
import "quill/dist/quill.snow.css";

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

let mention = ref(false);
let mentionWords = '';
let mentionPosition = 0;
function checkMention(event) {
  // console.log(event, event.target, event.selectionStart);
  if (mention.value) {
    if (event.inputType === 'deleteContentBackward') {
      mentionPosition -= 1;
      mentionWords = mentionWords.slice(0, -1);
      if (mentionPosition < 0) {
        mention.value = false;
      }
    } else {
      mentionWords += event.data;
      mentionPosition += 1;
    }
  }
  if (event.data == '@' && mention.value == false) {
    mention.value = true;
    showMention(event);
  } else if (event.data == ' ' || event.data == '　') {
    mention.value = false;
  }
  if (mention.value) {
    console.log('mention true');
    console.log(mentionWords);
    suggestions.value = aliasArray
      .filter(([alias, img]) => alias.includes(mentionWords))
      .map(([alias, img]) => [alias, img]);

    const originalText = quill.getSemanticHTML();
    const newText = originalText.substring(0, caretOffset) + originalText.substring(caretOffset + 1);
    console.log("New text after deletion:", newText);
    // quill.root.innerHTML = newText;

  } else {
    mentionWords = '';
    mentionPosition = 0;
  }
}

function initMention() {
  mention.value = false;
  mentionWords = '';
  mentionPosition = 0;
  caretOffset = 0;
  suggestions.value = [];
  selectedIndex.value = -1;
  dropdown.value.style.left = '-1000px';

}

let suggestions = ref([]);

function handleMove(event) {
  mention.value = false;
}
let selectedIndex = ref(-1);

function handleCross(event) {
  switch (event.keyCode) {
    case 37: // left
      mention.value = false;
      break;
    case 39: // right
      mention.value = false;
      break;
    case 38: // up
      selectedIndex.value -= 1
      break;
    case 40: // down
      selectedIndex.value += 1
      break;
    case "Enter":
      // console.log("Enter", suggestions.value[selectedIndex.value]);
      // console.log(event.data);
      if (suggestions.value[selectedIndex.value]) {
        // console.log("Selected value:", suggestions.value[selectedIndex.value]);
      }
      break;
    default:
      break;
  }
  // console.log("Enter", selectedIndex.value);
  // suggestions[0][2] = true;
}

const mentioning = (alias) => {
  console.log("Caret position:", caretOffset);
  let content = quill.getSemanticHTML();
  // console.log(content);
  let newText = content.slice(0, caretOffset) + alias[0] + content.slice(caretOffset);
  console.log('newText', newText);
}


let textarea = ref(null);
const dropdown = ref(null);
let caretOffset;
const showMention = (event) => {
  const cursorInfo = getCaretCoordinates(event.target);
  const cursorIndex = cursorInfo.offset;
  dropdown.value.style.top = `${cursorInfo.top + cursorInfo.height}px`;
  // dropdown.value.style.left = `${cursorInfo.left}px`;
  dropdown.value.style.left = '14px';
  caretOffset = getCaretCharacterOffsetWithin(event.target);
  // document.getElementById("name_" + messageID).focus();
  // quill.root.innerHTML = newText;
};

const getCaretCoordinates = (element) => {
  const value = element.value;
  const selection = window.getSelection();
  const range = selection.getRangeAt(0);
  const preCaretRange = range.cloneRange();
  preCaretRange.selectNodeContents(element);
  preCaretRange.setEnd(range.endContainer, range.endOffset);
  const offset = preCaretRange.toString().length;

  const rect = range.getBoundingClientRect();
  // console.log(rect);
  return {
    top: rect.top,
    left: rect.left,
    height: rect.height,
    offset: offset
  };
};

// カーソル位置を取得する関数
const getCaretCharacterOffsetWithin = (element) => {
    const doc = element.ownerDocument || element.document;
    const win = doc.defaultView || doc.parentWindow;
    let sel;
    if (typeof win.getSelection != "undefined") {
        sel = win.getSelection();
        if (sel.rangeCount > 0) {
            const range = win.getSelection().getRangeAt(0);
            const preCaretRange = range.cloneRange();
            preCaretRange.selectNodeContents(element);
            preCaretRange.setEnd(range.endContainer, range.endOffset);
            const fragment = preCaretRange.cloneContents();
            const div = document.createElement('div');
            div.appendChild(fragment);
            const preCaretText = div.innerHTML;
            // console.log(preCaretText); // HTML タグを含めた文字列
            // タグを含めたテキストをカウントする
            console.log('element.innerHTML', element.innerHTML);
            console.log('preCaretText', preCaretText);
            const preCaretTextWithTags = element.innerHTML.substring(0, preCaretText.length);
            console.log(preCaretTextWithTags.length);
            // タグを除外したテキストの長さを取得
            // const textWithoutTagsLength = preCaretTextWithTags.replace(/<[^>]+>/g, '').length;
            // console.log(textWithoutTagsLength);

            // カーソル位置を取得
            caretOffset = preCaretTextWithTags.length -3;
        }
    // } else if ((sel = doc.selection) && sel.type != "Control") {
    //     const textRange = sel.createRange();
    //     const preCaretTextRange = doc.body.createTextRange();
    //     preCaretTextRange.moveToElementText(element);
    //     preCaretTextRange.setEndPoint("EndToEnd", textRange);
    //     caretOffset = preCaretTextRange.text.length;
    }
    return caretOffset;
};

let quill;
onMounted(() => {
  quill = new Quill('#edit_' + messageID, {
    modules: {
      toolbar: '#toolbar_' + messageID
    },
    theme: 'snow'  // テーマを指定します（省略可能）
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

.dropdown {
  position: absolute;
  left: -1000px;
  width: 300px;
  height: 300px;
  border: 1px solid black;
  display: flex;
  align-items: flex-end; /* 子要素を下部に配置 */
  justify-content: center; /* 水平方向に中央揃え */
}

.selected {
  background-color: lightblue;
}
</style>
