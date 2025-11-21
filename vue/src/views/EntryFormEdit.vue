<script setup>
import { ref, onMounted } from 'vue'

import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  id: String,
  // code: String,
  formJson: String
})
document.title = 'フォーム編集'

const entryForm = ref(null)
const entryFormID = generateRandomCode(1)
const defaultForm = {
  entryFormID: entryFormID,
  entryFormTitle: '',
  askChoices: [{
    question: '',
    choices: [],
    sequence: 1
  }],
  askMultiChoices: [{
    question: '',
    choices: [],
    sequence: 2
  }],
  asks: [{
    question: '',
    sequence: 3
  }],
  dates: [{
    question: '',
    sequence: 4,
    dateType: 1, // 1 = dateTime, 2 = date, 3, time
    optionKey: false
  }]
}

// entryForm.value = defaultForm
// const entryFormJson = JSON.stringify(defaultForm)
// console.log()
const channel = ref(null)
const groups = ref([])
const aliases = ref([])
const forms = ref([])
const errorMessage = ref('')
onMounted(async () => {
  channel.value = await getIDB('channel', localStorage.getItem('channelID'))
  groups.value = await getIDBs('group', 'channelIDIndex', localStorage.getItem('channelID'), 10000)
  aliases.value = await getIDBs('alias', 'channelIDIndex', localStorage.getItem('channelID'), 10000)
  if (props.formJson) {
    try {
      const parsed = JSON.parse(props.formJson)
      entryForm.value = { ...entryForm.value, ...parsed }
    } catch (e) {
      console.error('entryFormパラメータのJSONパースに失敗しました:', e)
    }
  } else if (props.id) {
    entryForm.value = await getIDB('entryForm', props.id)
  } else {
    forms.value = await getAllIDBs('entryForm')
    console.log(forms.value)
  }
})

// ユーティリティ関数
function addChoice(array, index) {
  entryForm.value[array][index].choices.push('')
}
function removeChoice(array, index, choiceIndex) {
  entryForm.value[array][index].choices.splice(choiceIndex, 1)
}
function addQuestion(array) {
  entryForm.value[array].push({
    question: '',
    choices: [],
    sequence: entryForm.value[array].length + 1
  })
}
function removeQuestion(array, index) {
  entryForm.value[array].splice(index, 1)
}

async function submit(event) {
  event.preventDefault()
  const fd = new FormData();
  fd.append('userIDs', JSON.stringify(aliases.value.map(d => d.userID)));
  fd.append('channelID', localStorage.getItem("channelID"));
  fd.append('updatedBy', channel.value.myname);
  fd.append('csrf', localStorage.getItem('csrf'));
  fd.append('contents', JSON.stringify(entryForm.value));
  fd.append('pushTitle', 'entryForm');
  const res = await sendRequest('/ContentsPush/', fd);
  if (!res.csrf) errorMessage.value = res
  res.csrf && localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  // location.href = '/entryFormEdit/' + entryForm.value.entryFormID + '/'
}
</script>

<template>
<div id="drawer_column"><Drawer v-if="aliases" :aliases="aliases" :channel="channel" /></div>
  <div id="content">
    <h2 class="sp_head">フォーム編集・一覧</h2>
    <div v-if="errorMessage"> 
      <div class="errorMessage">{{errorMessage}}<br>データ取得に失敗しました。</div>
      <a href="/setting/"> データ設定ページ </a><br>
      <a href="/sign/"> サインインページ </a>
    </div>
    <form v-if="entryForm">
      <div>受付ID:{{entryForm.entryFormID}}</div>
      <label>タイトル:
        <input v-model="entryForm.entryFormTitle" type="text" />
      </label>

      <!-- Asks -->
      <div>
        <label>質問文:</label>
        <div v-for="(ask, i) in entryForm.asks" :key="i">
          <input v-model="ask.question" placeholder="質問文" type="text" />
          <label>順序:
            <input v-model.number="ask.sequence" type="number" min="1" />
          </label>
          <button @click.prevent="removeQuestion('asks', i)" v-if="entryForm.asks.length > 0">−質問削除</button>
        </div>
        <button @click.prevent="addQuestion('asks')">＋質問追加</button>
      </div>

      <!-- 単一選択 -->
      <div>
        <label>単一選択質問:</label>
        <div v-for="(ask, i) in entryForm.askChoices" :key="i">
          <input v-model="ask.question" placeholder="質問文" type="text" />
          <label>順序:
            <input v-model.number="ask.sequence" type="number" min="1" />
          </label>
          <div v-for="(choice, ci) in ask.choices" :key="ci">
            <input v-model="ask.choices[ci]" placeholder="選択肢" style="margin: 2px;" />
            <button @click.prevent="removeChoice('askChoices', i, ci)">−</button>
          </div>
          <button @click.prevent="addChoice('askChoices', i)">＋選択肢</button>
          <button @click.prevent="removeQuestion('askChoices', i)" v-if="entryForm.askChoices.length > 0">−質問削除</button>
        </div>
        <button @click.prevent="addQuestion('askChoices')">＋質問追加</button>
      </div>

      <!-- 複数選択 -->
      <div>
        <label>複数選択質問:</label>
        <div v-for="(ask, i) in entryForm.askMultiChoices" :key="i">
          <input v-model="ask.question" placeholder="質問文" type="text" />
          <label>順序:
            <input v-model.number="ask.sequence" type="number" min="1" />
          </label>
          <div v-for="(choice, ci) in ask.choices" :key="ci">
            <input v-model="ask.choices[ci]" placeholder="選択肢" style="margin: 2px;" />
            <button @click.prevent="removeChoice('askMultiChoices', i, ci)">−</button>
          </div>
          <button @click.prevent="addChoice('askMultiChoices', i)">＋選択肢</button>
          <button @click.prevent="removeQuestion('askMultiChoices', i)" v-if="entryForm.askMultiChoices.length > 0">−質問削除</button>
        </div>
        <button @click.prevent="addQuestion('askMultiChoices')">＋質問追加</button>
      </div>

      <!-- Dates -->
      <div>
        <label>日付:</label>
        <div v-for="(d, i) in entryForm.dates" :key="i">
          <input v-model="d.question" placeholder="質問文" type="text" />
          <label>順序:
            <input v-model.number="d.sequence" type="number" min="1" />
          </label>
          <div>
            <select v-model.number="d.dateType">
              <option :value="1">日時</option>
              <option :value="2">日付のみ</option>
              <option :value="3">時間のみ</option>
            </select>
          </div>
          <label>オプションキー:
            <input v-model="d.optionKey" type="checkbox" />
          </label>
          <button @click.prevent="removeQuestion('asks', i)" v-if="entryForm.asks.length > 0">−質問削除</button>
        </div>
        <button @click.prevent="addQuestion('asks')">＋質問追加</button>
      </div>

      <button type="submit" @click="submit">送信</button>
    </form>
    <div v-if="!entryForm">
      <a :href="'/entryFormEdit/?formJson=' + entryFormJson">新規作成</a>
    </div>
    <ul>
      <li v-for="f in forms">
        <a :href="'/entryFormEdit/' + f.entryFormID + '/'">{{ f.entryFormTitle }}</a>
      </li>
    </ul>

  </div>
  <div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>
</template>
