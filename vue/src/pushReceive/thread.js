import { useMessagesStore } from '../stores/messages.js';
import { getIDB, upsertData, deleteData } from '../my/indexDB.js';
import { getSubstring, removeHtmlTags } from '../my/strings.js';
import { textToHtml } from '../my/textToHtml.js';
export async function thread(pushData) {
  const messagesStore = useMessagesStore();
  const obj = {
    messageID: pushData[1],
    parentID: pushData[2],
    messageTxt: pushData[3],
    aliasName: pushData[4],
    aliasImg: pushData[5],
    createdAt: pushData[6]
  };
  console.log('thread messagesStore.messages', messagesStore.messages);
  console.log('messagesStore.messages.value[0].messageID', messagesStore.messages[0].messageID);
  console.log('messagesStore.messages.value[0].messageID', messagesStore.messages[0].messageID);

  if (messagesStore.messages.length === 1) {
    const threadHead = {
      parentID: messagesStore.messages[0].messageID,
      messageTxt: messagesStore.messages[0].messageTxt,
      aliasName: messagesStore.messages[0].aliasName,
      aliasImg: messagesStore.messages[0].aliasImg,
      createdAt: messagesStore.messages[0].createdAt,
      emojis: messagesStore.messages[0].emojis,
      channelID: pushData[7],
      title: getSubstring(removeHtmlTags(textToHtml(messagesStore.messages[0].messageTxt)), 0, 8),
      unreadFlg: true
    };
    console.log("配列は1件ですthreadHead", threadHead);
    upsertData(threadHead, 'threadHead', 'parentID', pushData[2])
      .catch((error) => {
        console.error(error);
      });
  }
  upsertData(obj, 'thread', 'messageID', pushData[1])
    .catch((error) => {
      console.error(error);
    });
  messagesStore.update(obj);

}
