function timeFormat(
  _fmt = 'YYYY/MM/DD hh:mm:ss',
  str
) {
  const _dt = str ? new Date(str) : new Date();
  return [
    ['YYYY', _dt.getFullYear()],
    ['MM', _dt.getMonth() + 1],
    ['DD', _dt.getDate()],
    ['hh', _dt.getHours()],
    ['mm', _dt.getMinutes()],
    ['ss', _dt.getSeconds()],
    ['iii',  _dt.getMilliseconds() ],
  ].reduce((s, a) => s.replace(a[0], `${a[1]}`.padStart(a[0].length, '0')), _fmt);
}

function getSubstring(str, start, end) {
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
  function byteSubstring(str, start, end) {
    let byteStart = byteOffset(str, start);
    let byteEnd = byteOffset(str, end);
    return str.substring(byteStart, byteEnd);
  }
  return byteSubstring(str, start, end);
}

function removeHtmlTags(html) {
  return html.replace(/<[^>]*>/g, '');
}

function getParam(name, url) {
  if (!url) url = window.location.href;
  name = name.replace(/[\[\]]/g, '\\$&');
  var regex = new RegExp('[?&]' + name + '(=([^&#]*)|&|#|$)'),
      results = regex.exec(url);
  if (!results) return null;
  if (!results[2]) return '';
  return decodeURIComponent(results[2].replace(/\+/g, ' '));
}
const constantStringCharacters = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz';
function generateRandomCode(codeLength) {
  let randomCode = '';
  for (let i = 0; i < codeLength; i++) {
    const randomIndex = Math.floor(Math.random() * constantStringCharacters.length);
    randomCode += constantStringCharacters[randomIndex];
  }
  return randomCode;
}
function base62Decode(str) {
  return str.split('').reverse().reduce((prev, curr, index) => {
    return prev + constantStringCharacters.indexOf(curr) * Math.pow(62, index);
  }, 0);
}

function urlBase64ToUint8Array(base64String) {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding)
    .replace(/\-/g, '+')
    .replace(/_/g, '/');
  const rawData = window.atob(base64);
  return Uint8Array.from([...rawData].map(char => char.charCodeAt(0)));
}
async function sendRequest(uri, fd) {
  const request = new Request(uri, {
    method: 'POST',
    body: fd,
  });
  try {
    const response = await fetch(request);

    if (response.ok) {
      const contentLength = response.headers.get("Content-Length");

      if (contentLength === "0" || contentLength === null) {
        console.log('no data');
        return null;
      } else {
        const data = await response.json();
        
        if (data && Object.keys(data).length > 0) {
          return data; // データが存在する場合は data を返す
        } else {
          console.log('データが存在しませんが、レスポンスは成功しました');
          return null;
        }
      }
    } else {
      console.error('Failed to findBookParent data', response.status);
      return null;
    }
  } catch (error) {
    console.error('Error findBookParent data:', error);
    return null;
  }
}
function editIDBLogging(title, channelID, aliasName, contents) {
  obj = {
    editLogID: generateRandomCode(8),
    title: title,
    channelID: channelID,
    aliasName: aliasName,
    contents: contents
  };
  upsertIDB(obj, 'editLog', 'editLogID', obj.editLogID)
    .catch((error) => {
      console.error(error);
    });
}
