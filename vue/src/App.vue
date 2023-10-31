<script setup>
import { RouterLink, RouterView } from 'vue-router'
import { ref, onUpdated } from 'vue'
import HelloWorld from './components/HelloWorld.vue'
import Form from './components/Form.vue'
import DrawerColumn from './components/DrawerColumn.vue'
// import ChannelMenu from './my/ChannelMenu.vue'

import {get_formated_time} from './my/get_formated_time.js'
import {subscribe} from './my/subscribe.js'
import {init} from './my/init.js'
import {subscription_post} from './my/subscription_post.js'

// let initLoad = true
const count = ref(0);
setInterval(() => {
  count.value += 1;
}, 1000);

let closedTime = null
// const msg4 = ref('')
const msg5 = ref('')
let initData = ref(null)
let channels = ref(Object)
const props = defineProps({
  conn: {
    type: Object,
    required: true
  },

})

// initData = init()

// (async function () {
//   const initData = await init()
//   console.log(initData)
//   channels = initData[1]
// })()

const param = {
  test1: 'POST',
  test2: 'hi',
}
const request = new Request('/Init/', {
  method: 'POST',
  body: param,
});
// fetch(request)
//   .then((response) => response.json())
// const response = await fetch(request);
// const resData = await response.json();
// console.log(resData);

fetch(request)
.then((response)=>{
  if(!response.ok){
    throw new Error();
  }
  return response.json()
})
.then((json)=>{
  console.log('json')
  console.log(json[1])
  channels = json[1]
})
.catch((reason)=>{
  console.log(reason)
});


// if (initData[1] !== undefined) {
//   channels = initData[1]
// }
// console.log(count)

channels = [[1,"channel face 1","description 1"],[1,"channel init 1","description 1"]]

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
        subscription_post(JSON.stringify(subscription))
      }
    });
  const { port1, port2 } = new MessageChannel();
  port1.onmessage = ev => {
    // $span.textContent = ev.data;
    console.log(ev.data)
  };
}


</script>

<template>

  <DrawerColumn :channels="channels" />

  <div id="content" style="padding-top: 30px;">

  <div>{{ count }} {{ data }}</div>

  <Form :conn="conn" />
  <img alt="Vue logo" class="logo" src="@/assets/logo.svg" width="125" height="125" />

  <div class="wrapper">
    <HelloWorld msg="You did it!" />

    <nav>
      <RouterLink to="/">Home</RouterLink>
      <RouterLink to="/about">About</RouterLink>
      <RouterLink to="/sign">Sign</RouterLink>
    </nav>
  </div>

  <RouterView />
  </div>

</template>

<style scoped>

</style>
