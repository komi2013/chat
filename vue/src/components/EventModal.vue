<template>
  <div class="modal">
    <div class="modal-content">
      <h2>イベントを追加</h2>

      <label for="timeStart">開始時間:{{localEvent.timeStart}}</label>
      <input type="datetime-local" v-model="localEvent.timeStart" />

      <label for="timeEnd">終了時間:</label>
      <input type="datetime-local" v-model="localEvent.timeEnd" />

      <label for="title">タイトル:</label>
      <input type="text" v-model="localEvent.title" />

      <label for="todo">ToDo（任意）:</label>
      <input type="text" v-model="localEvent.todo" />

      <label for="scheduleType">スケジュールタイプ:</label>
      <select v-model="localEvent.scheduleType">
        <option value="0">タイプ 0</option>
        <option value="1">タイプ 1</option>
        <option value="2">タイプ 2</option>
        <option value="3">タイプ 3</option>
      </select>

      <button @click="submit">投稿</button>
      <button @click="$emit('close')">閉じる</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
const props = defineProps({
  time: Object,
});
const emit = defineEmits(['submit', 'close']);

// モーダルでのイベント入力を管理
const localEvent = ref({ ...props.time });

console.log('props.time', props.time);

// イベントの投稿
const submit = () => {
  console.log('localEvent.value', localEvent.value);
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
