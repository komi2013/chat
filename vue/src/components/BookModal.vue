<template>
  <div class="modal">
    <div class="modal-content">
      <h2>イベントを追加</h2>
      <label for="timeStart">開始時間:</label>
      <input type="datetime-local" v-model="localEvent.timeStart" />
      <div>
        <div v-for="(ask, i) in reception.asks" :key="`ask-${i}`">
          <label>{{ ask.question }}</label>
          <input type="text" v-model="localEvent.answers[i]" />
        </div>

        <div v-for="(ask, i) in reception.askChoices" :key="`choice-${i}`">
          <label>{{ ask.question }}</label>
          <select v-model="localEvent.answers[reception.asks.length + i]">
            <option v-for="(choice, choiceIndex) in ask.choices?.slice(1)" :key="choiceIndex" :value="choice">
              {{ choice }}
            </option>
          </select>
        </div>

        <div v-for="(ask, i) in reception.askMultiChoices" :key="`multi-${i}`">
          <label>{{ ask.question }}</label>
          <select v-model="localEvent.answers[reception.asks.length + reception.askChoices.length + i]">
            <option v-for="(choice, choiceIndex) in ask.choices?.slice(1)" :key="choiceIndex" :value="choice">
              {{ choice }}
            </option>
          </select>
        </div>
      </div>
      <div>
        <label>メニュー:</label>
        <select v-model="localEvent.menuID">
          <option v-for="(serviceItem, serviceIndex) in reception.menus" :key="serviceItem.id" :value="serviceItem.menuID">
            {{ serviceItem.menuName }} - {{ serviceItem.price }}円
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
  reception: Object,
  menuID: Number
});
const emit = defineEmits(['submit', 'close']);

const localEvent = ref({
  ...props.time,
  receptionID: props.reception.receptionID,
  answers: [],   // askChoicesの回答を格納する配列
  serviceID: props.menuID   // 選択されたmenuのIDを格納
});

const reception = props.reception;

async function submit() {
  const fd = new FormData();
  fd.append('aliasName', localStorage.getItem('myname'));
  fd.append('csrf', localStorage.getItem('csrf'));
  fd.append('receptionID', localEvent.value.receptionID);
  fd.append('bookStart', localEvent.value.timeStart);
  fd.append('bookEnd', localEvent.value.timeEnd);
  fd.append('answers', JSON.stringify(localEvent.value.answers));
  fd.append('menuID', localEvent.value.menuID);

  const res = await sendRequest('/ReceptionBook/', fd);
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
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
  z-index: 10;
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
