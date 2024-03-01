const textToHtml = (rawText) => {
  if(!rawText) {
    return "";
  }
  let escaped = escapeHtml(rawText);
  let formatted = escaped.replace(/\n/g, "<br>");

  // リンクの置換
  formatted = replaceLinks(formatted);

  // コードブロックの置換
  formatted = replaceCodeBlocks(formatted);

  // 引用文の置換
  formatted = replaceBlockquotes(formatted);

  // 強調の置換
  formatted = replaceEmphasis(formatted);

  return formatted;
};


function escapeHtml(html) {
  return html.replace(/[&<>"']/g, function(match) {
    return {
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      '"': '&quot;',
      "'": '&#39;'
    }[match];
  });
}

// リンクの置換処理
const replaceLinks = (text) => {
  return text.replace(/\[([^\]]+)\]\(([^\)]+)\)/g, (match, p1, p2) => {
    return `<a href="${p2}">${p1}</a>`;
  });
};

// コードブロックの置換処理
const replaceCodeBlocks = (text) => {
  return text.replace(/```(.*?)```/g, '<code>$1</code>');
};

// 引用文の置換処理
const replaceBlockquotes = (text) => {
  return text.replace(/^>(.*)$/gm, '<blockquote>$1</blockquote>');
};

// 強調の置換処理
const replaceEmphasis = (text) => {
  return text.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
};

export { textToHtml };