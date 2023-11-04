export function subscription_post(subscription, userID) {
  const fd = new FormData()
  fd.append('subscription', subscription)
  fd.append('userID', userID)
  const request = new Request('/SetCookie/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .then((response) => response.json())
}