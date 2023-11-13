<script setup>
import { ref } from 'vue'
import DrawerColumn from '../components/DrawerColumn.vue'
const props = defineProps({
  msg: ref(''),
  channels: {
    type: Object,
    required: true
  },
})
props.msg = 
  ["message", "id2", "alias A", "/me.jpg", "09:30", 0, "message", "", ["aliasB","aliasC"]]
  
const userID = ref('');
const onClick = () => { console.log(userID.value) };

</script>

<template>
<DrawerColumn />
<div id="content">
<br>
<div>
  <table>
    <tr>
      <td rowspan="2" class="icon_td"><img v-if="msg[2]" :src="msg[2]" class="icon"></td>
      <td><span class="alias">{{msg[1]}}</span><span class="time">{{msg[3]}}</span></td>
      <td class="setting">
<!--         <span class="emoji">😄</span>
        <span class="reply">💬</span>
        <span class="others">⋮</span> -->
      </td>
    </tr>
    <tr><td colspan="2" class="msg">{{msg[5]}}</td></tr>
  </table>
</div>
<textarea class="msgBox"></textarea>
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

