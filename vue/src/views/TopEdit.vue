<script setup>
import { ref, onMounted } from 'vue'
import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue';

import { pushReceive } from '@/pushReceive/pushReceive.js';

document.title = 'フォーム編集'


// const errorMessage = ref('')
// const topLinks = ref([
//   {
//     topText: 'カレンダー',
//     topLink: '/calendar/'
//   },
//   {
//     topText: 'カレンダー',
//     topLink: '/calendar/'
//   }
// ])

const topLinks = ref([])
const channel = ref(null)
const groups = ref([])
const aliases = ref([])
const forms = ref([])
const errorMessage = ref('')
onMounted(async () => {
  channel.value = await getIDB('channel', localStorage.getItem('channelID'))
  groups.value = await getIDBs('group', 'channelIDIndex', localStorage.getItem('channelID'), 10000)
  aliases.value = await getIDBs('alias', 'channelIDIndex', localStorage.getItem('channelID'), 10000)
  // if (props.formJson) {
  //   try {
  //     const parsed = JSON.parse(props.formJson)
  //     entryForm.value = { ...entryForm.value, ...parsed }
  //   } catch (e) {
  //     console.error('entryFormパラメータのJSONパースに失敗しました:', e)
  //   }
  // } else if (props.id) {
  //   entryForm.value = await getIDB('entryForm', props.id)
  // } else {
  //   forms.value = await getAllIDBs('entryForm')
  //   console.log(forms.value)
  // }
})


// 追加・削除
function addLink() {
  topLinks.value.push({
    topText: '',
    topLink: ''
  })
}
function removeLink(index) {
  topLinks.value.splice(index, 1)
}

// i番目と i+1番目を入れ替える
function moveDown(index) {
  event.preventDefault()
  if (index < topLinks.value.length - 1) {
    const temp = topLinks.value[index]
    topLinks.value[index] = topLinks.value[index + 1]
    topLinks.value[index + 1] = temp
  }
}

// i番目と i-1番目を入れ替える（上に移動）
function moveUp(index) {
  event.preventDefault()
  if (index > 0) {
    const temp = topLinks.value[index]
    topLinks.value[index] = topLinks.value[index - 1]
    topLinks.value[index - 1] = temp
  }
}

// 仮の送信処理
async function submit(event) {
  event.preventDefault()
  const fd = new FormData();
  fd.append('pushNames', JSON.stringify(aliases.value.map(d => d.aliasName)));
  fd.append('channelID', localStorage.getItem("channelID"));
  fd.append('updatedBy', channel.value.myname);
  fd.append('csrf', localStorage.getItem('csrf'));
  fd.append('contents', JSON.stringify(topLinks.value));
  fd.append('pushTitle', 'topEdit');
  const res = await sendRequest('/ContentsPush/', fd);
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
}
</script>

<template>
<div id="drawer_column"><Drawer v-if="aliases" :aliases="aliases" :channel="channel" /></div>
  <div id="content">
    <h2 class="sp_head">リンク編集・一覧</h2>

    <div v-if="errorMessage"> 
      <div class="errorMessage">{{errorMessage}}<br>データ取得に失敗しました。</div>
      <a href="/setting/"> データ設定ページ </a><br>
      <a href="/sign/"> サインインページ </a>
    </div>

    <!-- 編集フォーム -->
    <form v-if="topLinks">
      <div v-for="(link, i) in topLinks" :key="i" style="border:1px solid #ccc; padding:8px; margin-bottom:6px;">
        <label>リンクテキスト:
          <input v-model="link.topText" placeholder="表示名" type="text" />
        </label><br>
        <label>リンクURL:
          <input v-model="link.topLink" placeholder="/example/" type="text" />
        </label><br>
        <button @click="moveUp(i)" :disabled="i===0">↑</button>
        <button @click="moveDown(i)" :disabled="i===topLinks.length-1">↓</button>
        <button @click.prevent="removeLink(i)" v-if="topLinks.length > 0">−リンク削除</button>
      </div>
      <button @click.prevent="addLink">＋リンク追加</button>
      <br><br>
      <button type="submit" @click="submit">送信</button>
    </form>

    <!-- 新規作成用 -->
    <div v-if="!topLinks || topLinks.length === 0">
      <button @click="addLink">新規リンク作成</button>
    </div>

    <!-- 一覧 -->
    <ul>
      <li v-for="(l, i) in topLinks" :key="'list-' + i">
        {{ l.topText }} → {{ l.topLink }}
      </li>
    </ul>
  </div>

  <div id="ad_right"> 
    <Advertisement /> <Advertisement /> <Advertisement /> 
  </div>
</template>

