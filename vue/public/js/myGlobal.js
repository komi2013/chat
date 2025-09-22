function timeFormat(
  _fmt = 'YYYY/MM/DD hh:mm:ss',
  str
) {
  const _dt = str ? new Date(str) : new Date();
  const daysOfWeek = ['日', '月', '火', '水', '木', '金', '土'];

  return [
    ['YYYY', _dt.getFullYear()],
    ['MM', _dt.getMonth() + 1],
    ['DD', _dt.getDate()],
    ['hh', _dt.getHours()],
    ['mm', _dt.getMinutes()],
    ['ss', _dt.getSeconds()],
    ['iii', _dt.getMilliseconds()],
    ['WWW', daysOfWeek[_dt.getDay()]], // 📌 曜日（ゼロ埋め不要）
  ].reduce((s, a) =>
    s.replace(a[0], a[0] === 'WWW' ? a[1] : `${a[1]}`.padStart(a[0].length, '0'))
  , _fmt);
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

function base62Encode(num) {
  let encoded = '';
  const base = constantStringCharacters.length;

  while (num > 0) {
    encoded = constantStringCharacters[num % base] + encoded;
    num = Math.floor(num / base);
  }

  return encoded || '0';  // numが0の場合は'0'を返す
}

function base62Decode(str) {
  return str.split('').reverse().reduce((prev, curr, index) => {
    return prev + constantStringCharacters.indexOf(curr) * Math.pow(62, index);
  }, 0);
}

function incrementBase62Smart(str) {
  if (!str) return '0'
  const charIndex = {};
  for (let i = 0; i < constantStringCharacters.length; i++) {
    charIndex[constantStringCharacters[i]] = i;
  }

  let runes = str.split("");
  let n = runes.length;
  let carry = true;

  // カウントアップ処理（末尾から）
  for (let i = n - 1; i >= 0 && carry; i--) {
    let index = charIndex[runes[i]];
    if (index < 61) {
      runes[i] = constantStringCharacters[index + 1];
      carry = false;
    } else {
      runes[i] = constantStringCharacters[0]; // 'z' -> '0'
    }
  }

  // すべて繰り上がった場合
  if (carry) {
    return "0".repeat(n + 1);
  }

  return runes.join("");
}

function urlBase64ToUint8Array(base64String) {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding)
    .replace(/\-/g, '+')
    .replace(/_/g, '/');
  const rawData = window.atob(base64);
  return Uint8Array.from([...rawData].map(char => char.charCodeAt(0)));
}

function createGetParams(params) {
  const queryString = Object.keys(params).map(key =>
    `${encodeURIComponent(key)}=${encodeURIComponent(params[key])}`
    ).join('&');
  return queryString;
}

async function sendRequest(uri, fd) {
  const request = new Request(uri, {
    method: 'POST',
    body: fd,
  });
  try {
    const response = await fetch(request)
    if (response.ok) {
      const data = await response.json()
      if (data && Object.keys(data).length > 0) {
        return data; // データが存在する場合は data を返す
      } else {
        console.error('no data')
        return response.text()
      }
    } else {
      console.error('response not ok', response.status)
      return response.text()
    }
  } catch (error) {
    console.error('Error fetch data:', error)
    return response.text()
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

const checkFaviconBadge = () => {
  if (localStorage.getItem('favicon')) {
    const favicon = document.querySelector('link[rel="icon"]')
    favicon.href = '/img/faviconAttention.png'
  }
}

const addFaviconBadge = () => {
  localStorage.setItem('favicon', "1")
  const favicon = document.querySelector('link[rel="icon"]')
  favicon.href = '/img/faviconAttention.png'
}

const revertFaviconBadge = () => {
  localStorage.removeItem("favicon")
  const favicon = document.querySelector('link[rel="icon"]')
  favicon.href = '/favicon.ico'
}

function sendErrorLog(payload) {
  try {
    // const LOG_API = '/LogFromJS/';
    fetch('/LogFromJS/', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        ...payload,
        url: window.location.href,
        csrf: localStorage.getItem('csrf')
      })
    }).catch(() => {
      // 送信失敗時は無限ループ防止のため何もせず握りつぶす
      console.warn('sendErrorLog: ログ送信に失敗（再送しません）');
    });
  } catch (e) {
    console.warn('sendErrorLog: 実行中に例外発生', e);
  }
}

