<script setup>
import { ref, onMounted, onBeforeMount } from "vue";
import Drawer from '@/components/Drawer.vue'
import SelectAlias from '@/components/SelectAlias.vue'

const channel = ref(null);
const aliases = ref([]);
const storeNames = ref([]);
const selectedStore = ref("");
const aliasNames = ref([]);
const message = ref("");

// onBeforeMount(async () => {
//   // await loadStores;
//   channel.value = await getIDB('channel', localStorage.getItem('channelID'));
//   aliases.value = await getIDBs('alias', 'channelIDIndex', localStorage.getItem('channelID'), 10000);
// });

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
};

onMounted(loadStores);

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
  </div>
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

ul {
  list-style-type: none;
  padding: 0;
}

li {
  padding: 5px;
  border-bottom: 1px solid #ccc;
}
</style>
