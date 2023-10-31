// export const PublicKeyGet = () => {
export function init() {
  const param = {
    test1: 'POST',
    test2: 'hi',
  }
  const request = new Request('/Init/', {
    method: 'POST',
    body: param,
  });
  fetch(request)
    .then((response) => response.json())
    // .then((info) => console.log(info))
}