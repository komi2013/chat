<script setup>
import { ref, reactive, onMounted } from 'vue'

import Advertisement from '@/components/Advertisement.vue'
import Drawer from '@/components/Drawer.vue'
import Image from '@/components/Image.vue';
import SelectAlias from '@/components/SelectAlias.vue'

import { pushReceive } from '@/pushReceive/pushReceive.js'

const props = defineProps({
  id: String,
  code: String,
  reception: String
})
document.title = '受付設定'

const channelID = localStorage.getItem("channelID")

const reception = ref(null)
const defaultReception = {
  receptionID: channelID,
  channelID: channelID,
  receptionTitle: '',
  adminNames: [''],
  joinNames: [''],
  passcodes: [{passkey: '', usageLimit: 1, passStart: '', passEnd: '' }],
  askChoices: [{
    question: '',
    choices: [],
    sequence: 0
  }],
  askMultiChoices: [{
    question: '',
    choices: [],
    sequence: 0
  }],
  asks: [{
    question: '',
    sequence: 0
  }],
  facilities: [{ 
    facilityName: '',
    capacity: 0,
    bookable: false
  }],
  openTimes: [{ limitStart: '', limitEnd: '' }],
  shifts: [{
    aliasNames: [''],
    shiftStart: '',
    shiftEnd: '',
    open: 1,
    skill: '',
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
    paidOptions: [{
      itemID: 1,
      price: 0
    }],
    freeOptions: [[]], // itemID　[オレンジジュース、メロンソーダ], [パン, ご飯]
    freeMultiOptions: [],
    forBookType: ''  // 1 = book only, 2 = book & at shop
  }],
  skills: [''],
  staffSkills: [{ aliasName: '', skills: [''] }],
  workStaffNeed: false,
  itemDetails: [{
    itemID: 1,
    itemName: '',
    imgPath: '',
    choices: [[]] // choices: [['硬い','普通','柔らかい'],['油多め','普通','油少なめ']]
  }],
  waitConfigs: [{
    guestRange: [1, 5],
    waitRatio: 0
  }]
}
reception.value = defaultReception

async function findReception() {
  const fd = new FormData()
  fd.append('receptionID', props.id)
  fd.append('channelID', localStorage.getItem('channelID'))
  fd.append('aliasName', channel.value.myname)
  fd.append('csrf', localStorage.getItem('csrf'))
  fd.append('code', props.code)
  const res = await sendRequest('/ReceptionGet/', fd)
  if (!res.csrf) errorMessage.value = res
  res.csrf && localStorage.setItem('csrf', res.csrf)
  res.pushContents.forEach(content => {
    pushReceive(content)
  })
  console.log('res.reception', res.reception.receptionID)
  if (!res.reception) { return defaultReception }
  res.reception.menus = res.menu.menus || []
  res.reception.itemDetails = res.menu.itemDetails || []

  res.reception.menus.forEach(menu => {
    menu.paidOptions = menu.paidOptions || []
    menu.freeOptions = menu.freeOptions || []
    menu.freeMultiOptions = menu.freeMultiOptions || []
  })
  return res.reception
}

const channel = ref(null)
const groups = ref([])
const aliases = ref([])
const found = ref(false)
const errorMessage = ref('')
onMounted(async () => {
  channel.value = await getIDB('channel', channelID)
  groups.value = await getIDBs('group', 'channelIDIndex', channelID, 10000)
  aliases.value = await getIDBs('alias', 'channelIDIndex', channelID, 10000)
  if (props.reception) {
    try {
      const parsed = JSON.parse(props.reception)
      reception.value = { ...reception.value, ...parsed }
    } catch (e) {
      console.error('receptionパラメータのJSONパースに失敗しました:', e)
    }
  } else if (props.id) {
    // reception.value = await getIDB('reception', props.id);
    reception.value = await findReception()
    // console.log('reception.value', reception.value)
  // } else {
  //   reception.value = await findReception()
  }
  found.value = true
})

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
  if (!confirm("実行▶️")) {
    return;
  }
  event.preventDefault()
  const fd = new FormData();
  console.log('reception', reception.value)
  const cleanedReception = removeEmpty(reception.value);
  const cleanedMenu = removeEmpty({
    menus: reception.value.menus || [],
    itemDetails: reception.value.itemDetails || []
  })
  fd.append('reception', JSON.stringify(cleanedReception))
  fd.append('menu', JSON.stringify(cleanedMenu))
  fd.append('channelID', channel.value.channelID)
  fd.append('aliasName', channel.value.myname)
  fd.append('csrf', localStorage.getItem('csrf'))
  const res = await sendRequest('/ReceptionEdit/', fd)
  if (!res.csrf) errorMessage.value = res
  res.csrf && localStorage.setItem('csrf', res.csrf)
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
}

// 再帰的に空データを削除する関数
function removeEmpty(obj) {
  if (Array.isArray(obj)) {
    return obj
      .map(item => removeEmpty(item))         // 子要素も処理
      .filter(item => {                       // 空要素は削除
        if (item === null || item === undefined) return false;
        if (typeof item === 'string' && item.trim() === '') return false;
        if (Array.isArray(item) && item.length === 0) return false;
        if (typeof item === 'object' && Object.keys(item).length === 0) return false;
        return true;
      });
  } else if (typeof obj === 'object' && obj !== null) {
    const newObj = {};
    Object.entries(obj).forEach(([key, value]) => {
      const cleaned = removeEmpty(value);
      if (
        cleaned === null ||
        cleaned === undefined ||
        (typeof cleaned === 'string' && cleaned.trim() === '') ||
        (Array.isArray(cleaned) && cleaned.length === 0) ||
        (typeof cleaned === 'object' && Object.keys(cleaned).length === 0)
      ) {
        // 空はスキップ
      } else {
        newObj[key] = cleaned;
      }
    });
    return newObj;
  }
  return obj;
}


</script>

<template>
<div id="drawer_column"><Drawer v-if="aliases" :aliases="aliases" :channel="channel" /></div>
  <form id="content">
    <h2 class="sp_head">受付フォーム</h2>
    <div v-if="errorMessage"> 
      <div class="errorMessage">{{errorMessage}}<br>データ取得に失敗しました。</div>
      <a href="/setting/"> データ設定ページ </a><br>
      <a href="/sign/"> サインインページ </a>
    </div>
    <label>タイトル:
      <input v-model="reception.receptionTitle" type="text" />
    </label>

    <div>
      <label>管理者:</label>
      <SelectAlias v-if="found"
        :channel="channel"
        :aliases="aliases"
        :groups="groups"
        :editable="true"
        :placeholder="'管理者'"
        v-model="reception.adminNames"
        />
    </div>

    <div>
      <label>加入者:</label>
      <SelectAlias v-if="found"
        :channel="channel"
        :aliases="aliases"
        :groups="groups"
        :editable="true"
        :placeholder="'管理者'"
        v-model="reception.joinNames"
        />
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
        <Image v-model="item.imgPath" />
        <!-- <input v-model="item.imgPath" placeholder="画像パス" type="text" /> -->

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
          <button @click.prevent="m.paidOptions.push({ itemID: m.paidOptions.length, price: 0 })">＋有料オプション</button>
        </div>

        <!-- Free Options (2次元配列: グループごとに1つ選択) -->
        <div>
          <label>無料オプション:</label>
          <div v-for="(group, gIdx) in m.freeOptions" :key="'group-'+gIdx" style="margin-bottom:10px; border:1px dashed #aaa; padding:5px;">
            <label>無料オプショングループ {{ gIdx + 1 }}</label>
            <div v-for="(opt, idx) in group" :key="'opt-'+gIdx+'-'+idx">
              <select v-model="m.freeOptions[gIdx][idx]">
                <option disabled value="">商品</option>
                <option v-for="(item, i) in reception.itemDetails" :value="item.itemID">
                  {{ item.itemName }}
                </option>
              </select>
              <button @click.prevent="m.freeOptions[gIdx].splice(idx, 1)">−</button>
            </div>
            <button @click.prevent="m.freeOptions[gIdx].push(0)">＋選択肢</button>
            <button @click.prevent="m.freeOptions.splice(gIdx, 1)" style="margin-left:10px;">グループ削除</button>
          </div>
          <button @click.prevent="m.freeOptions.push([])" style="margin-top:10px;">＋オプショングループ</button>
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
          <button @click.prevent="m.freeMultiOptions.push(m.freeMultiOptions.length)">＋無料複数オプション</button>
        </div>

        <div>
          <label for="bookType">予約可能範囲</label>
          <select id="bookType" v-model.number="m.forBookType" >
            <option value="" >予約不可</option>
            <option value="1">予約のみ</option>
            <option value="2">予約 店頭</option>
          </select>
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

    <!-- Asks (自由回答) -->
    <div>
      <label>アンケート質問 (自由回答):</label>
      <div v-for="(ask, i) in reception.asks" :key="'ask-'+i" style="margin-bottom:10px;">
        <input v-model="ask.question" placeholder="質問文" type="text" />

        <button @click.prevent="() => {
          reception.asks.splice(i, 1);
        }" v-if="reception.asks.length > 1">−</button>
      </div>

      <button @click.prevent="() => {
        reception.asks.push({ question: '', sequence: reception.asks.length });
      }">＋質問</button>
    </div>

    <hr />

    <!-- AskChoices (単一選択) -->
    <div>
      <label>アンケート質問 (単一選択):</label>
      <div v-for="(ask, i) in reception.askChoices" :key="'choice-'+i" style="margin-bottom:10px;">
        <input v-model="ask.question" placeholder="質問文" type="text" />

        <div v-for="(choice, choiceIndex) in ask.choices" :key="'choice-'+i+'-'+choiceIndex">
          <input v-model="ask.choices[choiceIndex]" placeholder="選択肢" style="margin: 2px;" />
          <button @click.prevent="() => {
            ask.choices.splice(choiceIndex, 1);
          }"> − </button>
        </div>

        <button @click.prevent="() => {
          ask.choices.push('');
        }">＋単一選択肢</button>

        <button @click.prevent="() => {
          reception.askChoices.splice(i, 1);
        }" v-if="reception.askChoices.length > 1">− 質問削除</button>
      </div>

      <button @click.prevent="() => {
        reception.askChoices.push({ question: '', choices: [], sequence: reception.askChoices.length });
      }">＋単一選択質問</button>
    </div>

    <hr />

    <!-- AskMultiChoices (複数選択) -->
    <div>
      <label>アンケート質問 (複数選択):</label>
      <div v-for="(ask, i) in reception.askMultiChoices" :key="'multi-'+i" style="margin-bottom:10px;">
        <input v-model="ask.question" placeholder="質問文" type="text" />

        <div v-for="(choice, choiceIndex) in ask.choices" :key="'multi-choice-'+i+'-'+choiceIndex">
          <input v-model="ask.choices[choiceIndex]" placeholder="選択肢" style="margin: 2px;" />
          <button @click.prevent="() => {
            ask.choices.splice(choiceIndex, 1);
          }"> − </button>
        </div>

        <button @click.prevent="() => {
          ask.choices.push('');
        }">＋複数選択肢</button>

        <button @click.prevent="() => {
          reception.askMultiChoices.splice(i, 1);
        }" v-if="reception.askMultiChoices.length > 1">− 質問削除</button>
      </div>

      <button @click.prevent="() => {
        reception.askMultiChoices.push({ question: '', choices: [], sequence: reception.askMultiChoices.length });
      }">＋複数選択質問</button>
    </div>

    <!-- Facilities -->
    <div>
      <label>施設:</label>
      <div v-for="(f, i) in reception.facilities" :key="i">
        <input v-model="f.facilityName" placeholder="施設名" type="text" />
        <input v-model.number="f.capacity" type="number" placeholder="数" />
        <label><input type="checkbox" v-model="f.bookable" true-value="1" false-value="0" /> 予約可能</label>
        <button @click.prevent="reception.facilities.splice(i, 1)" v-if="reception.facilities.length > 1">−</button>
      </div>
      <button @click.prevent="addObjectItem('facilities', { facilityName: '', capacity: 1 })">＋施設</button>
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
        <input v-model="s.skill" placeholder="役割" type="text" /><br>
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
        aliasNames: [''], shiftStart: '', shiftEnd: '', open: 1, skill: '', fix: false
      })">＋シフト</button>
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
