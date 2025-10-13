<script setup>
import { ref, onMounted, computed } from "vue"

import Advertisement from '@/components/Advertisement.vue'
import Drawer from '@/components/Drawer.vue'

import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  id: String,
  code: String,
  entryForm: Object
})

document.title = 'フォーム回答'

const entryForm = ref(null)
const channel = ref(null)
const aliases = ref([])
const errorMessage = ref('')

// 回答を保存するオブジェクト
const answers = ref({})

onMounted(async () => {
  channel.value = await getIDB('channel', localStorage.getItem('channelID'))
  aliases.value = await getIDBs('alias', 'channelIDIndex', localStorage.getItem('channelID'), 10000)

  entryForm.value = await getIDB('entryForm', props.id)

  // 回答用の初期化
  if (entryForm.value) {
    for (const ask of entryForm.value.asks ?? []) {
      answers.value[`ask_${ask.sequence}`] = ""
    }
    for (const ask of entryForm.value.askChoices ?? []) {
      answers.value[`choice_${ask.sequence}`] = ""
    }
    for (const d of entryForm.value.dates ?? []) {
      answers.value[`date_${d.sequence}`] = ""
    }
    for (const ask of entryForm.value.askMultiChoices ?? []) {
      answers.value[`multi_${ask.sequence}`] = []
    }
  }
})

// sequence順に並べ替えた全質問を返す
const orderedQuestions = computed(() => {
  const merged = []
  for (const ask of entryForm.value?.asks ?? []) {
    merged.push({ ...ask, type: "text" })
  }
  for (const ask of entryForm.value?.askChoices ?? []) {
    merged.push({ ...ask, type: "select" })
  }
  for (const d of entryForm.value?.dates ?? []) {
    merged.push({ ...d, type: "date" })
  }
  for (const ask of entryForm.value?.askMultiChoices ?? []) {
    merged.push({ ...ask, type: "checkbox" })
  }
  return merged.sort((a, b) => a.sequence - b.sequence)
})

async function submit(event) {
  event.preventDefault()
  let answerDatas = []
  let optionKeys = []
  for (const q of orderedQuestions.value) {
    let value = null
    if (q.type === "text") {
      value = answers.value[`ask_${q.sequence}`]
    } else if (q.type === "select") {
      value = answers.value[`choice_${q.sequence}`]
    } else if (q.type === "date") {
      value = answers.value[`date_${q.sequence}`]
    } else if (q.type === "checkbox") {
      value = answers.value[`multi_${q.sequence}`]
    }
    answerDatas.push({
      sequence: q.sequence,
      answer: value,
    })
    // optionKey: q.optionKey
    if (q.optionKey) {
      optionKeys.push(value)
    }
  }
  console.log('submitted', answerDatas)
  const contents = [
    '1',
    entryForm.value.entryFormID,
    answerDatas,
    optionKeys
  ]
  const fd = new FormData()
  fd.append('pushNames', JSON.stringify(aliases.value.map(d => d.aliasName)))
  fd.append('channelID', localStorage.getItem("channelID"))
  fd.append('updatedBy', channel.value.myname)
  fd.append('csrf', localStorage.getItem('csrf'))
  fd.append('contents', JSON.stringify(contents))
  fd.append('pushTitle', 'answer')
  const res = await sendRequest('/ContentsPush/', fd)
  if (!res.csrf) errorMessage.value = res
  res.csrf && localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  // location.href = '/entryFormEdit/' + entryForm.value.formID + '/'
}
</script>

<template>
  <Drawer v-if="aliases" :aliases="aliases" :channel="channel" />
  <form id="content" @submit="submit">
    <h2 class="sp_head">フォーム回答</h2>
    <div v-if="errorMessage">
      <div class="errorMessage">{{ errorMessage }}<br />送信エラーが発生しました。</div>
    </div>

    <div v-for="q in orderedQuestions" :key="q.sequence" class="question-block">
      <label class="question">{{ q.sequence }}. {{ q.question }}</label>

      <!-- テキスト回答 -->
      <div v-if="q.type === 'text'">
        <input type="text" v-model="answers[`ask_${q.sequence}`]" />
      </div>

      <!-- セレクト回答 -->
      <div v-else-if="q.type === 'select'">
        <select v-model="answers[`choice_${q.sequence}`]">
          <option disabled value="">選択してください</option>
          <option v-for="(choice, ci) in q.choices" :key="ci" :value="choice">
            {{ choice }}
          </option>
        </select>
      </div>

      <!-- 日付回答 -->
      <div v-else-if="q.type === 'date'">
        <input
          v-if="q.dateType === 1"
          type="datetime-local"
          v-model="answers[`date_${q.sequence}`]"
        />
        <input
          v-else-if="q.dateType === 2"
          type="date"
          v-model="answers[`date_${q.sequence}`]"
        />
        <input
          v-else-if="q.dateType === 3"
          type="time"
          v-model="answers[`date_${q.sequence}`]"
        />
      </div>

      <!-- チェックボックス回答 -->
      <div v-else-if="q.type === 'checkbox'">
        <label v-for="(choice, ci) in q.choices" :key="ci">
          <input type="checkbox" :value="choice" v-model="answers[`multi_${q.sequence}`]" />
          {{ choice }}
        </label>
      </div>
    </div>

    <button type="submit">送信</button>
  </form>
  <div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>
</template>

<style scoped>
.question-block {
  margin: 1rem 0;
}
.question {
  display: block;
  font-weight: bold;
  margin-bottom: 0.5rem;
}
</style>
