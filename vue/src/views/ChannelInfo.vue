<script setup>
import { ref, computed, onMounted } from 'vue'

import QRCode from 'qrcode';
import Quill from 'quill';
import "quill/dist/quill.snow.css";

import DrawerColumn from '../components/DrawerColumn.vue'
import { get_formated_time } from '../my/get_formated_time.js';
import { generateRandomCode } from '../my/strings.js';
import { markdownToHtml, htmlToMarkdown } from '../my/markdown.js';

const props = defineProps({
  id: '',
})

let aliass;
const aliasName = ref(null);
const aliasImg = ref(null);

const channel = ref('');
async function fetchChannel() {
  if (props.id) {
    try {
      channel.value = await getIDB('channel', props.id);
      aliasName.value = channel.value.aliasName;
      for (let i = 0; i < channel.value.allAliases.length; i++) {
        if (channel.value.allAliases[i][0] == channel.value.aliasName) {
          aliasImg.value = channel.value.allAliases[i][1];
        }
      }
    } catch (error) {
      console.error(error);
      channel.value = null;
    }    
  }
}

// let inputFile = false;
// function showInputFile () {
//   inputFile = true;
// }

// const updatePreview = () => {
//   const selectedAlias = aliass.value.find(alias => alias.aliasName === aliasName.value);
//   console.log(selectedAlias);
//   if (selectedAlias) {
//     aliasImg.value = selectedAlias.aliasImg;
//   }
// };

const fileInputRef = ref(null);
function triggerFileInput() {
  const fileInput = fileInputRef.value;
  if (fileInput) {
    fileInput.click();
  }
}

const previewAndUpload = (event) => {
  const file = event.target.files[0];
  if (file) {
    resizeAndPreviewImage(file);
  }
};

const resizeAndPreviewImage = (file) => {
  const reader = new FileReader();
  reader.onload = (event) => {
    const img = new Image();
    img.onload = () => {
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      const MAX_SIZE = 50;
      let width = img.width;
      let height = img.height;
      if (width > height) {
        if (width > MAX_SIZE) {
          height *= MAX_SIZE / width;
          width = MAX_SIZE;
        }
      } else {
        if (height > MAX_SIZE) {
          width *= MAX_SIZE / height;
          height = MAX_SIZE;
        }
      }
      canvas.width = width;
      canvas.height = height;
      ctx.drawImage(img, 0, 0, width, height);
      const resizedDataURL = canvas.toDataURL('image/jpeg'); // 画像を縮小してDataURLに変換
      aliasImg.value = resizedDataURL;
    };
    aliasImg.value = event.target.result;
  };
  reader.readAsDataURL(file);
};

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
  fd.append('aliasImg', aliasImg.value);
  let userIDS = [];
  channel.value.allAliases.forEach(item => {
    userIDS.push(item[2]);
  });
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
  fd.append('contents', JSON.stringify(channel.value));
  const userIDs = channel.value.allAliases.map(entry => entry[2]);
  fd.append('userIDs', JSON.stringify(userIDs));
  try {
    const response = await fetch('/PrivateAdd/', {
      method: 'POST',
      body: fd,
    });
    const json = await response.json();
    invitationCode.value = `${window.location.origin}/communityJoin/${json[1]}/`;
    invitationQR.value = await QRCode.toDataURL(invitationCode.value);
  } catch (error) {
    console.error(error);
  }
};


let quill;
onMounted(() => {
  console.log('fileInputRef onMounted', fileInputRef.value);
  quill = new Quill('#description', {
    modules: {
      toolbar: '#toolbar',
    },
    theme: 'snow'
  });
  fetchChannel().then(() => {
    quill.root.innerHTML = markdownToHtml(channel.value.channelDescription, channel.value);
  });
});

</script>



<template>
<DrawerColumn />

<div id="content">

<br><br>

<div class="form-container">
  <input type="text" v-model="channel.channelName" placeholder="グループ名">
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
    <button class="attachment" @click="attach">
      🌄
    </button>
  </div>
  <div id="description" markdownToHtml></div>
  <div class="files" v-if="fileInfo[0]" v-html="fileInfo"></div>
  <br>
  <img v-if="aliasImg" :src="aliasImg" @click="triggerFileInput" class="new-alias-img">
  <span v-if="!aliasImg" @click="triggerFileInput" class="new-alias-img" > 🌄 </span>
  <input type="file" ref="fileInputRef" @change="previewAndUpload" accept="image/*" style="display: none;">
  <span> {{ aliasName }} </span>

  <button @click="channelPost">▶️</button><br>
  <button @click="invite"> <span>✉️</span> <span>招待URL</span> </button>
  <div> {{invitationCode}} </div>
  <div> <img :src="invitationQR"></div>
  <div> <a :href="'/groupAlias/' + props.id + '/'"><button> 👥 ✏️ </button></a> </div>
</div>
<input type="file" style="position: fixed; left: -300px;" multiple id="fileInput">


</div>
</template>

<style>

.new-alias-img {
  max-width: 50px;
  max-height: 50px;
  vertical-align: middle;
  font-size: 30px;
}
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


.form-container {
  max-width: 300px;
  margin: auto;
}

/*.editLeft {
  display: inline-block;
  width: 69%;
}
*/
/*.editText .ql-container.ql-snow {
  border: 1px solid #d1d5db;
  border-bottom-width: 0;
}
.editText .ql-editor {
  padding: 4px 0px;
}
*/
.files {
  border-top: none;
  border-right: 1px solid #d1d5db;
  border-bottom: 1px solid #d1d5db;
  border-left: 1px solid #d1d5db;
}
.ql-snow.ql-toolbar {
  padding: 8px 0px;
}
.ql-snow.ql-toolbar .attachment {
  font-size: 12px;
  padding-top: 0px;
}

input[type="text"] {
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

