<script setup>
import { RouterLink, RouterView } from 'vue-router'
import { ref, onUpdated } from 'vue'
import HelloWorld from './components/HelloWorld.vue'
import Form from './components/Form.vue'
import ChannelMenu from './components/ChannelMenu.vue'
// import ChannelMenu from './my/ChannelMenu.vue'

import {get_formated_time} from './my/get_formated_time.js'

// let initLoad = true
const count = ref(0);
// setInterval(() => {
//   count.value += 1;
// }, 1000);

let closedTime = null
// const msg4 = ref('')
const msg5 = ref('')
let data = ref(null)
let channels = ref(Object)
const props = defineProps({
  conn: {
    type: Object,
    required: true
  },
  // data: {
  //   type: String,
  //   required: true
  // },
})


  // async function loadYourData() {
  //    //load data
  //    data = 'chiane'
  // }

  // onUpdated(async () => {
  //   await loadYourData()
  //   console.log(data)
  // })

if (window["WebSocket"]) {
  // console.log(get_formated_time())
  props.conn = new WebSocket("wss://" + document.location.host + "/ws");
  props.conn.onclose = function (evt) {
    // var item = document.createElement("div");
    // item.innerHTML = "<b>Connection closed.</b>";
    closedTime = get_formated_time()
    console.log(closedTime)
  };



  props.conn.onmessage = function (evt) {
    // console.log(evt.data)
    // console.log(JSON.parse(evt.data))
    const receiveData = JSON.parse(evt.data)

    switch(receiveData[0]){
      case 'init':
        props.conn.send(JSON.stringify(['init2',closedTime]))
        data = 'loadqq'
        count.value += 10
        channels = [
          {channel_id:1,msg:data,from:'from yade'},
          {channel_id:2,msg:'msg 22',from:'from yad2'}
          ]
        break;
      case 'load':
        // props.conn.send(JSON.stringify(['init2',closedTime]))
        data = 'load'
        break;
      case 'another1':
      case 'another2':
        console.log('another2 case')
        break;
      default:
        console.log('default case')
        // this.data = 'load'
        // console.log(this.data)
        // Vue.set(this, data, 'load')
        break;
    }

    var messages = evt.data.split('\n');

    // msg4.value = msg4.value + evt.data;
    for (var i = 0; i < messages.length; i++) {
      var item = document.createElement("div");
      item.innerText = messages[i];
    }
  };
} else {
  var item = document.createElement("div");
  item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
}
</script>

<template>
  <!-- <Conn v-slot="Conn"> -->
  <header>
    <div>{{ count }} {{ data }}</div>
    <ChannelMenu :data="data" :channels="channels" />
    <Form :conn="conn" />
    <img alt="Vue logo" class="logo" src="@/assets/logo.svg" width="125" height="125" />

    <div class="wrapper">
      <HelloWorld msg="You did it!" />

      <nav>
        <RouterLink to="/">Home</RouterLink>
        <RouterLink to="/about">About</RouterLink>
      </nav>
    </div>
  </header>

  <RouterView />
  <!-- </Conn> -->
</template>

<style scoped>
header {
  line-height: 1.5;
  max-height: 100vh;
}

.logo {
  display: block;
  margin: 0 auto 2rem;
}

nav {
  width: 100%;
  font-size: 12px;
  text-align: center;
  margin-top: 2rem;
}

nav a.router-link-exact-active {
  color: var(--color-text);
}

nav a.router-link-exact-active:hover {
  background-color: transparent;
}

nav a {
  display: inline-block;
  padding: 0 1rem;
  border-left: 1px solid var(--color-border);
}

nav a:first-of-type {
  border: 0;
}

@media (min-width: 1024px) {
  header {
    display: flex;
    place-items: center;
    padding-right: calc(var(--section-gap) / 2);
  }

  .logo {
    margin: 0 2rem 0 0;
  }

  header .wrapper {
    display: flex;
    place-items: flex-start;
    flex-wrap: wrap;
  }

  nav {
    text-align: left;
    margin-left: -1rem;
    font-size: 1rem;

    padding: 1rem 0;
    margin-top: 1rem;
  }
}
</style>
