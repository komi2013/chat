const openDatabase = () => {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('chat', 75);
    request.onerror = (event) => {
      reject(`Error opening database: ${event.target.error}`);
    };
    request.onupgradeneeded = (event) => {
      const db = event.target.result;

      // オブジェクトストアを削除
      if (db.objectStoreNames.contains('stampCode')) {
        db.deleteObjectStore('stampCode');
      }

      const transaction = event.target.transaction;
      transaction.onerror = (event) => {
        console.error('Error in upgrade transaction:', event.target.error);
      };
      transaction.oncomplete = (event) => {
        console.log('Upgrade transaction completed');
      };
      const tables = [
        ['channel', 'channelID'],
        ['thread', 'messageID'],
        ['threadHead', 'parentID'],
        ['bookmark','messageID'],
        ['timestampCode','code'],
        ['timestamp','timestampID'],
        ['ticket','ticketID'],        
      ];
      tables.forEach(([tableName, keyPath]) => {
        let objectStore;
        if (db.objectStoreNames.contains(tableName)) {
          objectStore = transaction.objectStore(tableName);
        } else {
          objectStore = db.createObjectStore(tableName, { keyPath, autoIncrement: false });
        }
        if (tableName === 'channel' && !objectStore.indexNames.contains('displayStatusIndex')) {
          objectStore.createIndex('displayStatusIndex', 'displayStatus', { unique: false });
        }
        if (tableName === 'message' && !objectStore.indexNames.contains('channelIDIndex')) {
          objectStore.createIndex('channelIDIndex', 'channelID', { unique: false });
        }
        if (tableName === 'thread' && !objectStore.indexNames.contains('parentIDIndex')) {
          objectStore.createIndex('parentIDIndex', 'parentID', { unique: false });
        }
        if (tableName === 'timestampCode' && !objectStore.indexNames.contains('channelIDIndex')) {
          objectStore.createIndex('channelIDIndex', 'channelID', { unique: false });
        }
        if (tableName === 'timestamp' && !objectStore.indexNames.contains('channelIDIndex')) {
          objectStore.createIndex('channelIDIndex', 'channelID', { unique: false });
        }
        if (tableName === 'timestamp' && !objectStore.indexNames.contains('channelID_aliasName')) {
          objectStore.createIndex('channelID_aliasName', ['channelID', 'aliasName'], { unique: false });
        }
        if (tableName === 'ticket' && !objectStore.indexNames.contains('ticketIDIndex')) {
          objectStore.createIndex('ticketIDIndex', 'ticketID', { unique: false });
        }
      });
    };

    request.onsuccess = (event) => {
      const db = event.target.result;
      resolve(db);
    };
  });
};

async function getIDB(table, id) {
  try {
    const db = await openDatabase();
    const transaction = db.transaction([table], 'readonly');
    const objectStore = transaction.objectStore(table);
    const getRequest = objectStore.get(id);

    return new Promise((resolve, reject) => {
      getRequest.onsuccess = (event) => {
        const data = event.target.result;
        if (data) {
          resolve(data); // データが存在する場合は解決
        } else {
          reject(`${table} by ${id} not found`); // データが存在しない場合は拒否
        }
      };

      getRequest.onerror = (event) => {
        reject(`Error getting data: ${event.target.error}`);
      };
    });
  } catch (error) {
    return Promise.reject(error);
  }
}
// getIDBs('thread', 'parentIDIndex', props.message_id)
async function getIDBs(table, key, id, limit = 5, offset = 0, sortOrder = 'desc') {
  const db = await openDatabase();
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([table], 'readonly');
    const objectStore = transaction.objectStore(table);
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
      reject(`Error fetching data: ${event.target.error}`);
    };
  });
}

async function getByMulti(table, keys, values, limit = 5, offset = 0, sortOrder = 'desc') {
  const db = await openDatabase();
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([table], 'readonly');
    const objectStore = transaction.objectStore(table);
    const indexName = keys.join('_');
    const index = objectStore.index(indexName);
    try {
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
        reject(`Error fetching data: ${event.target.error}`);
      };
    } catch (error) {
      console.error("Values:", values);
      reject(error);
    }
  });
}

async function upsertIDB(data, table, key, objKey) {
  const db = await openDatabase(table, key);
  const objectStore = db.transaction([table], 'readwrite').objectStore(table);
  return new Promise((resolve, reject) => {
    const existingDataRequest = objectStore.get(objKey);
    existingDataRequest.onsuccess = async () => {
      const existingData = existingDataRequest.result;
      if (existingData) {
        const putRequest = objectStore.put(data);
        putRequest.onsuccess = () => {
          resolve('Data updated successfully');
        };
        putRequest.onerror = (event) => {
          reject(`Error updating data: ${event.target.error}`);
        };
      } else {
        const addRequest = objectStore.add(data);
        addRequest.onsuccess = () => {
          resolve('Data inserted successfully');
        };
        addRequest.onerror = (event) => {
          reject(`Error inserting data: ${event.target.error}`);
        };
      }
    };
    existingDataRequest.onerror = (event) => {
      reject(`Error checking existing data: ${event.target.error}`);
    };
  });
}

async function deleteData(table, key, objKey) {
  const db = await openDatabase(table, key);
  const objectStore = db.transaction([table], 'readwrite').objectStore(table);
  return new Promise((resolve, reject) => {
    const deleteRequest = objectStore.delete(objKey);
    deleteRequest.onsuccess = () => {
      resolve('Data deleted successfully');
    };
    deleteRequest.onerror = (event) => {
      reject(`Error deleting data: ${event.target.error}`);
    };
  });
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
      reject(`Error fetching data: ${event.target.error}`);
    };
  });
}

async function updOne(table, key, columnName, columnValue) {
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
      putRequest.onsuccess = () => {
        console.log(`${columnName} updated successfully`);
      };
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

export { openDatabase, getIDB, getIDBs, upsertIDB, deleteData, getAllIDBs, updOne, getByMulti };
