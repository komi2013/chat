<script setup>
import { ref, onMounted } from 'vue'

import Drawer from '@/components/Drawer.vue'
import Advertisement from '@/components/Advertisement.vue'

import { pushReceive } from '@/pushReceive/pushReceive.js'

// Solana送金用のインポート（サーバーサイドで処理するため、フロントエンドでは不要）


document.title = '広告設定'

// const weekdays = ['', '日', '月', '火', '水', '木', '金', '土']

// --- reactive 変数 ---
const ads = ref([])
// const adPrices = ref([])
const previewBanner = ref({})  // 各広告ごとに画像を保持 { adID: base64 }
const previewSquare = ref({})
const errorMessage = ref('')
const fetched = ref(false)
const jpycBalance = ref(null) // JPYC残高
const userWalletAddress = ref('') // ユーザーのウォレットアドレス
const loadingBalance = ref(false) // 残高取得中フラグ

// --- 初期データ取得 ---
onMounted(async () => {
  await findAds()
  fetched.value = true
})

let systemWalletAddress
let systemFeeWalletAddress // システム利用料を受け取るウォレットアドレス
let systemFeePayerPublicKey // Fee Payerの公開鍵（バックエンドから取得）
let jpycCheckURL
let jpycMintAddress // JPYC SPL Token Mint Address
const errors = ref([]);
async function findAds() {
  const fd = new FormData()
  fd.append('csrf', localStorage.getItem('csrf'))
  const res = await sendRequest('/AdGet/', fd)
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
  ads.value = ((res.ads && res.ads.length) ? res.ads : [undefined])
    .map(apiAd => convertApiAdToUiAd(apiAd))

  ads.value.forEach(ad => {
    if (ad.pathSquare) previewSquare.value[ad.adID] = ad.pathSquare

    const adID = ad.adID || ''  // ← adID 無い場合は仮ID
    if (!errors.value[adID]) {
      errors.value[adID] = {
        square: null,
        banner: null,
      }
    }
  })
  systemWalletAddress = res.systemWalletAddress
  systemFeeWalletAddress = res.systemFeeWalletAddress || res.systemWalletAddress // デフォルトはシステムウォレット
  systemFeePayerPublicKey = res.systemFeePayerPublicKey // バックエンドから取得
  jpycCheckURL = res.jpycCheckURL
  jpycMintAddress = res.jpycMintAddress || "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v" // デフォルト値（実際のJPYC Mintアドレスに置き換え）
  
  // AdGetから返される残高とウォレットアドレスを使用
  if (res.jpycBalance !== undefined) {
    jpycBalance.value = res.jpycBalance
  }
  if (res.solanaWalletAddress) {
    userWalletAddress.value = res.solanaWalletAddress
  }
}


const invoiceButton = ref(true)
function convertApiAdToUiAd(apiAd) {
  apiAd = apiAd || {}
  const adStart = apiAd.adStart ? apiAd.adStart.slice(0, 16) : ''
  const adEnd   = apiAd.adEnd ? apiAd.adEnd.slice(0, 16) : ''
  let adEndDays = 1
  if (adStart && adEnd) {
    const start = new Date(adStart)
    const end = new Date(adEnd)
    adEndDays = Math.ceil((end - start) / (1000 * 60 * 60 * 24))
  }
  const invoicedAt = apiAd.invoicedAt ? new Date(apiAd.invoicedAt) : null
  const paidAt     = apiAd.paidAt ? new Date(apiAd.paidAt) : null
  if (invoicedAt && !isNaN(invoicedAt.getTime())) {
    if (!paidAt || paidAt < invoicedAt) {
      invoiceButton.value = false
    } else {
      invoiceButton.value = true
    }
  }
  return {
    adID: apiAd.adID || '',
    pathBanner: apiAd.pathBanner || '',
    pathSquare: apiAd.pathSquare || '',
    adText: apiAd.adText || '',
    adLink: apiAd.adLink || '',
    adStart,
    adEnd,
    adEndDays,
    latitude: apiAd.latitude || 0,
    longitude: apiAd.longitude || 0,
    distance: apiAd.distance > -1 ? apiAd.distance : -1,
    adYen: apiAd.adYen || 0,
    userID: apiAd.userID || '',
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
// async function findAdPrices(index) {
//   const ad = ads.value[index]
//   const endDate = new Date(ad.adStart)
//   endDate.setDate(endDate.getDate() + Number(ad.adEndDays))
//   const adEndFormatted = endDate.toISOString().slice(0, 16)
//   const fd = new FormData()
//   fd.append('csrf', localStorage.getItem('csrf'))
//   fd.append('pathSquare', ad.pathSquare)
//   fd.append('latitude', ad.latitude)
//   fd.append('longitude', ad.longitude)
//   fd.append('adStart', ad.adStart)
//   fd.append('adEnd', adEndFormatted)
//   fd.append('distance', ad.distance)
//   const res = await sendRequest('/AdPriceGet/', fd)
//   if (!res.csrf) {
//     errorMessage.value = res
//     return
//   }
//   localStorage.setItem('csrf', res.csrf)
//   if (Array.isArray(res.pushContents)) {
//     for (const content of res.pushContents) {
//       await pushReceive(content)
//     }
//   }
//   if (res.error) { errorMessage.value = res.error }
//   if (Array.isArray(res.adPrices) && res.adPrices.length > 0) {
//     const sorted = res.adPrices.sort((a, b) => b.adPriceYen - a.adPriceYen)
//     ad.adYen = sorted[0].adYen || 0
//   } else {
//     ad.adYen = 0
//   }
//   // adPrices.value = res.adPrices
// }

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
  ad.distance = 0
}

// --- 登録 ---
const payment = ref(false)
async function submitAd(index) {
  if (!confirm('広告設定の登録・更新')) return
  const ad = ads.value[index]
  const adID = ad.adID ?? ''
  const fd = new FormData()
  fd.append('csrf', localStorage.getItem('csrf'))

  // const bannerBase64 = await toBase64IfNeeded(previewBanner.value[ad.adID])
  const squareBase64 = await toBase64IfNeeded(previewSquare.value[adID])

  if (!squareBase64) {
    // if (!errors.value[adID]) errors.value[adID] = {}
    errors.value[adID].square = "画像を選択してください。"
    return
  }

  const startLocal = new Date(ad.adStart)     // ローカルとして扱われる
  const endLocal = new Date(startLocal)
  endLocal.setDate(endLocal.getDate() + Number(ad.adEndDays))
  const adEnd = endLocal.toISOString()
  const adEndLocalText = timeFormat('YYYY-MM-DDThh:mm', endLocal)

  fd.append('adID', adID)
  // fd.append('previewBanner', bannerBase64)
  // fd.append('adText', ad.adText)
  fd.append('previewSquare', squareBase64)
  fd.append('adLink', ad.adLink)
  fd.append('latitude', ad.latitude)
  fd.append('longitude', ad.longitude)
  fd.append('adStart', ad.adStart)
  fd.append('adEnd', adEndLocalText)
  fd.append('distance', ad.distance)
  fd.append('payment', payment.value ? 'true' : '')
  const res = await sendRequest('/AdEdit/', fd)
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
  // console.log(res.ad.adYen)
  ad.adID = res.ad.adID
  ad.adYen = res.ad.adYen
}

async function invoiceAd(index) {
  if (!confirm('請求書を発行します。Solanaのアドレスを登録してから請求書発行お願いします')) return
  const ad = ads.value[index]
  const fd = new FormData()
  fd.append('csrf', localStorage.getItem('csrf'))
  fd.append('adID', ad.adID)
  const res = await sendRequest('/AdInvoice/', fd)
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

// --- JPYC残高取得 ---
async function loadJpycBalance() {
  try {
    loadingBalance.value = true
    const fd = new FormData()
    fd.append('csrf', localStorage.getItem('csrf'))
    const res = await sendRequest('/SolanaJpycBalance/', fd)
    
    if (!res.csrf) {
      errorMessage.value = res
      jpycBalance.value = null
      userWalletAddress.value = ''
      return
    }
    localStorage.setItem('csrf', res.csrf)
    
    if (Array.isArray(res.pushContents)) {
      for (const content of res.pushContents) {
        await pushReceive(content)
      }
    }
    
    if (res.error) {
      // ウォレットが作成されていない場合
      jpycBalance.value = null
      userWalletAddress.value = ''
    } else {
      jpycBalance.value = res.balance || 0
      userWalletAddress.value = res.walletAddress || ''
    }
  } catch (error) {
    console.error('JPYC残高取得エラー:', error)
    jpycBalance.value = null
    userWalletAddress.value = ''
  } finally {
    loadingBalance.value = false
  }
}

// --- ウォレット作成 ---
async function createWallet() {
  if (!confirm('Solanaウォレットを作成しますか？')) return
  
  try {
    const fd = new FormData()
    fd.append('csrf', localStorage.getItem('csrf'))
    const res = await sendRequest('/SolanaWalletCreate/', fd)
    
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
    
    if (res.error) {
      errorMessage.value = res.error
    } else {
      alert(res.message || 'ウォレットが作成されました')
      userWalletAddress.value = res.solanaWalletAddress
      // 残高を再取得
      await loadJpycBalance()
    }
  } catch (error) {
    console.error('ウォレット作成エラー:', error)
    errorMessage.value = 'ウォレット作成に失敗しました'
  }
}

// --- ガスレス決済（Fee Relayer）によるSolana送金処理 ---
async function payForAd(index) {
  const ad = ads.value[index]
  if (!ad || !ad.adID) {
    errorMessage.value = '広告が特定できません。'
    return
  }
  
  if (!userWalletAddress.value) {
    errorMessage.value = 'ウォレットが作成されていません。先にウォレットを作成してください。'
    return
  }

  // 広告料金をJPYCに変換（1円 = 1 JPYC）
  const jpycAmount = ad.adYen // 円単位
  const jpycAmountLamports = Math.floor(jpycAmount * 1_000_000) // JPYCは通常decimals=6

  if (jpycAmountLamports <= 0) {
    errorMessage.value = '送金金額が0以下です。'
    return
  }

  // システム利用料（1 JPYC）
  const systemFeeAmount = 1_000_000 // 1 JPYC = 1,000,000 lamports (decimals=6)

  if (!confirm(`¥${ad.adYen?.toLocaleString() || 0} (${(jpycAmountLamports / 1_000_000).toFixed(6)} JPYC) + システム利用料 1 JPYC を送金しますか？`)) {
    return
  }

  try {
    // バックエンドへ送信（サーバーサイドでトランザクションを作成・署名・送信）
    const fd = new FormData()
    fd.append('csrf', localStorage.getItem('csrf'))
    fd.append('adID', ad.adID)
    
    const res = await sendRequest('/PaymentExecute/', fd)
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
    
    if (res.error) {
      errorMessage.value = res.error
    } else if (res.signature) {
      alert(`支払いが完了しました。\nトランザクション: ${res.signature}`)
      // 広告情報を再取得
      await findAds()
      // JPYC残高を再取得
      await loadJpycBalance()
    } else {
      errorMessage.value = '支払い処理が完了しましたが、トランザクション署名が取得できませんでした。'
    }
  } catch (error) {
    console.error('Solana送金エラー:', error)
    if (error.message) {
      errorMessage.value = `送金エラー: ${error.message}`
    } else {
      errorMessage.value = '送金処理中にエラーが発生しました。'
    }
  }
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
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  if (res.error) { errorMessage.value = res.error }
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
          <br>
        </label>
        <div v-if="errors[ad.adID]?.square" class="errorMessage">
          {{ errors[ad.adID].square }}
        </div>
        <img v-if="previewSquare[ad.adID]" :src="previewSquare[ad.adID]" style="border:1px solid #ccc; width:250px; height:250px;" />
        <br /><br />

        <label>広告リンク<br />
          <input v-model="ad.adLink" placeholder="https://sample.com/item?af=1" required pattern="https://.*" 
            title="https:// で始まる URL を入力してください" class="wide-text" />
        </label>

        <br /><br />

        <label>経緯度:<a href="https://maps.google.com/" target="_blank">Googleマップ</a><br>
          設定なしの場合は場所関係なしに広告表示されるので金額は高めになります<br />
          <input v-model="ad.coordinateInput" pattern="^-?\d+(\.\d+)?,\s*-?\d+(\.\d+)?$"
                 title="緯度と経度は「35.77111, 139.57111」の形式で入力してください"
                 @input="parseCoordinates(ad)" class="wide-text" />
        </label>

        <div v-if="ad.latitude && ad.longitude" style="font-size:14px;">
          ➤ 緯度: <strong>{{ ad.latitude }}</strong><br />
          ➤ 経度: <strong>{{ ad.longitude }}</strong>
        </div>

        <div v-if="ad.latitude && ad.longitude">
          <label>半径約: <input type="number" v-model="ad.distance" min="0" /> km</label>
        </div>

        <br /><br />

        <label>開始日時:
          <input type="datetime-local" v-model="ad.adStart" required />
        </label>
        <span> ~ </span>
        <label>終了日数:
          <input type="number" v-model="ad.adEndDays" class="input-number" required min="1" /> 日後
        </label>

        <div>
          見積り価格: ¥{{ ad.adYen?.toLocaleString() || 0 }}
        </div>

        <div class="centralize"><button type="submit">仮登録・仮更新</button></div>
        <div v-if="invoiceButton" class="centralize"><button type="button" @click="invoiceAd(index)">請求書発行</button></div>
        <div class="centralize"><button type="button" @click="payForAd(index)">Solanaで支払う</button></div>
        <div class="centralize"><button type="button" @click="deleteAd(index)">削除</button></div>
      </form>
    </div>
    <div class="ads"><span>システムSolanaアドレス口座: </span>{{systemWalletAddress}}</div>
    <div class="ads" v-if="jpycCheckURL">
      <a :href="jpycCheckURL" target="_blank">システムSolanaアドレス口座履歴URL</a>
    </div>
    <div class="ads">
      <div v-if="userWalletAddress">
        <div>
          <span>あなたのウォレットアドレス: </span>{{userWalletAddress}}
        </div>
        <div style="margin-top: 8px;">
          <span>手持ちのJPYC残高: </span>
          <strong v-if="jpycBalance !== null">{{ jpycBalance.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 6 }) }} JPYC</strong>
          <span v-else>取得中...</span>
          <button 
            type="button" 
            @click="loadJpycBalance()" 
            :disabled="loadingBalance"
            style="margin-left: 8px; padding: 4px 8px;"
          >
            {{ loadingBalance ? '取得中...' : '更新' }}
          </button>
        </div>
      </div>
      <div v-else style="margin-top: 8px;">
        <button 
          type="button" 
          @click="createWallet()"
          style="padding: 8px 16px;"
        >
          Solanaウォレットを作成
        </button>
      </div>
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
.input-number {
  width: 40px;
}
/*.centralize div {
  margin-top: -10px
}
*/
</style>
