import { ref } from 'vue'

import { pushReceive } from '@/pushReceive/pushReceive.js'

const errorMessage = ref('')
let ads = []

export async function loadAdvertisements() {
  let allAds = await getAllIDBs('advertisement')
  allAds = await removeExpiredAds(allAds)
  window.advertisements = allAds
  console.log('window.advertisements', window.advertisements)
  const lastUpdate = localStorage.getItem('adPublicGotAt')
  const now = Date.now()
  const interval = 10 * 60 * 1000 // 10分をミリ秒で定義
  if (!lastUpdate || (now - parseInt(lastUpdate, 10)) > interval) {
    await AdPublicGetAndMerge()
    localStorage.setItem('adPublicGotAt', now.toString())
  }
}

async function removeExpiredAds(adList) {
  const now = new Date()
  const validAds = []
  for (const ad of adList) {
    const endTime = new Date(ad.adEnd)
    if (isNaN(endTime.getTime()) || endTime > now) {
      validAds.push(ad)
    } else {
      await deleteIDB('advertisement', 'advertisementID', ad.advertisementID)
    }
  }
  return validAds
}

async function AdPublicGetAndMerge() {
  if (!localStorage.getItem('csrf')) { return }
  const fd = new FormData()
  fd.append('csrf', localStorage.getItem('csrf'))
  const res = await sendRequest('/AdPublicGet/', fd)
  if (!res.csrf) {
    console.error('AdPublicGet エラー:', res)
    return
  }
  localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  if (res.error) { console.error('AdPublicGet エラー:', res.error) }
  const serverAds = (res.ads || []).map(ad => normalizeAd(ad))
  const existing = await getAllIDBs('advertisement')
  const merged = mergeAds(existing, serverAds)
  const filtered = await removeExpiredAds(merged)
  window.advertisements = filtered
  for (const ad of filtered) {
    await upsertIDB(ad, 'advertisement', 'advertisementID', ad.advertisementID)
  }
  // console.log(`✅ 広告を ${filtered.length} 件保持しました。`)
}

function normalizeAd(ad) {
  if (!ad.adStart || !ad.adEnd || !ad.advertisementID) {
    console.error("normalizeAd error: adStart または adEnd, ID が不足しています", {
      adStart: ad.adStart,
      adEnd: ad.adEnd
    });
    return ad
  }
  const normalized = { ...ad }
  normalized.adStart = timeFormat('YYYY-MM-DDThh:mm:ss', ad.adStart)
  normalized.adEnd = timeFormat('YYYY-MM-DDThh:mm:ss', ad.adEnd)
  return normalized
}

function mergeAds(existing, newAds) {
  const map = new Map()
  for (const ad of existing) {
    const key = ad.advertisementID || ad._id || ad.adLink
    map.set(key, ad)
  }
  for (const ad of newAds) {
    const key = ad.advertisementID || ad._id || ad.adLink
    map.set(key, ad)
  }
  return Array.from(map.values())
}

