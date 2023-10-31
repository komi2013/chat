export function subscription_post(jsonData) {
  const fd = new FormData()
  fd.append('json', jsonData)
  const request = new Request('/SubscriptionPost/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .then((response) => response.json())
}