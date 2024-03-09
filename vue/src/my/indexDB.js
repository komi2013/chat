const openDatabase = () => {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('chat', 30);

    request.onerror = (event) => {
      reject(`Error opening database: ${event.target.error}`);
    };

    request.onupgradeneeded = (event) => {
      const db = event.target.result;
      const tables = [
        ['channel', 'channelID'],
        ['alias', 'aliasName'],
        ['message', 'messageID'],
        ['thread', 'messageID'],
        ['threadHead', 'parentID']
      ];
      tables.forEach(([tableName, keyPath]) => {
        if (!db.objectStoreNames.contains(tableName)) { // オブジェクトストアが存在しない場合のみ作成する
          const objectStore = db.createObjectStore(tableName, { keyPath, autoIncrement: false });
          if (tableName === 'message' && !objectStore.indexNames.contains('channelIDIndex')) {
            objectStore.createIndex('channelIDIndex', 'channelID', { unique: false });
          }
          if (tableName === 'thread' && !objectStore.indexNames.contains('parentIDIndex')) {
            objectStore.createIndex('parentIDIndex', 'parentID', { unique: false });
          }
        }
      });
      
      // アップグレード完了後にresolveを呼び出す
      resolve(db);
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
          reject('Data not found'); // データが存在しない場合は拒否
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

async function getIDBs(table, key, id, limit = 10) {
  const db = await openDatabase();
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([table], 'readonly');
    const objectStore = transaction.objectStore(table);
    const index = objectStore.index(key);

    const range = IDBKeyRange.only(id);
    const request = index.openCursor(range, 'prev');

    const result = [];

    request.onsuccess = (event) => {
      const cursor = event.target.result;
      if (cursor && result.length < limit) {
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


async function upsertData(data, table, key, objKey) {
  const db = await openDatabase(table, key);
  const objectStore = db.transaction([table], 'readwrite').objectStore(table);
  return new Promise((resolve, reject) => {
    const existingDataRequest = objectStore.get(objKey);
    console.log('existingDataRequest', existingDataRequest);
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

export { openDatabase, getIDB, getIDBs, upsertData, deleteData, getAllIDBs };
