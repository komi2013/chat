<script setup>
import { ref, onMounted } from "vue";

import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue'
import SelectAlias from '@/components/SelectAlias.vue'

const channel = ref(null);
const aliases = ref([]);
const storeNames = ref([]);
const selectedStore = ref("");
const aliasNames = ref([]);
const message = ref("");
document.title = '広告設定'

const loadStores = async () => {
  try {
    storeNames.value = await getObjectStoreNames();
  } catch (error) {
    message.value = `データベースエラー: ${error}`;
  }
};

const deleteStoreData = async () => {
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
  await loadStores
  // logs.value = await getSortedIDBs('log', 'updatedAtIndex', limit, offset, 'desc');
  logs.value = await fetchLog()
  console.log('logs', logs.value)
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


// const expandedRows = ref({}) // 各行ごとの展開状態
// function toggleRow(index) {
//   expandedRows.value[index] = !expandedRows.value[index]
// }


</script>

<template>
  <Drawer />
  <div class="content">
    <h2>IndexedDB ストア削除</h2>
    <p class="message">{{ message }}</p>
    <label for="storeSelect">削除するストアを選択:</label>
    <select v-model="selectedStore" id="storeSelect">
      <option value="" disabled>選択してください</option>
      <option v-for="store in storeNames" :key="store" :value="store">
        {{ store }}
      </option>
    </select>

    <button @click="deleteStoreData">データ削除</button>
    <p class="message">{{ message }}</p>
    <h2>現在のストア一覧</h2>
    <ul>
      <li v-for="[storeName] in indexedDBStores" :key="storeName">
        {{ storeName }}
      </li>
    </ul>
    <h2>Cookie・LocalStorage・IndexedDB をすべて削除</h2>
    <button @click="deleteAllStorage" style="background-color: crimson; color: white; padding: 10px;">
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
  color: #444;
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
button {
  padding: 3px 8px;
  border: 1px solid #666;
  background: #eee;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8rem;
}

</style>
