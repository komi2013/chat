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
const nickname = ref(null)
const nickImg = ref(null)
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
  nicknames.value = res.nicknames
}

const coordinateInput = ref(null)
const fetched = ref(false)
const errorMessage = ref('')
const toLink = ref('')
onMounted(async () => {
  toLink.value = localStorage.getItem('TO')
  localStorage.removeItem("TO")
  await findUser()
  if (user.value && user.value.latitude) {
    coordinateInput.value = `${user.value.latitude}, ${user.value.longitude}`

  }
  if (user.value) {
    fetched.value = true
  }
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
const noticesStore = useNoticesStore()
async function submitUser(index) {
  const fd = new FormData()
  fd.append('csrf', localStorage.getItem('csrf'))
  fd.append('latitude', user.value.latitude)
  fd.append('longitude', user.value.longitude)
  fd.append('mail', user.value.mail)
  fd.append('telephone', user.value.telephone)
  fd.append('nickname', nickname.value ?? "")
  fd.append('nickImg', nickImg.value)
  const res = await sendRequest('/UserEdit/', fd)
  if (!res.csrf) errorMessage.value = res
  res.csrf && localStorage.setItem('csrf', res.csrf)
  res.pushContents.forEach(content => {
    pushReceive(content)
  })
  localStorage.setItem('nickname', res.nickname)
  noticesStore.setNotice(res.message)
}

</script>

<template>
  <Drawer />
  <div id="content">
    <div>
      <form @submit.prevent="handleSubmit" v-if="fetched">
        <h2 class="sp_head">ユーザーページ</h2>
        <div v-if="user && user.latitude && toLink">
          <br><br><a :href="toLink"> 招待参加ページ </a><br><br>
        </div>
        <div v-if="errorMessage"> <div class="errorMessage">{{errorMessage}}</div> </div>
        <label>経緯度: <a href="https://maps.google.com/" target="_blank">Googleマップ</a>の右クリックで取得できます<br />
          <input v-model="coordinateInput"
                 required pattern="^-?\d+(\.\d+)?,\s*-?\d+(\.\d+)?$"
                 title="緯度と経度は「35.77, 139.57」の形式で入力してください"
                 @input="parseCoordinates" 
                 placeholder="35.72300346964341, 139.52507136879356" 
                 class="wide-text" />
        </label>

        <div v-if="user && user.latitude" style="margin-top: 5px; font-size: 14px;">
          ➤ 緯度（latitude）: <strong>{{ user.latitude }}</strong><br />
          ➤ 経度（longitude）: <strong>{{ user.longitude }}</strong>
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

        <div>
          <input type="text" v-model="nickname" placeholder="ニックネーム" class="divText">
        </div>
        <PeopleImg v-model="nickImg" />
        <div class="centralize">
          <button type="submit" class="wide-text" @click="submitUser">更新</button>
        </div>
      </form>
      <ul>
        <li v-for="nick in nicknames">
          <img v-if="nick.nickImg && nick.nickImg.charAt(0) != ','" 
            :src="nick.nickImg" class="min-icon">
          <span v-if="nick.nickImg && nick.nickImg.charAt(0) == ','"
            :style="'background-color:' + nick.nickImg.split(',')[2] "
            class="min-icon">
              <span>{{nick.nickImg.split(',')[1]}}</span>
          </span>
          <span>{{nick.nickname}}</span>
        </li>
      </ul>
    </div>
  </div>
  <div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>
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
  max-width: 26px;
  height: 26px;
  max-height: 26px;
  border-radius: 4px;
  display: inline-flex;
  vertical-align: middle;
  justify-content: center;
  align-items: center;
}

.divText {
  margin: 4px;
  padding: 4px;
}

</style>
