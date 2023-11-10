import {urlBase64ToUint8Array} from './urlBase64ToUint8Array.js'
import {subscription_post} from './subscription_post.js'

export function subscribe() {
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
        JSON.stringify(subscription)
      );
      subscription_post(JSON.stringify(subscription))
    })
    .catch(err => console.error(err));
}