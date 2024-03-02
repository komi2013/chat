<template>
<div>
  <div class="editLeft">
    <span :id="'link_' + messageID" class="markdown">📎</span>
    <span :id="'bold_' + messageID" class="markdown" @click="toggleStrong(messageID)">B</span>
    <span :id="'quote_' + messageID" class="markdown">&gt;&gt;</span>
    <span :id="'code_' + messageID" class="markdown">&lt;/&gt;</span>
  </div>
  <div class="editRight">
    <button v-if="messageID" @click="cancelEdit(message, messageID)">⬅</button>
    <button @click="msgUpsert(messageID)">▶️</button>
  </div>
</div>
<div contenteditable="true"
    class="chgble"
    :id="'editable_' + messageID"
    v-html="textToHtml(message.messageTxt)"
    @click="highlightTags($event, messageID)"
  >
</div>
</template>

<script setup>
import { ref, defineProps } from 'vue';
import { useMessagesStore } from '../stores/messages.js';

import { textToHtml } from '../my/textToHtml.js';

const props = defineProps({
  message: Object,
  channel: Object,
  parent_id: String
});

const messageID = props.message.messageID;

const messagesStore = useMessagesStore();

const cancelEdit = (message, messageID) => {
  message.editFlg = false;
  messagesStore.update(message, message.messageID);
};

const msgUpsert = (messageId) => {
  const messageData = document.getElementById('editable_' + messageId).innerHTML;
  const uri = messageId ? (props.parent_id ? '/ThreadEdit/' : '/MessageEdit/') : (props.parent_id ? '/ThreadPost/' : '/MessagePost/');
  const fd = new FormData();
  fd.append('parentID', props.parent_id);
  fd.append('channelID', props.channel.channelID);
  fd.append('messageID', messageId);
  fd.append('messageTxt', htmlToMarkdown(messageData));
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
};

const addSelected = (element) => {
  element.classList.add('selected');
};
const removeSelected = (element) => {
  element.classList.remove('selected');
};

let textRange = ref(''); 

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
  width: 70%;
}

.editRight {
  text-align: right;
  display: inline-block;
  width: 30%;
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
</style>
