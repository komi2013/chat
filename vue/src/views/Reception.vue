<script setup>
import { ref, onMounted } from 'vue';
import SelectGroup from '../components/SelectGroup.vue';

const props = defineProps({
  id: '', // Reception ID
});


const channel = ref('');
const groups = ref([]);
const selectedGroup = ref('');
const reception = ref(null);
const selectedMenu = ref(null);

async function fetchChannel() {
  try {
    const data = await getIDB('channel', localStorage.channelID);
    channel.value = data;
    for (let i = 0; i < data.groupAliases.length; i++) {
      groups.value.push([data.groupAliases[i][0], data.groupAliases[i][1]]);
    }
    console.log('groups', groups);
  } catch (error) {
    console.log('error', error);
    channel.value = null;
  }
}

async function fetchReception() {
  try {
    const fd = new FormData();
    fd.append('receptionID', props.id);
    const data = await sendRequest('/ReceptionGet/', fd); // データ取得リクエスト
    if (data) {
      reception.value = data;
      selectedGroup.value = data.adminGroup || '';
      console.log('Reception data:', reception.value);
    }
  } catch (error) {
    console.error('Error fetching reception data:', error);
  }
}

function selectMenu(menu) {
  selectedMenu.value = menu;
  console.log('Selected menu:', selectedMenu.value);
}

// 初期化
onMounted(() => {
  fetchChannel();
  fetchReception();
});
</script>

<template>
  <div>
    <h1>Reception: {{ reception?.id }}</h1>

    <div v-if="reception">
      <!-- メニューのリスト -->
      <h2>メニュー</h2>
      <ul>
        <li v-for="menu in reception.menus" :key="menu.id">
          <div>
            <input v-model="menu.menuName" type="text" />
            <input v-model.number="menu.price" type="number" />
          </div>
          <div v-if="menu.paidOptions?.length">
            <h4>有料オプション:</h4>
            <ul>
              <li v-for="(option, index) in menu.paidOptions" :key="index">
                <label>オプション{{ index + 1 }}:</label>
                <input v-model="menu.paidOptions[index][0]" placeholder="オプションID" type="number" />
                <input v-model.number="menu.paidOptions[index][1]" placeholder="価格" type="number" />
              </li>
              <button @click="menu.paidOptions.push([null, null])">+ オプションを追加</button>
            </ul>
          </div>
          <div v-if="menu.freeOptions?.length">
            <h4>無料オプション:</h4>
            <ul>
              <li v-for="(option, index) in menu.freeOptions" :key="index">
                <label>無料オプション{{ index + 1 }}:</label>
                <input v-model="menu.freeOptions[index]" placeholder="無料オプションID" type="number" />
              </li>
              <button @click="menu.freeOptions.push(null)">+ 無料オプションを追加</button>
            </ul>
          </div>
        </li>
        <button @click="reception.menus.push({ id: Date.now(), menuName: '', price: 0, items: [] })">
          + メニューを追加
        </button>
      </ul>

      <!-- 部屋リスト -->
      <h2>部屋</h2>
      <ul>
        <li v-for="(room, index) in reception.rooms" :key="index">
          <label>部屋名:</label>
          <input v-model="room[1]" type="text" />
          <label>定員:</label>
          <input v-model.number="room[0]" type="number" />
          <button @click="reception.rooms.splice(index, 1)">削除</button>
        </li>
        <button @click="reception.rooms.push([0, ''])">+ 部屋を追加</button>
      </ul>

      <!-- 入室パスコード -->
      <h2>入室パスコード</h2>
      <ul>
        <li v-for="(pass, index) in reception.enterPass" :key="index">
          <label>開始時間:</label>
          <input v-model="pass.passStart" type="datetime-local" />
          <label>終了時間:</label>
          <input v-model="pass.passEnd" type="datetime-local" />
          <ul>
            <li v-for="(code, codeIndex) in pass.passcodes" :key="codeIndex">
              <label>パスコード{{ codeIndex + 1 }}:</label>
              <input v-model="pass.passcodes[codeIndex]" type="text" />
              <button @click="pass.passcodes.splice(codeIndex, 1)">削除</button>
            </li>
            <button @click="pass.passcodes.push('')">+ パスコードを追加</button>
          </ul>
          <button @click="reception.enterPass.splice(index, 1)">削除</button>
        </li>
        <button @click="reception.enterPass.push({ passStart: '', passEnd: '', passcodes: [] })">
          + パスを追加
        </button>
      </ul>

      <!-- アイテムリスト -->
      <h2>アイテムリスト</h2>
      <ul>
        <li v-for="(item, index) in reception.itemDetails" :key="item.itemId">
          <div>
            <label>アイテム名:</label>
            <input v-model="item.itemName" type="text" />
          </div>
          <div>
            <label>画像パス:</label>
            <input v-model="item.imgPath" type="text" />
          </div>
          <div v-if="item.choices">
            <h4>選択肢:</h4>
            <ul>
              <li v-for="(choice, choiceIndex) in item.choices" :key="choiceIndex">
                <input v-model="item.choices[choiceIndex]" type="text" placeholder="選択肢" />
                <button @click="item.choices.splice(choiceIndex, 1)">削除</button>
              </li>
              <button @click="item.choices.push('')">+ 選択肢を追加</button>
            </ul>
          </div>
          <div v-else>
            <button @click="addChoicesToItem(item)">選択肢を追加</button>
          </div>
          <button @click="removeItem(index)">アイテムを削除</button>
        </li>
      </ul>
      <button @click="addItem">+ アイテムを追加</button>

      <SelectGroup 
        :groups="groups"
        v-model="selectedGroup"
      />

    </div>
    <div v-else>
      <p>データをロードしています...</p>
    </div>
  </div>
</template>

<style scoped>
h1, h2, h3, h4 {
  margin: 10px 0;
}

ul {
  list-style: none;
  padding: 0;
}

ul li {
  margin: 5px 0;
}

button {
  padding: 5px 10px;
  margin: 5px 0;
  cursor: pointer;
}

select {
  padding: 5px;
  margin-bottom: 15px;
}
</style>
