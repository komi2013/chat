import { ref } from 'vue';

export const isEmojiOpen = ref(false);

export const selectedEmoji = ref('');

export const selectedMessageId = ref(null);

export const openEmoji = (messageId) => {
  selectedMessageId.value = messageId;
  isEmojiOpen.value = true;
};

export const closeEmoji = () => {
  isEmojiOpen.value = false;
  selectedMessageId.value = null;
};

export const selectEmoji = (emoji) => {
  selectedEmoji.value = emoji;
  closeEmoji();
};

export function calcEmoji(emojis, aliasName) {
  if (!Array.isArray(emojis) || emojis.length === 0) {
    return [];
  }
  const result = emojis.reduce((acc, { aliasName: name, emoji: value }) => {
    const existingItem = acc.find(item => item.emoji === value);
    if (existingItem) {
      existingItem.count++;
      existingItem.selected = existingItem.selected || (name === aliasName);
    } else {
      acc.push({
        emoji: value,
        count: 1,
        selected: name === aliasName
      });
    }
    console.log(name, '=', aliasName)
    return acc;

  }, []);
  return result;
}

export const emojiPath = (str) => {
  const filePathRegex = /^\/[\w.-]+(\/[\w.-]+)*$/;
  return filePathRegex.test(str);
};


export const isEmojiedOpen = ref(false);
export const openEmojied = (messageId) => {
  isEmojiedOpen.value = true;
  selectedMessageId.value = messageId;
};

export const closeEmojied = () => {
  isEmojiedOpen.value = false;
  selectedMessageId.value = null;
};

export const emojiRanges = [
  [0x1F300, 0x1F5FF],
  [0x1F600, 0x1F64F],
  [0x1F680, 0x1F6FF],
  [0x1F900, 0x1F9FF],
  [0x1FA70, 0x1FAFF],
  [0x2600, 0x26FF],
  [0x2702, 0x27B0],
];

export const getRandomEmoji = () => {
  const validCodePoints = emojiRanges.flatMap(([min, max]) =>
    Array.from({ length: max - min + 1 }, (_, i) => min + i)
  );
  const randomCodePoint =
    validCodePoints[Math.floor(Math.random() * validCodePoints.length)];
  return String.fromCodePoint(randomCodePoint);
};

export const getRandomColor = () => {
  const randomValue = () => Math.floor(Math.random() * 256);
  const r = randomValue();
  const g = randomValue();
  const b = randomValue();
  return rgbToHex(r, g, b);
};

const rgbToHex = (r, g, b) => {
  return `#${((1 << 24) + (r << 16) + (g << 8) + b)
    .toString(16)
    .slice(1)}`;
};

export const isEmojiInRange = (input) => {
  const codePoint = input.codePointAt(0);
  return emojiRanges.some(([min, max]) => codePoint >= min && codePoint <= max);
};

export const validateEmoji = (input) => {
  if ( !isEmojiInRange(input) && input.length > 1 ) {
    return false;
  } else {
    return true;
  }
};

export const rotateEmoji = (emoji) => {
  masterEmojis = masterEmojis.filter(e => e !== emoji);
  masterEmojis.unshift(emoji);
  const imageEmojis = masterEmojis.filter(e => e.startsWith('/img/'));
  const normalEmojis = masterEmojis.filter(e => !e.startsWith('/img/')).slice(0, 35);
  masterEmojis = [...normalEmojis, ...imageEmojis];
  localStorage.setItem('emojis', JSON.stringify(masterEmojis));
};

export let masterEmojis = JSON.parse(localStorage.getItem('emojis')) || [
  '🙇','😁','🤔','😂','🤣','😱','😭','😅','👍','👌',
  '/img/arigatou.png','/img/kakunin.png','/img/odaijini.png','/img/soudesune.png',
  '/img/naruhodo.png','/img/shouchi.png',
  '👎','👏','💪','🤝','✅','☑️','🎉','💖','🔥','🎶',
  '😜','😋','😇','😊','😎','🥰','🤩','🚫'
];
