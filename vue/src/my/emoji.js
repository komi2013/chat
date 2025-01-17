import { ref } from 'vue';

export const isEmojiOpen = ref(false);

export const selectedEmoji = ref('');

export const selectedMessageId = ref(null);

const openEmoji = (messageId) => {
  selectedMessageId.value = messageId;
  isEmojiOpen.value = true;
};

const closeEmoji = () => {
  isEmojiOpen.value = false;
  selectedMessageId.value = null;
};

const selectEmoji = (emoji) => {
  selectedEmoji.value = emoji;
  console.log(selectedEmoji.value);
  closeEmoji();
};

function calcEmoji(emojis, aliasName) {
  if (!emojis) {
    return [];
  }
  const result = emojis.reduce((acc, [name, value]) => {
    const existingItem = acc.find(item => item[0] === value);

    if (existingItem) {
      existingItem[1]++;
      existingItem[2] = existingItem[2] || (name === aliasName);
    } else {
      acc.push([value, 1, name === aliasName]);
    }

    return acc;
  }, []);

  // return result.sort((a, b) => b[1] - a[1]);
  return result;
}

const emojiPath = (str) => {
  const filePathRegex = /^\/[\w.-]+(\/[\w.-]+)*$/;
  return filePathRegex.test(str);
};


export const isEmojiedOpen = ref(false);
export const openEmojied = (messageId) => {
  console.log('openEmojied');
  isEmojiedOpen.value = true;
  selectedMessageId.value = messageId;
};

export const closeEmojied = () => {
  console.log('closeEmojied');
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


export { openEmoji, closeEmoji, selectEmoji, calcEmoji, emojiPath };