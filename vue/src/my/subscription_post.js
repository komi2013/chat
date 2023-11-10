export function subscription_post(userID, alias) {
  const fd = new FormData()
  fd.append('subscription', localStorage.getItem('subscription'))
  fd.append('userID', userID)
  fd.append('alias', alias)
  const request = new Request('/SetCookie/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .then((response) => response.json())
}