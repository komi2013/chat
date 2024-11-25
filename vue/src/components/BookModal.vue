<template>
  <div class="modal">
    <div class="modal-content">
      <h2>イベントを追加</h2>
      <label for="timeStart">開始時間:</label>
      <input type="datetime-local" v-model="localEvent.timeStart" />

<!--       <label for="timeEnd">終了時間:</label>
      <input type="datetime-local" v-model="localEvent.timeEnd" />
 -->
      <div v-for="(question, index) in bookPattern.times[0].askChoices" :key="index">
        <label>{{ question[0] }}</label>
        <select v-model="localEvent.answers[index]">
          <option v-for="(choice, choiceIndex) in question.slice(1)" :key="choiceIndex" :value="choice">
            {{ choice }}
          </option>
        </select>
      </div>

      <div>
        <label>メニュー:</label>
        <select v-model="localEvent.menuID">
          <option v-for="(menuItem, menuIndex) in bookPattern.menus" :key="menuIndex" :value="menuIndex">
            {{ menuItem.name }} - {{ menuItem.price }}円
          </option>
        </select>
      </div>

      <button @click="submit">投稿</button>
      <button @click="$emit('close')">閉じる</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
const props = defineProps({
  time: Object,
  bookPattern: Object
});
const emit = defineEmits(['submit', 'close']);

const localEvent = ref({
  ...props.time,
  bookPatternID: props.bookPattern.id,
  answers: [],   // askChoicesの回答を格納する配列
  menuID: null   // 選択されたmenuのIDを格納
});

const bookPattern = props.bookPattern;

console.log('props.time', props.time);
console.log('props.bookPattern', props.bookPattern);

const submit = () => {
  console.log('localEvent.value', localEvent.value);
  const fd = new FormData();
  fd.append('bookPatternID', localEvent.value.bookPatternID);
  fd.append('bookStart', localEvent.value.timeStart);
  fd.append('bookEnd', localEvent.value.timeEnd);
  fd.append('answers', JSON.stringify(localEvent.value.answers));
  fd.append('menuID', localEvent.value.menuID);
  sendRequest('/BookAdd/', fd);
  emit('close');
};
</script>

<style scoped>
.modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 2;
}

.modal-content {
  background-color: white;
  padding: 20px;
  border-radius: 5px;
  max-width: 400px;
  width: 100%;
}

.modal-content input,
.modal-content select {
  width: 100%;
  margin-bottom: 10px;
}
</style>
