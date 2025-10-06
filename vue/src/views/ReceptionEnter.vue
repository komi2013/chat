<script setup>
import { ref, onMounted } from 'vue'

import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue';
import { pushReceive } from '@/pushReceive/pushReceive.js'

const props = defineProps({
  id: String,
  passkey: String
})

document.title = '受付手続き'

const reception = ref(null)
const waitMinutes = ref(0)
const inputGuests = ref(1) // 入力された人数（初期値1）
const assignedFacility = ref(null) // 案内可能施設
const errorMessage = ref('')
let channel
let myname

async function findReception() {
  const fd = new FormData()
  fd.append('channelID', localStorage.getItem('channelID'))
  fd.append('aliasName', myname)

  fd.append('csrf', localStorage.getItem('csrf'))
  fd.append('receptionID', props.id)
  fd.append('code', props.passkey)
  fd.append('codeType', '1')
  const res = await sendRequest('/ReceptionGet/', fd)
  if (!res.csrf) errorMessage.value = res
  if (res.error) errorMessage.value = res.error
  res.csrf && localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }

  reception.value = res.reception
  calculateWaitMinutes()
}

function calculateWaitMinutes() {
  if (!reception.value) return

  const guestCount = parseInt(inputGuests.value)
  const queues = reception.value.queues ? [...reception.value.queues] : []
  const facilities = reception.value.facilities || []
  const now = new Date()

  assignedFacility.value = null

  // --- すでに待機列がある場合はそれを優先 ---
  if (queues.length > 0) {
    queues.push({
      waitingGuest: guestCount
    })
    let totalWait = 0
    queues.forEach(queue => {
      const guests = queue.waitingGuest
      const config = reception.value.waitConfigs.find(cfg => {
        const [min, max] = cfg.guestRange
        return guests >= min && (max === 0 || guests <= max)
      })
      const waitRatio = config ? config.waitRatio : 1
      totalWait += waitRatio
    })
    waitMinutes.value = totalWait
    return
  }

  // --- 待機列が無い場合のみ、新規割り当てを試みる ---
  if (!isNaN(guestCount) && guestCount > 0) {
    let remainingGuests = guestCount

    const availableFacilities = facilities.filter(f => {
      if (f.currentCode) return false

      const config = reception.value.waitConfigs.find(cfg => {
        const [min, max] = cfg.guestRange
        return guestCount >= min && (max === 0 || guestCount <= max)
      })
      const waitRatio = config ? config.waitRatio : 1
      const leaveTime = new Date(now.getTime() + waitRatio * 60000)

      const conflicts = (f.bookTimes || []).some(bt => {
        const start = new Date(bt.bookStart)
        const end = new Date(bt.bookEnd)
        const inReservedNow = now >= start && now <= end
        const overlapWithFutureBooking = leaveTime > start && now < end
        return inReservedNow || overlapWithFutureBooking
      })
      if (conflicts) return false

      const current = f.currentGuests || 0
      const available = f.capacity - current
      return available > 0
    })

    if (availableFacilities.length > 0) {
      assignedFacility.value = availableFacilities
        .filter(f => f.capacity >= guestCount)
        .sort((a, b) => a.capacity - b.capacity)[0] ||
        availableFacilities.sort((a, b) => a.capacity - b.capacity)[0]

      if (assignedFacility.value) {
        const current = assignedFacility.value.currentGuests || 0
        const available = assignedFacility.value.capacity - current
        const toAssign = Math.min(remainingGuests, available)
        assignedFacility.value.currentGuests = current + toAssign
        remainingGuests -= toAssign
      }
    }
  }

  // --- 待機時間を再計算 ---
  let totalWait = 0
  queues.forEach(queue => {
    const guests = queue.waitingGuest
    const config = reception.value.waitConfigs.find(cfg => {
      const [min, max] = cfg.guestRange
      return guests >= min && (max === 0 || guests <= max)
    })
    const waitRatio = config ? config.waitRatio : 1
    totalWait += waitRatio
  })

  waitMinutes.value = totalWait
}

// IDB経由でchannel取得 → reception取得
onMounted(async () => {
  channel = await getIDB('channel', localStorage.getItem('channelID'))
  myname = channel ? channel.myname : ''
  await findReception()
})

async function queueUp() {
  if (!confirm("実行▶️")) return
  const fd = new FormData()
  fd.append('channelID', localStorage.getItem('channelID'))
  fd.append('aliasName', myname)
  fd.append('csrf', localStorage.getItem('csrf'))
  fd.append('receptionID', props.id)
  fd.append('code', props.passkey)
  fd.append('codeType', '1')
  fd.append('guestCount', inputGuests.value)
  fd.append('editType', 1)
  const res = await sendRequest('/ReceptionQueueEdit/', fd)
  if (!res.csrf) errorMessage.value = res
  if (res.error) errorMessage.value = res.error
  res.csrf && localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  location.href = ''
}

async function queueRemove(queueName) {
  if (!confirm("実行▶️")) return
  const fd = new FormData()
  fd.append('channelID', localStorage.getItem('channelID'))
  fd.append('aliasName', myname)
  fd.append('csrf', localStorage.getItem('csrf'))
  fd.append('receptionID', props.id)
  fd.append('queueName', queueName)
  fd.append('editType', 2)
  const res = await sendRequest('/ReceptionQueueEdit/', fd)
  if (!res.csrf) errorMessage.value = res
  if (res.error) errorMessage.value = res.error
  res.csrf && localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  location.href = ''
}


</script>

<template>
  <Drawer />
  <h2 class="sp_head">受付待ち時間</h2>
  <div v-if="errorMessage"> 
    <div class="errorMessage">{{ errorMessage }}</div>
  </div>

  <div id="content" v-if="reception">
    <div style="margin-bottom: 10px;">
      <label for="guestCount">何名さまですか？</label>
      <input
        id="guestCount"
        type="number"
        v-model.number="inputGuests"
        min="1"
        style="margin-left: 5px; width: 60px;"
      />
      <button @click="calculateWaitMinutes" style="margin-left: 10px;">再計算</button>
    </div>

    <!-- 案内できる施設がある場合 -->
    <div v-if="assignedFacility">
      <p>
        ✅ ご案内可能: <strong>{{ assignedFacility.facilityName }}</strong> 
        (定員 {{ assignedFacility.capacity }} 名)
      </p>
    </div>

    <!-- 待ち時間表示 -->
    <p v-else-if="waitMinutes > 0">
      お待ち時間は <strong>約 {{ waitMinutes }} 分</strong> です
    </p>

    <!-- 並ぶボタン -->
    <div v-if="!assignedFacility || reception.adminNames">
      <button @click="queueUp">並ぶ</button>
    </div>

    <!-- スタッフのみ queues 表示 -->
    <div v-if="reception.adminNames">
      <h3>現在の待機列</h3>
      <div v-for="(q, idx) in reception.queues" :key="idx">
        <button @click="queueRemove(q.queueName)">削除</button>{{ q.queueName }} - {{ q.waitingGuest }} 名 (受付: {{ new Date(q.queuedAt).toLocaleTimeString() }})
      </div>
    </div>
  </div>

  <div id="ad_right">
    <Advertisement /> <Advertisement /> <Advertisement />
  </div>
</template>
