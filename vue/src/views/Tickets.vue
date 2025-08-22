<script setup>
import { ref, onMounted } from 'vue';

import Drawer from '@/components/Drawer.vue'
import Advertisement from '@/components/Advertisement.vue';

const tickets = ref([]);

document.title = 'チケット一覧'

async function fetchTickets() {
  tickets.value = await getAllIDBs('ticket');
}

onMounted(() => {
  fetchTickets();
});

</script>

<template>
<Drawer />
<div id="content">
  <h1 class="sp_head">チケット一覧</h1>

  <a href="/ticket/">作成</a>

  <ul>
    <li v-for="ticket in tickets" :key="ticket.ticketID">
      <a :href="'/ticket/' + ticket.ticketID + '/'">{{ ticket.title }}</a>
    </li>
  </ul>

  <p v-if="!tickets.length">チケットがありません</p>
</div>
<div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>
</template>

<style scoped>

ul {
  list-style-type: none;
  padding: 0;
}

li {
  margin-bottom: 10px;
}

a {
  color: #007bff;
  text-decoration: none;
}

a:hover {
  text-decoration: underline;
}
</style>
