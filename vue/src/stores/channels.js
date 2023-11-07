import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

export const useChannelsStore = defineStore({
  id: 'channels',
  state: () => ({
    channels: [],
  }),
  actions: {
    insert(data) {
      this.channels = data;
    },
    update() {
      this.count--;
    },
    adding(add) {
      this.count = this.count + add;
    },
  },
});
