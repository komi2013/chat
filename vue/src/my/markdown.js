export const htmlToMarkdown = (html) => {
  let markdown = html;
  // markdown = markdown.replace(/<br>/g, "\n");
  markdown = reverseEmphasis(markdown);
  markdown = reverseStrikethrough(markdown);
  markdown = reverseBlockquotes(markdown);
  markdown = reverseCodeBlocks(markdown);
  markdown = reverseLinks(markdown);
  markdown = reverseColors(markdown);
  markdown = reverseMention(markdown);
  markdown = reverseParagraph(markdown);
  return markdown;
};

const reverseEmphasis = (html) => {
  return html.replace(/<strong>([^<]+)<\/strong>/g, '＊太＊$1・＊太＊');
};

const reverseStrikethrough = (html) => {
  return html.replace(/<s>(.*?)<\/s>/g, '〜〜$1・〜〜');
};

const reverseBlockquotes = (html) => {
  return html.replace(/<blockquote>([^]*?)<\/blockquote>/g, '＜quote＞$1＜・quote＞');
};

const reverseCodeBlocks = (html) => {
  return html
    .replace(/<pre class="ql-syntax" spellcheck="false">/g, '<pre>')
    .replace(/<pre>([^]*?)<\/pre>/g, '｀｀｀$1・｀｀｀');
};

const reverseLinks = (html) => {
  return html
    .replace(/ rel="noopener noreferrer" target="_blank"/g, '')
    .replace(/<a href="([^]*?)">([^]*?)<\/a>/g, '「$2」（$1）');
};

const reverseColors = (html) => {
  return html.replace(/<span style="color: red;">(.*?)<\/span>/g, '色＊赤$1赤＊色');
};

const reverseMention = (html) => {
  const replacedHtml = html.replace(/<span class="mention" .*?>(.*?)<\/span>/g, '$1');
  return replacedHtml.replace(/<span contenteditable="false">@(.*?)<\/span>/g, '＠＠$1・＠＠');
};

const reverseParagraph = (html) => {
  return html
    .replace(/<p><br><\/p>/g, '')
    .replace(/<p>([^]*?)<\/p>/g, '＊p＊$1・＊p＊');
};

export const markdownToHtml = (markdown, channel) => {
  if (!markdown) {
    return ''; 
  }
  let html = markdown.replace(/<[^>]*>/g, '');
  html = html.replace(/\n/g, '');
  html = applyEmphasis(html);
  html = applyStrikethrough(html);
  html = applyBlockquotes(html);
  html = applyCodeBlocks(html);
  html = applyAttach(html, channel);
  html = applyLinks(html);
  html = applyColors(html);
  html = applyMention(html, channel);
  html = applyParagraphs(html);
  return html;
};

const applyEmphasis = (markdown) => {
  return markdown.replace(/＊太＊(.*?)・＊太＊/g, '<strong>$1</strong>');
};

const applyStrikethrough = (markdown) => {
  return markdown.replace(/〜〜(.*?)・〜〜/g, '<s>$1</s>');
};

const applyBlockquotes = (markdown) => {
  return markdown.replace(/＜quote＞([^]*?)＜・quote＞/g, '<blockquote>$1</blockquote>');
};

const applyCodeBlocks = (markdown) => {
  return markdown.replace(/｀｀｀([\s\S]*?)・｀｀｀/g, '<pre class="ql-syntax" spellcheck="false">$1</pre>');
};

const applyAttach = (markdown, channel) => {
  const updatedFilePath = `/upload/${channel.channelID}/${channel.aliasName}/$1`;
  return markdown.replace(/＊f＊([^]*?)・＊f＊/g, `<p><a href="${updatedFilePath}" download>$1</a></p>`);
  // return markdown.replace(/「＊([^]*?)＊」（＊([^]*?)＊）/g, '<p><a href="$2" download>$1</a></p>');
};

const applyLinks = (markdown) => {
  return markdown.replace(/「([^]*?)」（([^]*?)）/g, '<a href="$2" target="_blank">$1</a>');
};

const applyColors = (markdown) => {
  return markdown.replace(/色＊赤([^]*?)赤＊色/g, '<span style="color: red;">$1</span>');
};

const applyMention = (markdown, channel) => {
  const aliasName = channel.aliasName;
  const channelID = channel.channelID;
  const regex = /＠＠([^]*?)・＠＠/g;
  const atName = '＠＠' + aliasName + '・＠＠';
  
  if (markdown.includes(atName)) {
    markdown = markdown.replace(new RegExp(atName, 'g'), 
      `<a href="/thread/${channelID}/@${aliasName}/"><span class="mentionme" contenteditable="false">@${aliasName}</span></a>`);
  }

  return markdown.replace(regex, (match, p1) => {
    const sortedNames = [aliasName, p1].sort();
    const sortedMention = `${sortedNames[0]}@${sortedNames[1]}`;
    
    return `<a href="/thread/${channelID}/${sortedMention}/"><span class="mentioned" contenteditable="false">@${p1}</span></a>`;
  });
};

const applyParagraphs = (markdown) => {
  return markdown.replace(/＊p＊([^]*?)・＊p＊/g, '<p>$1</p>');
};

export const removeMark = (html) => {
  html = html.replace(/<[^>]*>/g, '');
  html = html.replace(/\n/g, '');
  html = removeEmphasis(html);
  html = removeStrikethrough(html);
  html = removeBlockquotes(html);
  html = removeCodeBlocks(html);
  html = removeLinks(html);
  html = removeColors(html);
  html = removeMention(html);
  html = removeParagraphs(html);
  return html;
};

const removeEmphasis = (markdown) => {
  return markdown.replace(/＊太＊(.*?)・＊太＊/g, '$1');
};

const removeStrikethrough = (markdown) => {
  return markdown.replace(/〜〜(.*?)・〜〜/g, '$1');
};

const removeBlockquotes = (markdown) => {
  return markdown.replace(/＜quote＞([^]*?)＜・quote＞/g, '$1');
};

const removeCodeBlocks = (markdown) => {
  return markdown.replace(/｀｀｀([\s\S]*?)・｀｀｀/g, '$1');
};

const removeLinks = (markdown) => {
  return markdown.replace(/「([^]*?)」（([^]*?)）/g, '$1');
};

const removeColors = (markdown) => {
  return markdown.replace(/色＊赤([^]*?)赤＊色/g, '$1');
};

const removeMention = (markdown) => {
  return markdown.replace(/＠＠([^]*?)・＠＠/g, '$1');
};

const removeParagraphs = (markdown) => {
  return markdown.replace(/＊p＊([^]*?)・＊p＊/g, '$1');
};
