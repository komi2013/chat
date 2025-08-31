export async function answer(pd) {
  const channelID = pd[2]
  const updatedBy = pd[3]               // 回答者
  const from = pd[4][0]  // 1 = entryForm, 2 = reception
  const ID = pd[4][1]
  const answers = pd[4][2]           // [{ answer: value }, { answer: value } ...]
  const optionKeys = pd[4][3]
  let optionKeyStr = ""
  if (Array.isArray(optionKeys) && optionKeys.length > 0) {
    optionKeyStr = optionKeys.join("")  // 配列を文字列化
  }
  const answerID = channelID + from + updatedBy + ID + optionKeyStr

  console.log('answerData', answers)
  // const preAnswerID = channelID + from + updatedBy + ID
  // answerBy を付与した新しい配列を作成
  // const answersWithUser = Array.isArray(answerData)
  //   ? answerData.map(a => ({ ...a, answerBy: updatedBy }))
  //   : [{ ...answerData, answerBy: updatedBy }]

  // const entryForm = await getIDB('entryForm', entryFormID)
  // if (!Array.isArray(entryForm.answers)) {
  //   entryForm.answers = []
  // }

  // entryForm.answers.push(...answersWithUser)
  // for (const a of answers) {
    // const answerID = preAnswerID + a.sequence
  const answerData = {
    answerID: answerID,
    channelID: channelID,
    from: from,
    askID: ID,
    answerBy: updatedBy,
    answers: answers
  }
  await upsertIDB(answerData, 'answer', 'answerID', answerID)
  // }
}

// [
//   "A",
//   [
//     {
//       "answer": "A",
//       "sequence": 1
//     },
//     {
//       "answer": [
//         "C",
//         "D"
//       ],
//       "sequence": 2
//     },
//     {
//       "answer": "oooo",
//       "sequence": 3
//     }
//   ]
// ]