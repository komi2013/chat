<script setup>
import { ref, onMounted } from 'vue'

import DrawerTweet from '@/components/DrawerTweet.vue'

const props = defineProps({
  name: String
})

function tF(a, b = null){ return timeFormat(a, b) }

const nickData = ref(null)
const errorMessage = ref('')
const fetched = ref(false)

async function fetchNickData() {
  const fd = new FormData()
  fd.append('nickname', props.name)
  fd.append('csrf', localStorage.getItem('csrf') || '')

  const res = await sendRequest('/NicknameGet/', fd)
  res.csrf && localStorage.setItem('csrf', res.csrf)
  if (res.error) {
    errorMessage.value = res.error
    return
  }
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  nickData.value = res.nickData
  document.title = nickData.value.nickname
  fetched.value = true
}

onMounted(async () => {
  await fetchNickData()
})

</script>

<template>
  <DrawerTweet />
  <div id="content">
    <div v-if="!fetched">Loading…</div>
    <div v-if="errorMessage">{{ errorMessage }}</div>
    <div v-if="fetched">
      <div>
        <br>
        <h2>{{ nickData.nickname }}</h2>
        <div class="icon_td">
          <img v-if="nickData.nickImg && nickData.nickImg.charAt(0) != ','" 
            :src="nickData.nickImg" class="icon-img">
          <span v-if="nickData.nickImg && nickData.nickImg.charAt(0) == ','"
            class="icon-span" 
            :style="'background-color:' + nickData.nickImg.split(',')[2] ">
            {{nickData.nickImg.split(',')[1]}}</span>
        </div>
        <p>👍: {{ nickData.good }}</p>
        <p>👎: {{ nickData.bad }}</p>
        <p>⚠️: {{ nickData.report }}</p>
        <p>自己紹介:</p>
        <p> {{ nickData.nickBio }} </p>
        <p>作成日: {{ tF('YYYY/MM/DD', nickData.createdAt.toLocaleString('ja-JP')) }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.nickname-view {
  padding: 20px;
  max-width: 600px;
  margin: auto;
}

.icon_td {
  width: 50px;
  vertical-align: top;
  text-align: center;
  display: inline-block;
}
.icon-span {
  border-radius: 10%;
  display: inline-block;
  height: 28px;
  width: 28px;
}
.icon-img {
  border-radius: 10%;
  display: inline-block;
  max-height: 30px;
  max-width: 30px;
}

</style>
