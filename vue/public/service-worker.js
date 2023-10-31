self.addEventListener('push', event => {
  console.log('[Service Worker] Push Received.');
  console.log(`[Service Worker] Push had this data: "${event.data.text()}"`);

  const title = 'Test Webpush';
  const options = {
    body: event.data.text(),
  };
  let port = event.ports[0];
  port.postMessage(event.data.text())
  event.waitUntil(self.registration.showNotification(title, options));
});
