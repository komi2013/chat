<script setup>
import { ref, onMounted } from "vue";

import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue'
import SelectAlias from '@/components/SelectAlias.vue'

const channel = ref(null);
const aliases = ref([]);
const storeNames = ref([]);
const selectedStore = ref("");
const message = ref("");
document.title = '設定'

const loadStores = async () => {
  try {
    console.log('dd')
    storeNames.value = await getObjectStoreNames()
    console.log(storeNames.value)
  } catch (error) {
    message.value = `データベースエラー: ${error}`;
  }
};

const deleteStoreData = async () => {
  if (!confirm("削除")) return
  if (!selectedStore.value) {
    message.value = "削除するストアを選択してください。";
    return;
  }
  try {
    message.value = await clearObjectStore(selectedStore.value);
    await loadStores();
  } catch (error) {
    message.value = `エラー: ${error}`;
  }
};

const deleteAllStorage = async () => {
  if (!confirm("削除")) return
  const allStores = await getObjectStoreNames();

  for (const store of allStores) {
    await clearObjectStore(store);
  }
  localStorage.clear();
  message.value = 'すべてのIndexedDBストア、LocalStorage、Cookieを削除しました。';
  await loadStores();
  deleteIndexedDB();
};

const logs = ref([])
const limit = 50
let offset = 0
const loadingMore = ref(false)
onMounted(async () => {
  await loadStores()
  // logs.value = await getSortedIDBs('log', 'updatedAtIndex', limit, offset, 'desc');
  logs.value = await fetchLog()
  // console.log('logs', logs.value)
})

async function fetchLog() {
  const result = await getSortedIDBs('log', 'updatedAtIndex', limit, offset, 'desc')
  // offset を進める
  offset += limit
  return result
}

async function fetchLogMore() {
  loadingMore.value = true
  const nextLogs = await fetchLog()
  logs.value = [...logs.value, ...nextLogs] // 既存に追加
  loadingMore.value = false
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


// 既存 import & 変数省略（そのままでOK）
// ===== ▼ 追加：ストアから全データを取得してダウンロードする関数 =====

async function downloadStoreData() {
  if (!selectedStore.value) {
    message.value = "ダウンロードするストアを選択してください。";
    return;
  }

  try {
    const storeName = selectedStore.value;
    const data = await exportStoreData(storeName);

    const filename = `${storeName}_backup_${new Date()
      .toISOString()
      .replace(/[:.]/g, "-")}.json`;

    downloadJSON(filename, data);

    message.value = `ストア「${storeName}」のデータをダウンロードしました ✔️`;
  } catch (error) {
    message.value = `ダウンロードエラー: ${error}`;
  }
}

// ===== ▼ IndexedDB のデータをすべて取得（openDatabase を使用する版）=====
async function exportStoreData(storeName) {
  const db = await openDatabase();

  return new Promise((resolve, reject) => {
    if (!db.objectStoreNames.contains(storeName)) {
      return reject(`"${storeName}" は存在しません。`);
    }

    const transaction = db.transaction([storeName], "readonly");
    const objectStore = transaction.objectStore(storeName);
    const request = objectStore.getAll();

    request.onsuccess = () => {
      resolve(request.result); // ストア内の全レコード
    };
    request.onerror = (event) => {
      reject(`データ取得エラー: ${event.target.error}`);
    };

    transaction.onerror = (event) => {
      reject(`トランザクションエラー: ${event.target.error}`);
    };
  });
}

// ===== ▼ JSON ダウンロード処理 =====
function downloadJSON(filename, data) {
  const blob = new Blob([JSON.stringify(data, null, 2)], {
    type: "application/json",
  });

  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");

  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();

  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

// ===== ▼ JSON をインポートしてストアへ書き込む =====
async function importStoreData(storeName, jsonData) {
  const db = await openDatabase();

  return new Promise((resolve, reject) => {
    if (!db.objectStoreNames.contains(storeName)) {
      return reject(`"${storeName}" は存在しません。`);
    }

    const tx = db.transaction([storeName], "readwrite");
    const store = tx.objectStore(storeName);

    // JSON の配列データを1件ずつ put（追加 or 上書き）
    for (const record of jsonData) {
      store.put(record);
    }

    tx.oncomplete = () => resolve(`"${storeName}" のデータをインポートしました ✔️`);
    tx.onerror = event => reject(`トランザクションエラー: ${event.target.error}`);
  });
}

const importFile = ref(null);

function onJsonFileSelect(event) {
  importFile.value = event.target.files[0];
}

async function startImport() {
  if (!selectedStore.value) {
    message.value = "インポート先のストアを選択してください。";
    return;
  }
  if (!importFile.value) {
    message.value = "JSONファイルを選択してください。";
    return;
  }

  try {
    const text = await importFile.value.text();
    const json = JSON.parse(text);

    const result = await importStoreData(selectedStore.value, json);

    message.value = result;
  } catch (error) {
    message.value = `インポートエラー: ${error}`;
  }
}

</script>

<template>
  <div id="drawer_column"><Drawer /></div>
  <div class="content">
    <h2>IndexedDB ストア削除</h2>
    <label for="storeSelect">ストアを選択:</label>
    <select v-model="selectedStore" id="storeSelect">
      <option value="" disabled>選択してください</option>
      <option v-for="store in storeNames" :key="store" :value="store">
        {{ store }}
      </option>
    </select>
    <br>
    <button @click="deleteStoreData">データ削除</button>
    <button @click="downloadStoreData">データダウンロード</button>
    <br>
    <input type="file" @change="onJsonFileSelect" accept="application/json" />
    <button @click="startImport">インポートする</button>

    <p class="message">{{ message }}</p>
    <h2>現在のストア一覧</h2>
    <ul>
      <li v-for="[storeName] in indexedDBStores" :key="storeName">
        {{ storeName }}
      </li>
    </ul>
    <h2>Cookie・LocalStorage・IndexedDB をすべて削除</h2>
    <button @click="deleteAllStorage">
      💥 Cookie・LocalStorage・IndexedDB をすべて削除
    </button>

    <div class="logs-table">
      <table>
        <thead>
          <tr>
            <th>pushTitle</th>
            <!-- <th>channelID</th> -->
            <th>updatedBy</th>
            <th>updatedAt</th>
            <th>preContents</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(log, i) in logs" :key="i">
            <td>{{ log.pushTitle }}</td>
            <!-- <td>{{ log.channelID }}</td> -->
            <td>{{ log.updatedBy }}</td>
            <td>{{ log.updatedAt }}</td>
            <td>
              <div class="pre-cell">
                <pre contenteditable="true">{{ JSON.stringify(log.preContents, null, 2) }}</pre>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <!-- More ボタン -->
      <div class="load-more">
        <button @click="fetchLogMore" :disabled="loadingMore">
          {{ loadingMore ? 'Loading...' : 'More' }}
        </button>
      </div>
    </div>

  </div>
  <div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>
</template>

<style scoped>
.content {
  font-family: Arial, sans-serif;
  text-align: center;
  max-width: 500px;
  margin: auto;
  padding: 20px;
}

h1, h2 {
  color: #333;
}

select, button {
  margin: 10px;
  padding: 8px;
  font-size: 16px;
}

button {
  background-color: #f44336;
  color: white;
  border: none;
  cursor: pointer;
  padding: 10px;
  border-radius: 5px;
}

button:hover {
  background-color: #d32f2f;
}

.message {
  font-size: 14px;
  color: red;
}


.logs-table {
  overflow-x: auto;
}
table {
  border-collapse: collapse;
  width: 100%;
}
th, td {
  border: 1px solid #ccc;
  /*padding: 6px 10px;*/
  text-align: left;
  font-size: 0.9rem;
}
th {
  background: #f5f5f5;
}

.pre-cell {
  max-height: 60px;
  max-width: 260px;
  white-space: nowrap;
  overflow-x: hidden;
  text-overflow: ellipsis;
}

.pre-cell pre {
  margin: 0px;
}

.pre-cell.expanded {
  white-space: pre-wrap;
  overflow: visible;
  text-overflow: unset;
  max-width: none;
}

</style>
