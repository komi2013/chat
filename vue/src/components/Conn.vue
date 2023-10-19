<script setup>
import { ref } from 'vue'
import Form from './components/Form.vue'
const msg4 = ref('')
defineProps({
  msg1: {
    type: String,
    required: true
  },
})
var conn;
if (window["WebSocket"]) {
  conn = new WebSocket("wss://" + document.location.host + "/ws");
  conn.onclose = function (evt) {
    var item = document.createElement("div");
    item.innerHTML = "<b>Connection closed.</b>";
  };
  Form.conn = conn
  console.log(conn)
  conn.onmessage = function (evt) {
    console.log(conn)
    var messages = evt.data.split('\n');
    msg4.value = msg4.value + evt.data;
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