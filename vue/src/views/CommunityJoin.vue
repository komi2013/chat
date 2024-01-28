<script setup>
import { ref, computed, onMounted } from 'vue'
import DrawerColumn from '../components/DrawerColumn.vue'
import { useMessagesStore } from '../stores/messages.js';
import { useChannelsStore } from '../stores/channels.js';
import { get_formated_time } from '../my/get_formated_time.js';

const props = defineProps({
  channel_id: '',
  code: '',
})
console.log(props.channel_id)


const aliass = ref('');
const aliasName = ref(null);
const openRequest = indexedDB.open('chat', 12);

openRequest.onerror = function(event) {
  console.error("Database error: " + event.target.errorCode);
};

openRequest.onsuccess = function(event) {
  const db = event.target.result;
  const transaction = db.transaction(['alias'], 'readonly');
  const objectStore = transaction.objectStore('alias');
  const getRequest = objectStore.getAll();
  getRequest.onsuccess = function(event) {
    aliass.value = event.target.result;
  };
  getRequest.onerror = function(event) {
    console.error("Error getting data: " + event.target.errorCode);
  };
};

const join = () => {
  const fd = new FormData()
  fd.append('channelID', props.channel_id);
  fd.append('code', props.code);
  fd.append('aliasName', aliasName.value)
  const request = new Request('/CommunityMatch/', {
      method: 'POST',
      body: fd,
  });
  fetch(request)
  .then((response)=>{
    if(!response.ok){
      throw new Error();
    }
    return response.json()
  })
  .then((json)=>{
    // channels.value = json[1]
    // console.log(json[1]);
  })
  .catch((reason)=>{
    console.log(reason)
  });

}

</script>



<template>
<DrawerColumn />
<div id="content">
  <div class="alias-list">
    <label v-for="alias in aliass" :key="alias.aliasName" 
      :class="{ 'alias-item': true, 'selected': aliasName === alias.aliasName }">
      <input type="radio" v-model="aliasName" :value="alias.aliasName" class="alias-radio">
      <div class="alias-info">
        <span class="alias-name">{{ alias.aliasName }}</span>
        <img :src="alias.aliasImg" alt="❌" class="alias-image">
      </div>
    </label>
  </div>
  <div v-if="aliasName">
    Selected Alias: {{ aliasName }}
  </div>
  <button @click="join">▶️</button>
</div>
</template>

<style>
.alias-list {
  display: flex;
  flex-wrap: wrap;
}

.alias-item {
  margin: 10px;
  border: 1px solid #ccc;
  padding: 10px;
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.alias-item:hover {
  background-color: #f0f0f0;
}

.selected {
  background-color: #4CAF50; /* Add your desired background color for selected items */
  color: #fff; /* Add your desired text color for selected items */
}

.alias-radio {
  display: none; /* Hide the default radio button */
}

.alias-info {
  display: flex;
  align-items: center;
}

.alias-name {
  margin-right: 10px;
  font-weight: bold;
}

.alias-image {
  max-width: 30px;
  max-height: 30px;
  border-radius: 50%;
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
  .headTitle {
    margin-left: 50px;
  }
}
</style>

