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

