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

export { openEmoji, closeEmoji, selectEmoji, calcEmoji, emojiPath };