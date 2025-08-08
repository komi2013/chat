const indexedDBStores = [
  ['advertisement', 'advertisementID'],
  ['alias', 'aliasID'],
  ['bookmark', 'messageID'],
  ['bookPattern', 'bookPatternID'],
  ['calendar', 'calendarID'],
  ['channel', 'channelID'],
  ['chunk', 'chunkID'],
  ['group', 'groupID'],
  ['log', 'logID'],
  ['reception', 'receptionID'],
  ['receptionOrder', 'receptionOrderID'],
  ['shiftStaff', 'shiftStaffID'],
  ['thread', 'messageID'],
  ['threadHead', 'parentID'],
  ['timestamp', 'timestampID'],
  ['timestampCode', 'code'],
  ['ticket', 'ticketID'],
];

const openDatabase = () => {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('chat', 105);

    request.onerror = (event) => {
      reject(`Error opening database: ${event.target.error}`);
    };

    request.onupgradeneeded = (event) => {
      setupDatabaseSchema(event.target.result, event.target.transaction);
    };

    request.onsuccess = (event) => {
      resolve(event.target.result);
    };
  });
};

const setupDatabaseSchema = (db, transaction) => {
  console.log('Database upgrade triggered');

  // もし `schedule` ストアがあれば削除
  if (db.objectStoreNames.contains('schedule')) {
    db.deleteObjectStore('schedule');
  }

  // 各テーブルの作成・インデックス作成
  indexedDBStores.forEach(([tableName, keyPath]) => {
    let objectStore;

    if (db.objectStoreNames.contains(tableName)) {
      objectStore = transaction.objectStore(tableName);
    } else {
      objectStore = db.createObjectStore(tableName, { keyPath, autoIncrement: false });
    }

    // インデックスの作成
    const indexConfigs = {
      alias: [['channelIDIndex', 'channelID']],
      channel: [['displayStatusIndex', 'displayStatus']],
      chunk: [['chunkPassIndex', 'chunkPass']],
      group: [['channelIDIndex', 'channelID']],
      shiftStaff: [['bookPatternIDIndex', 'bookPatternID']],
      thread: [
        ['parentIDIndex', 'parentID'],
        ['channelID_parentID', ['channelID', 'parentID']],
        ],
      threadHead: [
        ['parentIDIndex', 'parentID'],
        ['channelIDIndex', 'channelID']
      ],
      timestampCode: [['channelIDIndex', 'channelID']],
      timestamp: [
        ['channelIDIndex', 'channelID'],
        ['channelID_aliasName', ['channelID', 'aliasName']],
      ],
      ticket: [['statusIndex', 'status']],
    };
    if (indexConfigs[tableName]) {
      indexConfigs[tableName].forEach(([indexName, keyPath]) => {
        if (!objectStore.indexNames.contains(indexName)) {
          objectStore.createIndex(indexName, keyPath, { unique: false });
        }
      });
    }
  });

  transaction.onerror = (event) => {
    console.error('Error in upgrade transaction:', event.target.error);
  };

  transaction.oncomplete = () => {
    console.log('Database upgrade completed');
  };
};

async function getIDB(table, id) {
  try {
    const db = await openDatabase();
    const transaction = db.transaction([table], 'readonly');
    const objectStore = transaction.objectStore(table);
    const getRequest = objectStore.get(id);

    return new Promise((resolve) => {
      getRequest.onsuccess = (event) => {
        resolve(event.target.result);
      };
      getRequest.onerror = (event) => {
        console.error('Request error:', event.target.error, table, id);
        resolve(null);
      };
    });
  } catch (error) {
    console.error('Unexpected error in getIDB:', error, table, id);
    return null;
  }
}

// getIDBs('thread', 'parentIDIndex', props.message_id)
async function getIDBs(table, key, id, limit = 5, offset = 0, sortOrder = 'desc') {
  const db = await openDatabase();
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([table], 'readonly');
    const objectStore = transaction.objectStore(table);

    if (!objectStore.indexNames.contains(key)) {
      console.log(`Index "${key}" not found in table "${table}". Returning empty array.`);
      return resolve([]);
    }
    const index = objectStore.index(key);
    const range = IDBKeyRange.only(id);
    const direction = sortOrder === 'asc' ? 'next' : 'prev';
    const request = index.openCursor(range, direction);
    const result = [];
    let i = 0;
    request.onsuccess = (event) => {
      const cursor = event.target.result;
      if (cursor) {
        if (i >= offset && result.length < limit) {
          result.push(cursor.value);
        }
        i++;
        cursor.continue();
      } else {
        resolve(result);
      }
    };
    request.onerror = (event) => {
      console.error('getIDBs Request:', event.target.error, table, id);
      resolve([]);
    };
  });
}

async function getIDBbyMulti(table, keys, values, limit = 5, offset = 0, sortOrder = 'desc') {
  try {
    const db = await openDatabase();
    return new Promise((resolve) => {
      const transaction = db.transaction([table], 'readonly');
      const objectStore = transaction.objectStore(table);
      const indexName = keys.join('_');

      if (!objectStore.indexNames.contains(indexName)) {
        console.log(`Index "${indexName}" not found in table "${table}". Returning empty array.`);
        return resolve([]);
      }

      const index = objectStore.index(indexName);
      const range = IDBKeyRange.only(values);
      const direction = sortOrder === 'asc' ? 'next' : 'prev';
      const request = index.openCursor(range, direction);
      const result = [];
      let i = 0;

      request.onsuccess = (event) => {
        const cursor = event.target.result;
        if (cursor) {
          if (i >= offset && result.length < limit) {
            result.push(cursor.value);
          }
          i++;
          cursor.continue();
        } else {
          resolve(result);
        }
      };

      request.onerror = (event) => {
        console.error(`getIDBbyMulti Error fetching data from "${table}":`, event.target.error);
        resolve([]);
      };
    });
  } catch (error) {
    console.error("getIDBbyMulti Error:", error);
    return [];
  }
}


async function upsertIDB(data, table, key, objKey) {
  try {
    const db = await openDatabase(table, key);
    const transaction = db.transaction([table], 'readwrite');
    const objectStore = transaction.objectStore(table);

    await new Promise((resolve, reject) => {
      const request = objectStore.put(data);
      request.onsuccess = () => resolve('データを追加または更新しました');
      request.onerror = (event) => reject(new Error(`データ更新エラー: ${event.target.error}`));
    });

    return 'データを追加または更新しました';
  } catch (error) {
    console.error(`upsertIDBエラー: ${error}`, {
      table,
      key,
      objKey,
      data,
    });
    return null;
  }
}

async function deleteIDB(table, key, objKey) {
  try {
    const db = await openDatabase();
    const objectStore = db.transaction([table], 'readwrite').objectStore(table);
    const deleteRequest = objectStore.delete(objKey);

    return new Promise((resolve) => {
      deleteRequest.onsuccess = () => {
        resolve('データを削除しました');
      };
      deleteRequest.onerror = (event) => {
        console.error('データ削除エラー:', event.target.error, table, objKey);
        resolve(null);
      };
    });
  } catch (error) {
    console.error('deleteIDBの予期しないエラー:', error, table, objKey);
    return null;
  }
}

async function getAllIDBs(table) {
  const db = await openDatabase();
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([table], 'readonly');
    const objectStore = transaction.objectStore(table);
    const request = objectStore.openCursor();
    const result = [];
    request.onsuccess = (event) => {
      const cursor = event.target.result;
      if (cursor) {
        result.push(cursor.value);
        cursor.continue();
      } else {
        resolve(result);
      }
    };
    request.onerror = (event) => {
      // reject(`Error fetching data: ${event.target.error}`);
      console.error('getAllIDBs:', event.target.error, table);
      resolve([]);
    };
  });
}

async function updIDBone(table, key, columnName, columnValue) {
  const db = await openDatabase(table, key);
  const transaction = db.transaction([table], 'readwrite');
  const objectStore = transaction.objectStore(table);

  // 指定されたキーに対応するデータを取得
  const getRequest = objectStore.get(key);

  getRequest.onsuccess = () => {
    const data = getRequest.result;
    if (data) {
      // 特定の列の値を更新
      data[columnName] = columnValue;

      // 更新したデータを保存
      const putRequest = objectStore.put(data);
      // putRequest.onsuccess = () => {
      //   console.log(`${columnName} updated successfully`);
      // };
      putRequest.onerror = (event) => {
        console.error(`Error updating ${columnName}: ${event.target.error}`);
      };
    } else {
      console.error(`No data found for key: ${key}`);
    }
  };

  getRequest.onerror = (event) => {
    console.error(`Error getting data: ${event.target.error}`);
  };
}

const getObjectStoreNames = async () => {
  const db = await openDatabase();
  return Array.from(db.objectStoreNames); // ストア一覧を取得
};

const clearObjectStore = async (storeName) => {
  const db = await openDatabase();
  return new Promise((resolve, reject) => {
    if (!db.objectStoreNames.contains(storeName)) {
      return reject(`"${storeName}" は存在しません。`);
    }
    const transaction = db.transaction([storeName], 'readwrite');
    const objectStore = transaction.objectStore(storeName);
    const request = objectStore.clear(); // データ削除
    request.onsuccess = () => resolve(`"${storeName}" のデータを削除しました！`);
    request.onerror = (event) => reject(`エラー: ${event.target.error}`);
  });
};

const deleteIndexedDB = () => {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('chat', 101);
    request.onerror = (event) => {
      reject(`Error resetting database: ${event.target.error}`);
    };
    request.onupgradeneeded = (event) => {
      const db = event.target.result;
      const transaction = event.target.transaction;
      console.log('Resetting IndexedDB: Dropping and recreating all stores');
      Array.from(db.objectStoreNames).forEach((storeName) => {
        db.deleteObjectStore(storeName);
      });
    };
    request.onsuccess = (event) => {
      const db = event.target.result;
      db.close();
      resolve('Database has been reset successfully.');
    };
  });
};

async function queryIndexByValue({
    table,
    indexKey,         // メイン検索対象のインデックス
    value,            // 開始位置
    limit = 10000000,
    offset = 0,       // ← 追加
    direction = 'desc', // 'asc' or 'desc'
    filter = null     // 例: ['parentIDIndex', '0Lf']
  }) {
  const db = await openDatabase()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([table], 'readonly')
    const store = transaction.objectStore(table)
    if (filter && !store.indexNames.contains(filter[0])) {
      console.warn(`Filter index "${filter[0]}" does not exist in "${table}".`)
      return resolve([]);
    }
    if (!store.indexNames.contains(indexKey)) {
      console.warn(`Index "${indexKey}" does not exist in "${table}".`)
      return resolve([]);
    }
    const results = []
    let cursorRequest
    let skipped = 0

    const filterIndex = store.index(filter[0])
    const filterRange = IDBKeyRange.only(filter[1])
    cursorRequest = filterIndex.openCursor(filterRange, direction === 'asc' ? 'next' : 'prev')

    cursorRequest.onsuccess = (event) => {
      const cursor = event.target.result
      if (!cursor) return resolve(results)
      const record = cursor.value;
      if (filter) {
        const compareField = indexKey.replace('Index', '')
        const compareVal = record[compareField]
        if (direction === 'asc' && compareVal < value) return cursor.continue()
        if (direction === 'desc' && compareVal > value) return cursor.continue()
      }
      if (skipped < offset) {
        skipped++
        return cursor.continue()
      }
      results.push(record)
      if (results.length >= limit) {
        return resolve(results)
      }
      cursor.continue()
    };
    cursorRequest.onerror = (event) => {
      console.error('queryIndexByValue Error:', event.target.error)
      reject(event.target.error)
    }
  })
}

// async function countIDBs(table, key, id) {
//   const db = await openDatabase();

//   return new Promise((resolve, reject) => {
//     const transaction = db.transaction([table], 'readonly');
//     const objectStore = transaction.objectStore(table);

//     if (!objectStore.indexNames.contains(key)) {
//       console.log(`Index "${key}" not found in table "${table}". Returning count 0.`);
//       return resolve(0);
//     }

//     const index = objectStore.index(key);
//     const range = IDBKeyRange.only(id);

//     const countRequest = index.count(range);

//     countRequest.onsuccess = () => {
//       resolve(countRequest.result);
//     };

//     countRequest.onerror = (event) => {
//       console.error('countIDBs Request Error:', event.target.error, table, id);
//       resolve(0);
//     };
//   });
// }


// const reception = 

//     {
//       "receptionID": "123456",
//       "adminNames": [
//         "mik2"
//       ],
//       "joinNames": [
//         "ivan1",
//         "mik2"
//       ],
//       "receptionTitle": "サロンの公開用予約リンク",
//       "facilities": [
//         {
//           "facilityCount": 4,
//           "facilityName": "perm"
//         }
//       ],
//       "shifts": [
//         {
//           "aliasNames": [
//             "mik2",
//             "ivan1"
//           ],
//           "shiftStart": "2025-04-15T10:00",
//           "shiftEnd": "2025-04-15T23:00",
//           "open": 1,
//           "role": "stylist",
//           "fix": true
//         },
//         {
//           "aliasNames": [
//             "ivan1",
//             "mik2"
//           ],
//           "shiftStart": "2025-04-17T15:00",
//           "shiftEnd": "2025-04-17T20:00",
//           "open": 1,
//           "role": "helper"
//         }
//       ],
//       "skills": [
//         "cut",
//         "perm"
//       ],
//       "staffSkills": [
//         {
//           "aliasName": "ivan1",
//           "skills": [
//             "perm",
//             "cut"
//           ]
//         },
//         {
//           "aliasName": "mik2",
//           "skills": [
//             "perm"
//           ]
//         }
//       ],
//       "workStaffNeed": true,
//       "workStaffs": [
//         {
//           "aliasName": "mik3",
//           "workStart": "2025-04-05T10:00",
//           "workEnd": "2025-04-05T23:00",
//           "seq": 2
//         },
//         {
//           "aliasName": "ivan1",
//           "workStart": "2025-04-05T15:00",
//           "workEnd": "2025-04-05T20:00",
//           "seq": 2
//         },
//         {
//           "aliasName": "mik2",
//           "workStart": "2025-04-05T15:00",
//           "workEnd": "2025-04-05T20:00",
//           "seq": 3
//         }
//       ]
//     }
// upsertIDB(reception, 'reception', 'receptionID', reception.receptionID);
