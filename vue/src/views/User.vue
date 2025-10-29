<script setup>
import { ref, onMounted } from 'vue'

import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue';
import NoticePopup from '@/components/NoticePopup.vue';
import PeopleImg from '@/components/PeopleImg.vue';

import { pushReceive } from '@/pushReceive/pushReceive.js';
import { useNoticesStore } from '@/stores/notices.js';

document.title = 'ユーザー設定'

const user = ref(null)
const nicknames = ref([])
const activeIndex = ref(0)
const nickname = ref(localStorage.getItem('nickname'))
const nickImg = ref('')
const nickBio = ref('')
const coordinateInput = ref(null)
const fetched = ref(false)
const errorMessage = ref('')
const noticesStore = useNoticesStore()
const nicknameForm = ref(true)

async function findUser() {
  const fd = new FormData()
  fd.append('csrf', localStorage.getItem('csrf'))
  const res = await sendRequest('/UserGet/', fd)
  if (!res.csrf) errorMessage.value = res
  res.csrf && localStorage.setItem('csrf', res.csrf)
  res.pushContents.forEach(content => {
    pushReceive(content)
  })
  user.value = res.user

  const fetchedNicks = res.nicknames || []
  fetchedNicks.forEach((n, i) => {
    nicknames.value.push({
      nickname: n.nickname || '',
      nickImg: n.nickImg || '',
      nickBio: n.nickBio || ''   // ← bio も保持
    })
    if (i === 2) nicknameForm.value = false
  })

  // ✅ ローカルストレージのnicknameに一致するデータを反映
  const storedName = localStorage.getItem('nickname')
  const found = nicknames.value.find(n => n.nickname === storedName)

  if (found) {
    nickname.value = found.nickname
    nickImg.value = found.nickImg
    nickBio.value = found.nickBio || ''
  } else if (nicknames.value.length > 0) {
    // 一致しない場合は最初のデータを表示
    nickname.value = nicknames.value[0].nickname
    nickImg.value = nicknames.value[0].nickImg
    nickBio.value = nicknames.value[0].nickBio || ''
  }

  console.log('nicknames', nicknames.value)
}

const toLink = ref('')
const isTO = ref(false)
onMounted(async () => {
  await findUser()
  if (user.value && user.value.latitude) {
    coordinateInput.value = `${user.value.latitude}, ${user.value.longitude}`
  }
  if (localStorage.getItem('TO')) {
    isTO.value = true
    if (user.value && user.value.latitude) {
      toLink.value = localStorage.getItem('TO')
      localStorage.removeItem("TO")      
    }
  }
  if (user.value) fetched.value = true
})

function parseCoordinates() {
  if (!coordinateInput.value) return
  const parts = coordinateInput.value.split(',').map(s => s.trim())
  if (parts.length !== 2) {
    user.value.latitude = ''
    user.value.longitude = ''
    return
  }

  const lat = Number(parts[0])
  const lng = Number(parts[1])
  if (!isNaN(lat) && !isNaN(lng)) {
    user.value.latitude = lat.toFixed(2)
    user.value.longitude = lng.toFixed(2)
  } else {
    user.value.latitude = ''
    user.value.longitude = ''
  }
}

async function submitUser(index) {
  const fd = new FormData()
  fd.append('csrf', localStorage.getItem('csrf'))
  fd.append('latitude', user.value.latitude)
  fd.append('longitude', user.value.longitude)
  fd.append('mail', user.value.mail)
  fd.append('telephone', user.value.telephone)
  fd.append('nickname', nickname.value ?? "")
  fd.append('nickImg', nickImg.value)
  fd.append('nickBio', nickBio.value)
  const res = await sendRequest('/UserEdit/', fd)
  if (!res.csrf) errorMessage.value = res
  res.csrf && localStorage.setItem('csrf', res.csrf)
  res.pushContents.forEach(content => {
    pushReceive(content)
  })
  localStorage.setItem('nickname', res.nickname)
  toLink.value = localStorage.getItem('TO')
  localStorage.removeItem("TO")    
  noticesStore.setNotice(res.message)
}

async function switchNickname(selectedName) {
  const selected = nicknames.value.find(n => n.nickname === selectedName)
  if (!selected) return
  nickname.value = selected.nickname
  nickImg.value = selected.nickImg
  nickBio.value = selected.nickBio
}

</script>

<template>
  <Drawer />
  <div id="content">
    <div>
      <form v-if="fetched">
        <h2 class="sp_head">ユーザーページ</h2>
        <div v-if="isTO">
          <br><br><a :href="toLink"> 招待参加ページ </a><br>
          <span>必須項目を登録してから参加ページにお願いします</span>
          <br>
        </div>
        <div v-if="errorMessage"> 
          <div class="errorMessage">{{ errorMessage }}</div> 
        </div>

        <label>経緯度: <a href="https://maps.google.com/" target="_blank">Googleマップ</a>の右クリックで取得できます<br />
          <input v-model="coordinateInput"
                 required pattern="^-?\d+(\.\d+)?,\s*-?\d+(\.\d+)?$"
                 title="緯度と経度は「35.77, 139.57」の形式で入力してください"
                 @input="parseCoordinates" 
                 placeholder="35.72300346964341, 139.52507136879356" 
                 class="wide-text" />
        </label>

        <div v-if="user && user.latitude" style="margin-top: 5px; font-size: 14px;">
          ➤ 緯度: <strong>{{ user.latitude }}</strong><br />
          ➤ 経度: <strong>{{ user.longitude }}</strong>
        </div>

        <div class="centralize">
          <span>　ーーー　オプション　ーーー　</span>
        </div>

        <div>
          <input type="text" v-model="user.mail" placeholder="メール" class="divText">
        </div>

        <div>
          <input type="text" v-model="user.telephone" placeholder="電話番号" class="divText">
        </div>

        <div class="nickname-editor">
          <input type="text" v-model="nickname" placeholder="ニックネーム" class="divText">
          <PeopleImg v-model="nickImg" />
        </div>

        <textarea v-model="nickBio"></textarea>

        <div class="centralize">
          <button type="button" class="wide-text" @click="submitUser(activeIndex)">このニックネームを更新</button>
        </div>
      </form>

      <div v-for="(nick, i) in nicknames" :key="i">
        <span v-if="nick.nickname !== nickname" class="select-name" @click="switchNickname(nick.nickname)" >
          ⬜
        </span>
        <span v-if="nick.nickname == nickname" class="select-name" >✅</span>
        <img v-if="nick.nickImg && nick.nickImg.charAt(0) != ','" 
             :src="nick.nickImg" class="min-icon">
        <span v-else-if="nick.nickImg && nick.nickImg.charAt(0) == ','"
              :style="'background-color:' + nick.nickImg.split(',')[2]"
              class="min-icon">
          <span>{{ nick.nickImg.split(',')[1] }}</span>
        </span>
        <span>{{ nick.nickname }}</span>
        <!-- ✅ 現在のnickname以外だけに◯を表示 -->
      </div>

    </div>
  </div>

  <div id="ad_right">
    <Advertisement /> <Advertisement /> <Advertisement />
  </div>

  <NoticePopup />
</template>

<style scoped>
.wide-text {
  width: 90%;
  max-width: 400px;
  margin: 4px;
  padding: 4px;
}
.centralize {
  text-align: center;
  width: 100%;
}
.min-icon {
  width: 26px;
  height: 26px;
  border-radius: 4px;
  display: inline-flex;
  justify-content: center;
  align-items: center;
}
.divText {
  margin: 4px;
  padding: 4px;
}
.nickname-tabs {
  display: flex;
  justify-content: center;
  margin: 8px 0;
}
.nickname-tab {
  margin: 0 4px;
  padding: 6px 10px;
  border: 1px solid #aaa;
  background: #f8f8f8;
  border-radius: 4px;
  cursor: pointer;
}
.nickname-tab.active {
  background: #cde7ff;
  border-color: #409eff;
  font-weight: bold;
}
.nickname-editor {
  border: 1px solid #ddd;
  padding: 10px;
  margin-top: 8px;
  border-radius: 6px;
}

.select-name {
  margin: 6px;
}
</style>
