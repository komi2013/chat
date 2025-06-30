<script setup>
import { ref, reactive, onMounted } from 'vue'

import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue';

const props = defineProps({
  id: String,
  code: String
})

const reception = ref({
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
    specifyNameFlag: 0,
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
    itemID: 0,
    itemName: '',
    imgPath: '',
    choices: [[]]
  }],
  waitConfigs: [{
    guestRange: [1, 5],
    waitRatio: 0
  }]
})

const channel = ref(null)

onMounted(async () => {
  channel.value = await getIDB('channel', localStorage.getItem('channelID'))
  if (props.id) {
    await findReception()
  }
})

async function findReception() {
  const fd = new FormData()
  fd.append('receptionID', props.id)
  fd.append('channelID', localStorage.getItem('channelID'))
  fd.append('aliasName', channel.value.myname)
  fd.append('csrf', localStorage.getItem('csrf'))
  fd.append('code', props.code)

  const res = await sendRequest('/ReceptionGet/', fd)
  res.csrf && localStorage.setItem('csrf', res.csrf)

  res.pushContents.forEach(content => {
    pushReceive(content)
  })

  reception.value = res.reception || {}
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

function submit() {
  event.preventDefault()
  console.log(reception.value);
}

</script>

<template>
<Drawer />
  <form id="content">
    <h2>受付フォーム</h2>

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
        <input v-model="reception.adminNames[i]" type="text" />
        <button v-if="reception.adminNames.length > 1" @click.prevent="removeItem('adminNames', i)">−</button>
      </div>
      <button @click.prevent="addArrayItem('adminNames')">＋</button>
    </div>

    <div>
      <label>加入者:</label>
      <div v-for="(name, i) in reception.joinNames" :key="i">
        <input v-model="reception.joinNames[i]" type="text" />
        <button v-if="reception.joinNames.length > 1" @click.prevent="removeItem('joinNames', i)">−</button>
      </div>
      <button @click.prevent="addArrayItem('joinNames')">＋</button>
    </div>

    <div>
      <label>スキル名:</label>
      <div v-for="(skill, i) in reception.skills" :key="i">
        <input v-model="reception.skills[i]" type="text" />
        <button v-if="reception.skills.length > 1" @click.prevent="removeItem('skills', i)">−</button>
      </div>
      <button @click.prevent="addArrayItem('skills')">＋</button>
    </div>

    <div>
      <label>スタッフ保有のスキル:</label>
      <div v-for="(staff, i) in reception.staffSkills" :key="i">
        <input v-model="staff.aliasName" placeholder="スタッフ名" type="text" />
        <div v-for="(s, j) in staff.skills" :key="j">
          <input v-model="staff.skills[j]" placeholder="スキル" type="text" />
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

    <!-- Menus -->
    <div>
      <label>メニュー:</label>
      <div v-for="(m, i) in reception.menus" :key="i">
        <!-- <input v-model.number="m.menuID" type="number" placeholder="メニューID" /> -->
        メニューID: {{m.menuID}}<br>
        <input v-model="m.menuName" placeholder="メニュー名" type="text" /><br>
        価格: <input v-model.number="m.price" type="number" /><br>
        予約プリペイド価格: <input v-model.number="m.prepaidPrice" type="number" /><br>
        <input v-model="m.needSkill" placeholder="必要スキル名" type="text" /><br>
        <input v-model="m.needFacility" placeholder="必要設備名" type="text" /><br>
        指名フラグ: <input v-model.number="m.specifyNameFlag" type="number" /><br>
        所要時間(分): <input v-model.number="m.spendMinute" type="number" /><br>
        <button @click.prevent="reception.menus.splice(i, 1)" v-if="reception.menus.length > 1">−</button>
      </div>
      <button @click.prevent="addObjectItem('menus', {
        menuID: reception.menus.length, menuName: '', price: 0, prepaidPrice: 0, needSkill: '', needFacility: '',
        specifyNameFlag: 0, spendMinute: 0, items: [], paidOptions: [], freeOptions: [], freeMultiOptions: []
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
        <input type="text"
          v-model="s.aliasNames[0]"
          @blur="s.aliasNames = s.aliasNames[0].split(',').map(s => s.trim())"
          placeholder="担当者 (カンマ区切り)"
        />
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
        <input v-model="s.seatName" placeholder="座席名" type="text" />
        <input v-model.number="s.capacity" type="number" placeholder="定員" />
        <input v-model="s.currentCode" placeholder="現在コード" type="text" />
        <button @click.prevent="reception.seats.splice(i, 1)" v-if="reception.seats.length > 1">−</button>
      </div>
      <button @click.prevent="addObjectItem('seats', {
        seatName: '', capacity: 1, passcodes: [], currentCode: ''
      })">＋座席</button>
    </div>

    <!-- ItemDetails -->
    <div>
      <label>アイテム詳細:</label>
      <div v-for="(item, i) in reception.itemDetails" :key="i">
        <input v-model.number="item.itemID" type="number" placeholder="ID" />
        <input v-model="item.itemName" placeholder="名前" type="text" />
        <input v-model="item.imgPath" placeholder="画像パス" type="text" />
        <input type="text"
          v-model="item.choices[0]"
          placeholder="選択肢 (カンマ区切り)"
          @blur="item.choices[0] = item.choices[0].split(',').map(s => s.trim())"
        />
        <button @click.prevent="reception.itemDetails.splice(i, 1)" v-if="reception.itemDetails.length > 1">−</button>
      </div>
      <button @click.prevent="addObjectItem('itemDetails', {
        itemID: 0, itemName: '', imgPath: '', choices: [[]]
      })">＋アイテム</button>
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
