<template>
  <div class="tweet-form">
    <div id="tweet-toolbar">
      <button class="ql-bold"></button>
      <button class="ql-italic"></button>
      <button class="ql-link"></button>
    </div>
    <div id="tweet-editor" class="tweet-editor"></div>
    <button @click="submitTweet">投稿</button>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import Quill from 'quill';
import 'quill/dist/quill.snow.css';
// import { sendRequest } from '@/my/sendRequest.js';

const props = defineProps({
  parentID: String,
  channelID: String
});

let quill;

onMounted(() => {
  quill = new Quill('#tweet-editor', {
    modules: { toolbar: '#tweet-toolbar' },
    theme: 'snow',
  });
});

const submitTweet = async () => {
  const messageTxt = quill.root.innerHTML.trim();
  const plainText = quill.getText().trim();

  if (!plainText) return alert('内容を入力してください。');
  if (plainText.length > 500) return alert('500文字以内で入力してください。');

  const fd = new FormData();
  // fd.append('channelID', props.channelID);
  fd.append('parentID', props.parentID ?? '')
  // fd.append('nickname', localStorage.getItem('nickname'));
  // fd.append('nickImg', localStorage.getItem('nickImg') || '');
  fd.append('messageTxt', messageTxt);
  fd.append('csrf', localStorage.getItem('csrf'));

  const res = await sendRequest('/TweetPost/', fd);

  if (res && res.status === 425) {
    alert('投稿制限中です。20時間後に再度お試しください。');
    return;
  }

  if (res && res.messageID) {
    alert('投稿完了しました！');
    quill.root.innerHTML = '';
  } else if (res && res.error) {
    alert(res.error);
  }

  res.csrf && localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }

};
</script>

<style scoped>
.tweet-form {
  margin: 10px;
}
#tweet-toolbar {
  border: 1px solid #ccc;
  border-bottom: none;
}
#tweet-editor {
  border: 1px solid #ccc;
  height: 150px;
}
button {
  margin-top: 10px;
}
</style>
