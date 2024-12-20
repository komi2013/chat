<script setup>
import { ref, reactive, onMounted } from 'vue';

// Props と リアクティブデータ
const props = defineProps({
  id: '', // Reception ID
  code: '',
});

const reception = ref(null); // Reception データ
const selectedOptions = reactive({});
async function fetchReception() {
  try {
    const fd = new FormData();
    fd.append('receptionID', props.id);
    fd.append('code', props.code);
    const data = await sendRequest('/ReceptionCheck/', fd);
    if (data) {
      reception.value = data;
      data.menus.forEach(menu => {
        selectedOptions[menu.id] = {
          paidOptions: [],
          freeOptions: menu.freeOptions ? Array(menu.freeOptions.length).fill(null) : [],
          freeMultiOptions: []
        };
      });

      console.log('Reception data:', reception.value);
    }
  } catch (error) {
    console.error('Error fetching reception data:', error);
  }
}

function getItemName(itemId) {
  const item = reception.value.itemDetails.find(item => item.itemId === itemId);
  return item ? item.itemName : '不明なアイテム';
}
const orderHistories = ref([]);
async function order(menu) {
  const fd = new FormData();
  fd.append('receptionID', props.id);
  fd.append('code', props.code);
  fd.append('menuID', menu.id);
  fd.append('freeOptions', JSON.stringify(selectedOptions[menu.id].freeOptions));
  fd.append('paidOptions', JSON.stringify(selectedOptions[menu.id].paidOptions));
  fd.append('freeMultiOptions', JSON.stringify(selectedOptions[menu.id].freeMultiOptions));
  const data = await sendRequest('/ReceptionOrder/', fd);
  const receptionOrder = {
    receptionOrderID: data[1] + '_' + data[2] + '_' + data[5],
    tableName: data[1],
    menuID: data[2],
    itemDetailIDs: data[3],
    price: data[4]
  }
  upsertIDB(receptionOrder, 'receptionOrder', 'receptionOrderID', receptionOrder.receptionOrderID)
    .catch((error) => {
      console.error(error);
    });

  const orderHistory = {
    menuName: getMenuName(data[1], reception.value.menus),
    totalPrice: data[4],
    itemNames: getItemNames(data[3], reception.value.itemDetails)
  };


  orderHistories.value.push(orderHistory);
  selectedOptions[menu.id] = {
    paidOptions: [],
    freeOptions: Array(menu.freeOptions?.length).fill(null),
    freeMultiOptions: []
  };
}

// メニュー名を取得する関数
function getMenuName(menuId, menus) {
  const menu = menus.find(m => m.id === menuId);
  return menu ? menu.menuName : "不明なメニュー";
}

function getItemNames(itemIds, itemDetails) {
  console.log(itemIds);
  if (Array.isArray(itemIds) && itemIds.length > 0) {
    return itemIds.map(itemId => {
      const item = itemDetails.find(detail => detail.itemId === itemId);
      return item ? item.itemName : `不明なアイテム (${itemId})`;
    });
  }
  return ['データがありません'];
}

onMounted(() => {
  fetchReception();
});
</script>

<template>
  <div>
    <h1>メニュー一覧</h1>

    <!-- メニューリスト -->
    <div v-if="reception && reception.menus.length">
      <div v-for="menu in reception.menus" :key="menu.id" class="menu-item">
        <h2>{{ menu.menuName }}</h2>
        <p>価格: {{ menu.price }}円</p>

        <h3>アイテム:</h3>
        <ul>
          <li v-for="itemId in menu.items" :key="itemId">
            {{ getItemName(itemId) }}
          </li>
        </ul>

        <!-- 有料オプション -->
        <h3>有料オプション:</h3>
        <ul v-if="menu.paidOptions?.length">
          <li v-for="(option, index) in menu.paidOptions" :key="index">
            <label>
              <input
                type="checkbox"
                :value="option[0]"
                v-model="selectedOptions[menu.id].paidOptions"
              />
              {{ getItemName(option[0]) }}: {{ option[1] }}円
            </label>
          </li>
        </ul>

        <!-- 無料オプション (複数パターンで単一選択) -->
        <h3>無料オプション:</h3>
        <ul v-if="menu.freeOptions?.length">
          <li v-for="(optionArray, patternIndex) in menu.freeOptions" :key="patternIndex">
            <h4>パターン {{ patternIndex + 1 }}</h4>
            <label v-for="optionId in optionArray" :key="optionId">
              <input
                type="radio"
                :value="optionId"
                :name="`freeOption-${menu.id}-pattern-${patternIndex}`"
                v-model="selectedOptions[menu.id].freeOptions[patternIndex]"
                @click="handleMenuSelect(menu)"
              />
              {{ getItemName(optionId) }}
            </label>
          </li>
        </ul>

        <!-- 無料複数選択オプション -->
        <h3>無料複数選択オプション:</h3>
        <ul v-if="menu.freeMultiOptions?.length">
          <li v-for="optionId in menu.freeMultiOptions" :key="optionId">
            <label>
              <input
                type="checkbox"
                :value="optionId"
                v-model="selectedOptions[menu.id].freeMultiOptions"
              />
              {{ getItemName(optionId) }}
            </label>
          </li>
        </ul>

        <button @click="order(menu)">カートに追加</button>
      </div>
    </div>

    <div v-else>
      <p>メニューが見つかりません。</p>
    </div>

    <h1>履歴</h1>
    <div v-for="(order, index) in orderHistories" :key="index" class="order">
      <h2>注文 {{ index + 1 }}</h2>
      <p>メニュー名: {{ order.menuName }}</p>
      <p>合計価格: {{ order.totalPrice }}円</p>
      <h3>選択されたアイテム:</h3>
      <ul>
        <li v-for="item in order.itemNames" :key="item">
          {{ item }}
        </li>
      </ul>
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
