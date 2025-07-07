<script setup>
import { ref, reactive, onMounted } from 'vue'

import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue';

const props = defineProps({
  id: String,
  code: String,
  reception: String
})

const reception = ref(null)
const defaultReception = {
  receptionID: '',
  channelID: '',
  receptionTitle: '',
  adminNames: [''],
  joinNames: [''],
  passcodes: [{ passkey: '', usageLimit: 1, passStart: '', passEnd: '' }],
  asks: [''],
  askChoices: [[]],
  askMultiChoices: [[]],
  facilities: [{ facilityName: '', facilityCount: 1 }],
  openTimes: [{ limitStart: '', limitEnd: '' }],
  shifts: [{
    aliasNames: [''],
    shiftStart: '',
    shiftEnd: '',
    open: 1,
    role: '',
    fix: false
  }],
  menus: [{
    menuID: 1,
    menuName: '',
    price: 0,
    prepaidPrice: 0,
    needSkill: '',
    needFacility: '',
    specifyNameFlag: false,
    spendMinute: 0,
    items: [],
    paidOptions: [],
    freeOptions: [],
    freeMultiOptions: []
  }],
  skills: [''],
  staffSkills: [{ aliasName: '', skills: [''] }],
  workStaffNeed: false,
  seats: [{
    seatName: '',
    capacity: 1,
    passcodes: [],
    currentCode: ''
  }],
  itemDetails: [{
    itemID: 1,
    itemName: '',
    imgPath: '',
    choices: [] // choices: [['硬い','普通','柔らかい'],['油多め','普通','油少なめ']]
  }],
  waitConfigs: [{
    guestRange: [1, 5],
    waitRatio: 0
  }]
}

// const add = {
//   receptionID: "3eHg",
//   channelID: '3eHg',
//   receptionTitle: "テスト予約カレンダー",
//   adminNames: ["管理グループA"],
//   menus: [
//     { menuID: 1, menuName: "カット", needSkill: "カット技術", needFacility: "施術室A" },
//     { menuID: 2, menuName: "カラー", needSkill: "カラー技術", needFacility: "施術室B" },
//   ],
//   books: [
//     {
//       bookStart: "2025-07-04T10:00",
//       bookEnd: "2025-07-04T11:00",
//       menuID: 1,
//     },
//     {
//       bookStart: "2025-07-04T15:00",
//       bookEnd: "2025-07-04T16:00",
//       menuID: 2,
//     },
//   ],
//   workStaffs: [
//     {
//       aliasName: "スタッフA",
//       workStart: "2025-07-04T09:00",
//       workEnd: "2025-07-04T17:00",
//     },
//   ],
//   staffSkills: [
//     {
//       aliasName: "スタッフA",
//       skills: ["カット技術", "カラー技術"],
//     },
//     {
//       aliasName: "スタッフB",
//       skills: ["カット技術"],
//     },
//   ],
//   facilities: [
//     {
//       facilityName: "施術室A",
//       facilityCount: 1,
//     },
//     {
//       facilityName: "施術室B",
//       facilityCount: 1,
//     },
//   ],
// };

const add = {
  receptionID: "3eHg",
  channelID: "3eHg",
  receptionTitle: "サンプル受付",
  menus: [
    {
      menuID: 1,
      menuName: "ラーメン",
      price: 800,
      items: [101, 102],
      paidOptions: [
        { itemID: 201, price: 100 },
        { itemID: 202, price: 150 }
      ],
      freeOptions: [
        [301, 302], // パターン1
        [303, 304]  // パターン2
      ],
      freeMultiOptions: [401, 402]
    },
    {
      menuID: 2,
      menuName: "チャーハン",
      price: 700,
      items: [103],
      paidOptions: [],
      freeOptions: [],
      freeMultiOptions: []
    }
  ],
  itemDetails: [
    {
      itemID: 101,
      itemName: "麺の硬さ",
      choices: [
        ["硬め", "普通", "柔らかめ"]
      ]
    },
    {
      itemID: 102,
      itemName: "スープの濃さ",
      choices: [
        ["濃いめ", "普通", "薄め"]
      ]
    },
    {
      itemID: 103,
      itemName: "チャーハンサイズ",
      choices: [
        ["小", "中", "大"]
      ]
    },
    {
      itemID: 201,
      itemName: "味玉",
      // 有料オプションなので choices不要
    },
    {
      itemID: 202,
      itemName: "チャーシュー追加",
    },
    {
      itemID: 301,
      itemName: "ネギあり",
    },
    {
      itemID: 302,
      itemName: "ネギなし",
    },
    {
      itemID: 303,
      itemName: "ごまあり",
    },
    {
      itemID: 304,
      itemName: "ごまなし",
    },
    {
      itemID: 401,
      itemName: "紅ショウガ",
    },
    {
      itemID: 402,
      itemName: "辛味ダレ",
    }
  ]
};


reception.value = defaultReception

const channel = ref(null)
const groups = ref([])
const aliases = ref([])
onMounted(async () => {
  channel.value = await getIDB('channel', localStorage.getItem('channelID'))
  groups.value = await getIDBs('group', 'channelIDIndex', localStorage.getItem('channelID'), 10000)
  aliases.value = await getIDBs('alias', 'channelIDIndex', localStorage.getItem('channelID'), 10000)
  if (props.id) {
    reception.value = await getIDB('reception', props.id);
  } else {
    reception.value = await findReception();
  }
  if (props.reception) {
    try {
      const parsed = JSON.parse(props.reception)
      reception.value = { ...reception.value, ...parsed }
    } catch (e) {
      console.error('receptionパラメータのJSONパースに失敗しました:', e)
    }
  }
  console.log(reception.value)
  reception.value.receptionID = channel.value.channelID;
  reception.value.channelID = channel.value.channelID;
})

async function findReception() {
  const fd = new FormData()
  fd.append('receptionID', localStorage.getItem('channelID'))
  fd.append('channelID', localStorage.getItem('channelID'))
  fd.append('aliasName', channel.value.myname)
  fd.append('csrf', localStorage.getItem('csrf'))
  fd.append('code', props.code)

  const res = await sendRequest('/ReceptionGet/', fd)
  res.csrf && localStorage.setItem('csrf', res.csrf)

  res.pushContents.forEach(content => {
    pushReceive(content)
  })
  return res.reception?.receptionID || defaultReception
}

// ユーティリティ関数
function addArrayItem(field) {
  reception.value[field].push('')
}
function addObjectItem(field, item) {
  reception.value[field].push(item)
}
function removeItem(field, index) {
  reception.value[field].splice(index, 1)
}
function addNestedItem(array, index, nestedField, initValue = '') {
  reception.value[array][index][nestedField].push(initValue)
}

async function submit() {
  event.preventDefault()
  const fd = new FormData();
  fd.append('reception', JSON.stringify(add));
  fd.append('channelID', channel.value.channelID);
  fd.append('aliasName', channel.value.myname);
  fd.append('csrf', localStorage.getItem('csrf'));
  const res = await sendRequest('/ReceptionEdit/', fd);
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
}

</script>

<template>
<Drawer />
  <form id="content">
    <h2 class="sp_head">受付フォーム</h2>

    <label>受付ID:
      <input v-model="reception.receptionID" type="text" />
    </label>

    <label>チャネルID:
      <input v-model="reception.channelID" type="text" />
    </label>

    <label>タイトル:
      <input v-model="reception.receptionTitle" type="text" />
    </label>

    <div>
      <label>管理者:</label>
      <div v-for="(name, i) in reception.adminNames" :key="i">
        <select v-model="reception.adminNames[i]">
          <option disabled value="">アカウント名</option>
          <option v-for="alias in aliases" :key="alias.aliasName" :value="alias.aliasName">
            {{ alias.aliasName }}
          </option>
        </select>
        <button v-if="reception.adminNames.length > 1" @click.prevent="removeItem('adminNames', i)">−</button>
      </div>
      <button @click.prevent="addArrayItem('adminNames')">＋</button>
    </div>

    <div>
      <label>加入者:</label>
      <div v-for="(name, i) in reception.joinNames" :key="i">
        <select v-model="reception.joinNames[i]">
          <option disabled value="">アカウント名</option>
          <option v-for="alias in aliases" :key="alias.aliasName" :value="alias.aliasName">
            {{ alias.aliasName }}
          </option>
        </select>
        <button v-if="reception.joinNames.length > 1" @click.prevent="removeItem('joinNames', i)">−</button>
      </div>
      <button @click.prevent="addArrayItem('joinNames')">＋</button>
    </div>

    <div>
      <label>スキル:</label>
      <div v-for="(skill, i) in reception.skills" :key="i">
        <input v-model="reception.skills[i]" type="text" />
        <button v-if="reception.skills.length > 1" @click.prevent="removeItem('skills', i)">−</button>
      </div>
      <button @click.prevent="addArrayItem('skills')">＋</button>
    </div>

    <div>
      <label>スタッフ保有のスキル:</label>
      <div v-for="(staff, i) in reception.staffSkills" :key="i">
        <select v-model="staff.aliasName">
          <option disabled value="">アカウント名</option>
          <option v-for="alias in aliases" :key="alias.aliasName" :value="alias.aliasName">
            {{ alias.aliasName }}
          </option>
        </select>
        <div v-for="(s, j) in staff.skills" :key="j">
          <select v-model="staff.skills[j]">
            <option disabled value="">スキル</option>
            <option v-for="skill in reception.skills" :value="skill">
              {{ skill }}
            </option>
          </select>

          <button @click.prevent="staff.skills.splice(j, 1)" v-if="staff.skills.length > 1">−</button>
        </div>
        <button @click.prevent="addNestedItem('staffSkills', i, 'skills')">＋スキル</button>
        <button @click.prevent="removeItem('staffSkills', i)" v-if="reception.staffSkills.length > 1">−スタッフ</button>
      </div>
      <button @click.prevent="addObjectItem('staffSkills', { aliasName: '', skills: [''] })">＋スタッフ</button>
    </div>

    <div>
      <label>スタッフと予約を自動調整　美容師など:
        <input type="checkbox" v-model="reception.workStaffNeed" />
      </label>
    </div>

    <!-- ItemDetails -->
    <div>
      <label>アイテム詳細:</label>
      <div v-for="(item, i) in reception.itemDetails" :key="i" style="margin-left: 6px;">
        <input v-model="item.itemID" type="number" placeholder="ID" />
        <input v-model="item.itemName" placeholder="名前" type="text" />
        <input v-model="item.imgPath" placeholder="画像パス" type="text" />

        <!-- choices: 二重配列 -->
        <div>
          <label>利用方法:</label>
          <div v-for="(group, groupIndex) in item.choices" :key="groupIndex" style="margin: 10px 0; padding: 5px; border: 1px dashed #aaa;">
            
            <div>
              <label>利用方法 {{ groupIndex + 1 }}</label>
              <template v-for="(choice, choiceIndex) in group" :key="choiceIndex" style="margin-bottom: 5px;">
                <input v-model="item.choices[groupIndex][choiceIndex]" placeholder="選択肢" />
                <button @click.prevent="item.choices[groupIndex].splice(choiceIndex, 1)" v-if="group.length > 1">−</button>
              </template>
              <button @click.prevent="item.choices[groupIndex].push('')">＋選択肢</button>
              <button @click.prevent="item.choices.splice(groupIndex, 1)" style="margin-left: 10px;" v-if="item.choices.length > 0">利用方法削除</button>
            </div>

          </div>
          <button @click.prevent="item.choices.push(['硬い', '普通', '柔らかい'])" style="margin-top: 10px;">＋利用方法追加</button>
        </div>

        <br>
        <button @click.prevent="reception.itemDetails.splice(i, 1)" v-if="reception.itemDetails.length > 1">−</button>
      </div>
      <button @click.prevent="addObjectItem('itemDetails', {
        itemID: reception.itemDetails.length+1, itemName: '', imgPath: '', choices: []
      })">＋アイテム</button>
    </div>

    <!-- Menus -->
    <div>
      <label>メニュー:</label>
      <div v-for="(m, i) in reception.menus" :key="i" style="margin-left: 6px;">
        メニューID: {{m.menuID}}<br>
        <input v-model="m.menuName" placeholder="メニュー名" type="text" /><br>
        価格: <input v-model.number="m.price" type="number" /><br>
        予約プリペイド価格: <input v-model.number="m.prepaidPrice" type="number" /><br>
        <select v-model="m.needSkill">
          <option disabled value="">スキル</option>
          <option v-for="skill in reception.skills" :value="skill">
            {{ skill }}
          </option>
        </select><br>
        <input v-model="m.needFacility" placeholder="必要設備" type="text" /><br>
        <div>
          <label>指名フラグ:
            <input type="checkbox" v-model="m.specifyNameFlag" />
          </label>
        </div>
        所要時間(分): <input v-model="m.spendMinute" type="number" /><br>
        <!-- Items (int配列) -->
        <div>
          <label>商品の選択:</label>
          <div v-for="(item, idx) in m.items" :key="idx">
            <select v-model="m.items[idx]">
              <option disabled value="">商品</option>
              <option v-for="(item, i) in reception.itemDetails" :value="item.itemID">
                {{ item.itemName }}
              </option>
            </select>
            <button @click.prevent="m.items.splice(idx, 1)">−</button>
          </div>
          <button @click.prevent="m.items.push(0)">＋アイテム</button>
        </div>

        <!-- Paid Options (itemIDと価格) -->
        <div>
          <label>有料オプション:</label>
          <div v-for="(opt, idx) in m.paidOptions" :key="idx">
            <select v-model="opt.itemID">
              <option disabled value="">商品</option>
              <option v-for="(item, i) in reception.itemDetails" :value="item.itemID">
                {{ item.itemName }}
              </option>
            </select>
            価格: <input v-model.number="opt.price" type="number" />
            <button @click.prevent="m.paidOptions.splice(idx, 1)">−</button>
          </div>
          <button @click.prevent="m.paidOptions.push({ itemID: 0, price: 0 })">＋有料オプション</button>
        </div>

        <!-- Free Options (int配列) -->
        <div>
          <label>無料オプション:</label>
          <div v-for="(opt, idx) in m.freeOptions" :key="idx">
            <select v-model="m.freeOptions[idx]">
              <option disabled value="">商品</option>
              <option v-for="(item, i) in reception.itemDetails" :value="item.itemID">
                {{ item.itemName }}
              </option>
            </select>
            <button @click.prevent="m.freeOptions.splice(idx, 1)">−</button>
          </div>
          <button @click.prevent="m.freeOptions.push(0)">＋無料オプション</button>
        </div>

        <!-- Free Multi Options (int配列) -->
        <div>
          <label>無料複数オプション:</label>
          <div v-for="(opt, idx) in m.freeMultiOptions" :key="idx">
            <select v-model="m.freeMultiOptions[idx]">
              <option disabled value="">商品</option>
              <option v-for="(item, i) in reception.itemDetails" :value="item.itemID">
                {{ item.itemName }}
              </option>
            </select>
            <button @click.prevent="m.freeMultiOptions.splice(idx, 1)">−</button>
          </div>
          <button @click.prevent="m.freeMultiOptions.push(0)">＋無料複数オプション</button>
        </div>

        <button @click.prevent="reception.menus.splice(i, 1)" v-if="reception.menus.length > 1">−</button>
      </div>

      <button @click.prevent="addObjectItem('menus', {
        menuID: reception.menus.length,
        menuName: '',
        price: 0,
        prepaidPrice: 0,
        needSkill: '',
        needFacility: '',
        specifyNameFlag: false,
        spendMinute: 0,
        items: [],
        paidOptions: [],
        freeOptions: [],
        freeMultiOptions: []
      })">＋メニュー</button>
    </div>

    <!-- Asks -->
    <div>
      <label>アンケート質問:</label>
      <div v-for="(ask, i) in reception.asks" :key="i">
        <input v-model="reception.asks[i]" placeholder="質問文" type="text" />

        <template v-for="(choice, choiceIndex) in reception.askChoices[i]" :key="choiceIndex" >
          <input v-model="reception.askChoices[i][choiceIndex]" placeholder="選択肢" style="margin: 2px;" />

          <button @click.prevent="() => {
            reception.askChoices[i].splice(choiceIndex, 1);
          }"> − </button>
        </template>
        <button @click.prevent="() => {
          reception.askChoices[i].push('');
        }">＋単一選択肢</button>

        <template v-for="(choice, choiceIndex) in reception.askMultiChoices[i]" :key="choiceIndex" >
          <input v-model="reception.askMultiChoices[i][choiceIndex]" placeholder="選択肢" style="margin: 2px;" />

          <button @click.prevent="() => {
            reception.askMultiChoices[i].splice(choiceIndex, 1);
          }"> − </button>
        </template>
        <button @click.prevent="() => {
          reception.askMultiChoices[i].push('');
        }">＋複数選択肢</button>
        <button @click.prevent="() => {
          reception.asks.splice(i, 1);
          reception.askChoices.splice(i, 1);
          reception.askMultiChoices.splice(i, 1);
        }" v-if="reception.asks.length > 1">−</button>
      </div>
      <button @click.prevent="() => {
        reception.asks.push('');
        reception.askChoices.push([]);
        reception.askMultiChoices.push([]);
      }">＋質問</button>
    </div>

    <!-- Facilities -->
    <div>
      <label>施設:</label>
      <div v-for="(f, i) in reception.facilities" :key="i">
        <input v-model="f.facilityName" placeholder="施設名" type="text" />
        <input v-model.number="f.facilityCount" type="number" placeholder="数" />
        <button @click.prevent="reception.facilities.splice(i, 1)" v-if="reception.facilities.length > 1">−</button>
      </div>
      <button @click.prevent="addObjectItem('facilities', { facilityName: '', facilityCount: 1 })">＋施設</button>
    </div>

    <!-- OpenTimes -->
    <div>
      <label>営業利用可能時間帯:</label>
      <div v-for="(t, i) in reception.openTimes" :key="i">
        開始：<input type="datetime-local" v-model="t.limitStart" /> 〜 終了：<input type="datetime-local" v-model="t.limitEnd" />
        <button @click.prevent="reception.openTimes.splice(i, 1)" v-if="reception.openTimes.length > 1">−</button>
      </div>
      <button @click.prevent="addObjectItem('openTimes', { limitStart: '', limitEnd: '' })">＋時間帯</button>
    </div>

    <!-- Shifts -->
    <div>
      <label>シフトが必要な期間:</label>
      <div v-for="(s, i) in reception.shifts" :key="i">
        <input v-model="s.role" placeholder="役割" type="text" /><br>
        開始：<input type="datetime-local" v-model="s.shiftStart" /> 〜 終了：<input type="datetime-local" v-model="s.shiftEnd" />
        <template v-for="(name, key) in s.aliasNames" >
          <input v-model="s.aliasNames[key]" style="margin: 2px;" />

          <button @click.prevent="() => {
            s.aliasNames.splice(key, 1);
          }"> − </button>
        </template>
        <button @click.prevent="() => {
          s.aliasNames.push('');
        }">＋シフトスタッフ</button>

        <label><input type="checkbox" v-model="s.open" true-value="1" false-value="0" /> 公開</label>
        <label><input type="checkbox" v-model="s.fix" /> 固定</label>
        <button @click.prevent="reception.shifts.splice(i, 1)" v-if="reception.shifts.length > 1">−</button>
      </div>
      <button @click.prevent="addObjectItem('shifts', {
        aliasNames: [''], shiftStart: '', shiftEnd: '', open: 1, role: '', fix: false
      })">＋シフト</button>
    </div>

    <!-- Seats -->
    <div>
      <label>座席:</label>
      <div v-for="(s, i) in reception.seats" :key="i">
        <input v-model="s.seatName" placeholder="座席名" type="text" /><br>
        定員：<input v-model="s.capacity" type="number" /><br>
        <input v-model="s.currentCode" placeholder="現在コード" type="text" /><br>
        <button @click.prevent="reception.seats.splice(i, 1)" v-if="reception.seats.length > 1">−</button>
      </div>
      <button @click.prevent="addObjectItem('seats', {
        seatName: '', capacity: 1, passcodes: [], currentCode: ''
      })">＋座席</button>
    </div>

    <!-- WaitConfigs -->
    <div>
      <label>待機設定:</label>
      <div v-for="(cfg, i) in reception.waitConfigs" :key="i">
        待機できる最小人数:<input v-model.number="cfg.guestRange[0]" type="number" /><br>
        待機できる最大人数:<input v-model.number="cfg.guestRange[1]" type="number" /><br>
        この範囲の人数の団体ゲストの待機単位(分):<input v-model.number="cfg.waitRatio" type="number" />
        <button @click.prevent="reception.waitConfigs.splice(i, 1)" v-if="reception.waitConfigs.length > 1">−</button>
      </div>
      <button @click.prevent="addObjectItem('waitConfigs', {
        guestRange: [1, 5], waitRatio: 0
      })">＋設定</button>
    </div>

    <button type="submit" @click="submit">送信</button>
  </form>
  <div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>
</template>

<style scoped>
/*form {
  max-width: 700px;
  margin: 2rem auto;
  padding: 1rem;
  background: #f9f9f9;
  border: 1px solid #ccc;
  border-radius: 8px;
}*/

input[type="text"],
input[type="email"],
input[type="password"] {
  width: 90%;
  padding: 0.5rem;
  margin-top: 0.2rem;
}

select {
  margin: 6px 0px;
  padding: 6px;
}
/*input[type="number"] {
  width: 40px;
}
*/
button {
  margin: 0.2rem;
  padding: 0.4rem 0.6rem;
  background-color: #42b983;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}
button:hover {
  background-color: #2d9365;
}
label {
  display: block;
  font-weight: bold;
  margin-top: 1rem;
}
</style>
