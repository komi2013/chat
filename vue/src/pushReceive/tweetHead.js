
import { htmlToMarkdown, markdownToHtml, removeMark } from '@/my/markdown.js';

export async function tweetHead(pd) {
  const parentID = pd[2]
  const updatedBy = pd[3]
  let head = pd[4]
  let tweetHead = {}
  let displayStatus = 1
  if (head.messageTxt.includes('＠＠' + localStorage.getItem('nickname') + '・＠＠')) {
    displayStatus = 2
  }
  tweetHead.displayStatus = displayStatus
  tweetHead.title = getSubstring(removeMark(head.messageTxt), 0, 14)
  tweetHead.parentID = head.parentID
  tweetHead.createdAt = head.createdAt

  await upsertIDB(tweetHead, 'tweetHead', 'parentID', tweetHead.parentID)

}
