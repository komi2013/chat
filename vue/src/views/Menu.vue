<script setup>
import { ref, reactive, onMounted } from 'vue';

// Props と リアクティブデータ
const props = defineProps({
  id: '', // Reception ID
  code: '',
});

const reception = ref(null); // Reception データ
const cart = ref([]); // カート内のメニュー

// メニューごとの選択状態を管理するオブジェクト
const selectedOptions = reactive({});

// データ取得関数
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
          freeOption: null,
          freeMultiOptions: []
        };
      });

      console.log('Reception data:', reception.value);
    }
  } catch (error) {
    console.error('Error fetching reception data:', error);
  }
}

// アイテム名を取得する関数
function getItemName(itemId) {
  const item = reception.value.itemDetails.find(item => item.itemId === itemId);
  return item ? item.itemName : '不明なアイテム';
}

function order(menu) {
  const cartItem = {
    ...menu,
    selectedPaidOptions: [...selectedOptions[menu.id].paidOptions],
    selectedFreeOption: selectedOptions[menu.id].freeOption,
    selectedFreeMultiOptions: [...selectedOptions[menu.id].freeMultiOptions],
  };

  const fd = new FormData();
  fd.append('receptionID', props.id);
  fd.append('code', props.code);
  fd.append('freeOptions', selectedOptions[menu.id].freeOption);
  fd.append('paidOptions', selectedOptions[menu.id].paidOptions);
  fd.append('freeMultiOptions', selectedOptions[menu.id].freeMultiOptions);
  const data = sendRequest('/ReceptionOrder/', fd);
  console.log(data);
  // cart.value.push(cartItem);

  // 選択状態をリセット
  selectedOptions[menu.id] = {
    paidOptions: [],
    freeOption: null,
    freeMultiOptions: []
  };

  console.log('Cart updated:', cart.value);
}

// カートから削除する関数
function removeFromCart(index) {
  cart.value.splice(index, 1);
}

// 初期化処理
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

        <!-- 無料オプション (単一選択) -->
        <h3>無料オプション:</h3>
        <ul v-if="menu.freeOptions?.length">
          <li v-for="(optionArray, index) in menu.freeOptions" :key="index">
            <label v-for="optionId in optionArray" :key="optionId">
              <input
                type="radio"
                :value="optionId"
                name="freeOption-{{ menu.id }}"
                v-model="selectedOptions[menu.id].freeOption"
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

    <!-- カート表示 -->
    <h1>カート</h1>
    <div v-if="cart.length">
      <div v-for="(cartItem, index) in cart" :key="index" class="cart-item">
        <h2>{{ cartItem.menuName }}</h2>
        <p>価格: {{ cartItem.price }}円</p>

        <h3>選択された有料オプション:</h3>
        <ul>
          <li v-for="option in cartItem.selectedPaidOptions" :key="option">
            {{ getItemName(option) }}
          </li>
        </ul>

        <h3>選択された無料オプション:</h3>
        <p>{{ getItemName(cartItem.selectedFreeOption) || '未選択' }}</p>

        <h3>選択された無料複数選択オプション:</h3>
        <ul>
          <li v-for="option in cartItem.selectedFreeMultiOptions" :key="option">
            {{ getItemName(option) }}
          </li>
        </ul>

        <button @click="removeFromCart(index)">削除</button>
      </div>
    </div>
    <div v-else>
      <p>カートが空です。</p>
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
