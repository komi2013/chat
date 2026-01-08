export async function advertisement(pd) {
  // pd index mapping:
  // pd[0] = "advertisement"
  // pd[1] = adStart (ISO string)
  // pd[2] = adEnd   (ISO string)
  // pd[3] = adLink
  // pd[4] = pathSquare

  if (!Array.isArray(pd) || pd.length < 3) return
  if (pd[0] !== 'advertisement') return

  const adStart = new Date(pd[1])
  const adEnd   = new Date(pd[2])

  if (isNaN(adStart.getTime()) || isNaN(adEnd.getTime())) return

  // 同一広告を上書きできる deterministic key
  const advertisementID =
    adStart.toISOString() + '_' +
    adEnd.toISOString()

  const adData = {
    advertisementID,
    adStart,
    adEnd,
    adLink: pd[3] || '',
    pathSquare: pd[4] || '',
    updatedAt: new Date()
  }

  await upsertIDB(
    adData,
    'advertisement',
    'advertisementID',
    advertisementID
  )
}
