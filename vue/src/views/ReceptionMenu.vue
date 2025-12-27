<script setup>
import { ref, computed, onMounted } from 'vue'

import Advertisement from '@/components/Advertisement.vue'
import DrawerReception from '@/components/DrawerReception.vue'

import { pushReceive } from '@/pushReceive/pushReceive.js'

const props = defineProps({
  id: String, // Reception ID
  code: String,
  codeType: String,
})

document.title = 'メニュー'

const reception = ref(null); // Reception データ
const seatName = ref('');
const selectedOptions = ref([]);
const selectedChoices = ref([]);
let nickname
async function findReception() {
  const fd = new FormData()
  fd.append('csrf', localStorage.getItem('csrf'))
  fd.append('receptionID', props.id)
  fd.append('code', props.code)
  fd.append('codeType', props.codeType) // 1 = before enter, 2 = take QR code, 3 = ordering
  const res = await sendRequest('/ReceptionGet/', fd)
  if (!res.csrf) {
    errorMessage.value = res
    return
  }
  localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  if (res.error) { errorMessage.value = res.error }
  if (props.codeType == '2') location.href = `/ReceptionMenu/${props.id}/${props.code}/3/`
  res.reception.menus = res.menu.menus || []
  res.reception.itemDetails = res.menu.itemDetails || []
  const itemMap = {}
  res.reception.itemDetails.forEach(item => {
    itemMap[item.itemID] = item
  })

  res.reception.menus.forEach(menu => {
    if (!selectedChoices.value[menu.menuID]) {
      selectedChoices.value[menu.menuID] = {};
    }
    menu.itemDetails = menu.items.map(itemID => itemMap[itemID]);
    menu.itemDetails.forEach(item => {
      if (item.choices) {
        item.choices.forEach((group, groupIndex) => {
          const key = `item-${item.itemID}-${groupIndex}`;
          selectedChoices.value[menu.menuID][key] = group[0] ?? ''
        });
      }
    });

    const freeOptions = (menu.freeOptions || []).map(optArray => optArray[0] ?? null); 
    selectedOptions.value[menu.menuID] = {
      paidOptions: [],                      // デフォルト選択なし
      freeOptions,                          // 各パターンで最初の要素をデフォルトに
      freeMultiOptions: []                  // 複数選択は空配列でスタート
    };

  });
  reception.value = res.reception
  seatName.value = res.facilityName
  nickname = res.nickname
}

function getItemName(itemID) {
  const item = reception.value.itemDetails.find(item => item.itemID === itemID);
  return item ? item.itemName : '不明なアイテム';
}

function getItemDetail(itemID) {
  return reception.value.itemDetails.find(item => item.itemID === itemID) || null;
}

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

const errorMessage = ref('')
onMounted(async() => {
  await findReception()
});

const receptionOrders = ref([]);

async function addCart(menu) {
  console.log('menu', menu)
  const receptionOrder = createReceptionOrder(menu)
  receptionOrders.value.push(receptionOrder)
}

function createReceptionOrder(menu) {
  let seq = incrementBase62Smart(localStorage.getItem('cartItemSeq'));
  localStorage.setItem('cartItemSeq', seq);
  const itemDetailsMaster = reception.value.itemDetails;
  const opts = selectedOptions.value[menu.menuID] || {
    freeOptions: [],
    paidOptions: [],
    freeMultiOptions: []
  };
  const pureChoices = Object.entries(selectedChoices.value[menu.menuID] || {})
    .filter(([_, val]) => val !== '' && val !== null && val !== undefined)
    .reduce((obj, [key, val]) => {
      obj[key] = val;
      return obj;
    }, {});
  const paidOptions = (opts.paidOptions || []).map(id => {
    const optDef = menu.paidOptions.find(o => o.itemID === id);
    return {
      itemID: id,
      price: optDef ? optDef.price : 0
    };
  });
  const paidOptionsPrice = paidOptions.reduce((sum, opt) => sum + opt.price, 0);
  const price = menu.price + paidOptionsPrice;
  let order = {
    receptionOrderID: `${menu.menuID}${nickname}${seq}`,
    seatName: seatName.value,
    menuID: menu.menuID,
    menuName: menu.menuName,
    itemChoices: pureChoices,
    freeOptions: opts.freeOptions,
    paidOptions,
    freeMultiOptions: opts.freeMultiOptions,
    price,
    items: (menu.itemDetails || []).map(d => ({
      itemID: d.itemID,
      itemName: d.itemName,
      imgPath: d.imgPath,
      choices: Array.isArray(d.choices) ? d.choices : []
    }))
  };
  order = Object.fromEntries(
    Object.entries(order).filter(([_, v]) => {
      if (Array.isArray(v)) return v.length > 0;       // 空配列は消す
      if (v && typeof v === 'object') return Object.keys(v).length > 0; // 空オブジェクトは消す
      return v !== null && v !== undefined && v !== ''; // null/undefined/空文字は消す
    })
  );

  return order;
}

function removeOrder(receptionOrderID) {
  receptionOrders.value = receptionOrders.value.filter(
    order => order.receptionOrderID !== receptionOrderID
  );
}

const totalPrice = computed(() => {
  return receptionOrders.value.reduce((sum, order) => {
    return sum + Number(order.price || 0);
  }, 0);
});


async function order() {
  if (!confirm("注文")) { return }
  const fd = new FormData()
  fd.append('csrf', localStorage.getItem('csrf'))
  fd.append('receptionID', props.id)
  fd.append('receptionOrders', JSON.stringify(receptionOrders.value))
  fd.append('code', props.code)
  // fd.append('codeType', 3) // 1 = before enter, 2 = at seat
  const res = await sendRequest('/ReceptionOrder/', fd)
  if (!res.csrf) {
    errorMessage.value = res
    return
  }
  localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  if (res.error) { errorMessage.value = res.error }
}

</script>

<template>
<div id="drawer_column"><DrawerReception :id="id" :code="code" /></div>
  <div id="content">
    <h1 class="sp_head">メニュー一覧</h1>
    <div>{{seatName}}</div>
    <div v-if="errorMessage"> 
      <div class="errorMessage">{{errorMessage}}<br>データ取得に失敗しました。</div>
      <a href="/setting/"> データ設定ページ </a><br>
      <a href="/sign/"> サインインページ </a>
    </div>
    <!-- メニューリスト -->
    <div v-if="reception && reception.menus.length">
      <div v-for="menu in reception.menus" class="menu-item">
        <h2>{{ menu.menuName }}</h2>
        <div>価格: {{ menu.price }}円</div>

        <div v-for="item in menu.itemDetails" :key="itemId">
          <div style="display: inline-block;">
            <img :src="item.imgPath">
          </div>

          <div>
            {{ item.itemName }}
            <div v-if="item.choices" v-for="(group, groupIndex) in item.choices" :key="'group-' + groupIndex">
              <label v-for="choice in group" :key="choice" >
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
        </div>

        <!-- 有料オプション -->
        <div v-if="menu.paidOptions?.length" class="option-list">
          <h3>有料オプション:</h3>
          <label
            v-for="(option, index) in menu.paidOptions"
            :key="index"
            class="option-label"
            :class="{ selected: selectedOptions[menu.menuID].paidOptions.includes(option.itemID) }"
          >
            <input
              type="checkbox"
              :value="option.itemID"
              v-model="selectedOptions[menu.menuID].paidOptions"
              class="hidden-radio"
            />
            <img 
              v-if="getItemDetail(option.itemID)?.imgPath" 
              :src="getItemDetail(option.itemID).imgPath" 
              :alt="getItemDetail(option.itemID).itemName" 
              class="option-img"
            />
            <div class="option-name">
              {{ getItemDetail(option.itemID)?.itemName }} <br />
              <span class="option-price">{{ option.price }}円</span>
            </div>
          </label>
        </div>

        <!-- 無料オプション (複数パターンで単一選択) -->
        <div 
          v-if="menu.freeOptions?.length" 
          v-for="(optionArray, patternIndex) in menu.freeOptions" 
          :key="patternIndex"
          class="option-group"
        >
          <h3>無料オプション:</h3>

          <div class="option-list">
            <label 
              v-for="optionId in optionArray" 
              :key="optionId" 
              class="option-label"
              :class="{ selected: selectedOptions[menu.menuID].freeOptions[patternIndex] === optionId }"
            >
              <input
                type="radio"
                :value="optionId"
                :name="`freeOption-${menu.menuID}-pattern-${patternIndex}`"
                v-model="selectedOptions[menu.menuID].freeOptions[patternIndex]"
                class="hidden-radio"
              />
              <img 
                v-if="getItemDetail(optionId)?.imgPath" 
                :src="getItemDetail(optionId).imgPath" 
                :alt="getItemDetail(optionId).itemName" 
                class="option-img"
              />
              <div class="option-name">{{ getItemDetail(optionId)?.itemName }}</div>
            </label>
          </div>
        </div>

        <!-- 無料複数選択オプション -->
        <div v-if="menu.freeMultiOptions?.length" class="option-list">
          <h3>無料複数選択オプション:</h3>
          <label
            v-for="optionId in menu.freeMultiOptions"
            :key="optionId"
            class="option-label"
            :class="{ selected: selectedOptions[menu.menuID].freeMultiOptions.includes(optionId) }"
          >
            <input
              type="checkbox"
              :value="optionId"
              v-model="selectedOptions[menu.menuID].freeMultiOptions"
              class="hidden-radio"
            />
            <img
              v-if="getItemDetail(optionId)?.imgPath"
              :src="getItemDetail(optionId).imgPath"
              :alt="getItemDetail(optionId).itemName"
              class="option-img"
            />
            <div class="option-name">{{ getItemDetail(optionId)?.itemName }}</div>
          </label>
        </div>

        <button @click="addCart(menu)">カートに追加</button>
      </div>
    </div>
    <div v-else>
      <p>メニューが見つかりません。</p>
    </div>

    <div class="reception-orders">
      <h2>カート内の注文 {{seatName}} </h2>
      <div
        v-if="receptionOrders.length > 0"
        v-for="order in receptionOrders"
        :key="order.receptionOrderID"
        class="order-item"
        >
        <h3>{{ order.menuName }}</h3>
        <!-- 削除ボタン -->
        <button class="delete-btn" @click="removeOrder(order.receptionOrderID)">
          ❌ 削除
        </button>

        <!-- 注文アイテム -->
        <div
          v-for="item in order.items"
          :key="item.itemID"
          class="item-detail"
          >
          <img
            v-if="item.imgPath"
            :src="item.imgPath"
            :alt="item.itemName"
            class="item-img"
          />
          <span>{{ item.itemName }}</span>
          <div v-if="item.choices && item.choices.length > 0">
            <div v-for="(group, groupIndex) in item.choices" :key="groupIndex">
              <span v-if="order.itemChoices && order.itemChoices[order.menuID]">
                {{
                  order.itemChoices[order.menuID][`item-${item.itemID}-${groupIndex}`]
                }}
              </span>
            </div>
          </div>
        </div>

        <!-- オプション -->
        <div class="options">
          <!-- 無料オプション -->
          <div v-if="order.freeOptions?.length > 0">
            <div>無料オプション:</div>
            <div class="option-list">
              <div
                v-for="id in order.freeOptions"
                :key="'free-' + id"
                class="option-item"
              >
                <img
                  v-if="getItemDetail(id)?.imgPath"
                  :src="getItemDetail(id).imgPath"
                  :alt="getItemDetail(id).itemName"
                  class="option-img"
                />
                <div>{{ getItemDetail(id)?.itemName }}</div>
              </div>
            </div>
          </div>

          <!-- 有料オプション -->
          <div v-if="order.paidOptions?.length > 0">
            <div>有料オプション:</div>
            <div class="option-list">
              <div
                v-for="id in order.paidOptions"
                :key="'paid-' + id.itemID"
                class="option-item"
                >
                <img
                  v-if="getItemDetail(id.itemID)?.imgPath"
                  :src="getItemDetail(id.itemID).imgPath"
                  :alt="getItemDetail(id.itemID).itemName"
                  class="option-img"
                />
                <div>{{ getItemDetail(id.itemID)?.itemName }} {{id.price}} 円 </div>
              </div>
            </div>
          </div>

          <!-- 複数選択オプション -->
          <div v-if="order.freeMultiOptions?.length > 0">
            <p><strong>複数選択オプション:</strong></p>
            <div class="option-list">
              <div
                v-for="id in order.freeMultiOptions"
                :key="'multi-' + id"
                class="option-item"
              >
                <img
                  v-if="getItemDetail(id)?.imgPath"
                  :src="getItemDetail(id).imgPath"
                  :alt="getItemDetail(id).itemName"
                  class="option-img"
                />
                <div>{{ getItemDetail(id)?.itemName }}</div>
              </div>
            </div>
          </div>
          <div>メニュー価格: {{order.price}}円</div>
        </div>
      </div>
      <p v-else>カートに商品はありません。</p>
      <!-- ✅ 合計金額 -->
      <div class="total-price">
        <strong>合計: {{ totalPrice }} 円</strong>
      </div>
      <button @click="order()">注文</button>
    </div>

  </div>
  <div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>
</template>

<style scoped>

.hidden-radio {
  display: none;
}

.option-group {
  margin-bottom: 1rem;
}

.option-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.option-label {
  display: flex;
  flex-direction: column;
  align-items: center;
  cursor: pointer;
  border: 2px solid transparent;
  border-radius: 8px;
  padding: 5px;
  transition: border-color 0.2s, background-color 0.2s;
}

.option-label:hover {
  border-color: #999;
}

.option-label.selected {
  border-color: #007bff;
  background-color: #e6f0ff;
}

.option-img {
  width: 120px;
  height: 120px;
  max-width: 250px;
  max-height: 250px;
  object-fit: cover;
  border-radius: 6px;
}

.option-name {
  margin-top: 5px;
  font-size: 0.9rem;
  text-align: center;
}


/* メニュー */
.menu-item {
  border: 1px solid #ccc;
  margin-bottom: 10px;
  padding: 10px;
  border-radius: 5px;
}

/* ボタン */
button {
  margin-top: 10px;
  cursor: pointer;
  width: 100%; /* 横幅いっぱいに */
  padding: 10px;
  font-size: 1rem;
  border-radius: 4px;
  border: none;
  background: #007bff;
  color: white;
}

/* カート内 */
.reception-orders {
  margin: 0; /* インデントなし */
}
.order-item {
  border: 1px solid #ccc;
  padding: 10px;
  margin-bottom: 12px;
  border-radius: 5px;
}

/* 商品詳細 */
.item-detail {
  display: flex;
  flex-direction: column; /* 縦並びに */
  align-items: flex-start;
  gap: 6px;
}
.item-img {
  max-width: 250px;
  max-height: 250px;
  width: 100%;
  height: auto;
  object-fit: cover;
  border-radius: 4px;
}

/* オプション表示 */
.options {
  margin-top: 6px;
  font-size: 0.9rem;
}

/* スマホ用の微調整 */
@media (max-width: 600px) {
  h2, h3, h4 {
    font-size: 1rem;
    margin: 6px 0;
  }
  .menu-item, .order-item {
    padding: 8px;
  }
}
</style>

