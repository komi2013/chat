<script setup>
import { ref, computed, onMounted } from 'vue'
import DrawerColumn from '../components/DrawerColumn.vue'
import { useMessagesStore } from '../stores/messages.js';
import { get_formated_time } from '../my/get_formated_time.js';
const messagesStore = useMessagesStore()
const messages = computed(() => {
  // messagesStore.messages.forEach(row => {
  //   row[9] = get_formated_time('hh:mm',row[9])
  // })
  return messagesStore.messages
})
const props = defineProps({
  id: '',
})

console.log(props.id)


//     arr = append(arr, r.MessageID)  0
//     arr = append(arr, r.ChannelID)  1
//     arr = append(arr, r.MessageTxt) 2
//     arr = append(arr, r.MessageType)3
//     arr = append(arr, r.From)       4
//     arr = append(arr, r.FromImg)    5
//     arr = append(arr, r.EditFlg)    6
//     arr = append(arr, r.ParentID)   7
//     arr = append(arr, r.Emojis)     8
//     arr = append(arr, r.CreatedAt)  9
const pushAction = () => {
  const fd = new FormData()
  fd.append('channelID', props.id)
  fd.append('messageTxt', document.getElementById("msgText").value)
  fd.append('messageType', 1)
  fd.append('editFlg', 1)
  const request = new Request('/MessagePost/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .then((response) => response.json())
    .then((json)=>{
      // when status not 1
    })
    .catch((reason)=>{
      console.log(reason)
    })
}
const msgText = ref(null)
onMounted(() => {
  msgText.value.addEventListener('input', function () {
    this.style.height = 'auto';
    this.style.height = (this.scrollHeight) + 'px';
  })
})
</script>

<template>
<DrawerColumn />
<div id="content">
<br>
<!--   <button @click="incrementChildCount">Increment Child Count</button>
  <div>
    <p>Count: {{ countString }}</p>
    <button @click="counterStore.adding(3)">Increment</button>
    <button @click="counterStore.decrement">Decrement</button>
  </div> -->
<div v-for="d in messages">
  <table>
    <tr>
      <td rowspan="2" class="icon_td"><img v-if="d[5]" :src="d[5]" class="icon"></td>
      <td>
        <span class="alias">{{d[4]}}</span>
        <span class="time">{{get_formated_time('hh:mm',d[9]) }}</span>
      </td>
      <td class="setting">
        <span class="emoji"> <RouterLink to="/emoji/1"> 😄 </RouterLink> </span>
        <span class="reply"> <RouterLink to="/reply/1"> 💬 </RouterLink> </span>
        <span class="others"> &nbsp; ⋮ &nbsp; </span>
      </td>
    </tr>
    <tr><td colspan="2" class="msg">{{d[2]}}</td></tr>
  </table>
<!--   <div class="box">
    <img v-if="d[2]" :src="d[2]" class="icon">
    <span class="alias">{{d[1]}}</span>
    <span class="">{{d[3]}}</span><br>
    <div >{{d[5]}}</div>
  </div> -->
</div>
<!-- <RouterLink to="/channel/abc/" >channel abc</RouterLink> -->

<div class="msgBox">
  <div><span>📎</span><span style="font: bold;">B</span></div>
  <textarea id="msgText" ref="msgText" ></textarea>
  <div style="text-align: right"><button @click="pushAction">▶️</button></div>
</div>

</div>
</template>

<style>

.icon {
  max-width: 50px;
  max-height: 50px;
}

.icon_td {
  width: 50px;
}

.setting {
  text-align: right;
}

.box {
  display: flex;
}

.alias {
  margin: 2px;
}

.time {
  margin: 2px;
}

.emoji {
  margin: 2px;
}

.reply {
  margin: 2px;
}

.others {
  margin: 2px;
}

.msgBox textarea {
  width: 100%;
  border: none;
  height: 50px;
}

@media screen and (min-width : 701px) { 
  .msgBox {
    position: fixed;
    bottom: 10px;
    width: 300px;
  }
}

@media screen and (max-width : 700px) {
  .msgBox {
    position: fixed;
    bottom: 10px;
    width: 300px;
  }
}
</style>

