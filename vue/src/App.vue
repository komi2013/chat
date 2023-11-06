<script setup>
import { RouterLink, RouterView } from 'vue-router'
import { ref, onUpdated } from 'vue'
import HelloWorld from './components/HelloWorld.vue'
import Form from './components/Form.vue'
import DrawerColumn from './components/DrawerColumn.vue'
// import ChannelMenu from './my/ChannelMenu.vue'

import {get_formated_time} from './my/get_formated_time.js'
// import {subscribe} from './my/subscribe.js'
import {init} from './my/init.js'
// import { useCounterStore } from './my/store';
// const store = useCounterStore();
// const count = store.count;

// const updateCount = (newCount) => {
//   count.value = newCount;
// };

let closedTime = null
// const msg4 = ref('')
const msg5 = ref('')
let initData = ref(null)

const props = defineProps({
  conn: {
    type: Object,
    required: true
  },
  message: ref('')
})


setInterval(() => {console.log(navigator.onLine)}, 1000)
const msgs = ref('')
const channels = ref([[1,"channel face 1","description 1"],[1,"channel init 1","description 1"]])
const param = {
  test1: 'POST',
  test2: 'hi',
}
const request = new Request('/Init/', {
  method: 'POST',
  body: param,
});
fetch(request)
.then((response)=>{
  if(!response.ok){
    throw new Error();
  }
  return response.json()
})
.then((json)=>{
  channels.value = json[1]
  msgs.value = json[2]
})
.catch((reason)=>{
  console.log(reason)
});

if ('serviceWorker' in navigator) {
  const { port1, port2 } = new MessageChannel();
  port1.onmessage = ev => {
    // $span.textContent = ev.data;
    console.log(ev.data)
  };
}

const messageFromHome = ref('');

const handleCustomEvent = (message) => {
  messageFromHome.value = message;
};


</script>

<template>

  <DrawerColumn :channels="channels" />
  
  <div id="content">
  <div>
    <router-view @custom-event="handleCustomEvent" />
  </div>
    <RouterView />
  </div>
      <RouterLink to="/">Home</RouterLink>
      <RouterLink to="/about">About</RouterLink>
      <RouterLink to="/sign" >Sign</RouterLink>
      <RouterLink to="/channel/" >channel q</RouterLink>
</template>

<style scoped>

</style>
