<script setup>
import { ref, onMounted } from 'vue'

import Drawer from '@/components/Drawer.vue'
import Advertisement from '@/components/Advertisement.vue'

import { pushReceive } from '@/pushReceive/pushReceive.js'
// import { sendRequest } from '@/my/api'


document.title = '広告設定'

const weekdays = ['', '日', '月', '火', '水', '木', '金', '土']

// --- reactive 変数 ---
const ads = ref([])
// const adPrices = ref([])
const previewBanner = ref({})  // 各広告ごとに画像を保持 { adID: base64 }
const previewSquare = ref({})
const errorMessage = ref('')
const fetched = ref(false)

// --- 初期データ取得 ---
onMounted(async () => {
  await findAds()
  // if (ads.value.length > 0) await findAdPrices(0)
  fetched.value = true
})

let systemWalletAddress
let jpycCheckURL
async function findAds() {
  const fd = new FormData()
  fd.append('csrf', localStorage.getItem('csrf'))
  const res = await sendRequest('/AdGet/', fd)
  if (!res.csrf) {
    errorMessage.value = res
    return
  }
  localStorage.setItem('csrf', res.csrf)
  res.pushContents.forEach(content => pushReceive(content))
  ads.value = ((res.ads && res.ads.length) ? res.ads : [undefined])
    .map(apiAd => convertApiAdToUiAd(apiAd))

  // 各広告ごとの初期画像をセット
  ads.value.forEach(ad => {
    if (ad.pathBanner) previewBanner.value[ad.adID] = ad.pathBanner
    if (ad.pathSquare) previewSquare.value[ad.adID] = ad.pathSquare
  })
  ads.value.push(convertApiAdToUiAd({}))
  systemWalletAddress = res.systemWalletAddress
  jpycCheckURL = res.jpycCheckURL
}

// --- APIデータをUI形式に変換 ---
function convertApiAdToUiAd(apiAd) {
  apiAd = apiAd || {}
  const { day: adStartDay, hour: adStartHour } = splitDayHourFromString(apiAd.adStart)
  const { day: adEndDay, hour: adEndHour } = splitDayHourFromString(apiAd.adEnd)

  const coordinateInput = apiAd.latitude ? `${apiAd.latitude}, ${apiAd.longitude}` : ''
  return {
    adID: apiAd.adID || '',
    pathBanner: apiAd.pathBanner || '',
    pathSquare: apiAd.pathSquare || '',
    adText: apiAd.adText || '',
    adLink: apiAd.adLink || '',
    coordinateInput,
    latitude: apiAd.latitude || 0,
    longitude: apiAd.longitude || 0,
    adStartDay,
    adStartHour,
    adEndDay,
    adEndHour,
    distance: apiAd.distance || 0,
    userID: apiAd.userID || '',
    adYen: apiAd.adYen || 0,
  }
}

// --- 日時分割 ---
function splitDayHourFromString(input) {
  const str = String(input).padStart(3, '0')
  const day = parseInt(str[0], 10) || 0
  const hour = parseInt(str.slice(1), 10) || 0
  return { day, hour }
}

// --- 価格取得 ---
async function findAdPrices(index) {
  const ad = ads.value[index]
  const fd = new FormData()
  fd.append('csrf', localStorage.getItem('csrf'))
  fd.append('pathSquare', ad.pathSquare)
  fd.append('latitude', ad.latitude)
  fd.append('longitude', ad.longitude)
  fd.append('adStart', `${ad.adStartDay}${String(ad.adStartHour).padStart(2, '0')}`)
  fd.append('adEnd', `${ad.adEndDay}${String(ad.adEndHour).padStart(2, '0')}`)
  fd.append('distance', ad.distance)
  const res = await sendRequest('/AdPriceGet/', fd)
  if (!res.csrf) {
    errorMessage.value = res
    return
  }

  localStorage.setItem('csrf', res.csrf)
  res.pushContents.forEach(content => pushReceive(content))

  if (Array.isArray(res.adPrices) && res.adPrices.length > 0) {
    const sorted = res.adPrices.sort((a, b) => b.adPriceYen - a.adPriceYen)
    ad.adYen = sorted[0].adYen || 0
  } else {
    ad.adYen = 0
  }
  // adPrices.value = res.adPrices
}

// --- 経緯度処理 ---
function parseCoordinates(ad) {
  const parts = ad.coordinateInput.split(',').map(s => s.trim())
  if (parts.length !== 2) return (ad.latitude = ad.longitude = 0)

  const lat = Number(parts[0])
  const lng = Number(parts[1])
  if (!isNaN(lat) && !isNaN(lng)) {
    ad.latitude = lat.toFixed(2)
    ad.longitude = lng.toFixed(2)
  } else {
    ad.latitude = ad.longitude = 0
  }
}

// --- 登録 ---
const payment = ref(false)
async function submitAd(index) {
  const ad = ads.value[index]
  const fd = new FormData()
  fd.append('csrf', localStorage.getItem('csrf'))

  const bannerBase64 = await toBase64IfNeeded(previewBanner.value[ad.adID])
  const squareBase64 = await toBase64IfNeeded(previewSquare.value[ad.adID])
  fd.append('adID', ad.adID ?? '')
  fd.append('previewBanner', bannerBase64)
  fd.append('previewSquare', squareBase64)
  fd.append('adText', ad.adText)
  fd.append('adLink', ad.adLink)
  fd.append('latitude', ad.latitude)
  fd.append('longitude', ad.longitude)
  fd.append('adStart', `${ad.adStartDay}${String(ad.adStartHour).padStart(2, '0')}`)
  fd.append('adEnd', `${ad.adEndDay}${String(ad.adEndHour).padStart(2, '0')}`)
  fd.append('distance', ad.distance)
  fd.append('payment', payment.value ? 'true' : '')

  const res = await sendRequest('/AdEdit/', fd)
  if (!res.csrf) {
    errorMessage.value = res
    return
  }

  localStorage.setItem('csrf', res.csrf)
  res.pushContents.forEach(content => pushReceive(content))
}

async function invoiceAd(index) {
  const ad = ads.value[index]
  const fd = new FormData()
  fd.append('csrf', localStorage.getItem('csrf'))
  fd.append('adID', ad.adID)
  fd.append('nextPayment', '1')

  const res = await sendRequest('/AdEdit/', fd)
  if (!res.csrf) {
    errorMessage.value = res
    return
  }

  localStorage.setItem('csrf', res.csrf)
  res.pushContents.forEach(content => pushReceive(content))
}


// --- 削除 ---
async function deleteAd(index) {
  const ad = ads.value[index]
  if (!ad || !ad.adID) {
    errorMessage.value = '削除する広告が特定できません。'
    return
  }

  if (!confirm('この広告を削除しますか？')) return

  const fd = new FormData()
  fd.append('csrf', localStorage.getItem('csrf'))
  fd.append('adID', ad.adID)

  const res = await sendRequest('/AdDelete/', fd)
  if (!res.csrf) {
    errorMessage.value = res
    return
  }

  localStorage.setItem('csrf', res.csrf)
  if (res.pushContents) res.pushContents.forEach(content => pushReceive(content))

  ads.value.splice(index, 1)
}

// --- 画像処理 ---
async function toBase64IfNeeded(value) {
  if (isBase64Image(value)) return value
  return await imageUrlToBase64(value)
}

function isBase64Image(str) {
  return typeof str === 'string' && str.startsWith('data:image/')
}

function imageUrlToBase64(imageUrl) {
  return new Promise(resolve => {
    if (!imageUrl) return resolve('')
    const img = new Image()
    img.crossOrigin = 'Anonymous'
    img.onload = () => {
      const canvas = document.createElement('canvas')
      canvas.width = img.width
      canvas.height = img.height
      const ctx = canvas.getContext('2d')
      ctx.drawImage(img, 0, 0)
      resolve(canvas.toDataURL('image/png'))
    }
    img.onerror = () => resolve('')
    img.src = imageUrl
  })
}

function handleTrim(event, adID, targetW, targetH, type) {
  const file = event.target.files[0]
  if (!file) return

  const reader = new FileReader()
  reader.onload = () => {
    const img = new Image()
    img.onload = () => {
      const canvas = document.createElement('canvas')
      canvas.width = targetW
      canvas.height = targetH
      const ctx = canvas.getContext('2d')

      const sx = Math.max(0, (img.width - targetW) / 2)
      const sy = Math.max(0, (img.height - targetH) / 2)
      const sw = Math.min(img.width, targetW)
      const sh = Math.min(img.height, targetH)
      ctx.drawImage(img, sx, sy, sw, sh, 0, 0, targetW, targetH)

      const base64 = canvas.toDataURL('image/png')
      if (type === 'banner') previewBanner.value[adID] = base64
      else previewSquare.value[adID] = base64
    }
    img.src = reader.result
  }
  reader.readAsDataURL(file)
}

</script>

<template>
  <div id="drawer_column"><Drawer /></div>
  <div id="content">
    <br>
    <div v-if="errorMessage" class="errorMessage">{{ errorMessage }}</div>
    <div v-for="(ad, index) in ads" :key="ad.adID || index" class="ads">
      <form @submit.prevent="submitAd(index)">
        <label>正方形画像（250x250）:<br />
          <input type="file" accept="image/*" @change="e => handleTrim(e, ad.adID, 250, 250, 'square')" />
        </label>
        <br />
        <img v-if="previewSquare[ad.adID]" :src="previewSquare[ad.adID]" style="border:1px solid #ccc; width:250px; height:250px;" />
        <br /><br />

        <label>広告リンク<br />
          <input v-model="ad.adLink" placeholder="https://sample.com/item?af=1" required pattern="https://.*" class="wide-text" />
        </label>

        <br /><br />

        <label>経緯度:<br />
          <input v-model="ad.coordinateInput" pattern="^-?\\d+(\\.\\d+)?,\\s*-?\\d+(\\.\\d+)?$"
                 title="緯度と経度は「35.77, 139.57」の形式で入力してください"
                 @input="parseCoordinates(ad)" class="wide-text" />
        </label>

        <div v-if="ad.latitude && ad.longitude" style="font-size:14px;">
          ➤ 緯度: <strong>{{ ad.latitude }}</strong><br />
          ➤ 経度: <strong>{{ ad.longitude }}</strong>
        </div>

        <br />

        <label>半径約: <input type="number" v-model="ad.distance" /> km</label>

        <br /><br />

        <label>開始曜日:
          <select v-model="ad.adStartDay">
            <option v-for="(name, i) in weekdays" :key="i" :value="i">{{ name }}</option>
          </select>
        </label>

        <label>終了曜日:
          <select v-model="ad.adEndDay">
            <option v-for="(name, i) in weekdays" :key="i" :value="i">{{ name }}</option>
          </select>
        </label>

        <div>見積り価格: ¥{{ ad.adYen?.toLocaleString() || 0 }}</div>

        <div class="centralize"><button type="button" @click="findAdPrices(index)">確認</button></div>
        <div class="centralize"><button type="submit">登録・更新</button></div>
        <div class="centralize"><button type="button" @click="invoiceAd(index)">請求書発行</button></div>
        <div class="centralize"><button type="button" @click="deleteAd(index)">削除</button></div>
      </form>
    </div>
    <div class="ads"><span>システムJPYCアドレス口座: </span>{{systemWalletAddress}}</div>
    <div class="ads">
      <a :href="jpycCheckURL" target="_blank">システムJPYCアドレス口座履歴URL</a>
    </div>
  </div>
  <div id="ad_right"><Advertisement /><Advertisement /><Advertisement /></div>
</template>

<style scoped>
.wide-text {
  width: 100%;
  max-width: 400px;
}
.ads {
  padding: 1rem;
  margin-bottom: 1rem;
}
.centralize {
  text-align: center;
  width: 100%;
}
.centralize button {
  /*background-color: blue;*/
  width: 50%;
  margin: 12px;
  padding: 4px;
}

/*.centralize div {
  margin-top: -10px
}
*/
</style>
