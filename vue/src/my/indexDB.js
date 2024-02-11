const INDEX_DB_VERSION = 22;
function openDatabase(table, key) {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('chat', INDEX_DB_VERSION);

    request.onerror = (event) => {
      reject(`Error opening database: ${event.target.error}`);
    };

    request.onsuccess = (event) => {
      const db = event.target.result;
      resolve(db);
    };

    request.onupgradeneeded = (event) => {
      const db = event.target.result;
      const tables = [
        ['channel', 'channelID'],
        ['alias', 'aliasName'],
        ['message', 'messageID']
      ];
      tables.forEach(([tableName, keyPath]) => {
        const objectStore = db.createObjectStore(tableName, { keyPath, autoIncrement: false });
        if (tableName === 'message' && !objectStore.indexNames.contains('channelIDIndex')) {
          objectStore.createIndex('channelIDIndex', 'channelID', { unique: false });
        }
      });
    };
  });
}

function getIDB(table, id) {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('chat',INDEX_DB_VERSION);

    request.onerror = (event) => {
      reject(`Error opening database: ${event.target.error}`);
    };

    request.onsuccess = (event) => {
      const db = event.target.result;
      const transaction = db.transaction([table], 'readonly');
      const objectStore = transaction.objectStore(table);

      const getRequest = objectStore.get(id);

      getRequest.onsuccess = (event) => {
        const data = event.target.result;
        resolve(data);
      };

      getRequest.onerror = (event) => {
        reject(`Error getting data: ${event.target.error}`);
      };
    };
  });
}

async function getIDBs(table, key, id) {
  const db = await openDatabase('chat', INDEX_DB_VERSION);
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([table], 'readonly');
    const objectStore = transaction.objectStore(table);
    const index = objectStore.index(key);

    const range = IDBKeyRange.only(id);
    const request = index.openCursor(range, 'prev');

    const result = [];

    request.onsuccess = (event) => {
      const cursor = event.target.result;
      if (cursor && result.length < 10) {
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

export { openDatabase, getIDB, getIDBs, upsertData, deleteData };
