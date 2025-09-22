<script setup>
import { ref, defineProps, onMounted } from 'vue'

import Advertisement from '@/components/Advertisement.vue'

const props = defineProps({
  id: String, // Reception ID
  code: String,
})

onMounted(async () => {

})

</script>

<template>
  <div id="drawer_column">
    <label for="drawer_check" class="pc_disp_none for_drawer">≡</label>
    <input id="drawer_check" type="checkbox" class="pulling pc_disp_none">
    <table id="drawer">
      <tr><td><a href="/" > ホーム </a></td></tr>
      <tr><td style="text-align: center;"> <Advertisement /> </td></tr>
      <tr><td><a :href="`/receptionMenu/${id}/${code}/`" > メニュー一覧 </a></td></tr>
      <tr><td><a :href="`/receptionOrder/${id}/${code}/`" > 注文履歴 </a></td></tr>
    </table>
  </div>
</template>

<style scoped>

#drawer td {
  background-color: #EEEEEE;
}

</style>
