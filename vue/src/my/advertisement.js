import { ref } from 'vue'

const errorMessage = ref('')
let ads = []

export async function loadAdvertisements() {
  // IndexedDB から取得
  let allAds = await getAllIDBs('advertisement')

  // 🔴 期限切れ広告を削除
  allAds = await removeExpiredAds(allAds)

  window.advertisements = allAds

  // 🔄 足りない場合はサーバーから取得してマージ
  if (!window.advertisements || window.advertisements.length < 5) {
    await AdPublicGetAndMerge()
  }
}

/**
 * 🧹 期限切れ広告の削除処理
 */
async function removeExpiredAds(adList) {
  const now = new Date()
  const validAds = []

  for (const ad of adList) {
    try {
      const endTime = new Date(ad.adEnd)
      if (isNaN(endTime.getTime()) || endTime > now) {
        // 有効
        validAds.push(ad)
      } else {
        // 期限切れ → IndexedDB から削除
        console.log(`🗑 削除: ${ad.advertisementID} (${ad.adEnd})`)
        await deleteIDB('advertisement', 'advertisementID', ad.advertisementID)
      }
    } catch (e) {
      console.error('adEnd 解析エラー:', e, ad)
    }
  }

  return validAds
}

async function AdPublicGetAndMerge() {
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

  // 再度期限切れを除外
  const filtered = await removeExpiredAds(merged)

  window.advertisements = filtered

  // 🟢 IndexedDB に upsert
  for (const ad of filtered) {
    await upsertIDB(ad, 'advertisement', 'advertisementID', ad.advertisementID)
  }

  console.log(`✅ 広告を ${filtered.length} 件保持しました。`)
}

/**
 * 🧩 adStart / adEnd を ISO8601形式へ変換
 */
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

/**
 * 🕐 Go形式 (例: 223 = 月曜23時) → 現在週の日時(Date)
 */
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

/**
 * 🧠 既存データとサーバーデータをマージ
 */
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

/**
 * 🧹 IndexedDB からデータ削除
 */
async function deleteIDB(table, key, objKey) {
  try {
    const db = await openDatabase()
    const objectStore = db.transaction([table], 'readwrite').objectStore(table)
    const deleteRequest = objectStore.delete(objKey)

    return new Promise((resolve) => {
      deleteRequest.onsuccess = () => {
        console.log(`🗑 deleteIDB: ${table} / ${objKey}`)
        resolve('データを削除しました')
      }
      deleteRequest.onerror = (event) => {
        console.error('データ削除エラー:', event.target.error, table, objKey)
        resolve(null)
      }
    })
  } catch (error) {
    console.error('deleteIDBの予期しないエラー:', error, table, objKey)
    return null
  }
}
