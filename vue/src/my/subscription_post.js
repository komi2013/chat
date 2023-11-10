export function subscription_post(subscription, userID, alias) {
  const fd = new FormData()
  fd.append('subscription', subscription)
  fd.append('userID', userID)
  fd.append('alias', alias)
  const request = new Request('/SetCookie/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .then((response) => response.json())
}