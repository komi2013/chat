const indexedDBStores = [
  ['advertisement', 'advertisementID'],
  ['alias', 'aliasID'],
  ['answer', 'answerID'],
  ['bookmark', 'messageID'],
  ['calendar', 'calendarID'],
  ['channel', 'channelID'],
  ['chunk', 'chunkID'],
  ['entryForm', 'entryFormID'],
  ['group', 'groupID'],
  ['log', 'logID'],
  ['pushDuplication', 'pushDuplicationID'],
  ['receptionOrder', 'receptionOrderID'],
  ['thread', 'messageID'],
  ['threadHead', 'parentID'],
  ['timestamp', 'timestampID'],
  ['timestampCode', 'code'],
  ['ticket', 'ticketID'],
  ['tweetHead', 'parentID'],
];

const openDatabase = () => {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('chat', 120);

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
      answer: [['askIDIndex', 'askID']],
      channel: [['displayStatusIndex', 'displayStatus']],
      chunk: [['chunkPassIndex', 'chunkPass']],
      group: [['channelIDIndex', 'channelID']],
      log: [['updatedAtIndex', 'updatedAt']],
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
    console.log('Unexpected error in getIDB:', error, table, id);
    return null;
  }
}

// getIDBs('thread', 'parentIDIndex', props.message_id)
// async function getIDBs(table, key, id, limit = 5, offset = 0, sortOrder = 'desc') {
//   const db = await openDatabase();
//   return new Promise((resolve, reject) => {
//     const transaction = db.transaction([table], 'readonly');
//     const objectStore = transaction.objectStore(table);

//     if (!objectStore.indexNames.contains(key)) {
//       console.log(`Index "${key}" not found in table "${table}". Returning empty array.`);
//       return resolve([]);
//     }
//     const index = objectStore.index(key);
//     const range = IDBKeyRange.only(id);
//     const direction = sortOrder === 'asc' ? 'next' : 'prev';
//     const request = index.openCursor(range, direction);
//     const result = [];
//     let i = 0;
//     request.onsuccess = (event) => {
//       const cursor = event.target.result;
//       if (cursor) {
//         if (i >= offset && result.length < limit) {
//           result.push(cursor.value);
//         }
//         i++;
//         cursor.continue();
//       } else {
//         resolve(result);
//       }
//     };
//     request.onerror = (event) => {
//       console.error('getIDBs Request:', event.target.error, table, id);
//       resolve([]);
//     };
//   });
// }

function getIDBs(table, index, id, limit = 5, offset = 0, sort = 'desc') {
  return queryIDB({
    table,
    indexKey: index,
    match: id ?? null,
    reverse: sort === 'desc',
    limit,
    offset,
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

async function getSortedIDBs(table, sortKey = 'updatedAt', limit = 10, offset = 0, sortOrder = 'desc') {
  const db = await openDatabase();
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([table], 'readonly');
    const objectStore = transaction.objectStore(table);

    // インデックス存在チェック
    if (!objectStore.indexNames.contains(sortKey)) {
      console.log(`Index "${sortKey}" not found in table "${table}". Returning empty array.`);
      return resolve([]);
    }

    const index = objectStore.index(sortKey);
    const direction = sortOrder === 'asc' ? 'next' : 'prev';

    // 全件対象なので range は null
    const request = index.openCursor(null, direction);

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
      console.error('getSortedIDBs Request:', event.target.error, table);
      resolve([]);
    };
  });
}

async function queryIDB({
  table,
  indexKey,
  match = null,              // IDBKeyRange.only()
  gte = null,                // >=
  lte = null,                // <=
  reverse = false,           // desc=true
  limit = Infinity,
  offset = 0,
  filterIndex = null,        // ['parentIDIndex', '0Lf']
  filterCompare = null       // { field: 'updatedAt', op: '>=', value: 123 }
}) {
  const db = await openDatabase();

  return new Promise(resolve => {
    const tx = db.transaction([table], 'readonly');
    const store = tx.objectStore(table);

    if (!store.indexNames.contains(indexKey)) {
      console.warn(`Index "${indexKey}" missing in "${table}".`);
      return resolve([]);
    }

    // Secondary filter index
    let idx = store.index(indexKey);

    // 1️⃣ Determine range safely
    let range = null;
    try {
      if (match !== null) {
        range = IDBKeyRange.only(match);
      } else if (gte !== null && lte !== null) {
        range = IDBKeyRange.bound(gte, lte);
      } else if (gte !== null) {
        range = IDBKeyRange.lowerBound(gte);
      } else if (lte !== null) {
        range = IDBKeyRange.upperBound(lte);
      }
    } catch (e) {
      console.warn(`queryIDB: Invalid key/range`, e);
      return resolve([]);
    }

    const direction = reverse ? 'prev' : 'next';
    const req = idx.openCursor(range, direction);

    const results = [];
    let skipped = 0;

    req.onsuccess = e => {
      const cursor = e.target.result;
      if (!cursor) return resolve(results);

      const record = cursor.value;

      // 2️⃣ Apply secondary filter via another index
      if (filterIndex) {
        const [fk, fv] = filterIndex;

        if (record[fk.replace('Index', '')] !== fv) {
          return cursor.continue();
        }
      }

      // 3️⃣ Apply custom comparison filters
      if (filterCompare) {
        const { field, op, value } = filterCompare;
        const v = record[field];

        if (op === '>=' && v < value) return cursor.continue();
        if (op === '<=' && v > value) return cursor.continue();
        if (op === '>'  && v <= value) return cursor.continue();
        if (op === '<'  && v >= value) return cursor.continue();
      }

      // 4️⃣ Offset
      if (skipped < offset) {
        skipped++;
        return cursor.continue();
      }

      // 5️⃣ Save
      results.push(record);
      if (results.length >= limit) return resolve(results);

      cursor.continue();
    };

    req.onerror = e => {
      console.error(`queryIDB Error:`, e.target.error);
      resolve([]);
    };
  });
}
