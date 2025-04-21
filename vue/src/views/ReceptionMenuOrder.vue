<script setup>
import { ref, reactive, onMounted } from 'vue';

const props = defineProps({
  id: '',
  apiKey: '',
});

async function fetchReception() {
  try {
    const fd = new FormData();
    fd.append('receptionID', props.id);
    const reception = await sendRequest('/ReceptionGet/', fd);
    if (reception?.menus) {
      await fetchOrder(reception);
    }
  } catch (error) {
    console.error('Error fetching reception data:', error);
  }
}

async function fetchReceptionOrder() {
  try {
    const orders = await getAllIDBs('receptionOrder');
    return orders;
  } catch (error) {
    console.error('Error fetching reception orders:', error);
    return [];
  }
}

const groupedOrders = ref({});
async function fetchOrder(reception) {
  const orders = await fetchReceptionOrder();
  orders.forEach(order => {
    reception.menus.forEach(menu => {
      if (String(order.menuID) === String(menu.id)) {
        const orderHistory = {
          menuName: menu.name,
          totalPrice: order.price,
          itemNames: getItemNames(order.itemDetailIDs, reception.itemDetails),
          tableName: order.tableName
        };
        if (!groupedOrders.value[order.tableName]) {
          groupedOrders.value[order.tableName] = [];
        }
        groupedOrders.value[order.tableName].push(orderHistory);
      }
    });
  });
}

function getItemNames(itemDetailIDs, itemDetails) {
  if (!itemDetailIDs || itemDetailIDs.length === 0 || !itemDetails) {
    return [];
  }

  return itemDetailIDs.map(id => {
    const matchingItem = itemDetails.find(item => item.itemId === id);
    return matchingItem ? matchingItem.itemName : `Unknown Item (${id})`;
  });
}

onMounted(() => {
  fetchReception();
});

async function deleteOrders(tableName) {
  try {
    const fd = new FormData();
    fd.append('receptionID', props.id);
    fd.append('apiKey', props.apiKey);
    fd.append('tableName', tableName);
    const reception = await sendRequest('/ReceptionDelete/', fd);
  } catch (error) {
    console.error('Error fetching reception data:', error);
  }
}

</script>

<template>

  <div>
    <h1>注文履歴</h1>
    <div v-for="(orders, tableName) in groupedOrders" :key="tableName" class="order-group">
      <h2>{{ tableName }}</h2>
      <div v-for="(order, index) in orders" :key="index" class="order">
        <p>メニュー名: {{ order.menuName }}</p>
        <p>合計価格: {{ order.totalPrice }}円</p>
        <h3>選択されたアイテム:</h3>
        <ul>
          <li v-for="item in order.itemNames" :key="item">
            {{ item }}
          </li>
        </ul>
      </div>
      <!-- 削除ボタン -->
      <button @click="deleteOrders(tableName)">このテーブルの履歴を削除</button>
    </div>
  </div>
</template>

<style scoped>
.menu-item, .cart-item {
  border: 1px solid #ccc;
  margin-bottom: 15px;
  padding: 10px;
  border-radius: 5px;
}

button {
  margin-top: 10px;
  cursor: pointer;
}
</style>
