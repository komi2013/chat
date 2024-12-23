export function subscriptionRegister() {
  navigator.serviceWorker.register('/service-worker.js');
  navigator.serviceWorker.ready
    .then(function(registration) {
      return registration.pushManager.getSubscription();
    })
    .then(function(subscription) {
      if (!subscription) {
        subscribe()
      } else {
        localStorage.setItem('subscription', JSON.stringify(subscription))
      }
    });
}

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
      console.log(JSON.stringify(subscription))
      localStorage.setItem('subscription', JSON.stringify(subscription))
    })
    .catch(err => console.error(err));
}

export function urlBase64ToUint8Array(base64String) {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding)
    .replace(/\-/g, '+')
    .replace(/_/g, '/');
  const rawData = window.atob(base64);
  return Uint8Array.from([...rawData].map(char => char.charCodeAt(0)));
}


