<script setup>
import { ref } from 'vue'
import DrawerColumn from '../components/DrawerColumn.vue'
import {subscription_post} from '../my/subscription_post.js'
import { useChannelsStore } from '../stores/channels.js'
import { useMessagesStore } from '../stores/messages.js'
import {subscriptionRegister} from '../my/subscribe.js'
import { upsertData } from '../my/indexDB.js'


</script>

<template>
  <DrawerColumn />
<div id="content">
<br><br><br><br><br>hi
    <script src="https://accounts.google.com/gsi/client" async></script>
    <script>
      function handleCredentialResponse(response) {
        console.log("Encoded JWT ID token: " + response.credential);
      }
      window.onload = function () {
        google.accounts.id.initialize({
          client_id: "163867041258-k0h1g9s3qdt5b2qbo3ajlmh7fah8citl.apps.googleusercontent.com"
          callback: handleCredentialResponse
        });
        google.accounts.id.renderButton(
          document.getElementById("buttonDiv"),
          { theme: "outline", size: "large" }  // customization attributes
        );
        google.accounts.id.prompt(); // also display the One Tap dialog
      }
    </script>
    <div id="buttonDiv"></div>

</div>
</template>