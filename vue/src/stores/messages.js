import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

export const useMessagesStore = defineStore({
  id: 'messages',
  state: () => ({
    messages: [],
  }),
  actions: {
    insert(data) {
      console.log(data)
      this.messages = data;
    },
    update() {
      this.count--;
    },
    adding(add) {
      this.count = this.count + add;
    },
  },
});
