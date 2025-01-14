<script setup>
import { ref, computed, onMounted } from 'vue'
import DrawerColumn from '../components/DrawerColumn.vue'

// if (!localStorage.csrf) {
//   localStorage.setItem('TO', window.location.href);
//   location.href = '/sign.html';
// }

const props = defineProps({
  channel_id: '',
  code: '',
  toAliasName: ''
})

const aliases = ref('');
const aliasName = ref(null);
const aliasImg = ref(null);
async function fetchAlias() {
  try {
    aliases.value = await getAllIDBs('alias');
  } catch (error) {
    aliases.value = [];
  }
}
// fetchAlias();

// const channel = ref('');
const channel = ref({
  channelDescription: '',
  channelName: ''
});
async function fetchChannel() {
  if (props.id) {
    try {
      channel.value = await getIDB('channel', props.id);
      aliasName.value = channel.value.aliasName;
    } catch (error) {
      console.error(error);
      channel.value = null;
    }    
  }
}

let inputFile = false;
function showInputFile () {
  inputFile = true;
}

const updatePreview = () => {
  const selectedAlias = aliases.value.find(alias => alias.aliasName === aliasName.value);
  console.log(selectedAlias);
  if (selectedAlias) {
    aliasImg.value = selectedAlias.aliasImg;
  }
};

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
      console.log(aliasImg.value);
    };
    img.src = event.target.result;
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

let aliasArray = channel.value.aliasArray || [];
const updateAliasArray = (name, image) => {
  if (image.startsWith('data:image')) {
    const randomFileName = generateRandomCode(1) + '.png';
    image = `/upload/${localStorage.userID}/${randomFileName}`;
  }
  const index = aliasArray.findIndex(entry => entry[0] === name);
  if (index !== -1) {
    aliasArray[index][1] = image;
  } else {
    aliasArray.push([name, image]);
  }
};


async function join() {
// `${window.location.origin}/communityJoin/${props.id}/${aliasName.value}/${json[1]}`;
  const fd = new FormData()
  fd.append('code', props.code);
  fd.append('aliasName', aliasName.value);
  fd.append('aliasImg', aliasImg.value);
  fd.append('userID', localStorage.userID);
  const channel = await sendRequest('/CommunityMatch/', fd);
  channel.aliasName = aliasName.value;

  console.log(channel);
  upsertIDB(channel, 'channel', 'channelID', channel.channelID)
    .catch((error) => {
      console.error(error);
    });
  const alias = {
    aliasID: channel.channelID + aliasName.value,
    channelID: channel.channelID,
    aliasName: aliasName.value,
    aliasImg: aliasImg.value,
    userID: localStorage.getItem("userID")
  }
  upsertIDB(alias, 'alias', 'aliasID', alias.aliasID)
    .catch((error) => {
      console.error(error);
    });
  // const request = new Request('/CommunityMatch/', {
  //     method: 'POST',
  //     body: fd,
  // });
  // fetch(request)
  //   .catch((reason)=>{
  //     console.log(reason)
  //   });

//     .then((response) => response.json())
//     .then((json)=>{
      
//   arr = append(arr, private.ChannelID)
//   arr = append(arr, private.Contents)
//       [
//   "7sUqxN0hDGbr",
//   "rookie",
//   "vzz6",
//   "[[\"sei1\",\"/me.jpg\",\"seijiro\"]]"
// ]
//     })
}

</script>



<template>
<DrawerColumn />

<div id="content">

<br><br>

<div class="form-container">
  <input type="text" v-model="aliasName" placeholder="グループ中の自分の名前" @input="showInputFile">
  <template v-if="inputFile">
    <img v-if="aliasImg" :src="aliasImg" class="new-alias-img"><br>
    <input type="file" @change="previewAndUpload" accept="image/*">
  </template>
  <div class="alias-list">
    <label v-for="alias in aliases" :key="alias.aliasName" 
      :class="{ 'alias-item': true, 'selected': aliasName === alias.aliasName }">
      <input type="radio" 
        v-model="aliasName"
        :value="alias.aliasName"
        class="alias-radio"
        @change="updatePreview">
      <div class="alias-info">
        <span class="alias-name">{{ alias.aliasName }}</span>
        <img :src="alias.aliasImg" alt="❌" class="alias-image">
      </div>
    </label>
  </div>
  <button @click="join">▶️</button><br>
</div>
<input type="file" style="position: fixed; left: -300px;" multiple id="fileInput">


</div>
</template>

<style>

.new-alias-img {
  max-width: 50px;
  max-height: 50px;
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

.files {
  border-top: none;
  border-right: 1px solid #d1d5db;
  border-bottom: 1px solid #d1d5db;
  border-left: 1px solid #d1d5db;
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

