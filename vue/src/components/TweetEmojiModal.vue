<template>
  <div class="modal" @click.self="closeModal">
    <div class="modal-content">

      <div class="emoji-list">
        <template v-for="emoji in masterEmojis" :key="emoji">
          <template v-if="emojiPath(emoji)">
            <span>
              <img
                class="emoji-img"
                :src="emoji"
                @click="selectEmoji(emoji)"
              />
            </span>
          </template>
          <template v-else>
            <span
              class="emoji"
              @click="selectEmoji(emoji)"
            >
              {{ emoji }}
            </span>
          </template>
        </template>
      </div>

      <input
        type="text"
        v-model="selectedEmoji"
        maxlength="2"
        class="emoji-input"
        placeholder="😀"
        @change="inputEmoji"
      />

      <button class="close-btn" @click="closeModal">×</button>

      <br>
      <span v-if="emojiValidErr" class="emoji-valid-err">
        絵文字か1文字にしてください
      </span>

    </div>
  </div>
</template>

<script setup>
import { ref, defineProps, defineEmits } from 'vue';
// import { sendRequest } from '@/my/fetch';
import { emojiPath } from '@/my/emoji.js';
import { validateEmoji, rotateEmoji, masterEmojis } from '@/my/emoji';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  tweetID: String,   // 対象ツイートID
  parentID: String,  // 親ドキュメントID
  participants: Array // 通知対象ユーザー
});

const emit = defineEmits(['closeEmoji']);

const selectedEmoji = ref('');
const emojiValidErr = ref(false);

// ========================================
// 🔹 絵文字をクリック選択したとき
// ========================================
const selectEmoji = async (emoji) => {
  await sendEmoji(emoji);
  rotateEmoji(emoji);
  emit('closeEmoji');
};

// ========================================
// 🔹 絵文字を直接入力したとき
// ========================================
const inputEmoji = async () => {
  if (!validateEmoji(selectedEmoji.value)) {
    emojiValidErr.value = true;
    return;
  }
  await sendEmoji(selectedEmoji.value);
  rotateEmoji(selectedEmoji.value);
  emit('closeEmoji');
};

// ========================================
// 🔹 API送信共通処理
// ========================================
async function sendEmoji(emoji) {
  const fd = new FormData();
  fd.append('channelID', localStorage.getItem('channelID'));
  fd.append('updatedBy', localStorage.getItem('myname'));
  fd.append('pushNames', JSON.stringify(props.participants));
  fd.append('tweetID', props.tweetID);
  fd.append('parentID', props.parentID);
  fd.append('emoji', emoji);
  fd.append('csrf', localStorage.getItem('csrf'));

  const res = await sendRequest('/TweetEmoji/', fd)
  if (res.csrf) localStorage.setItem('csrf', res.csrf);

  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      pushReceive(content);
    }
  }
}

// ========================================
// 🔹 モーダルを閉じる
// ========================================
const closeModal = () => {
  emit('closeEmoji');
};
</script>

<style scoped>
.modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 20;
}

.modal-content {
  background: #fff;
  padding: 20px;
  border-radius: 8px;
  width: 280px;
  max-height: 80vh;
  overflow-y: auto;
  position: relative;
}

.close-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
}

.emoji-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  justify-content: center;
}

.emoji {
  font-size: 20px;
  cursor: pointer;
  padding: 4px;
}

.emoji-img {
  width: 20px;
  height: 20px;
  cursor: pointer;
}

.emoji-input {
  width: 36px;
  text-align: center;
  font-size: 18px;
  margin-top: 10px;
}

.emoji-valid-err {
  color: red;
  font-size: 12px;
}
</style>
