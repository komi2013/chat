<script setup>
import { ref, computed, onMounted } from 'vue'
import DrawerColumn from '../components/DrawerColumn.vue'
import { htmlToMarkdown, markdownToHtml } from '../my/markdown.js';
import { getIDB, getIDBs, upsertData } from '../my/indexDB.js';
import Quill from 'quill';
import "quill/dist/quill.snow.css";

const props = defineProps({
  parent_id: ''
})

const threadHead = ref({});
let editTxt = ref({});
editTxt.value = markdownToHtml(threadHead.value.description);

async function fetchThreadHead() {
  try {
    threadHead.value = await getIDB('threadHead', props.parent_id);
  } catch (error) {
    console.log('no threadHead', error);
  }
}

const postThreadHead = () => {
  console.log(quill.root.innerHTML);
  const description = quill.root.innerHTML;
  console.log(description);
  const fd = new FormData();
  fd.append('channelID', threadHead.value.channelID);
  fd.append('parentID', threadHead.value.parentID);
  fd.append('title', threadHead.value.title);
  fd.append('description', description);
  const request = new Request('/ThreadHeadPost/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .then((response) => response.json())
    .catch((reason)=>{
      console.log(reason)
    })
}
const msgText = ref(null);
let quill;
onMounted(() => {
  // msgText.value.addEventListener('input', function () {
  //   this.style.height = 'auto';
  //   this.style.height = (this.scrollHeight) + 'px';
  // });
  // fetchData();
  fetchThreadHead();
  quill = new Quill('#edit', {
    modules: {
      toolbar: '#toolbar'
    },
    theme: 'snow'
  });
})


// function showNotification(title, body, icon) {
//   // ブラウザが通知をサポートしているかを確認
//   if (!("Notification" in window)) {
//     console.error("このブラウザは通知をサポートしていません");
//     return;
//   }

//   // ユーザーが通知を許可しているかを確認
//   if (Notification.permission === "granted") {
//     // 許可されている場合は通知を表示
//     new Notification(title, { body, icon });
//   } else if (Notification.permission !== "denied") {
//     // 許可が求められていない場合は、許可を求める
//     Notification.requestPermission().then(permission => {
//       if (permission === "granted") {
//         // 許可された場合は通知を表示
//         new Notification(title, { body, icon });
//       }
//     });
//   }
// }

// // 使用例
// showNotification("新しいメッセージ", "新着メッセージがあります", "/icon.png");


const text = '＊p＊こっちばっかに集中してしまう　＠＠ivan1・＠＠・＊p＊';

if (text.includes('＠＠ivan1・＠＠')) {
  console.log('＠＠ivan1＠＠が見つかりました');
  new Notification('title', { body: 'nbo', icon: '/me.jpg' });
} else {
  console.log('＠＠ivan1＠＠は見つかりませんでした');
}

</script>



<template>
<DrawerColumn />
<div id="content">
  <textarea class="headTitle" v-model="threadHead.title"></textarea>
</div>

  <div>
    <div class="editLeft" id="toolbar">
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
    <div id="edit"
      v-html="editTxt"
      >
    </div>
  </div>

<div @click="postThreadHead">➡️</div>
</template>

<style>
@media screen and (min-width : 701px) { 
  .headTitle {
    margin-left: 50px;
    display: flex;
    width: 300px;
  }
}

@media screen and (max-width : 700px) {
  .headTitle {
    margin-left: 50px;
    display: flex;
    width: 300px;
  }
  .headTitle div {
    display: table-cell;
  }
}

</style>

