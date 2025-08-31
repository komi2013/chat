export async function topEdit(pd) {
  const channelID = pd[2]
  const updatedBy = pd[3]
  const topLinks = pd[4]
  // const ID = pd[4][1]
  // const answers = pd[4][2]           // [{ answer: value }, { answer: value } ...]
  // const optionKeys = pd[4][3]
  // let optionKeyStr = ""
  // if (Array.isArray(optionKeys) && optionKeys.length > 0) {
  //   optionKeyStr = optionKeys.join("")  // 配列を文字列化
  // }
  // const answerID = channelID + from + updatedBy + ID + optionKeyStr
  console.log('topLinks', topLinks)
  localStorage.setItem('topLinks' + channelID, JSON.stringify(topLinks))
  

  // const answerData = {
  //   answerID: answerID,
  //   channelID: channelID,
  //   from: from,
  //   askID: ID,
  //   answerBy: updatedBy,
  //   answers: answers
  // }
  // await upsertIDB(answerData, 'answer', 'answerID', answerID)

}
