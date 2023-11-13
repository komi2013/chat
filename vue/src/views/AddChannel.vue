<script setup>
import { ref } from 'vue'
import DrawerColumn from '../components/DrawerColumn.vue'

const channelName = ref('')
const aliasName = ref('')
const channelDescription = ref('')

const userID = ref('')
const channelAdd = () => {
  console.log(userID.value)
  const fd = new FormData()
  fd.append('channelName', channelName.value)
  fd.append('aliasName', aliasName.value)
  fd.append('channelDescription', channelDescription.value)
  const request = new Request('/ChannelAdd/', {
    method: 'POST',
    body: fd,
  })
  fetch(request)
    .then((response) => response.json())
    .then((json)=>{
      // when status not 1
    })
    .catch((reason)=>{
      console.log(reason)
    })
}

</script>

<template>
<DrawerColumn />
<div id="content">
<br>
  <input type="text" v-model="channelName" placeholder="chnnnel name"><br>
  <input type="text" v-model="aliasName" placeholder="alias name"><br>
  <textarea v-model="channelDescription" placeholder="chnnnel description"></textarea><br>
  <button @click="channelAdd">▶️</button>
</div>
</template>

<style>

@media screen and (min-width : 701px) { 
/*  .msgBox {
    position: fixed;
    bottom: 0;
    width: 380px;
  }*/
}

@media screen and (max-width : 700px) {
/*  .msgBox {
    position: fixed;
    bottom: 0;
    width: 100%;
  }*/
}
</style>

