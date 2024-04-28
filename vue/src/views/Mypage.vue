<script setup>
// import { ref, computed, onMounted } from 'vue'

// import { get_formated_time } from '../my/get_formated_time.js';
import { getIDB } from '../my/indexDB.js';
import { getSubstring, getParam } from '../my/strings.js';


function subscribe() {
  navigator.serviceWorker.ready
    .then(function(registration) {
      const vapidPublicKey = 'BIN2Jc5Vmkmy-S3AUrcMlpKxJpLeVRAfu9WBqUbJ70SJOCWGCGXKY-Xzyh7HDr6KbRDGYHjqZ06OcS3BjD7uAm8';

      return registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(vapidPublicKey),
      });
    })
    .then(function(subscription) {
      console.log(
        JSON.stringify({
          subscription: subscription,
        })
      );
      pushSubscribe(subscription);
    })
    .catch(err => console.error(err));
}

function urlBase64ToUint8Array(base64String) {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding)
    .replace(/\-/g, '+')
    .replace(/_/g, '/');
  const rawData = window.atob(base64);
  return Uint8Array.from([...rawData].map(char => char.charCodeAt(0)));
}

if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register('/service-worker.js');
  navigator.serviceWorker.ready
    .then(function(registration) {
      return registration.pushManager.getSubscription();
    })
    .then(function(subscription) {
      if (!subscription) {
        subscribe();
      } else {
        console.log(
          JSON.stringify(subscription)
        );
        pushSubscribe(subscription);
      }
    });
  // navigator.serviceWorker.addEventListener('message', event => {
  //   const notificationData = event.data.notificationData;
  //   console.log(`Received data from Service Worker: "${notificationData}"`);
  // });
}
function pushSubscribe(subscription) {
  const fd = new FormData();
  fd.append('subscription', JSON.stringify(subscription));
  const request = new Request('/PushSubscribe/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      console.log(reason)
    })
}

</script>

<template>
<div>camera</div>
</template>

<style>

</style>

