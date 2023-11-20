<script setup>
import { ref, computed, onMounted } from 'vue'
import DrawerColumn from '../components/DrawerColumn.vue'
import { get_formated_time } from '../my/get_formated_time.js';

const props = defineProps({
  id: '',
})
console.log(props.id)
const channel = ref({
  channelID: '',
  channelName: '',
  channelDescription: '',
  updatedAt: ''
});

function getData(channelID) {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('chat',2);

    request.onerror = (event) => {
      reject(`Error opening database: ${event.target.error}`);
    };

    request.onsuccess = (event) => {
      const db = event.target.result;
      const transaction = db.transaction(['channel'], 'readonly');
      const objectStore = transaction.objectStore('channel');

      const getRequest = objectStore.get(channelID);

      getRequest.onsuccess = (event) => {
        const data = event.target.result;
        resolve(data);
      };

      getRequest.onerror = (event) => {
        reject(`Error getting data: ${event.target.error}`);
      };
    };
  });
}
async function fetchData() {
  try {
    const data = await getData(props.id);
    console.log('Data retrieved:', data);
    console.log(data.channelID);
    channel.value = data;
  } catch (error) {
    console.error(error);
    channel.value = null;
  }
}

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
  });
  fetchData();
})

const invite = () => {
  console.log('invite');
  // const fd = new FormData()
  // fd.append('csrf', '')
  // const request = new Request('/InvitationGet/', {
  //   method: 'POST',
  //   body: fd,
  // })
  // fetch(request)
  //   .then((response) => response.json())
  //   .then((json)=>{
  //     // when status not 1
  //   })
  //   .catch((reason)=>{
  //     console.log(reason)
  //   })
}

</script>



<template>
<DrawerColumn />
<div id="content">
<div class="headTitle">
  <RouterLink :to="'/channelInfo/' + channel.channelID"> {{ channel.channelName }} </RouterLink>
</div>
<div contenteditable="false">
  {{ channel.channelDescription }}
</div>
<div @click="invite"> <span>✉️</span> <span>招待URL</span> </div>
<div class="msgBox">
  <div><span>📎</span><span style="font-weight: bold;">B</span></div>
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
  .headTitle {
    margin-left: 50px;
  }
}
</style>

