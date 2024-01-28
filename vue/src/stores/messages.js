import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

export const useMessagesStore = defineStore({
  id: 'messages',
  state: () => ({
    messages: [],
  }),
  actions: {
    insert(data) {
      // console.log(data)
      this.messages = data;
    },
    update(data) {
      this.messages.push(data);
    },
    adding(add) {
      this.count = this.count + add;
    },
  },
});
