<script setup>
import { ref, onMounted } from 'vue'

import Advertisement from '@/components/Advertisement.vue'
import DrawerReception from '@/components/DrawerReception.vue'

import { pushReceive } from '@/pushReceive/pushReceive.js'

const props = defineProps({
  id: String,
  code: String
});

document.title = '注文履歴'

const reception = ref(null);
const receptionOrders = ref([]);
const errorMessage = ref('');
const totalPrice = ref(0);
const seatName = ref('')
const iamStaff = ref(false);
async function findReception() {
  const fd = new FormData()
  fd.append('channelID', localStorage.getItem('channelID'))
  fd.append('aliasName', myname)
  // ↑ for staff
  fd.append('csrf', localStorage.getItem('csrf'))
  fd.append('receptionID', props.id)
  fd.append('code', props.code)
  fd.append('codeType', '3') // 1 = before enter, 2 = take QR code at table, 3 = ordering
  const res = await sendRequest('/ReceptionGet/', fd)
  if (res.error) { errorMessage.value = res.error }
  res.csrf && localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }

  reception.value = res.reception || null;
  seatName.value = res.facilityName || ''
  console.log('seatName', seatName.value)

  // console.log('joinNames', reception.value.joinNames)
  iamStaff.value = Array.isArray(reception.value.joinNames) && reception.value.joinNames.length > 0
  if (reception.value) {
    await fetchOrders()
  }
}

async function fetchOrders() {
  const orders = await getAllIDBs('receptionOrder')
  const orderSeats = orders.filter(order => order.seatName === seatName.value)
  receptionOrders.value = orderSeats
  totalPrice.value = orderSeats.reduce((sum, o) => sum + o.price, 0)
}

function getItemDetail(id) {
  if (!reception.value || !reception.value.itemDetails) return null;
  return reception.value.itemDetails.find(item => String(item.itemID) === String(id)) || null;
}
const channel = ref(null)
let myname
onMounted(async() => {
  channel.value = await getIDB('channel', localStorage.getItem("channelID"))
  myname = channel.value ? channel.value.myname : ''
  findReception()
})

async function deleteOrder() {
  if (!confirm("会計終了")) { return }
  const fd = new FormData()
  fd.append('channelID', localStorage.getItem('channelID'))
  fd.append('aliasName', myname)
  // ↑ for staff
  fd.append('receptionID', props.id)
  fd.append('code', props.code)
  fd.append('csrf', localStorage.getItem('csrf'))
  const res = await sendRequest('/ReceptionOrderDelete/', fd)
  if (res.error) { errorMessage.value = res.error }
  res.csrf && localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  // location.href = ''
}

</script>

<template>
<div id="drawer_column"><DrawerReception :id="id" :code="code" /></div>
  <div id="content">
    <h1 class="sp_head">注文履歴</h1>

    <div v-if="errorMessage"> 
      <div class="errorMessage">{{ errorMessage }}</div>
    </div>

    <div class="reception-orders">
      <h2>カート内の注文 {{seatName}}</h2>
      <div
        v-if="receptionOrders.length > 0"
        v-for="order in receptionOrders"
        :key="order.receptionOrderID"
        class="order-item"
      >
        <h3>{{ order.menuName }}</h3>
        <div v-if="iamStaff"><strong>注文ID:</strong> {{ order.receptionOrderID }}</div>

        <!-- 注文アイテム -->
        <div
          v-for="item in order.items"
          :key="item.itemID"
          class="item-detail"
        >
          <img
            v-if="item.imgPath"
            :src="item.imgPath"
            :alt="item.itemName"
            class="item-img"
          />
          <span>{{ item.itemName }}</span>
          <div v-if="item.choices && item.choices.length > 0">
            <div v-for="(group, groupIndex) in item.choices" :key="groupIndex">
              <span v-if="order.itemChoices">
                {{ order.itemChoices[`item-${item.itemID}-${groupIndex}`] }}
              </span>
            </div>
          </div>
        </div>

        <!-- オプション -->
        <div class="options">
          <!-- 無料オプション -->
          <div v-if="order.freeOptions && order.freeOptions.length > 0">
            <p><strong>無料オプション:</strong></p>
            <div class="option-list">
              <div
                v-for="id in order.freeOptions"
                :key="'free-' + id"
                class="option-item"
              >
                <img
                  v-if="getItemDetail(id)?.imgPath"
                  :src="getItemDetail(id).imgPath"
                  :alt="getItemDetail(id).itemName"
                  class="option-img"
                />
                <div>{{ getItemDetail(id)?.itemName }}</div>
              </div>
            </div>
          </div>

          <!-- 有料オプション -->
          <div v-if="order.paidOptions && order.paidOptions.length > 0">
            <p><strong>有料オプション:</strong></p>
            <div class="option-list">
              <div
                v-for="opt in order.paidOptions"
                :key="'paid-' + opt.itemID"
                class="option-item"
              >
                <img
                  v-if="getItemDetail(opt.itemID)?.imgPath"
                  :src="getItemDetail(opt.itemID).imgPath"
                  :alt="getItemDetail(opt.itemID).itemName"
                  class="option-img"
                />
                <div>
                  {{ getItemDetail(opt.itemID)?.itemName }} (+{{ opt.price }}円)
                </div>
              </div>
            </div>
          </div>

          <!-- 複数選択オプション -->
          <div v-if="order.freeMultiOptions && order.freeMultiOptions.length > 0">
            <p><strong>複数選択オプション:</strong></p>
            <div class="option-list">
              <div
                v-for="id in order.freeMultiOptions"
                :key="'multi-' + id"
                class="option-item"
              >
                <img
                  v-if="getItemDetail(id)?.imgPath"
                  :src="getItemDetail(id).imgPath"
                  :alt="getItemDetail(id).itemName"
                  class="option-img"
                />
                <div>{{ getItemDetail(id)?.itemName }}</div>
              </div>
            </div>
          </div>
        </div>

        <div>メニュー価格: {{ order.price }}円</div>
      </div>

      <p v-else>カートに商品はありません。</p>

      <div class="total-price">
        <strong>合計: {{ totalPrice }} 円</strong>
      </div>
      <div v-if="iamStaff">
        <button @click="deleteOrder()">会計終了</button>
      </div>
    </div>
  </div>

  <div id="ad_right">
    <Advertisement /> 
    <Advertisement /> 
    <Advertisement />
  </div>
</template>

<style scoped>
.order-item {
  border: 1px solid #ccc;
  margin-bottom: 15px;
  padding: 10px;
  border-radius: 5px;
}
.item-img, .option-img {
  width: 40px;
  height: 40px;
  margin-right: 5px;
}
.option-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
.delete-btn {
  margin: 5px 0;
  cursor: pointer;
  color: red;
}
.total-price {
  margin-top: 20px;
  font-size: 18px;
}
</style>
