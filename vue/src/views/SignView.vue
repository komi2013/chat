<script setup>
import { ref } from 'vue'
import {subscribe} from '../my/subscribe.js'
import {subscription_post} from '../my/subscription_post.js'
// import DrawerColumn from '../components/DrawerColumn.vue'

const count = ref(0)
const userID = ref('')
// setInterval(() => {
//   count.value += 1;
// }, 1000);
const props = defineProps({
  channels: '',
})

// channels = [[1,"channel face 1","description 1"],[1,"channel init 1","description 1"]]

// function setCookie() {
//   console.log(userID.value)
//   const fd = new FormData()
//   fd.append('ss', userID.value)
//   const request = new Request('/SetCookie/', {
//     method: 'POST',
//     body: fd,
//   });
//   fetch(request)
//     .then((response) => response.json())
// }
function setCookie() {
  if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('service-worker.js');
    navigator.serviceWorker.ready
      .then(function(registration) {
        return registration.pushManager.getSubscription();
      })
      .then(function(subscription) {
        if (!subscription) {
          subscribe()
        } else {
          console.log(
            JSON.stringify(subscription)
          );
          subscription_post(JSON.stringify(subscription),userID.value)
        }
      });
  }
}

</script>

<template>

  <!-- <DrawerColumn :channels="channels" /> -->
  <input v-model="userID" />
  <button @click="setCookie">click</button>
  <div>{{count}}</div>
</template>