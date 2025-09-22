<template>
  <div>
    <img 
      v-if="imgSrc && imgSrc.charAt(0) !== ','" 
      :src="imgSrc" 
      @click="triggerFileInput" 
      class="new-img"
    >
    <span 
      v-else 
      @click="triggerFileInput" 
      class="new-img"
    > 🌄 </span>
    <input 
      type="file" 
      ref="fileInputRef" 
      @change="previewAndUpload" 
      accept="image/*" 
      style="display: none;"
    >
  </div>
</template>

<script setup>
import { ref } from 'vue';

const props = defineProps({
  modelValue: String,
  editable: Boolean
});

const emit = defineEmits(["update:modelValue"]);

const imgSrc = ref(getImgSrc());

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
      const MAX_SIZE = 250;
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
      imgSrc.value = resizedDataURL;
      emit("update:modelValue", imgSrc.value);
    };
    img.src = event.target.result;
  };
  reader.readAsDataURL(file);
};

function getImgSrc() {
  if (props.modelValue) {
    return props.modelValue;
  } else {
    return "";
  }
}
</script>

<style>
.new-img {
  max-height: 250px;
  max-width: 250px;
  margin: 6px;
  display: inline-block;
  font-size: 36px;
  cursor: pointer;
}
</style>
