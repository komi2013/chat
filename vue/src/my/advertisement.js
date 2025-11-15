import { ref } from 'vue'

import { pushReceive } from '@/pushReceive/pushReceive.js'

const errorMessage = ref('')
let ads = []

export async function loadAdvertisements() {
  let allAds = await getAllIDBs('advertisement')
  allAds = await removeExpiredAds(allAds)
  window.advertisements = allAds
  if (!window.advertisements || window.advertisements.length < 5) {
    await AdPublicGetAndMerge()
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
  if (res.error) {
    errorMessage.value = res.error
    console.error('AdPublicGet エラー:', res.error)
    return
  }
  if (res.csrf) localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
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
  const normalized = { ...ad }
  const now = new Date()
  const sixHoursLater = new Date(now.getTime() + 6 * 60 * 60 * 1000)
  if (typeof ad.adStart === 'number') {
    normalized.adStart = convertTimeCodeToDate(ad.adStart)
  } else if (!ad.adStart) {
    normalized.adStart = now.toISOString()
  } else if (ad.adStart instanceof Date) {
    normalized.adStart = ad.adStart.toISOString()
  } else if (typeof ad.adStart === 'string' && !isNaN(Date.parse(ad.adStart))) {
    normalized.adStart = new Date(ad.adStart).toISOString()
  } else {
    normalized.adStart = now.toISOString()
  }

  if (typeof ad.adEnd === 'number') {
    normalized.adEnd = convertTimeCodeToDate(ad.adEnd)
  } else if (!ad.adEnd) {
    normalized.adEnd = sixHoursLater.toISOString()
  } else if (ad.adEnd instanceof Date) {
    normalized.adEnd = ad.adEnd.toISOString()
  } else if (typeof ad.adEnd === 'string' && !isNaN(Date.parse(ad.adEnd))) {
    normalized.adEnd = new Date(ad.adEnd).toISOString()
  } else {
    normalized.adEnd = sixHoursLater.toISOString()
  }

  if (!normalized.advertisementID) {
    normalized.advertisementID = `${ad.adLink || 'no-link'}-${Date.now()}`
  }

  return normalized
}

function convertTimeCodeToDate(timeCode) {
  const rawDayOfWeek = Math.floor(timeCode / 100)
  const hour = timeCode % 100
  const targetDayOfWeek = (rawDayOfWeek - 1 + 7) % 7
  const now = new Date()
  const targetDate = new Date(now)
  const diff = targetDayOfWeek - now.getDay()
  targetDate.setDate(targetDate.getDate() + diff)
  targetDate.setHours(hour, 0, 0, 0)
  if (targetDate < now) {
    targetDate.setDate(targetDate.getDate() + 7)
  }
  return targetDate.toISOString()
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

