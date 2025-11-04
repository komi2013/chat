<script setup>
import { ref, onMounted } from 'vue'

import Drawer from '@/components/Drawer.vue'
import Advertisement from '@/components/Advertisement.vue';

import { pushReceive } from '@/pushReceive/pushReceive.js';

const weekdays = ['', '日', '月', '火', '水', '木', '金', '土']

const ads = ref([])
const adPrices = ref([])
const previewBanner = ref('')
const previewSquare = ref('')

document.title = '広告設定'

async function findAds() {
  const fd = new FormData()
  fd.append('csrf', localStorage.getItem('csrf'))
  const res = await sendRequest('/AdGet/', fd)
  if (!res.csrf) errorMessage.value = res
  res.csrf && localStorage.setItem('csrf', res.csrf)
  res.pushContents.forEach(content => {
    pushReceive(content)
  })
  ads.value = ((res.ads && res.ads.length) ? res.ads : [undefined])
    .map(apiAd => convertApiAdToUiAd(apiAd))
  if (ads.value.length > 0) {
    const firstAd = ads.value[0]
    if (firstAd.pathBanner) previewBanner.value = firstAd.pathBanner
    if (firstAd.pathSquare) previewSquare.value = firstAd.pathSquare
  }
}

function convertApiAdToUiAd(apiAd) {
  apiAd = apiAd || {}
  const { day: adStartDay, hour: adStartHour } = splitDayHourFromString(apiAd.adStart);
  const { day: adEndDay, hour: adEndHour } = splitDayHourFromString(apiAd.adEnd);

  const coordinateInput = apiAd.latitude ? `${apiAd.latitude}, ${apiAd.longitude}` : ''
  return {
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
    userID: apiAd.userID || ''
  }
}

function splitDayHourFromString(input) {
  const str = String(input).padStart(3, '0');

  const day = parseInt(str[0], 10) || 0;
  const hour = parseInt(str.slice(1), 10) || 0;

  return { day, hour }
}

async function findAdPrices() {

  const ad = ads.value[0]
  const fd = new FormData();

  fd.append('csrf', localStorage.getItem('csrf'));
  fd.append('pathBanner', ad.pathBanner);
  fd.append('pathSquare', ad.pathSquare);
  fd.append('latitude', ad.latitude);
  fd.append('longitude', ad.longitude);
  fd.append('adStart', `${ad.adStartDay}${String(ad.adStartHour).padStart(2, '0')}`);
  fd.append('adEnd', `${ad.adEndDay}${String(ad.adEndHour).padStart(2, '0')}`);
  fd.append('distance', ad.distance);
  const res = await sendRequest('/AdPriceGet/', fd);
  if (!res.csrf) errorMessage.value = res
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });

  let endDay = ad.adEndDay;
  if (endDay < ad.adStartDay) {
    endDay += 7;
  }
  const durationHours = (endDay * 24 + ad.adEndHour) - (ad.adStartDay * 24 + ad.adStartHour);
  if (Array.isArray(res.adPrices) && res.adPrices.length > 0) {
    const sorted = res.adPrices.sort((a, b) => b.adPriceYen - a.adPriceYen);
    // ad.adYen = Math.pow(2 * ad.distance + 1, 2) * sorted[0].adPriceYen * durationHours;
    let basePrice = Math.pow(2 * ad.distance + 1, 2) * sorted[0].adPriceYen * durationHours;
    let bonus = 0;

    if (previewBanner.value) {
      bonus += basePrice * 2;
    }

    if (previewSquare.value) {
      bonus += basePrice * 2;
    }
    ad.adYen = basePrice + bonus;
  }
  adPrices.value = res.adPrices;
}

const fetched = ref(false)
const errorMessage = ref('')
onMounted(async () => {
  await findAds()
  await findAdPrices()
  fetched.value = true
})

function parseCoordinates(ad) {
  const parts = ad.coordinateInput.split(',').map(s => s.trim())
  if (parts.length !== 2) {
    ad.latitude = ''
    ad.longitude = ''
    return
  }

  const lat = Number(parts[0])
  const lng = Number(parts[1])

  if (!isNaN(lat) && !isNaN(lng)) {
    ad.latitude = lat.toFixed(2)
    ad.longitude = lng.toFixed(2)
  } else {
    ad.latitude = ''
    ad.longitude = ''
  }
}

async function submitAd(index) {
  const ad = ads.value[index]
  const fd = new FormData()
  fd.append('csrf', localStorage.getItem('csrf'))
  const bannerBase64 = isBase64Image(previewBanner.value)
    ? previewBanner.value
    : await imageUrlToBase64(previewBanner.value);
  const squareBase64 = isBase64Image(previewSquare.value)
    ? previewSquare.value
    : await imageUrlToBase64(previewSquare.value);
  fd.append('adID', ad.adID)
  fd.append('previewBanner', bannerBase64)
  fd.append('previewSquare', squareBase64)
  fd.append('adText', ad.adText)
  fd.append('adLink', ad.adLink)
  fd.append('latitude', ad.latitude)
  fd.append('longitude', ad.longitude)
  fd.append('adStart', `${ad.adStartDay}${String(ad.adStartHour).padStart(2, '0')}`)
  fd.append('adEnd', `${ad.adEndDay}${String(ad.adEndHour).padStart(2, '0')}`)
  fd.append('distance', ad.distance)
  const res = await sendRequest('/AdEdit/', fd)
  if (!res.csrf) errorMessage.value = res
  res.csrf && localStorage.setItem('csrf', res.csrf)
  res.pushContents.forEach(content => {
    pushReceive(content)
  })
}

function isBase64Image(str) {
  return typeof str === 'string' && str.trim() !== '' && str.startsWith('data:image/');
}

function imageUrlToBase64(imageUrl) {
  return new Promise((resolve, reject) => {
    if (typeof imageUrl !== 'string' || imageUrl.trim() === '') {
      return resolve('');
    }

    const img = new Image();
    img.crossOrigin = 'Anonymous'; // CORS対応
    img.onload = () => {
      try {
        const canvas = document.createElement('canvas');
        canvas.width = img.width;
        canvas.height = img.height;
        const ctx = canvas.getContext('2d');
        ctx.drawImage(img, 0, 0);
        const base64 = canvas.toDataURL('image/png');
        resolve(base64);
      } catch (error) {
        reject(error);
      }
    };
    img.onerror = () => resolve('');
    img.src = imageUrl;
  });
}

function handleTrim(event, targetW, targetH, type) {
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

      // 元画像サイズから中央をトリミング
      const sx = Math.max(0, (img.width - targetW) / 2)
      const sy = Math.max(0, (img.height - targetH) / 2)

      const sw = Math.min(img.width, targetW)
      const sh = Math.min(img.height, targetH)

      ctx.drawImage(img, sx, sy, sw, sh, 0, 0, targetW, targetH)

      const base64 = canvas.toDataURL('image/png')

      if (type === 'banner') {
        previewBanner.value = base64
        // ad.value.pathBanner = base64
      } else {
        previewSquare.value = base64
        // ad.value.pathSquare = base64
      }
    }
    img.src = reader.result
  }

  reader.readAsDataURL(file)
}

function removeImage(type) {
  if (type === 'banner') {
    previewBanner.value = '';
  } else if (type === 'square') {
    previewSquare.value = '';
  }
}

</script>

<template>
<Drawer />
<div id="content">
  <br><br>
  <div>
    <div v-if="errorMessage"> 
      <div class="errorMessage">{{errorMessage}}<br>データ取得に失敗しました。</div>
      <a href="/setting/"> データ設定ページ </a><br>
      <a href="/sign/"> サインインページ </a>
    </div>
    <div v-for="(ad, index) in ads" :key="index" class="ads">
      <form @submit.prevent="submitAd(index)">
        <div>
        <!--         <label>バナー画像（300x50）:<br />
                  <input type="file" accept="image/*" @change="e => handleTrim(e, 300, 50, 'banner')" />
                </label>
                <img v-if="previewBanner" :src="previewBanner" style="border: 1px solid #ccc" />
                <button type="button" @click="removeImage('banner')">削除</button>
                <br /><br /> -->

          <label>正方形画像（250x250）:<br />
            <input type="file" accept="image/*" @change="e => handleTrim(e, 250, 250, 'square')" />
          </label>
          <img v-if="previewSquare" :src="previewSquare" style="border: 1px solid #ccc" />
          <button type="button" @click="removeImage('square')">削除</button>
        </div>
        <br />

          <!--       <label>広告文<br />
                  <input v-model="ad.adText" placeholder="20%割引" class="wide-text" />
                </label>
                <br /><br />
           -->
        <label>広告リンク<br />
          <input v-model="ad.adLink" placeholder="https://sample.com/item?af=1" class="wide-text"
           required pattern="https://.*" />
        </label>
        <br /><br />

        <label>経緯度:<br />
          <input v-model="ad.coordinateInput"
                 required pattern="^-?\d+(\.\d+)?,\s*-?\d+(\.\d+)?$"
                 title="緯度と経度は「35.77, 139.57」の形式で入力してください"
                 @input="parseCoordinates(ad)" 
                 placeholder="35.72300346964341, 139.52507136879356" 
                 class="wide-text" />
        </label>

        <div v-if="ad.latitude && ad.longitude" style="margin-top: 5px; font-size: 14px;">
          ➤ 緯度（latitude）: <strong>{{ ad.latitude }}</strong><br />
          ➤ 経度（longitude）: <strong>{{ ad.longitude }}</strong>
        </div>
        <br />

        <label>半径約:
          <input type="number" v-model="ad.distance" width="3" /> km
        </label>

        <br /><br />

        <label>開始 曜日:
          <select v-model="ad.adStartDay" required>
            <option v-for="(name, i) in weekdays" :key="i" :value="i">{{ name }}</option>
          </select>
        </label>

        <label> 時刻:
          <select v-model="ad.adStartHour">
            <option v-for="h in 24" :key="h - 1" :value="h - 1">{{ String(h - 1).padStart(2, '0') }}</option>
          </select>
        </label>

        <br />

        <label>終了 曜日:
          <select v-model="ad.adEndDay" required>
            <option v-for="(name, i) in weekdays" :key="i" :value="i">{{ name }}</option>
          </select>
        </label>

        <label> 時刻:
          <select v-model="ad.adEndHour">
            <option v-for="h in 24" :key="h - 1" :value="h - 1">{{ String(h - 1).padStart(2, '0') }}</option>
          </select>
        </label>
        <div> {{ad.adYen}} </div>
        <button type="button" @click="findAdPrices">確認</button>
        <button type="submit">登録</button>
      </form>
    </div>
    <h2>広告価格リスト</h2>
    <ul>
      <li v-for="(ad, index) in adPrices" :key="index">
        <p><strong>緯度範囲:</strong> {{ ad.latitudeSouth }} ～ {{ ad.latitudeNorth }}</p>
        <p><strong>経度範囲:</strong> {{ ad.longitudeWest }} ～ {{ ad.longitudeEast }}</p>
        <p><strong>広告表示時間:</strong> {{ ad.adStart }} ～ {{ ad.adEnd }} 秒</p>
        <p><strong>価格:</strong> ¥{{ ad.adPriceYen }}</p>
        <hr />
      </li>
    </ul>
  </div>
</div>
<div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>
</template>

<style scoped>
.wide-text {
  width: 100%;
  max-width: 400px;
}

.ads {
  border: 1px solid #ccc;
  padding: 1rem;
  margin-bottom: 1rem;
}
</style>
