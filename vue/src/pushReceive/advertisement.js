export async function advertisement(pd) {
  // pd index mapping (based on your Go append order):
  // pd[0] = "advertisement"
  // pd[1] = adStart (ISO string)
  // pd[2] = adEnd   (ISO string)
  // pd[3] = adLink
  // pd[4] = pathSquare

  // const adStart = new Date(pd[1]);
  // const adEnd   = new Date(pd[2]);

  const key =
    adStart.toISOString() +
    timeFormat(':mm_') +
    generateRandomCode(5)

  const adData = {
    advertisementID: key,
    adStart: new Date(pd[1]),              // direct Date
    adEnd: new Date(pd[2]),                // direct Date
    adLink: pd[3] || '',
    pathSquare: pd[4] || '',
    updatedAt: new Date()
  };

  upsertIDB(
    adData,
    'advertisement',
    'advertisementID',
    adData.advertisementID
  )
}
