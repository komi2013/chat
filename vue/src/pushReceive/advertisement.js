
export async function advertisement(pd) {

  const key = pd[2] + timeFormat(':mm_') + generateRandomCode(5);
  const adData = {
    advertisementID: key,
    adStart: convertTimeCodeToDate(pd[2], 1),                            // 開始時間 未定なので0
    adEnd: convertTimeCodeToDate(pd[3]),                     // 終了時間
    adLink: pd[4] || '',                   // リンクURL
    pathSquare: pd[5] || '',               // square画像パス
    updatedAt: new Date()   // 更新時刻
  };

  upsertIDB(adData, 'advertisement', 'advertisementID', adData.advertisementID);

  function convertTimeCodeToDate(timeCode, past = 0) {
      const rawDayOfWeek = Math.floor(timeCode / 100);  // Go基準の曜日
      const hour = timeCode % 100;

      // 曜日補正: Goの 1=日曜 → JSの 0=日曜
      const targetDayOfWeek = (rawDayOfWeek - 1 + 7) % 7;

      // 現在時刻にpast(時間)を加算
      const now = new Date();
      now.setHours(now.getHours() - past);

      // 今週の対象曜日・時間をDateで作成
      const targetDate = new Date(now);
      const currentDayOfWeek = now.getDay();
      let diff = targetDayOfWeek - currentDayOfWeek;
      targetDate.setDate(targetDate.getDate() + diff);
      targetDate.setHours(hour, 0, 0, 0);
      console.log('targetDate1', targetDate);
      console.log('now', now);
      // 過去だった場合は次週にずらす
      if (targetDate < now) {
          targetDate.setDate(targetDate.getDate() + 7);
      }
      console.log('targetDate2', targetDate);
      return targetDate;
  }
}

// '["H","advertisement",513,225,"https://sample.com","/data/ad/original.png"]'

// subscription: '{"endpoint":"https://fcm.googleapis.com/fcm/send/cumMdnjqD_E:APA91bHrwoLX4bgTmOnvcTynsUoXjA7zblLpa2_ipgjZIjaa0GAL6f0IFPlVEsJPuqe7WTIkT7eQVqtJS6JTZG7SrII1FDyn-q7ymnm86HAcXTxvfMZOVUbDzwBFlblMNzyUox5S0MRD","expirationTime":null,"keys":{"p256dh":"BCp-qkjjCDHOjbT8LAEikyyA2enOKC7LexAiFYgiS23LYnzg9itzgCc9iOoVSafV_6Rzc3km8qKKLwDyQZ8B53U","auth":"QXPZ65HXSkXAmwh3SWNj7Q"}}'

// subscription: '{"endpoint":"https://fcm.googleapis.com/fcm/send/cumMdnjqD_E:APA91bHrwoLX4bgTmOnvcTynsUoXjA7zblLpa2_ipgjZIjaa0GAL6f0IFPlVEsJPuqe7WTIkT7eQVqtJS6JTZG7SrII1FDyn-q7ymnm86HAcXTxvfMZOVUbDzwBFlblMNzyUox5S0MRD","expirationTime":null,"keys":{"p256dh":"BCp-qkjjCDHOjbT8LAEikyyA2enOKC7LexAiFYgiS23LYnzg9itzgCc9iOoVSafV_6Rzc3km8qKKLwDyQZ8B53U","auth":"QXPZ65HXSkXAmwh3SWNj7Q"}}'

// https://youtu.be/TKaXcfIeHT0
// 22:32 40:11 
