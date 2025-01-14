self.addEventListener('push', event => {
  console.log(`Push Received.....: "${event.data.text()}"`);

  const title = 'Test Webpush';
  const options = {
    body: event.data.text(),
  };

  // Send a message to the client (main page)
  self.clients.matchAll().then(clients => {
    clients.forEach(client => {
      client.postMessage({
        type: 'push',
        notificationData: event.data.text(),
      });
    });
  });

  // event.waitUntil(self.registration.showNotification(title, options));
});
