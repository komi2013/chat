<script setup>
import { ref, onMounted, nextTick, watch } from 'vue'
import Quill from 'quill'
import 'quill/dist/quill.snow.css'

import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue';
import NoticePopup from '@/components/NoticePopup.vue';
import PeopleImg from '@/components/PeopleImg.vue';

import { htmlToMarkdown, markdownToHtml } from '@/my/markdown.js'
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
  user.value = res.user
  const fetchedNicks = res.nicknames || []
  fetchedNicks.forEach((n, i) => {
    nicknames.value.push({
      nickname: n.nickname || '',
      nickImg: n.nickImg || '',
      nickBio: n.nickBio || ''
    })
    if (i === 2) nicknameForm.value = false
  })
  const storedName = localStorage.getItem('nickname')
  const found = nicknames.value.find(n => n.nickname === storedName)
  if (found) {
    nickname.value = found.nickname
    nickImg.value = found.nickImg
    nickBio.value = found.nickBio || ''
  } else if (nicknames.value.length > 0) {
    nickname.value = nicknames.value[0].nickname
    nickImg.value = nicknames.value[0].nickImg
    nickBio.value = nicknames.value[0].nickBio || ''
  }
}

let quillBio
function initQuillBio() {
  const toolbarOptions = [
    ['bold', 'strike'],
    [{ 'list': 'ordered' }, { 'list': 'bullet' }],
    ['blockquote', 'code-block'],
    [{ 'color': [] }, { 'background': [] }]
  ]
  quillBio = new Quill('#nickBioEditor', {
    theme: 'snow',
    modules: { toolbar: toolbarOptions }
  })

  if (nickBio.value) {
    quillBio.root.innerHTML = markdownToHtml(nickBio.value)
  }
  quillBio.on('text-change', (delta, oldDelta, source) => {
    if (source !== 'user') return
    const text = quillBio.getText().trimEnd()
    if (text.length > 200) {
      quillBio.deleteText(200, text.length)
      return
    }
    const urlRegex = /(https?:\/\/[^\s]+)/g
    let match
    while ((match = urlRegex.exec(text)) !== null) {
      const url = match[0]
      const index = match.index
      const formats = quillBio.getFormat(index, url.length)
      if (!formats.link) {
        quillBio.formatText(index, url.length, 'link', url)
      }
    }
    const html = quillBio.root.innerHTML
    nickBio.value = htmlToMarkdown(html)
  })
}

watch(nickBio, (newVal) => {
  if (quillBio && newVal !== htmlToMarkdown(quillBio.root.innerHTML)) {
    quillBio.root.innerHTML = markdownToHtml(newVal || '')
  }
})

const toLink = ref('')
const isTO = ref(false)
onMounted(async () => {
  await findUser()
  if (user.value && user.value.latitude) {
    coordinateInput.value = `${user.value.latitude}, ${user.value.longitude}`
  } else {
    // await getGeolocation()
  }
  if (localStorage.getItem('TO')) {
    isTO.value = true
    if (user.value && user.value.latitude) {
      toLink.value = localStorage.getItem('TO')
      // localStorage.removeItem("TO")
    }
  }
  if (user.value) fetched.value = true
})

async function getGeolocation() {
  if (!navigator.geolocation) {
    console.warn("Geolocation is not supported");
    return;
  }

  navigator.geolocation.getCurrentPosition(
    (pos) => {
      const lat = pos.coords.latitude.toFixed(6)
      const lng = pos.coords.longitude.toFixed(6)
      coordinateInput.value = `${lat}, ${lng}`
      parseCoordinates()  // ← user.latitude / longitude に反映
    },
    (err) => {
      console.warn("Geolocation error:", err.message)
    },
    {
      enableHighAccuracy: true,
      timeout: 5000,
      maximumAge: 0
    }
  )
}


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

const nicknameFormVisible = ref(false)
const editingMode = ref('')

const editableNickname = ref('')

const MAX_NICKNAMES = 3

async function openNicknameForm(mode) {
  if (mode === 'new' && nicknames.value.length >= MAX_NICKNAMES) {
    alert('ニックネームは3つまで作成できます。')
    return
  }

  editingMode.value = mode
  nicknameFormVisible.value = true

  if (mode === 'edit') {
    editableNickname.value = nickname.value
  } else {
    editableNickname.value = ''
    nickImg.value = ''
    nickBio.value = ''
    if (quillBio) quillBio.root.innerHTML = ''
  }
  await nextTick()
  initQuillBio()
}

async function submitUser(index) {
  const fd = new FormData()
  fd.append('csrf', localStorage.getItem('csrf'))
  fd.append('latitude', user.value.latitude)
  fd.append('longitude', user.value.longitude)
  fd.append('mail', user.value.mail ?? '')
  fd.append('telephone', user.value.telephone ?? '')
  fd.append('walletAddress', user.value.walletAddress ?? '')
  fd.append('nickImg', nickImg.value)
  fd.append('nickBio', nickBio.value)
  if (nicknameFormVisible.value && !editableNickname.value) {
    alert('ニックネームの入力してください')
    return
  }
  if (editingMode.value === 'new') {
    fd.append('nickname', editableNickname.value)
  } else {
    fd.append('nickname', nickname.value ?? '')
  }

  const res = await sendRequest('/UserEdit/', fd)
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
  localStorage.setItem('nickname', res.nickname)
  nicknameFormVisible.value = false
  noticesStore.setNotice(res.message)
}

async function switchNickname(selectedName) {
  const selected = nicknames.value.find(n => n.nickname === selectedName)
  if (!selected) return
  nickname.value = selected.nickname
  nickImg.value = selected.nickImg
  nickBio.value = selected.nickBio
  if (quillBio) quillBio.root.innerHTML = markdownToHtml(selected.nickBio || '')
  nicknameFormVisible.value = false
}

</script>

<template>
  <div id="drawer_column"><Drawer /></div>
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

        <label>経緯度: 
          <input type="button" value="GEOデータ再取得" @click="getGeolocation">&nbsp;
          <a href="https://maps.google.com/" target="_blank">Googleマップ</a>の右クリックで取得できます
          <br />
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
          <input type="text" v-model="user.mail" placeholder="メール" class="wide-text">
        </div>
        <div>
          <input type="text" v-model="user.telephone" placeholder="電話番号" class="wide-text">
        </div>
        <div v-if="!nicknameFormVisible" class="centralize">
          <button v-if="nicknames.length > 0" type="button" class="wide-text" @click="openNicknameForm('edit')">
            ニックネームの絵文字、画像を変更
          </button>
          <button type="button" class="wide-text" @click="openNicknameForm('new')">
            新しいニックネームを作成
          </button>
        </div>

        <div v-if="nicknameFormVisible" class="nickname-editor">
          <h3 v-if="editingMode === 'edit'">ニックネームの編集</h3>
          <h3 v-if="editingMode === 'new'">新しいニックネームを作成</h3>
          <div>ニックネームは登録後は変更できません<br>３つまで登録できます</div>
          <input
            type="text"
            v-model="editableNickname"
            :disabled="editingMode === 'edit'"
            placeholder="ニックネーム"
            class="divText"
          />
          <PeopleImg v-model="nickImg" />
          <div class="bio-editor">
            <div id="nickBioEditor" class="quill-editor"></div>
          </div>
          <div class="centralize">
            <button type="button" class="wide-text" @click="nicknameFormVisible = false">
              キャンセル
            </button>
          </div>
        </div>

        <div>
          <input type="text" v-model="user.walletAddress" placeholder="JPYCアドレス" class="wide-text">
        </div>

        <div class="centralize">
          <button type="button" class="wide-text" @click="submitUser(activeIndex)">更新</button>
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
        <span>&nbsp;{{ nick.nickname }}</span>
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
  vertical-align: middle;
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
.quill-editor {
  border: 1px solid #ccc;
  border-radius: 6px;
  padding: 6px;
  margin: 4px;
  background: #fff;
}
.ql-toolbar.ql-snow {
  border: 1px solid #ccc;
  border-bottom: none;
}
.bio-editor {
  margin-top: 10px;
}

</style>
