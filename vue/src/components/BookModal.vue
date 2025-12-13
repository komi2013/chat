<template>
  <div class="modal">
    <div class="modal-content">
      <h2>イベントを追加</h2>
      <div for="timeStart">開始時間:</div>
      <input type="datetime-local" v-model="localEvent.timeStart" />
      <div>
        <div v-for="(ask, i) in reception.asks" :key="`ask-${i}`">
          <div>{{ ask.question }}</div>
          <input type="text" v-model="localEvent.answers[i]" />
        </div>

        <div v-for="(ask, i) in reception.askChoices" :key="`choice-${i}`">
          <div>{{ ask.question }}</div>
          <select v-model="localEvent.answers[reception.asks?.length + i]">
            <option v-for="(choice, choiceIndex) in ask.choices?.slice(1)" :key="choiceIndex" :value="choice">
              {{ choice }}
            </option>
          </select>
        </div>

        <div v-for="(ask, i) in reception.askMultiChoices" :key="`multi-${i}`">
          <div>{{ ask.question }}</div>
          <select v-model="localEvent.answers[reception.asks?.length + reception.askChoices.length + i]">
            <option v-for="(choice, choiceIndex) in ask.choices?.slice(1)" :key="choiceIndex" :value="choice">
              {{ choice }}
            </option>
          </select>
        </div>
      </div>
      <div>
        <div>メニュー:</div>

        <select v-model="localEvent.menuID">
          <template v-for="(serviceItem, serviceIndex) in reception.menus" :key="serviceItem.menuID">
            <option v-if="serviceItem.forBookType" :value="serviceItem.menuID" >
              {{ serviceItem.menuName || '無題メニュー' }} - {{ serviceItem.price }}円
            </option>
          </template>
        </select>

      </div>

      <button @click="submit">投稿</button>
      <button @click="$emit('close')">閉じる</button>
    </div>
  </div>
  <div v-if="errorMessage" class="errorMessage">{{ errorMessage }}</div>
  <NoticePopup />
</template>

<script setup>
import { ref, computed } from 'vue'

import NoticePopup from '@/components/NoticePopup.vue'

import { useNoticesStore } from '@/stores/notices.js'

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

const errorMessage = ref('')
const reception = props.reception;

async function submit() {
  if (!confirm("予約")) return
  const fd = new FormData();
  fd.append('aliasName', localStorage.getItem('myname'));
  fd.append('csrf', localStorage.getItem('csrf'));
  fd.append('receptionID', localEvent.value.receptionID);
  fd.append('bookStart', localEvent.value.timeStart);
  fd.append('bookEnd', localEvent.value.timeEnd);
  fd.append('answers', JSON.stringify(localEvent.value.answers));
  fd.append('menuID', localEvent.value.menuID);
  const res = await sendRequest('/ReceptionBook/', fd)
  if (!res.csrf) {
    errorMessage.value = res
    return
  }
  localStorage.setItem('csrf', res.csrf)
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  if (res.error) {
    const noticesStore = useNoticesStore()
    noticesStore.setNotice(res.error)
    return 
  }
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
