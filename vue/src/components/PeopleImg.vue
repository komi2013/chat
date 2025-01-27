<template>
  <div>
    <span
      @click="emojiImg = true"
      :style="{ backgroundColor: !emojiImg ? 'silver' : 'transparent' }"
      class="toggleEmoji"
    >
      😃
    </span>
    <span
      @click="emojiImg = false"
      :style="{ backgroundColor: emojiImg ? 'silver' : 'transparent' }"
      class="toggleEmoji"
    >
      🌄
    </span>
  </div>
  <div v-if="emojiImg">
    <div class="emojiDisplay">
      <input 
        type="text" 
        v-model="selectedEmoji" 
        maxlength="2" 
        class="emoji-input"
        @change="validateEmoji"
      />
      <input 
        id="color-picker"
        type="color" 
        v-model="selectedColor" 
        class="emoji-input"
        @change="validateEmoji"
      />
      <br>
      <span v-if="emojiValidErr" class="emoji-valid-err">絵文字か1文字にしてください</span>
    </div>
    <div class="emojiDisplay">
      <span class="people-img" :style="'background-color:' + selectedColor ">
        <span>{{selectedEmoji}}</span>
      </span>
    </div>
  </div>
  <div v-if="!emojiImg">
    <img v-if="aliasImg && aliasImg.charAt(0) != ','" :src="aliasImg" @click="triggerFileInput" class="new-alias-img">
    <span v-else @click="triggerFileInput" class="new-alias-img" > 🌄 </span>
    <input type="file" ref="fileInputRef" @change="previewAndUpload" accept="image/*" style="display: none;">
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';

import { emojiRanges, getRandomEmoji, getRandomColor } from '@/my/emoji';

const props = defineProps({
  alias: Object,
  modelValue: String,
  editable: Boolean
});

const emojiImg = ref(!props.modelValue || props.modelValue.charAt(0) == ',');

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
      const resizedDataURL = canvas.toDataURL('image/jpeg');
      aliasImg.value = resizedDataURL;
      emit("update:modelValue", aliasImg.value);
    };
    aliasImg.value = event.target.result;
    emit("update:modelValue", aliasImg.value);
  };
  reader.readAsDataURL(file);
};

function getAliasImg() {
  if (props.modelValue) {
    const paths = props.modelValue.split(',');
    // console.log('getAliasImg', aliasImg.value);
    selectedEmoji.value = paths[1];
    selectedColor.value = paths[2];
    return props.modelValue;
  } else {
    return "," + selectedEmoji.value + "," + selectedColor.value;
  }
}

const selectedEmoji = ref(getRandomEmoji());
const selectedColor = ref(getRandomColor());
const aliasImg = ref(getAliasImg());
emit("update:modelValue", aliasImg.value);

const emit = defineEmits(["update:modelValue"]);
const emojiValidErr = ref(false);
const validateEmoji = () => {
  if ( !isEmojiInRange(selectedEmoji.value) && selectedEmoji.value.length > 1 ) {
    emojiValidErr.value = true;
    return
  }
  emojiValidErr.value = false;
  aliasImg.value = "," + selectedEmoji.value + "," + selectedColor.value;
  emit("update:modelValue", aliasImg.value);
};

</script>

<style>

.people-img {
	width: 24px;
	vertical-align: middle;
	display: inline-table;
	text-align: center;
	border-radius: 5px;
}

.toggleEmoji {
  display: inline-block;
  text-align: center;
  width: 50%;
  padding: 6px 0px 6px 0px;
}


.emojiDisplay {
  width: 44%;
  display: inline-block;
  margin: 6px;
}

.emoji-input {
  width: 40px;
  height: 40px;
  margin: 6px;
  font-size: 20px;
  text-align: center;
  border: 1px solid #ccc;
  border-radius: 4px;
}
#color-picker {
  position: absolute;
  width: 60px;
  height: 44px;
}
.emoji-valid-err {
  position: absolute;
  color: red;
  left: 10px;
}

</style>
