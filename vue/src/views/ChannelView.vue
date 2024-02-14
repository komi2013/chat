<script setup>
import { ref, computed, onMounted } from 'vue'
import DrawerColumn from '../components/DrawerColumn.vue'
import { useMessagesStore } from '../stores/messages.js';
import { useChannelsStore } from '../stores/channels.js';
import { get_formated_time } from '../my/get_formated_time.js';
import { getIDB, getIDBs, upsertData } from '../my/indexDB.js';
import { isEmojiOpen, selectedMessageId, openEmoji, closeEmoji, selectEmoji, calcEmoji, emojiPath } from '../my/emoji.js';
import { isOtherOpen, otherMessageId, openOther, closeOther, selectOther, activeEdit, cancelEdit, adjustHeight, textareaRefs } from '../my/other.js';
import EmojiModal from '../components/EmojiModal.vue';
import OtherModal from '../components/OtherModal.vue';

const props = defineProps({
  id: '',
})

const channel = ref('');
const messagesStore = useMessagesStore();
const messages = computed(() => {
  return messagesStore.messages;
})

async function fetchData() {
  try {
    const data = await getIDB('channel', props.id);
    channel.value = data;
  } catch (error) {
    channel.value = null;
  }
}

const fetchMessageData = () => {
  return new Promise((resolve, reject) => {
    getIDBs('message', 'channelIDIndex', props.id)
      .then((data) => {
        const latest = data.reverse();
        latest.forEach(message => {
          messagesStore.insert(message);
        });
        resolve();
      })
      .catch((error) => {
        reject(error);
      });
  });
};

const msgText = ref(null);
const msgUpsert = (messageTxt, messageId) => {
  const messageData = messageTxt ? messageTxt : document.getElementById('editable_' + messageId).innerHTML;
  console.log(messageData);
  let uri = messageId ? '/MessageEdit/' : '/MessagePost/';
  const fd = new FormData();
  fd.append('channelID', props.id);
  fd.append('messageID', messageId);
  fd.append('messageTxt', htmlToMarkdown(messageData));
  fd.append('messageType', 1);
  fd.append('editFlg', 1);
  fd.append('aliasName', channel.value.aliasName);
  const request = new Request(uri, {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      alert(reason)
    })
}

const contentRef = ref(null);
onMounted(() => {
  fetchData();
  fetchMessageData()
  .then(() => {
      const content = contentRef.value;
      content.scrollTop = content.scrollHeight;
    });
});

const clickEmoji = (messageId, emoji) => {
  const fd = new FormData();
  fd.append('channelID', props.id);
  fd.append('messageID', messageId);
  fd.append('emoji', emoji[0]);
  fd.append('clicked', emoji[2] ? 1 : 0);
  fd.append('aliasName', channel.value.aliasName);
  const request = new Request('/MessageEdit/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      alert(reason)
    })
};

const formattedText = (rawText) => {
  let escaped = escapeHtml(rawText);
  let formatted = escaped.replace(/\n/g, "<br>");

  // リンクの置換
  formatted = replaceLinks(formatted);

  // コードブロックの置換
  formatted = replaceCodeBlocks(formatted);

  // 引用文の置換
  formatted = replaceBlockquotes(formatted);

  // 強調の置換
  formatted = replaceEmphasis(formatted);

  return formatted;
};


function escapeHtml(html) {
  return html.replace(/[&<>"']/g, function(match) {
    return {
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      '"': '&quot;',
      "'": '&#39;'
    }[match];
  });
}

// リンクの置換処理
const replaceLinks = (text) => {
  return text.replace(/\[([^\]]+)\]\(([^\)]+)\)/g, (match, p1, p2) => {
    return `<a href="${p2}">${p1}</a>`;
  });
};

// コードブロックの置換処理
const replaceCodeBlocks = (text) => {
  return text.replace(/```(.*?)```/g, '<code>$1</code>');
};

// 引用文の置換処理
const replaceBlockquotes = (text) => {
  return text.replace(/^>(.*)$/gm, '<blockquote>$1</blockquote>');
};

// 強調の置換処理
const replaceEmphasis = (text) => {
  return text.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
};

const htmlToMarkdown = (html) => {
  console.log(html);
  let markdown = html.replace(/\n/g, "");
  console.log(markdown);
  markdown = markdown.replace(/<br>/g, "\n");
  console.log(markdown);
  markdown = reverseEmphasis(markdown);
  markdown = reverseBlockquotes(markdown);
  markdown = reverseCodeBlocks(markdown);
  markdown = reverseLinks(markdown);
  return markdown;
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



<template>
<DrawerColumn />
<div id="content" ref="contentRef">
<div class="headTitle">
  <RouterLink :to="'/channelInfo/' + channel.channelID"> {{ channel.channelName }} </RouterLink>
</div>

<div class="text_body" id="text_body">
  <!-- ここにテキストが含まれると仮定 -->
  This is some text.
</div>

<button id="btn">Get Selected Text</button>
<div class="result"></div>

  <div v-for="message in messages" :key="message.messageID">
    <table>
      <tr>
        <td rowspan="2" class="icon_td">
          <img v-if="message.aliasImg" :src="message.aliasImg" class="icon">
        </td>
        <td>
          <span class="alias">{{ message.aliasName }}</span>
          <span class="time">{{ get_formated_time('hh:mm', message.createdAt) }}</span>
        </td>
        <td class="setting">
          <span class="emoji" @click="activeEdit(message, message.messageID, $event)"> 🖋 </span>
          <span class="emoji" @click="openEmoji(message.messageID)"> 😄 </span>
          <EmojiModal :key="message.messageID" v-if="isEmojiOpen && selectedMessageId === message.messageID" @selectEmoji="selectEmoji" @closeEmoji="closeEmoji" :channelID="channel.channelID" :messageId="message.messageID" :aliasName="channel.aliasName" />

          <span class="reply"> <RouterLink :to="'/reply/' + message.messageID"> 💬 </RouterLink> </span>
          <span class="others" @click="openOther(message.messageID)"> &nbsp; ⋮ &nbsp; </span>
          <OtherModal :key="message.messageID" v-if="isOtherOpen && otherMessageId === message.messageID" @selectOther="selectOther" @closeOther="closeOther" @activeEdit="activeEdit" :channelID="channel.channelID" :messageId="message.messageID" :aliasName="channel.aliasName" :message="message" />
        </td>
      </tr>
      <tr>
        <td v-if="message.editFlg" colspan="2" class="msg" >
          <div>
            <div class="editLeft">
              <span :id="'link_' + message.messageID" class="markdown">📎</span>
              <span :id="'bold_' + message.messageID" class="markdown" @click="toggleStrong(message.messageID)">B</span>
              <span :id="'quote_' + message.messageID" class="markdown">&quot; &quot;</span>
              <span :id="'code_' + message.messageID" class="markdown">&lt;/&gt;</span>
            </div>
            <div class="editRight">
              <button @click="cancelEdit(message, message.messageID)">⬅</button>
              <button @click="msgUpsert(null, message.messageID)">▶️</button>
            </div>
          </div>
          <div contenteditable="true"
              class="chgble"
              :id="'editable_' + message.messageID"
              v-html="formattedText(message.messageTxt)"
              @click="highlightTags($event, message.messageID)"
            >
          </div>
        </td>
        <td v-else colspan="2" class="msg">
          <div v-html="formattedText(message.messageTxt)"></div>
          <template v-for="d in calcEmoji(message.emojis, channel.aliasName)">
            <template v-if="emojiPath(d[0])">
              <span class="img-stamp" :class="{ 'selected': d[2] }"> <img :src="d[0]" class="emoji-img" @click="clickEmoji(message.messageID, d)" /> {{d[1]}} </span>
            </template>
            <template v-else>
              <span class="emoji-stamp" :class="{ 'selected': d[2] }" @click="clickEmoji(message.messageID, d)" >{{ d[0] }} {{d[1]}} </span>
            </template>
          </template>
        </td>
      </tr>
    </table>
  </div>
<!-- <RouterLink to="/channel/abc/" >channel abc</RouterLink> -->

<div class="msgBox">
  <div><span>📎</span><span style="font-weight: bold;">B</span></div>
  <div contenteditable="true" ref="msgData" class="chgble" ></div>
  <div style="text-align: right"><button @click="msgUpsert($refs.msgData.innerText)">▶️</button></div>
</div>

</div>
</template>

<style>

textarea {
  resize: none;
  height: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
}

#content {
  height: 100%;
  overflow-y: auto;
}

#content > div {
  display: flex;
  flex-direction: column-reverse;
}

.icon {
  max-width: 50px;
  max-height: 50px;
}
.icon_td {
  width: 50px;
  vertical-align: top;
}

.setting {
  text-align: right;
}

.box {
  display: flex;
}

.alias {
  margin: 2px;
}

.time {
  margin: 2px;
}

.emoji {
  margin: 2px;
}

.reply {
  margin: 2px;
}

.others {
  margin: 2px;
}

.markdown {
  padding: 4px;
}

.editLeft {
  display: inline-block;
  width: 70%;
}

.editRight {
  text-align: right;
  display: inline-block;
  width: 30%;
}

.msgBox textarea {
  width: 100%;
  border: none;
  height: 50px;
}

.emoji-stamp {
  display: inline-flex;
  align-items: center;
  padding: 1px 2px;
  border-radius: 5px;
  margin: 1px 2px;
  vertical-align: text-bottom;
  font-size: 14px;
  background-color: #92a7b54a;
}

.img-stamp {
  display: inline-flex;
  align-items: center;
  padding: 1px 2px;
  border-radius: 5px;
  margin: 1px 2px;
  background-color: #92a7b54a;
}

.emoji-img {
  max-width: 20px;
  max-height: 20px;
  padding: 2px;
}

.selected {
  border: 1px solid #3498db;
}

.chgble {
  outline: 1px solid blue;
  width: 90%;
  display: inline-block;
}
.msgBox {
  /*position: fixed;*/
  /*bottom: 10px;*/
  width: 90%;
}
@media screen and (min-width : 701px) { 

}

@media screen and (max-width : 700px) {
  .headTitle {
    margin-left: 50px;
  }
}
</style>

