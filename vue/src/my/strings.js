export function getSubstring(str, start, end) {
  // バイトオフセットを取得する
  function byteOffset(str, index) {
    let offset = 0;
    for (let i = 0; i < index; i++) {
      const code = str.charCodeAt(i);
      if (code >= 0x00 && code <= 0x7f) {
        offset += 1;
      } else if (code >= 0x80 && code <= 0x7ff) {
        offset += 2;
      } else if (code >= 0x800 && code <= 0xffff) {
        offset += 3;
      } else {
        offset += 4;
      }
    }
    return offset;
  }

  // バイトオフセットでサブストリングを取得する
  function byteSubstring(str, start, end) {
    let byteStart = byteOffset(str, start);
    let byteEnd = byteOffset(str, end);
    return str.substring(byteStart, byteEnd);
  }

  return byteSubstring(str, start, end);
}

export function removeHtmlTags(html) {
  return html.replace(/<[^>]*>/g, '');
}

export function getParam(name, url) {
  if (!url) url = window.location.href;
  name = name.replace(/[\[\]]/g, '\\$&');
  var regex = new RegExp('[?&]' + name + '(=([^&#]*)|&|#|$)'),
      results = regex.exec(url);
  if (!results) return null;
  if (!results[2]) return '';
  return decodeURIComponent(results[2].replace(/\+/g, ' '));
}

const chars = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz';
export function generateRandomCode(codeLength) {
  // const characters = '0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ';
  let randomCode = '';
  for (let i = 0; i < codeLength; i++) {
    const randomIndex = Math.floor(Math.random() * chars.length);
    randomCode += chars[randomIndex];
  }
  return randomCode;
}

export function base62Decode(str) {
  return str.split('').reverse().reduce((prev, curr, index) => {
    return prev + chars.indexOf(curr) * Math.pow(62, index);
  }, 0);
}