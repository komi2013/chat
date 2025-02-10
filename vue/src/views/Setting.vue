<!-- App.vue -->
<script setup>
import { ref, onMounted } from "vue";
// import { getObjectStoreNames, clearObjectStore, indexedDBStores } from "./indexeddb.js";

const storeNames = ref([]); // IndexedDBのストア一覧
const selectedStore = ref(""); // 選択されたストア
const message = ref(""); // メッセージ表示

// ストア一覧を取得
const loadStores = async () => {
  try {
    storeNames.value = await getObjectStoreNames();
  } catch (error) {
    message.value = `データベースエラー: ${error}`;
  }
};

// データ削除処理
const deleteStoreData = async () => {
  if (!selectedStore.value) {
    message.value = "削除するストアを選択してください。";
    return;
  }
  try {
    message.value = await clearObjectStore(selectedStore.value);
    await loadStores(); // 削除後にリスト更新
  } catch (error) {
    message.value = `エラー: ${error}`;
  }
};

// 初回ロード時にストアリストを取得
onMounted(loadStores);
</script>

<template>
  <div class="container">
    <h1>IndexedDB ストア削除</h1>

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
  </div>
</template>

<style scoped>
.container {
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

ul {
  list-style-type: none;
  padding: 0;
}

li {
  padding: 5px;
  border-bottom: 1px solid #ccc;
}
</style>
