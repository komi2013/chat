self.addEventListener('install', event => {
  self.skipWaiting();
});

self.addEventListener('activate', event => {
  event.waitUntil(self.clients.claim());
});

self.addEventListener('push', event => {
  const raw = event.data?.text() || "";
  console.log("SW push:", raw);

  event.waitUntil((async () => {
    const clientsArr = await self.clients.matchAll({ type: "window" });

    // Always forward to pages (keep old system working)
    clientsArr.forEach(c => {
      c.postMessage({
        type: "push",
        notificationData: raw
      });
    });

    // Try parse push JSON
    let arr;
    try { arr = JSON.parse(raw); } catch { return; }

    const sw = arr[arr.length - 1];
    if (!sw || typeof sw !== "object" || !sw.url) return;

    // Check if any visible tab exists
    const hasVisible = clientsArr.some(c => c.visibilityState === "visible");

    if (hasVisible) {
      return;
    }

    await self.registration.showNotification("Notification", {
      body: sw.message,
      icon: sw.icon,
      data: { url: sw.url },
      requireInteraction: true
    });
  })());
});

self.addEventListener("notificationclick", event => {
  const url = event.notification.data?.url;
  event.notification.close();
  if (!url) return;
  event.waitUntil(clients.openWindow(url));
});
