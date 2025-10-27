// tweetHeadPush.js
// import { getIDB, upsertIDB } from '@/my/idb.js'
// import { timeFormat } from '@/my/timeFormat.js'

// pushから受け取るパラメータ構造
// pd = [pushID, pushTitle, parentID, updatedBy, tweetHeadData]
export async function tweetHead(pd) {
  const parentID = pd[2]
  const updatedBy = pd[3]
  let head = pd[4]

  // 既存のtweetHeadを取得
  // const pre = await getIDB('tweetHead', newTweetHead.parentID)
  // let tweetHead

  // if (pre) {
  //   // 既存データを複製して編集
  //   tweetHead = JSON.parse(JSON.stringify(pre))
  // } else {
  //   // 新規データとして受信したものを使用
  //   tweetHead = JSON.parse(JSON.stringify(newTweetHead))
  // }

  // console.log('📨 受信したtweetHead', tweetHead)

  // === タイトル・表示状態の調整 ===
  // 例: 親IDが「@ユーザー名」形式ならタイトルを動的生成
  // if (parentID) {
  //   if (parentID.startsWith('@')) {
  //     const nameTail = parentID.slice(1)
  //     if (tweetHead.nickname !== updatedBy) {
  //       tweetHead.title = nameTail
  //     }
  //   } else if (parentID.includes('@')) {
  //     const [namePre, nameTail] = parentID.split('@')
  //     if (tweetHead.nickname === namePre) {
  //       tweetHead.title = nameTail
  //     } else if (tweetHead.nickname === nameTail) {
  //       tweetHead.title = namePre
  //     }
  //   }
  // }

  // 投稿者が自分自身の場合 → 未読扱いにするなど
  // const myname = localStorage.getItem('myname')
  let tweetHead
  let displayStatus = 1
  if (head.messageTxt.includes('＠＠' + localStorage.getItem('nickname') + '・＠＠')) {
    displayStatus = 2
  }
  tweetHead.displayStatus = displayStatus
  tweetHead.title = getSubstring(removeMark(head.messageTxt), 0, 30)
  tweetHead.parentID = head.parentID
  tweetHead.createdAt = head.createdAt


  // newThreadフラグは削除
  // delete tweetHead.newThread

  console.log('💾 保存前 tweetHead', tweetHead)

  // === IndexedDBにupsert ===
  await upsertIDB(tweetHead, 'tweetHead', 'parentID', tweetHead.parentID)

  // === ログ保存 ===
  // if (pre) {
  //   const logID = pd[1] + pd[2] + pd[3] + pd[0]
  //   const log = {
  //     logID,
  //     pushID: pd[0],
  //     pushTitle: pd[1],
  //     parentID: pd[2],
  //     updatedBy: pd[3],
  //     updatedAt: timeFormat(),
  //     preContents: pre
  //   }
  //   await upsertIDB(log, 'log', 'logID', logID)
  // }

  console.log('✅ tweetHead upsert 完了', tweetHead.parentID)
}
