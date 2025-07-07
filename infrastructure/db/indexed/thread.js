  async function generateTestData(baseData, totalRecords, batchSize = 100) {
    const maxParentID = 1000; // parentIDの最大件数

    for (let i = 0; i < totalRecords; i += batchSize) {
      const batch = [];
      for (let j = i; j < i + batchSize && j < totalRecords; j++) {
        // ランダムな parentID を生成（1000件までのランダムな値）
        const randomParentID = `parent${Math.floor(Math.random() * maxParentID)}`;

        // ベースデータの中からランダムに1つ選ぶ
        const base = baseData[Math.floor(Math.random() * baseData.length)];

        // ランダムな messageID を生成
        const randomMessageID = `msg${j}`;

        // 新しいデータオブジェクトを作成
        const newData = {
          ...base, // 元データをコピー
          messageID: randomMessageID,
          parentID: 'hQKPMch',
          createdAt: new Date().toISOString(), // 新しい作成日時
        };

        batch.push(newData); // バッチにデータを追加
      }

      // バッチごとにデータをアップサート
      await insertBatch(batch);
    }

    console.log('すべてのデータが挿入されました');
  }

  // バッチごとにIndexedDBにデータを挿入
  async function insertBatch(batch) {
    const promises = batch.map(data => {
      // 非同期の upsertIDB 呼び出しを処理
      console.log(data);
      return upsertIDB(data, 'thread', 'messageID', data.messageID)
        .catch((error) => {
          console.error('Error inserting data:', data, error);
        });
    });

    // すべての upsertIDB 処理が終わるまで待機
    await Promise.all(promises);
  }

  // 元データの例
  const baseData = [
    {
      "messageID": "hQKP1sV70e9",
      "parentID": "hQKPasd@sei1",
      "messageTxt": "＊p＊＠＠2kaime・＠＠ fafa・＊p＊",
      "aliasName": "sei1",
      "aliasImg": "/me.jpg",
      "createdAt": "2024/07/20 19:08:12",
      "channelID": "hQKP",
      "threadType": "1",
      "aliasNames": [
        "hQKPasd",
        "sei1"
      ],
      "backID": "undefined",
      "emojis": ""
    },
    {
      "messageID": "hQKP1sorfGf",
      "parentID": "hQKPMch",
      "messageTxt": "＊p＊test load ・＊p＊",
      "aliasName": "sei1",
      "aliasImg": "/me.jpg",
      "createdAt": "2024/09/13 06:47:46",
      "channelID": "hQKP",
      "threadType": "",
      "aliasNames": [
        "sei1"
      ],
      "backID": "",
      "emojis": ""
    }
  ];

  // 10万件のテストデータを生成
  generateTestData(baseData, 10000);
