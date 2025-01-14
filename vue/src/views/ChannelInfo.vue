<script setup>
import { ref, computed, onMounted } from 'vue'
import QRCode from 'qrcode';
import Quill from 'quill';
import "quill/dist/quill.snow.css";
import DrawerColumn from '../components/DrawerColumn.vue'
import { markdownToHtml, htmlToMarkdown } from '../my/markdown.js';

const props = defineProps({
  id: '',
})

let aliass;
const aliasName = ref(null);
const aliasImg = ref(null);

const channel = ref({
  channelDescription: '',
  channelName: ''
});

const channels = ref(null);
async function fetchAllChannel() {
  try {
    channels.value = await getAllIDBs('channel');
  } catch (error) {
    channels.value = [];
  }
}

async function fetchChannel() {
  if (props.id) {
    try {
      channel.value = await getIDB('channel', props.id);
      aliasName.value = channel.value.aliasName;
      for (let i = 0; i < channel.value.allAliases.length; i++) {
        if (channel.value.allAliases[i][0] == channel.value.aliasName) {
          aliasImg.value = channel.value.allAliases[i][1];
          console.log(aliasImg.value);
          if (aliasImg.value.charAt(0) === ',') {
              const parts = aliasImg.value.split(',');
              selectedEmoji.value = parts[1];
              selectedColor.value = parts[2];
          } else {
            emojiImg.value = false;
          }
        }
      }
    } catch (error) {
      console.error(error);
      channel.value = null;
    }    
  }
}

const attach = () => {
  const fileInput = document.getElementById('fileInput');
  if (fileInput) {
    fileInput.click();
  }
  fileInput.addEventListener('change', handleFileInputChange);
};
const fileInfo = ref({});
console.log(fileInfo.value);
const handleFileInputChange = (event) => {
  const files = event.target.files;
  const newFileInfo = document.createElement('div');
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    const fileContainer = document.createElement('div');
    if (file.type.startsWith('image/')) {
      const image = document.createElement('img');
      image.src = URL.createObjectURL(file);
      image.style.maxWidth = '50px';
      image.style.maxHeight = '50px';
      fileContainer.appendChild(image);
    } else {
      const fileName = document.createTextNode(file.name);
      fileContainer.appendChild(fileName);
    }
    newFileInfo.appendChild(fileContainer);
  }
  
  fileInfo.value = newFileInfo.outerHTML;
};

let clicked = false;
const channelPost = () => {
  if (clicked) {
    return;
  }
  clicked = true;
  const fd = new FormData();
  fd.append('channelID', props.id);
  fd.append('channelName', channel.value.channelName);
  fd.append('description', htmlToMarkdown(quill.root.innerHTML.replace(/\uFEFF/g, '')));
  fd.append('aliasName', aliasName.value);
  registerCombination();
  fd.append('aliasImg', aliasImg.value);
  let userIDS = [];
  if (channel.value.allAliases) {
    channel.value.allAliases.forEach(item => {
      userIDS.push(item[2]);
    });    
  }

  fd.append('userIDs', JSON.stringify(userIDS));
  const fileInput = document.getElementById('fileInput');
  if (fileInput && fileInput.files.length > 10) {
    alert('too many files');
    return;
  }
  if (fileInput && fileInput.files.length > 0) {
    for (const file of fileInput.files) {
      fd.append('files[]', file);
    }
  }

  const request = new Request('/ChannelEdit/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .then(function(response) {
      clicked = false;
    })
    .catch((reason)=>{
      alert(reason)
    })
}

const invitationCode = ref('');
const invitationQR = ref('');
const invite = async () => {
  console.log(channel.value.allAliases);
  const fd = new FormData();
  fd.append('channelID', props.id);
  fd.append('aliasName', aliasName.value);
  fd.append('channel', JSON.stringify(channel.value));
  const userIDs = channel.value.allAliases.map(entry => entry[2]);
  fd.append('userIDs', JSON.stringify(userIDs));
  try {
    const response = await fetch('/ChannelInvite/', {
      method: 'POST',
      body: fd,
    });
    const json = await response.json();
    invitationCode.value = `${window.location.origin}/communityJoin/${props.id}//${json[0]}/`;
    invitationQR.value = await QRCode.toDataURL(invitationCode.value);
  } catch (error) {
    console.error(error);
  }
};

let quill;
onMounted(() => {
  quill = new Quill('#description', {
    modules: {
      toolbar: '#toolbar',
    },
    theme: 'snow'
  });
  fetchChannel().then(() => {
    quill.root.innerHTML = markdownToHtml(channel.value.channelDescription, channel.value);
  });
  // setRandomDefaults();
  fetchAllChannel();
});

</script>



<template>
<DrawerColumn />

<div id="content">

<br><br>

<div class="form-container">
  <input type="text" v-model="channel.channelName" placeholder="グループ名" class="name">
  <div class="editLeft" id="toolbar">
    <button class="ql-bold"></button>
    <button class="ql-strike"></button>
    <button class="ql-blockquote"></button>
    <button class="ql-code-block"></button>
    <button class="ql-link"></button>
    <select class="ql-color">
      <option value="red">Red</option>
      <option value=""></option>
    </select>
  </div>
  <div id="description" markdownToHtml></div>
  <div v-if="props.id">{{aliasName}}</div>
  <div v-else>
    <div>名前の変更はできません</div>
    <input type="text" v-model="aliasName" placeholder="グループ内の名前" class="myname">
  </div>

  <button @click="channelPost">▶️</button><br>
  <button @click="invite"> <span>✉️</span> <span>招待URL</span> </button>
  <div> {{invitationCode}} </div>
  <div> <img :src="invitationQR"></div>
  <div> <a :href="'/groupAlias/' + props.id + '/'">
    <button> グループアカウント作成・編集 </button>
  </a> </div>
</div>
<input type="file" style="position: fixed; left: -300px;" multiple id="fileInput">

<ul v-if="channels">
  <li>全てのチャネルチーム一覧</li>
  <li v-for="d in channels" :key="d.channelID" class="channel_menu">
    <a :href="'/channel/' + d.channelID + '/'">{{ d.channelName }}</a>
  </li>
</ul>

</div>
</template>

<style>

.ql-snow.ql-toolbar {
  padding: 8px 0px;
}
.ql-snow.ql-toolbar .attachment {
  font-size: 12px;
  padding-top: 0px;
}

.myname {
  width: 100%;
  padding: 10px;
  margin-bottom: 10px;
  border: 1px solid #ccc;
  border-radius: 5px;
  box-sizing: border-box;
}

button {
  display: block;
  width: 100%;
  padding: 10px;
  border: none;
  border-radius: 5px;
  background-color: #007bff;
  color: #fff;
  cursor: pointer;
  transition: background-color 0.3s;
}

button:hover {
  background-color: #0056b3;
}

</style>

