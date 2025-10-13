<script setup>
import { ref, onMounted, computed } from "vue"

import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue'
import SelectPeople from '@/components/SelectPeople.vue';

import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  id: String,
  group: String
})
const channelID = localStorage.getItem("channelID")
document.title = '回答一覧'

const targetNames = ref([])
const channel = ref(null)
const groups = ref([])
const aliases = ref([])
const answers = ref([])
const errorMessage = ref("")
let limit = 100
let offset = 0
console.log('id', props.id)
onMounted(async () => {
  channel.value = await getIDB('channel', channelID)
  groups.value = await getIDBs('group', 'channelIDIndex', channelID, 10000)
  aliases.value = await getIDBs('alias', 'channelIDIndex', channelID, 10000)
  answers.value = await getIDBs("answer", "askIDIndex", props.id, limit, offset)
  console.log('answers.value', answers.value)
})

// answers配列をsequence順にソートして返す
function getOrderedAnswers(ansObj) {
  if (!ansObj.answers) return []
  return [...ansObj.answers].sort((a, b) => a.sequence - b.sequence)
}

const missingNames = computed(() => {
  const existing = new Set(answers.value.map(a => a.answerBy))
  return targetNames.value.filter(name => !existing.has(name))
})

</script>

<template>
<Drawer v-if="aliases" :aliases="aliases" :channel="channel" />
<div id="content">
  <h2 class="sp_head">回答一覧</h2>
  <div v-if="errorMessage" class="errorMessage">{{ errorMessage }}</div>

  <SelectPeople v-if="aliases"
    :aliases="aliases"
    :groups="groups"
    :placeholder="'参加ユーザー'"
    :nonDisplay="true"
    :searchGroup="group"
    v-model="targetNames"
    />
  <div v-if="missingNames.length > 0">
    <h3>まだ回答していない人</h3>
    <ul>
      <li v-for="name in missingNames" :key="name">{{ name }}</li>
    </ul>
  </div>
  <div v-if="missingNames.length === 0 && targetNames.length > 0">
    全員回答しています
  </div>
  <div v-if="answers.length === 0">回答がありません。</div>
  <table v-else class="answer-table">
    <thead>
      <tr>
        <th>回答者</th>
        <th v-for="n in Math.max(...answers.map(a => a.answers?.length || 0))" :key="n">
          回答{{ n }}
        </th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="ans in answers" :key="ans.answerID">
        <td>{{ ans.answerBy }}</td>
        <td v-for="a in getOrderedAnswers(ans)" :key="a.sequence">
          <span v-if="Array.isArray(a.answer)">
            {{ a.answer.join(", ") }}
          </span>
          <span v-else>
            {{ a.answer }}
          </span>
        </td>
      </tr>
    </tbody>
  </table>
</div>
</template>

<style scoped>
.errorMessage {
  color: red;
  margin: 1rem 0;
}
.answer-table {
  border-collapse: collapse;
  width: 100%;
  margin-top: 1rem;
}
.answer-table th,
.answer-table td {
  border: 1px solid #ccc;
  padding: 6px 10px;
  text-align: left;
}
.answer-table th {
  background: #f5f5f5;
}

</style>
