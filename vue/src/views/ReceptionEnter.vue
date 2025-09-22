<script setup>
import { ref, onMounted } from 'vue'

import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue';

import { pushReceive } from '@/pushReceive/pushReceive.js'

// need API which edit waiting list, response with waiting list code 
// should be connected with reservation 
//      from QR code or from web site which can check congested or not

const props = defineProps({
  id: String,
  passkey: String
})

document.title = '受付手続き'

const reception = ref(null)
const waitMinutes = ref(0)
const inputGuests = ref(1) // 入力された人数（初期値1）

async function findReception() {
  const fd = new FormData()
  fd.append('channelID', localStorage.getItem('channelID'))
  fd.append('aliasName', myname)

  fd.append('csrf', localStorage.getItem('csrf'))
  fd.append('receptionID', props.id)
  fd.append('code', props.passkey)
  fd.append('codeType', '1') // 1 = before enter, 2 = at seat
  const res = await sendRequest('/ReceptionGet/', fd)
  if (!res.csrf) errorMessage.value = res
  if (res.error) errorMessage.value = res.error
  res.csrf && localStorage.setItem('csrf', res.csrf)
  res.pushContents.forEach(content => {
    pushReceive(content);
  })

  reception.value = res.reception;

  calculateWaitMinutes();
}

function calculateWaitMinutes() {
  if (!reception.value) return

  const guestCount = parseInt(inputGuests.value)
  const queues = reception.value.queues ? [...reception.value.queues] : []
  // 🧮 キューに基づいて待機時間計算
  console.log('queues', queues)
  const facilityCapacity = reception.value.facilities?.reduce(
    (sum, f) => sum + f.facilityCount,
    0
  ) || Infinity

  console.log('facilityCapacity', guestCount, facilityCapacity);

  // 🧠 今回のゲストも含めて収容オーバーかどうかを判定
  if (!isNaN(guestCount) && guestCount > 0) {
    const overCapacity = guestCount > facilityCapacity

    if (overCapacity) {
      // キューに追加（順番待ち発生）
      queues.push({
        waitingGuest: guestCount,
        queueName: `番号札 #${queues.length + 1}`,
        queuedAt: new Date().toISOString()
      })
    // } else {
    //   // 空いてるなら即入店可 → 待機なし
    //   waitMinutes.value = 0
    //   return
    }
  }


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

  console.log('🕒 計算結果 waitMinutes:', waitMinutes.value)
}


// IDB経由でchannel取得 → reception取得
let channel
let myname
const errorMessage = ref('')
onMounted(async () => {
  channel = await getIDB('channel', localStorage.getItem('channelID'))
  myname = channel ? channel.myname : ''
  findReception()
})


</script>

<template>
  <Drawer />
  <h2 class="sp_head">受付待ち時間</h2>
  <div v-if="errorMessage"> <div class="errorMessage">{{errorMessage}}</div></div>
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
      <button @click="calculateWaitMinutes" style="margin-left: 10px;">
        再計算
      </button>
    </div>

    <p v-if="waitMinutes > 0">
      お待ち時間は <strong>約 {{ waitMinutes }} 分</strong> です
    </p>
  </div>
  <div id="ad_right">
    <Advertisement /> <Advertisement /> <Advertisement />
  </div>
</template>
