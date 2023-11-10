<script setup>
import { ref, computed } from 'vue'
import { useChannelsStore } from '../stores/channels.js';
const channelsStore = useChannelsStore()
const channels = computed(() => {
  return channelsStore.channels
})


</script>

<template>
  <div id="drawer_column">
    <label for="drawer_check" class="pc_disp_none emoji" style="position: absolute;" >≡</label>
    <input id="drawer_check" type="checkbox" class="pulling pc_disp_none">
    <table id="drawer">
      <tr><td><a href="/rank/">threads</a></td></tr>
      <tr><td><RouterLink to="/sign/" >Sign</RouterLink></td></tr>
      <tr v-for="d in channels">
        <td class="channel_menu">
          <RouterLink :to="'/channel/' + d[0]">{{ d[1] }}</RouterLink>
        </td>
      </tr>
      <tr><td><a href="/htm/mydata/" rel="nofollow"><span class="emoji">📊</span> chart</a></td></tr>
      <tr><td><a href="/htm/rule/" rel="nofollow"><span class="emoji">📏</span> rule</a></td></tr>
    </table>
  </div>
</template>

<style scoped>

#drawer td {
  background-color: #EEEEEE;
}
#drawer td a {
  width: 100%;
}
@media screen and (min-width : 901px) {
  #drawer {
    margin-top : -1px;
    background-color: white;
  }
}

@media screen and (max-width : 900px) {
  #drawer {
    width: 80%;
    overflow: scroll;
    position: absolute;
    z-index: 10;
    margin: 0;
    background-color: white;
    left: -100%;
    top : 51px;
    float: left;
  }
  .pulling {
    position: absolute;
    top: 0px;
    height: 40px;
    width: 50px;
    opacity: 0;
    z-index: 10;
  }
  .pulling:checked ~ #drawer{
    left: 0%;
  }
}
</style>
