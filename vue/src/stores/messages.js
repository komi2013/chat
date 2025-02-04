import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

export const useMessagesStore = defineStore({
  id: 'messages',
  state: () => ({
    messages: ref([]),
  }),
  actions: {
    insert(data) {
      const exists = this.messages.some(msg => msg.messageID === data.messageID);
      if (!exists) {
        this.messages.push(data);
      }
    },
    unshift(data, position) {
      this.messages.splice(position, 0, data);
    },
    update(data, messageID) {
      const index = this.messages.findIndex(message => message.messageID === messageID);
      if (index !== -1) {
        this.messages[index] = data;
      }
    },
    delete(messageID) {
      this.messages = this.messages.filter(message => message.messageID !== messageID);
    },
    deleteAll() {
      this.messages = ref([]); // 全てのメッセージを削除する
    },
    currentDisplay(currentID) {
      const message = this.messages.find(msg => {
        // メッセージのparentIDが存在する場合はそれを、存在しない場合はchannelIDをチェック
        const idToCheck = msg.parentID !== undefined ? msg.parentID : msg.channelID;
        // currentIDと一致するかどうかを確認
        return idToCheck === currentID;
      });
      // メッセージが見つかればtrue、見つからなければfalseを返す
      return !!message;
    },
    upOne(messageID, key, value) {
      const message = this.messages.find(message => message.messageID === messageID);
      if (message) {
        message[key] = value;
      }
    },
  },
});
