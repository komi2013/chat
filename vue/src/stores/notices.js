// store/noticeStore.js
import { defineStore } from 'pinia';

export const useNoticesStore = defineStore('notice', {
  state: () => ({
    notice: '', // noticeメッセージ
  }),
  actions: {
    setNotice(message) {
      this.notice = message;
      setTimeout(() => {
        this.notice = ''; // 1秒後にクリア
      }, 2000);
    },
  },
});
