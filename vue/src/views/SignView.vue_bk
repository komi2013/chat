<script setup>
import { ref } from 'vue'
import Drawer from '../components/Drawer.vue'

import {subscriptionRegister} from '../my/subscribe.js'

// const count = ref(0)
const userID = ref('')
const alias = ref('')

const props = defineProps({
  channels: '',
})

function setCookie() {
  const fd = new FormData()
  fd.append('userID', userID.value);
  const request = new Request('/TmpLogin/', {
      method: 'POST',
      body: fd,
  });
  fetch(request)
  .then((response)=>{
    console.log(response);
    location.href = '/pushSubscription/';
  })
  .catch((reason)=>{
    console.log(reason)
  });
}

function unregister() {
  // if ('serviceWorker' in navigator) {
  //   navigator.serviceWorker.getRegistrations().then(registrations => {
  //     for (const registration of registrations) {
  //       registration.unregister();
  //     }
  //   });
  // }
}

</script>

<template>
  <Drawer />
  <div id="content">
    <br><br>
  <input v-model="userID" placeholder="seijiro" />
  <br>
  <button @click="setCookie()">setCookie</button>
  <button @click="unregister()">unregister</button>
</div>
</template>