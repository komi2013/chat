import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

export const useBookmarksStore = defineStore({
  id: 'bookmarks',
  state: () => ({
    bookmarks: ref([]),
  }),
  actions: {
    insert(data) {
      this.bookmarks.push(data);
    },
    delete(messageID) {
      this.bookmarks = this.bookmarks.filter(bookmark => bookmark.messageID !== messageID);
    },
  },
});
