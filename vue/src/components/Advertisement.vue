<script setup>
import { ref, onMounted, onUnmounted } from 'vue';

const currentAd = ref(null);

function selectValidAd() {
  if (!window.advertisements || window.advertisements.length === 0) {
      return;
  }
  const now = new Date();
  const validAds = window.advertisements.filter(ad => {
    const start = new Date(ad.adStart);
    const end = new Date(ad.adEnd);
    return start <= now && now <= end;
  });
  
  if (validAds.length > 0) {
    const randomIndex = Math.floor(Math.random() * validAds.length);
    currentAd.value = validAds[randomIndex];
  } else {
    currentAd.value = null;
  }
}

let intervalId = null;

onMounted(() => {
  selectValidAd();
  intervalId = setInterval(selectValidAd, 60 * 1000); // 毎分実行
});

onUnmounted(() => {
  clearInterval(intervalId);
});
</script>

<template>
  <div v-if="currentAd">
    <a :href="currentAd.adLink" target="_blank" rel="noopener noreferrer">
      <img :src="currentAd.pathSquare" alt="広告" class="ad-image" />
    </a>
  </div>
</template>

<style scoped>
.ad-image {
  width: 150px;
  height: 150px;
  object-fit: cover;
  border: 1px solid #ccc;
  border-radius: 8px;
  transition: transform 0.3s;
}
.ad-image:hover {
  transform: scale(1.05);
}
</style>
