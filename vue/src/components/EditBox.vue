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

import { textToHtml } from '../my/textToHtml.js';

const props = defineProps({
  message: Object,
  channel: Object,
  parent_id: String
});

const messageID = props.message.messageID;

const messagesStore = useMessagesStore();

const editorContent = ref({});

const cancelEdit = (message, messageID) => {
  message.editFlg = false;
  messagesStore.update(message, message.messageID);
};

const msgUpsert = (messageId) => {
  console.log(editorContent.value[messageId]);
  // const messageData = document.getElementById('editable_' + messageId).innerHTML;
  // const uri = messageId ? (props.parent_id ? '/ThreadEdit/' : '/MessageEdit/') : (props.parent_id ? '/ThreadPost/' : '/MessagePost/');
  // const fd = new FormData();
  // fd.append('parentID', props.parent_id);
  // fd.append('channelID', props.channel.channelID);
  // fd.append('messageID', messageId);
  // fd.append('messageTxt', htmlToMarkdown(messageData));
  // fd.append('aliasName', props.channel.aliasName);
  // const request = new Request(uri, {
  //   method: 'POST',
  //   body: fd,
  // });
  // fetch(request)
  //   .catch((reason)=>{
  //     alert(reason)
  //   })
}

const htmlToMarkdown = (html) => {
  let markdown = html.replace(/\n/g, "");
  markdown = markdown.replace(/<br>/g, "\n");
  markdown = reverseEmphasis(markdown);
  markdown = reverseBlockquotes(markdown);
  markdown = reverseCodeBlocks(markdown);
  markdown = reverseLinks(markdown);
  markdown = unescapeHtml(markdown);
  return markdown;
};

function unescapeHtml(html) {
  return html.replace(/&amp;|&lt;|&gt;|&quot;|&#39;/g, function(match) {
    return {
      '&amp;': '&',
      '&lt;': '<',
      '&gt;': '>',
      '&quot;': '"',
      '&#39;': "'"
    }[match];
  });
}

// HTMLエスケープ関数
const escapeHtml = (unsafe) => {
  return unsafe.replace(/[&<"'\n]/g, (match) => {
    switch (match) {
      case '&':
        return '&amp;';
      case '<':
        return '&lt;';
      case '>':
        return '&gt;';
      case '"':
        return '&quot;';
      case "'":
        return '&#039;';
      case "\n":
        return '<br>';
    }
  });
};

const reverseEmphasis = (html) => {
  return html.replace(/<strong>([^<]+)<\/strong>/g, '**$1**');
};

const reverseBlockquotes = (html) => {
  return html.replace(/<blockquote>(.*?)<\/blockquote>/g, '> $1\n');
};

const reverseCodeBlocks = (html) => {
  return html.replace(/<code>(.*?)<\/code>/g, '```\n$1\n```');
};

const reverseLinks = (html) => {
  return html.replace(/<a href="(.*?)">(.*?)<\/a>/g, '[$2]($1)');
};


const highlightTags = (event, messageID) => {
  const target = event.target;
  if (target.tagName === 'STRONG') {
    addSelected(document.getElementById('bold_' + messageID));
  } else {
    removeSelected(document.getElementById('bold_' + messageID));
  }
  const selection = window.getSelection();
  textRange.value = selection.getRangeAt(0);
  selectedText.value = selection.toString();
};

const addSelected = (element) => {
  element.classList.add('selected');
};
const removeSelected = (element) => {
  element.classList.remove('selected');
};

let textRange = ref('');
let selectedText = ref(''); 

const toggleStrong = (messageID) => {
  const element = document.getElementById('editable_' + messageID);
  const selectedText = textRange.value.toString();
  const parentNode = textRange.value.commonAncestorContainer.parentElement;
  const isAlreadyStrong = parentNode.tagName === 'STRONG';

  if (isAlreadyStrong) {
    const strongNode = parentNode;
    const parent = strongNode.parentNode;
    while (strongNode.firstChild) {
      parent.insertBefore(strongNode.firstChild, strongNode);
    }
    parent.removeChild(strongNode);
    removeSelected(document.getElementById('bold_' + messageID));
  } else if (selectedText) {
    const strongText = '<strong>' + selectedText + '</strong>';
    textRange.value.deleteContents();
    textRange.value.insertNode(document.createRange().createContextualFragment(strongText));
    addSelected(document.getElementById('bold_' + messageID));
  }
};

const toggleBlockquote = (messageID) => {
  const element = document.getElementById('editable_' + messageID);
  const selectedText = escapeHtml(textRange.value.toString());
  const parentNode = textRange.value.commonAncestorContainer.parentElement;
  const isAlreadyBlockquote = parentNode.tagName === 'BLOCKQUOTE';

  if (isAlreadyBlockquote) {
    const blockquoteNode = parentNode;
    const parent = blockquoteNode.parentNode;
    while (blockquoteNode.firstChild) {
      parent.insertBefore(blockquoteNode.firstChild, blockquoteNode);
    }
    parent.removeChild(blockquoteNode);
    removeSelected(document.getElementById('quote_' + messageID));
  } else if (selectedText) {
    const blockquoteText = '<blockquote>' + selectedText + '</blockquote>';
    textRange.value.deleteContents();
    textRange.value.insertNode(document.createRange().createContextualFragment(blockquoteText));
    addSelected(document.getElementById('quote_' + messageID));
  }
};

const toggleCode = (messageID) => {
  const element = document.getElementById('editable_' + messageID);
  const selectText = escapeHtml(selectedText.value);
  const selectText2 = escapeHtml(textRange.value.toString());
  console.log(selectText);
  console.log(selectText2);
  const parentNode = textRange.value.commonAncestorContainer.parentElement;
  const isAlreadyCode = parentNode.tagName === 'CODE';

  if (isAlreadyCode) {
    const codeNode = parentNode;
    const parent = codeNode.parentNode;
    while (codeNode.firstChild) {
      parent.insertBefore(codeNode.firstChild, codeNode);
    }
    parent.removeChild(codeNode);
    removeSelected(document.getElementById('code_' + messageID));
  } else if (selectText) {
    const codeText = '<code>' + selectText + '</code>';
    textRange.value.deleteContents();
    textRange.value.insertNode(document.createRange().createContextualFragment(codeText));
    addSelected(document.getElementById('code_' + messageID));
  }
};



  // // リンク [リンクテキスト](URL)
  // formatted = formatted.replace(/\[([^\]]+)\]\(([^\)]+)\)/g, (match, p1, p2) => {
  //   return `<a href="${p2}">${p1}</a>`;
  // });

  // // コード `コード`　📄
  // formatted = formatted.replace(/`(.*?)`/g, '<code>$1</code>');

  // // 引用 > 引用文　󠀢”
  // formatted = formatted.replace(/^>(.*)$/gm, '<blockquote>$1</blockquote>');

  // // 強調 **強調**　<span style="font-weight: bold;">B</span>
  // formatted = formatted.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');


</script>

<style>
.markdown {
  padding: 4px;
}

.editLeft {
  text-align: left;
  display: inline-block;
  width: 80%;
}

.editRight {
  display: inline-block;
  width: 20%;
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
.msgBox {
  text-align: center;
  width: 100%;
}
.quote {
  padding: 3px;
  margin: 0 0 0 5px;
  border-left: 3px solid #ccc;
}
</style>
