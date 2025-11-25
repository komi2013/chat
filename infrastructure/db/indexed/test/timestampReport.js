
// stamper uri /timestampReport/いけるか/2024-09/

//admin uri timestampReport/いけるか/2024-09/sei1/

// 初期設定
const timestamps = [];
const channelID = "hQKP";
const aliasName = "sei1";

// 現在の年月を取得
const now = new Date();
const currentYear = now.getFullYear();
const currentMonth = now.getMonth(); // 0: January, 1: February, ...

// 当月の日付を生成する関数
function generateDatesForMonth(year, month) {
  const dates = [];
  const firstDay = new Date(year, month, 1);
  const lastDay = new Date(year, month + 1, 0); // 次の月の0日 = 今月の最終日

  for (let day = firstDay; day <= lastDay; day.setDate(day.getDate() + 1)) {
    const weekday = day.getDay();
    if (weekday !== 0 && weekday !== 6) { // 土日以外
      dates.push(new Date(day)); // 日付を追加
    }
  }

  return dates;
}

// 当月の平日を取得
const weekdays = generateDatesForMonth(currentYear, currentMonth);

// 平日データを生成して timestamps に追加
for (let date of weekdays) {
  const timeIn = `${date.toISOString().split('T')[0]}T09:00`;
  const timeOut = `${date.toISOString().split('T')[0]}T18:00`;
  const timestampID = `${timeIn}${aliasName}`;

  const newEntry = {
    timeIn,
    timeOut,
    channelID,
    aliasName,
    stampStatus: 1,
    timestampID
  };

  timestamps.push(newEntry);

  // データベースに登録
  // upsertIDB(newEntry, 'timestamp', 'timestampID', timestampID)
  //   .catch(error => {
  //     console.error(`Failed to upsert: ${timestampID}`, error);
  //   });
}

console.log(timestamps);
