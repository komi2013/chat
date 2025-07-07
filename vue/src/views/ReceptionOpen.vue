<script setup>
import { ref, onMounted } from 'vue'

import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue';

// Props: receptionID, menuID, code を受け取る
const props = defineProps({
  id: String,
  menuID: Number,
  passkey: String
})

// 状態定義
const channel = ref(null)
const reception = ref(null)
const waitMinutes = ref(0)
const inputGuests = ref(1) // 入力された人数（初期値1）

// 📡 Reception データ取得
async function findReception() {
  const fd = new FormData();
  fd.append('receptionID', props.id);
  fd.append('channelID', localStorage.getItem('channelID'));
  if (channel.value) {
    fd.append('aliasName', channel.value.myname);
  }
  fd.append('csrf', localStorage.getItem('csrf'));
  fd.append('passkey', props.passkey);
  const res = await sendRequest('/ReceptionGet/', fd);
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  })

  reception.value = res.reception;

  calculateWaitMinutes();
}

function calculateWaitMinutes() {
  if (!reception.value) return

  const targetMenu = reception.value.menus.find(
    m => m.menuID === Number(props.menuID)
  )
  if (!targetMenu) return

  const guestCount = parseInt(inputGuests.value)
  const queues = reception.value.queues ? [...reception.value.queues] : []

  const now = new Date()
  const activeBooks = reception.value.books.filter(book => {
    const start = new Date(book.bookStart)
    const end = new Date(book.bookEnd)
    return start <= now && now <= end && book.menuID === Number(props.menuID)
  })
  console.log('activeBooks', activeBooks);
  const totalBookedPeople = activeBooks.reduce((sum, book) => sum + (book.people || 1), 0)
  console.log('totalBookedPeople', totalBookedPeople);
  const facilityCapacity = reception.value.facilities?.reduce(
    (sum, f) => sum + f.facilityCount,
    0
  ) || Infinity

  console.log('facilityCapacity', facilityCapacity);

  // 🧠 今回のゲストも含めて収容オーバーかどうかを判定
  if (!isNaN(guestCount) && guestCount > 0) {
    const overCapacity = (totalBookedPeople + guestCount) > facilityCapacity

    if (overCapacity) {
      // キューに追加（順番待ち発生）
      queues.push({
        waitingGuest: guestCount,
        queueName: `番号札 #${queues.length + 1}`,
        queuedAt: now.toISOString()
      })
    } else {
      // 空いてるなら即入店可 → 待機なし
      waitMinutes.value = 0
      return
    }
  }

  // 🧮 キューに基づいて待機時間計算
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
onMounted(async () => {
  // channel.value = await getIDB('channel', localStorage.getItem('channelID'))
  findReception();
})
</script>

<template>

<Drawer />
  <div id="content" v-if="reception">
    <h2 class="sp_head">受付待ち時間</h2>

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
    <p v-else>すぐにご案内可能です 🙌</p>
  </div>
  <p v-else>URLが違います</p>
  <div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>
</template>
