<script setup>
import { ref, computed } from 'vue'
import { useMessagesStore } from '../stores/messages.js';
const messagesStore = useMessagesStore()
const messages = computed(() => {
  console.log(messagesStore.messages)
  return messagesStore.messages
})
const props = defineProps({
  id: '',
})

// const id = computed(() => context.attrs.id)
console.log(props.id)

// props.msgs = [
//   ["id1", "alias A", "/me.jpg", "09:00", 0, "message text", "",  [["aliasA","🙇"]]],
//   ["id2", "alias A", "/me.jpg", "09:30", 0, "message text,message textmessage textmessage textmessage text", "", [["aliasB","🙇"]]]
//   ]

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
    // .then((json)=>{

    //   let data = ["","6545c74e71aee3b6bb941191","micro plastic ","1","sei1","/me.jpg","1",0,null,"2023-11-10T14:30:23.352583421Z"]
    //   console.log(data)
    //   console.log(json)
    //   console.log(JSON.parse(json))
    //   // messagesStore.update(JSON.parse(json))
    //   messagesStore.update(data)

    // })
    .catch((reason)=>{
      console.log(reason)
    })
}


</script>

<template>
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
      <td><span class="alias">{{d[4]}}</span><span class="time">{{d[9]}}</span></td>
      <td class="setting">
        <span class="emoji"> <RouterLink to="/emoji/1"> 😄 </RouterLink> </span>
        <span class="reply"> <RouterLink to="/reply/1"> 💬 </RouterLink> </span>
        <span class="others">⋮</span>
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
<RouterLink to="/channel/abc/" >channel abc</RouterLink>
<button @click="pushAction">pushAction</button>

<textarea class="msgBox" id="msgText"></textarea>
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

@media screen and (min-width : 701px) { 
  .msgBox {
    position: fixed;
    bottom: 0;
    width: 380px;
  }
}

@media screen and (max-width : 700px) {
  .msgBox {
    position: fixed;
    bottom: 0;
    width: 100%;
  }
}
</style>

