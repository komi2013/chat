import DOMPurify from 'dompurify';
import { marked } from 'marked';
import TurndownService from 'turndown';

const turndown = new TurndownService({
  headingStyle: 'atx',
  codeBlockStyle: 'fenced',
  bulletListMarker: '-',
});

// Keep the app's mention and red-text markers while writing standard Markdown
// for the formats that Markdown already represents.
turndown.addRule('chatMentionLink', {
  filter: (node) => node.nodeName === 'A' && !!node.querySelector('.mentioned'),
  replacement: (_content, node) => {
    const name = node.querySelector('.mentioned').textContent.replace(/^@/, '');
    return `＠＠${name}・＠＠`;
  },
});
turndown.addRule('chatMention', {
  filter: (node) => node.nodeName === 'SPAN' && node.classList.contains('mentioned'),
  replacement: (_content, node) => {
    const name = node.textContent.replace(/^@/, '');
    return `＠＠${name}・＠＠`;
  },
});
turndown.addRule('redText', {
  filter: (node) => node.nodeName === 'SPAN' && (node.classList.contains('markdown-red') || node.getAttribute('style')?.toLowerCase().includes('color: red')),
  replacement: (content) => `色＊赤${content}赤＊色`,
});

/** Convert Quill HTML to the persisted chat format. */
export const htmlToMarkdown = (html) => turndown.turndown(html || '');

/** Render both the persisted full-width syntax and standard Markdown safely. */
export const markdownToHtml = (source, channel) => {
  if (!source) return '';
  const markdown = decodeLegacyFormat(String(source), channel);
  const rendered = marked.parse(markdown, { gfm: true, breaks: true });
  return DOMPurify.sanitize(rendered, {
    ADD_ATTR: ['download', 'target'],
    ALLOWED_URI_REGEXP: /^(?:(?:https?|mailto|tel):|\/(?!\/)|#)/i,
    FORBID_TAGS: ['style', 'svg', 'math', 'iframe', 'object', 'embed'],
    FORBID_ATTR: ['style'],
  });
};

const escapeMarkdownText = (value) => value.replace(/([\\`*_{}\[\]()#+.!|>~-])/g, '\\$1');
const markdownUrl = (value) => value.replace(/[\\()< >\t\r\n]/g, (character) => encodeURIComponent(character));

function decodeLegacyFormat(source, channel) {
  const savedChannelID = typeof localStorage === 'undefined' ? '' : localStorage.getItem('channelID') || '';
  const channelID = typeof channel === 'string' ? channel : channel?.channelID || savedChannelID;
  let markdown = source;

  // Attachments are rendered as ordinary Markdown links/images. DOMPurify
  // enforces the URL protocol allowlist after marked has generated HTML.
  markdown = markdown.replace(/＊img＊([\s\S]*?)・＊img＊/g, (_match, url) => `![Image](${markdownUrl(url.trim())})`);
  markdown = markdown.replace(/＊f＊([\s\S]*?)・＊f＊/g, (_match, url) => `[Download file](${markdownUrl(url.trim())})`);
  markdown = markdown.replace(/「([\s\S]*?)」（([\s\S]*?)）/g, (_match, label, url) => `[${escapeMarkdownText(label)}](${markdownUrl(url)})`);
  markdown = markdown.replace(/＠＠([\s\S]*?)・＠＠/g, (_match, name) => {
    const href = `/people/${encodeURIComponent(channelID)}/${encodeURIComponent(name)}/`;
    return `<a href="${href}"><span class="mentioned" contenteditable="false">@${name.replace(/[&<>"']/g, '')}</span></a>`;
  });
  markdown = markdown.replace(/色＊赤([\s\S]*?)赤＊色/g, '<span class="markdown-red">$1</span>');
  markdown = markdown.replace(/＊太＊([\s\S]*?)・＊太＊/g, '**$1**');
  markdown = markdown.replace(/〜〜([\s\S]*?)・〜〜/g, '~~$1~~');
  markdown = markdown.replace(/＜quote＞([\s\S]*?)＜・quote＞/g, (_match, quote) => quote.split('\n').map((line) => `> ${line}`).join('\n'));
  markdown = markdown.replace(/｀｀｀([\s\S]*?)・｀｀｀/g, (_match, code) => `\n\`\`\`\n${code}\n\`\`\`\n`);
  markdown = markdown.replace(/＊p＊([\s\S]*?)・＊p＊/g, '\n$1\n');
  return markdown;
}

export const removeMark = (value) => {
  return String(value || '')
    .replace(/<[A-Za-z!/][^>]*>/g, '')
    .replace(/＊(?:太|p)＊|・＊(?:太|p)＊/g, '')
    .replace(/〜〜|・〜〜|＜quote＞|＜・quote＞|｀｀｀|・｀｀｀|色＊赤|赤＊色/g, '')
    .replace(/＠＠|・＠＠/g, '')
    .replace(/「([\s\S]*?)」（[\s\S]*?）/g, '$1')
    .replace(/!?(\[)([^\]]*)\]\([^)]*\)/g, '$2')
    .replace(/[`*_~>#]/g, '')
    .replace(/\r?\n/g, ' ')
    .trim();
};
