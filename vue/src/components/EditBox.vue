<template>
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
    <div :id="'edit_' + messageID" @input="handleInput" v-html="editTxt[messageID]">
    </div>
  <div>
    <textarea ref="textarea" @input="handleInput" style="display: none;"></textarea>
    <div ref="dropdown" class="dropdown" >
      <input :list="'alias_' + messageID" :id="'name_' + messageID" />
      <datalist :id="'alias_' + messageID">
        <option v-for="item in aliasArray">{{ item[0] }}{{ item[1] }}</option>
      </datalist>
    </div>
  </div>
<!--   <div>
    <QuillEditor :toolbar="'#my-toolbar_' + messageID" contentType="html" v-model:content="editorContent[messageID]" @input="handleInput" @mousedown="handleMove" @keydown="handleCross">

    <textarea v-model="searchText" @input="updateDropdownItems"></textarea>
    <select v-model="selectedItem">
      <option v-for="item in filteredItems" :value="item.value">{{ item.label }}</option>
    </select>
    <p>選択されたアイテム: {{ selectedItem }}</p>
  </div> -->
<!-- <link href="https://cdn.jsdelivr.net/npm/quill@2.0.0-rc.2/dist/quill.snow.css" rel="stylesheet" /> -->
<!-- <script src="https://cdn.jsdelivr.net/npm/quill@2.0.0-rc.2/dist/quill.js"></script> -->

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
const cancelEdit = (message, messageID) => {
  message.editFlg = false;
  messagesStore.update(message, message.messageID);
};

const msgUpsert = (messageId) => {
  const messageData = htmlToMarkdown(editorContent.value[messageId]);
  const threadFlg = props.message.parentID;
  const uri = messageId ? (threadFlg ? '/ThreadEdit/' : '/MessageEdit/') : (threadFlg ? '/ThreadPost/' : '/MessagePost/');
  const fd = new FormData();
  fd.append('parentID', props.message.parentID);
  fd.append('channelID', props.channel.channelID);
  fd.append('messageID', messageId);
  fd.append('messageTxt', messageData);
  fd.append('aliasName', props.channel.aliasName);
  if (props.threadHead && props.threadHead.backID) {
    fd.append('backID', props.threadHead.backID);
  }

  // const request = new Request(uri, {
  //   method: 'POST',
  //   body: fd,
  // });
  // fetch(request)
  //   .catch((reason)=>{
  //     alert(reason)
  //   })
}

let aliasArray = [
  ['sei1', '/me.jpg'],
  ['ivan1', '/me.jpg'],
  ['group2', '/group.png']
];

function suggestAlias(text) {
  const suggestions = aliasArray.filter(([alias, img]) => alias.includes(text)).map(([alias, img]) => [alias, img]);
  return suggestions;
}

// テキスト入力が行われたときに呼び出される関数
let mention = false;
let mentionWords = '';
let mentionPosition = 0;
// function handleInput(event) {
//   console.log(event, event.target, event.selectionStart);
//   if (mention) {
//     if (event.inputType === 'deleteContentBackward') {
//       mentionPosition -= 1;
//       mentionWords = mentionWords.slice(0, -1);
//       if (mentionPosition < 0) {
//         mention = false;
//       }
//     } else {
//       mentionWords += event.data;
//       mentionPosition += 1;
//     }
//   }
//   if (mention) {
//     const suggestions = suggestAlias(mentionWords);
//     console.log(suggestions);
//   }
//   if (event.data == '@') {
//     mention = true;
//     mentionWords = '';
//     mentionPosition = 0;
//   } else if (event.data == ' ' || event.data == '　') {
//     mention = false;
//   }
// }

function handleMove(event) {
  mention = false;
}

function handleCross(event) {
  const keyCode = event.keyCode;
  if (keyCode === 37 || keyCode === 38 || keyCode === 39 || keyCode === 40) {
    mention = false;
  }
}

// const text = ref('');
// const showDropdown = ref(false);
// const dropdownPosition = ref({ top: 0, left: 0 });
// const dropdownItems = ref([]);

// function handleInput(event) {
//   // カーソルの位置を取得
//   const cursorIndex = event.target.selectionStart;
//   // カーソルの位置を基準にドロップダウンリストを表示する位置を計算
//   const textareaRect = event.target.getBoundingClientRect();
//   console.log(cursorIndex, textareaRect, event.target);
//   const cursorPos = getCaretCoordinates(event.target, cursorIndex);
//   const dropdownTop = textareaRect.top + cursorPos.top + cursorPos.height;
//   const dropdownLeft = textareaRect.left + cursorPos.left;
//   // ドロップダウンリストを表示
//   dropdownPosition.value = { top: dropdownTop, left: dropdownLeft };
//   showDropdown.value = true;
// }

// // テキストエリア内のカーソルの位置を取得する関数
// function getCaretCoordinates(element, index) {
//   console.log(element.value);
//   const textBeforeCursor = element.value.substring(0, index);

//   const span = document.createElement('span');
//   span.textContent = textBeforeCursor;
//   element.appendChild(span);
//   const rect = span.getBoundingClientRect();
//   const cursorPos = {
//     top: rect.top,
//     left: rect.width,
//     height: rect.height
//   };
//   span.remove();
//   return cursorPos;
// }


// function getCaretCoordinates(textarea) {
//   const value = textarea.value;
//   const selection = window.getSelection();
//   const range = selection.getRangeAt(0);
//   const preCaretRange = range.cloneRange();
//   preCaretRange.selectNodeContents(textarea);
//   preCaretRange.setEnd(range.endContainer, range.endOffset);
//   const offset = preCaretRange.toString().length;

//   const rect = range.getBoundingClientRect();
//   console.log(rect);
//   return {
//     top: rect.top,
//     left: rect.left,
//     height: rect.height,
//     offset: offset
//   };
// }

// function handleInput(event) {
//   const textarea = event.target;
//   const cursorInfo = getCaretCoordinates(textarea);
//   const cursorIndex = cursorInfo.offset;
//   console.log(cursorIndex);
// }

let textarea = ref(null);
const dropdown = ref(null);
const showDropdown = ref(false);
const dropdownItems = ref([]);

const handleInput = (event) => {
  
  const cursorInfo = getCaretCoordinates(event.target);
  const cursorIndex = cursorInfo.offset;

  // ドロップダウンリストを表示するための位置を設定
  dropdown.value.style.top = `${cursorInfo.top + cursorInfo.height}px`;
  dropdown.value.style.left = `${cursorInfo.left}px`;

  // テキストエリア内のカーソル位置に応じてドロップダウンリストを更新する処理を実装する

  // // ドロップダウンリストを表示する
  // showDropdown.value = true;

  // // テキストエリアからフォーカスが外れたときにドロップダウンリストを非表示にする
  // textarea.value.addEventListener('blur', () => {
  //   showDropdown.value = false;
  // });
  
  // const editableDiv = document.querySelector('#edit_' + messageID + ' .ql-editor');

  const caretOffset = getCaretCharacterOffsetWithin(event.target);
  console.log("Caret position:", caretOffset);
  let text = quill.getSemanticHTML();
  console.log(text);
  let newText = text.slice(0, caretOffset) + '<strong>komatsu</strong>' + text.slice(caretOffset);
  console.log('newText', newText);
  document.getElementById("name_" + messageID).focus();
  // quill.root.innerHTML = newText;
  // editableDiv.textContent = newText;
  // editorContent.value[messageID] = newText;
  // quill[messageID].setHTML(newText);
  // quill[messageID].root.innerHTML = newText;
  // console.log(editor.value[messageID]);
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
  console.log(rect);
  return {
    top: rect.top,
    left: rect.left,
    height: rect.height,
    offset: offset
  };
};

// カーソル位置を取得する関数
const getCaretCharacterOffsetWithin = (element) => {
    let caretOffset = 0;
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
            caretOffset = preCaretRange.toString().length;
        }
    } else if ((sel = doc.selection) && sel.type != "Control") {
        const textRange = sel.createRange();
        const preCaretTextRange = doc.body.createTextRange();
        preCaretTextRange.moveToElementText(element);
        preCaretTextRange.setEndPoint("EndToEnd", textRange);
        caretOffset = preCaretTextRange.text.length;
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
  /*display: none;*/
}
</style>
