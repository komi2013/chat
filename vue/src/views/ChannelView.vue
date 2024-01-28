<script setup>
import { ref, computed, onMounted } from 'vue'
import DrawerColumn from '../components/DrawerColumn.vue'
import { useMessagesStore } from '../stores/messages.js';
import { useChannelsStore } from '../stores/channels.js';
import { get_formated_time } from '../my/get_formated_time.js';
import { getIDB, getIDBs, upsertData } from '../my/indexDB.js';

const props = defineProps({
  id: '',
})

const channel = ref({
  channelID: '',
  channelName: '',
  channelDescription: '',
  updatedAt: ''
});
const messagesStore = useMessagesStore()
const messages = computed(() => {
  console.log(messagesStore.messages);
  return messagesStore.messages
})

async function fetchData() {
  try {
    const data = await getIDB('channel', props.id);
    console.log('Data retrieved:', data);
    console.log(data.channelID);
    channel.value = data;
  } catch (error) {
    console.error(error);
    channel.value = null;
  }
}

// const obj = {
//   messageID: 'A3',
//   channelID: 'fakeIDa',
//   messageTxt: 'oiiii',
//   messageType: 0,
//   aliasName: 'tekiotu',
//   aliasImg: 'no_img.png',
//   editFlg: 0,
//   parentID: '',
//   emojis: '',
//   createdAt: '2023-10-01'
// };
// upsertData(obj, 'message', 'messageID', 'A3')
//   .then((message) => {
//     console.log(message);  // 成功時のメッセージをログに表示
//   })
//   .catch((error) => {
//     console.error(error);  // エラー時のメッセージをログに表示
//   });


const fetchMessageData = async () => {
  try {
    const data = await getIDBs('message', 'channelIDIndex', props.id);
    console.log('IDB Data retrieved:', data);
    messages.value = data;
  } catch (error) {
    console.error(error);
    messages.value = null;
  }
};


const pushAction = () => {
  const fd = new FormData();
  fd.append('channelID', props.id);
  fd.append('messageTxt', document.getElementById("msgText").value);
  fd.append('messageType', 1);
  fd.append('editFlg', 1);
  fd.append('aliasName', 'sei2');
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

function getMessageData(channelID) {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('chat',12);

    request.onerror = (event) => {
      reject(`Error opening database: ${event.target.error}`);
    };

    request.onsuccess = (event) => {
      const db = event.target.result;
      const transaction = db.transaction(['message'], 'readonly');
      const objectStore = transaction.objectStore('message');

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

onMounted(() => {
  msgText.value.addEventListener('input', function () {
    this.style.height = 'auto';
    this.style.height = (this.scrollHeight) + 'px';
  });
  fetchData();
  fetchMessageData();
})

</script>



<template>
<DrawerColumn />
<div id="content">
<div class="headTitle">
  <RouterLink :to="'/channelInfo/' + channel.channelID"> {{ channel.channelName }} </RouterLink>
</div>

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
</div>
<!-- <RouterLink to="/channel/abc/" >channel abc</RouterLink> -->

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

