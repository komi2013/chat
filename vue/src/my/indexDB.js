function openDatabase(table, key) {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('chat', 21);

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
    const request = indexedDB.open('chat',21);

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

// async function getIDBs(table, key, channelID) {
//   const db = await openDatabase('message', 'messageID');
//   return new Promise((resolve, reject) => {
//     const transaction = db.transaction(['message'], 'readonly');
//     const objectStore = transaction.objectStore('message');
//     const channelIDIndex = objectStore.index('messageID');

//     const getRequest = channelIDIndex.getAll(IDBKeyRange.only(channelID));

//     getRequest.onsuccess = (event) => {
//       const messages = event.target.result;
//       resolve(messages);
//     };

//     getRequest.onerror = (event) => {
//       reject(`Error getting messages for channel ${channelID}: ${event.target.error}`);
//     };
//   });
// }


// async function getIDBs(table, key, id) {
//   return new Promise((resolve, reject) => {
//     const request = indexedDB.open('chat',21);
//     request.onerror = (event) => {
//       reject(`Error opening database: ${event.target.error}`);
//     };
//     request.onsuccess = (event) => {
//       const db = event.target.result;
//       const transaction = db.transaction([table], 'readonly');
//       const objectStore = transaction.objectStore(table);
//       const IDIndex = objectStore.index(key);
//       const getRequest = IDIndex.getAll(IDBKeyRange.only(id));
//       getRequest.onsuccess = (event) => {
//         const data = event.target.result;
//         resolve(data);
//       };
//       getRequest.onerror = (event) => {
//         console.error(`Error getting messages for channel ${channelID}: ${event.target.error}`);
//         reject(`Error getting messages for channel ${channelID}: ${event.target.error}`);
//       };
//     };
//   });
// }

async function getIDBs(table, key, id) {
  const db = await openDatabase('chat', 21);
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([table], 'readonly');
    const objectStore = transaction.objectStore(table);
    const index = objectStore.index(key);

    const range = IDBKeyRange.only(id);
    const request = index.openCursor(range, 'prev');

    const result = [];

    request.onsuccess = (event) => {
      const cursor = event.target.result;
      if (cursor && result.length < 5) {
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

export { openDatabase, getIDB, getIDBs, upsertData };
