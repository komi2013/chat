<script setup>
import { RouterLink, RouterView } from 'vue-router'
import { ref, onUpdated } from 'vue'
import { useMessagesStore } from './stores/messages.js'

setInterval(() => {console.log(navigator.onLine)}, 1000)

const messagesStore = useMessagesStore()
if ('serviceWorker' in navigator) {
  navigator.serviceWorker.addEventListener('message', event => {
    console.log(event.data.notificationData);
    const data = JSON.parse(event.data.notificationData);
    switch (data[0]) {
      case 'message':
        const messageData = data.slice(1)
        console.log(messageData)
        messagesStore.update(messageData)
        break
      case 'community_join':
  // arr = append(arr, channelID)
  // arr = append(arr, aliasName)
  // arr = append(arr, aliasImg)
        console.log('community_join', data)
        break
      case 'orange':
        console.log('This is an orange.')
        break
      default:
        console.log('Unknown fruit.')
    }

  })
}

</script>

<template>
  <RouterView />
</template>

<style scoped>

</style>
