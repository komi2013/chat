<script setup>
import { ref, computed, onBeforeMount } from 'vue'

</script>

<template>
  <div id="drawer_column">
    <label for="drawer_check" class="pc_disp_none for_drawer">≡</label>
    <input id="drawer_check" type="checkbox" class="pulling pc_disp_none">
    <table id="drawer">
      <tr><td><a href="/" > 🏠 ホーム </a></td></tr>
      <tr><td><a href="/channel/" > 👪 組織・チャネル </a></td></tr>
      <tr><td><a href="/calendar/">カレンダー</a></td></tr>
      <tr><td><a href="/sign/" > 🔒 ログイン </a></td></tr>
    </table>
  </div>
</template>

<style scoped>

#drawer td {
  background-color: #EEEEEE;
}
/*#drawer td a {
  display: inline-block;
  width: 100%;
}
*/

@media screen and (min-width : 701px) {
  #drawer {
    margin-top : -1px;
    background-color: white;
  }
}

@media screen and (max-width : 700px) {
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
    height: 50px;
    width: 50px;
    opacity: 0;
    z-index: 10;
  }
  .pulling:checked ~ #drawer{
    left: 0%;
  }
  .for_drawer {
    position: absolute;
    font-size: 40px;
    top: -10px;
    width: 50px;
    text-align: center;
  }
}
</style>
