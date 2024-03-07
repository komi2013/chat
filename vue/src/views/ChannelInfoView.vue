<script setup>
import { ref, computed, onMounted } from 'vue'
import DrawerColumn from '../components/DrawerColumn.vue'
import { get_formated_time } from '../my/get_formated_time.js';
import { getIDB, getIDBs, upsertData } from '../my/indexDB.js';
import QRCode from 'qrcode';
const props = defineProps({
  id: '',
})
const channel = ref({
  channelID: '',
  channelName: '',
  channelDescription: '',
  updatedAt: ''
});
const invitationCode = ref('');
const invitationQR = ref('');

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

const invite = async () => {
  const fd = new FormData();
  fd.append('csrf', '');
  fd.append('channelID', props.id);
  try {
    const response = await fetch('/InvitationGet/', {
      method: 'POST',
      body: fd,
    });
    const json = await response.json();
    console.log(json);

    invitationCode.value = 'https://' + location.host + '/communityJoin/' + props.id + '/' + json[1];
    invitationQR.value = await QRCode.toDataURL(invitationCode.value);
  } catch (error) {
    console.error(error);
  }
};

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
<div> {{invitationCode}} </div>
<div> <img :src="invitationQR"></div>

</div>
</template>

<style>


</style>

