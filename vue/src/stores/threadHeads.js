import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

export const useThreadHeadsStore = defineStore({
  id: 'threadHeads',
  state: () => ({
    threadHeads: ref([]),
  }),
  actions: {
    insert(data) {
      this.threadHeads.push(data);
    },
    // update() {
    //   this.count--;
    // },
    // adding(add) {
    //   this.count = this.count + add;
    // },
  },
});
