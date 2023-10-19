<script setup>
const props = defineProps({
  data: String,
  channels: {
    type: Object,
    required: true
  },
})
props.channels = [
  {channel_id:1,msg:props.data,from:'from yade'},
  {channel_id:2,msg:'msg 22',from:'from yad2'}
  ]
// console.log('test', props.data)
// props.conn.onmessage = function (evt) {
//   console.log(evt.data)
//   var messages = evt.data.split('\n');
//   // msg4.value = msg4.value + evt.data;
//   // for (var i = 0; i < messages.length; i++) {
//   //   var item = document.createElement("div");
//   //   item.innerText = messages[i];
//   //   // appendLog(item);
//   // }
// };
</script>

<template>
  <div class="channel_menu" v-for="d in channels">
    channel_id：{{d.channel_id}} msg {{d.msg}}
  </div>
  <div>{{data}}</div>
</template>

<style scoped>

</style>
