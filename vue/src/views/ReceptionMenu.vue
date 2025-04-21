<script setup>
import { ref, reactive, onMounted } from 'vue';

// Props と リアクティブデータ
const props = defineProps({
  id: '', // Reception ID
  code: '',
});

const reception = ref(null); // Reception データ
const selectedOptions = ref([]);
const selectedChoices = ref([]);
async function fetchReception() {
  try {
    const fd = new FormData();
    fd.append('receptionID', props.id);
    fd.append('code', props.code);
    const data = await sendRequest('/ReceptionCheck/', fd);
    if (data) {
      reception.value = data;
      data.menus.forEach(menu => {
        selectedOptions[menu.menuID] = {
          paidOptions: [],
          freeOptions: [],
          freeMultiOptions: []
        };
      });

      console.log('Reception data:', reception.value);
    }
  } catch (error) {
    console.error('Error fetching reception data:', error);
  }
}

async function findReception() {
  const fd = new FormData();
  fd.append('receptionID', props.id);
  fd.append('channelID', localStorage.getItem('channelID'));
  fd.append('aliasName', channel.value.myname);
  fd.append('csrf', localStorage.getItem('csrf'));
  fd.append('receptionID', props.id);
  fd.append('code', props.code);
  const res = await sendRequest('/ReceptionGet/', fd);
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
  const itemMap = {};
  res.reception.itemDetails.forEach(item => {
    itemMap[item.itemID] = item;
    // if (item.choices) {
    //   item.choices.forEach((group, groupIndex) => {
    //     selectedChoices.value['item-' + item.itemID + '-' + groupIndex] = '';
    //   });
    // }
  });

  res.reception.menus.forEach(menu => {
    if (!selectedChoices.value[menu.menuID]) {
      selectedChoices.value[menu.menuID] = {};
    }
    menu.itemDetails = menu.items.map(itemID => itemMap[itemID]);
    menu.itemDetails.forEach(item => {
      if (item.choices) {
        item.choices.forEach((group, groupIndex) => {
          const key = `item-${item.itemID}-${groupIndex}`;
          selectedChoices.value[menu.menuID][key] = '';
        });
      }
    });

    // その他の選択肢構造（必要なら）
    selectedOptions.value[menu.menuID] = {
      paidOptions: [],
      freeOptions: [],
      freeMultiOptions: []
    };

  });
// v-model="selectedChoices['item-choice-' + item.itemID]"
  reception.value = res.reception;
  console.log('menus', reception.value.menus);
  console.log('selectedChoices', selectedChoices.value);
  // return res.reception;
}

function getItemName(itemID) {
  const item = reception.value.itemDetails.find(item => item.itemID === itemID);
  return item ? item.itemName : '不明なアイテム';
}

function getItemDetail(itemID) {
  return reception.value.itemDetails.find(item => item.itemID === itemID) || null;
}

const orderHistories = ref([]);
async function addCart(menu) {

  const pureChoices = Object.entries(selectedChoices.value).reduce((acc, [menuID, group]) => {
    const filtered = Object.entries(group)
      .filter(([_, val]) => val !== '' && val !== null && val !== undefined)
      .reduce((obj, [key, val]) => {
        obj[key] = val;
        return obj;
      }, {});
    if (Object.keys(filtered).length > 0 && menuID == menu.menuID) {
      acc[menuID] = filtered;
    }
    return acc;
  }, {});

  const cartItem = {
    menuID: menu.menuID,
    itemChoices: pureChoices,
    freeOptions: selectedOptions.value[menu.menuID]?.freeOptions || [],
    paidOptions: selectedOptions.value[menu.menuID]?.paidOptions || [],
    freeMultiOptions: selectedOptions.value[menu.menuID]?.freeMultiOptions || []
  };

  // upsertIDB(cartItem, 'receptionOrder', 'receptionOrderID', receptionOrder.receptionOrderID);

  const fd = new FormData();
  fd.append('menuID', menu.menuID);
  fd.append('itemChoices', JSON.stringify(pureChoices));
  fd.append('freeOptions', JSON.stringify(selectedOptions.value[menu.menuID].freeOptions));
  fd.append('paidOptions', JSON.stringify(selectedOptions.value[menu.menuID].paidOptions));
  fd.append('freeMultiOptions', JSON.stringify(selectedOptions.value[menu.menuID].freeMultiOptions));
  fd.append('receptionID', props.id);
  fd.append('code', props.code);
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
  selectedOptions[menu.menuID] = {
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
      const item = itemDetails.find(detail => detail.itemID === itemId);
      return item ? item.itemName : `不明なアイテム (${itemId})`;
    });
  }
  return ['データがありません'];
}
const channel = ref(null);
onMounted(async() => {
  channel.value = await getIDB('channel', localStorage.getItem('channelID'));
  // groups.value = await getIDBs('group', 'channelIDIndex', channelID, 10000);
  // aliases.value = await getIDBs('alias', 'channelIDIndex', channelID, 10000);
  // fetchReception();
  await findReception();
});
</script>

<template>
  <div>
    <h1>メニュー一覧</h1>

    <!-- メニューリスト -->
    <div v-if="reception && reception.menus.length">
      <div v-for="menu in reception.menus" class="menu-item">
        <h2>{{ menu.menuName }}</h2>
        <p>価格: {{ menu.price }}円</p>

        <ul>
          <li v-for="item in menu.itemDetails" :key="itemId">
            {{ item.itemName }}
            <div v-if="item.choices">
              <div
                v-for="(group, groupIndex) in item.choices"
                :key="'group-' + groupIndex"
              >
                <p>選択{{ groupIndex + 1 }}</p>
                <label
                  v-for="choice in group"
                  :key="choice"
                >
                  <input
                    type="radio"
                    :name="'choice-' + item.itemID + '-' + groupIndex"
                    :value="choice"
                    v-model="selectedChoices[menu.menuID]['item-' + item.itemID + '-' + groupIndex]"
                  />
                  {{ choice }}
                </label>
              </div>
            </div>
          </li>
        </ul>

        <!-- 有料オプション -->
        <h3>有料オプション:</h3>
        <ul v-if="menu.paidOptions?.length">
          <li v-for="(option, index) in menu.paidOptions" :key="index">
            <label>
              <input
                type="checkbox"
                :value="option.itemID"
                v-model="selectedOptions[menu.menuID].paidOptions"
              />
              {{ getItemName(option.itemID) }}: {{ option.price }}円
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
                :name="`freeOption-${menu.menuID}-pattern-${patternIndex}`"
                v-model="selectedOptions[menu.menuID].freeOptions[patternIndex]"
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
                v-model="selectedOptions[menu.menuID].freeMultiOptions"
              />
              {{ getItemName(optionId) }}
            </label>
          </li>
        </ul>

        <button @click="addCart(menu)">カートに追加</button>
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
