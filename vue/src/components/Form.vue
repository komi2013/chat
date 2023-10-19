
<script setup>
import { ref } from 'vue'
// import { Conns } from 'Conn'
const msg3 = ref('hello')
// const msg4 = ref('')
// defineProps([conn])
const props = defineProps({
  conn: Object
})
    // var conn;
    // var msg = document.getElementById("msg");
    // var log = document.getElementById("log");

    // function appendLog(item) {
    //     var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
    //     log.appendChild(item);
    //     if (doScroll) {
    //         log.scrollTop = log.scrollHeight - log.clientHeight;
    //     }
    // }
    // props.conn.send("hiiiii");

    function onButtonClick() {
      // console.log(props)
      // console.log(props.conn)
      // console.log(msg3.value)
        if (!props.conn) {
            return false;
        }
        // if (!msg.value) {
        //     return false;
        // }
        props.conn.send(msg3.value);
        // msg.value = "";
        return false;
    }

    // if (window["WebSocket"]) {
    //     conn = new WebSocket("wss://" + document.location.host + "/ws");
    //     conn.onclose = function (evt) {
    //         var item = document.createElement("div");
    //         item.innerHTML = "<b>Connection closed.</b>";
    //         // appendLog(item);
    //     };

    // } else {
    //     var item = document.createElement("div");
    //     item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
    //     // appendLog(item);
    // }
</script>

<template>
  <!-- <div> {{msg4}} </div> -->

    <input type="submit" value="Sendko" @click="onButtonClick" />
    <input type="text" size="64" autofocus v-model="msg3" />

</template>