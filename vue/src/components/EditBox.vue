<template>
  <div>
    <div class="editLeft" :id="'toolbar_' + messageID">
      <button class="ql-bold"></button>
      <button class="ql-strike"></button>
      <button class="ql-blockquote"></button>
      <button class="ql-code-block"></button>
      <button class="ql-link"></button>
      <select class="ql-color">
        <option value="red">Red</option>
        <option value=""></option>
      </select>
      <button class="attachment" @click="attach(messageID)">
        🌄
      </button>
    </div>
    <div class="editRight ql-toolbar ql-snow">
      <button v-if="messageID" @click="msgUpsert(messageID, true)">🗑</button>
      <button @click="tasking" :class="{ 'task': task }">🔖</button>
      <button @click="tasking" :class="{ 'task': task }">👥</button>
      <button @click="msgUpsert(messageID, false)">▶️</button>
    </div>
    <div :id="'edit_' + messageID"
      v-html="editTxt[messageID]"
      >
    </div>
  </div>
  <div class="files" v-html="fileInfo[messageID]"></div>
  <input type="file" style="position: fixed; left: -300px;" multiple :id="'fileInput_' + messageID">
</template>

<script setup>
import { ref, defineProps, onMounted } from 'vue';
import { useMessagesStore } from '@/stores/messages.js';

import Quill from 'quill';
import "quill-mention";
import "quill/dist/quill.snow.css";

import { htmlToMarkdown, markdownToHtml } from '@/my/markdown.js';
import { userIDsByName } from '@/my/channelFunc';
import { pushReceive } from '@/pushReceive/pushReceive.js';

const props = defineProps({
  channel: Object,
  aliases: Array,
  groups: Array,
  message: Object,
  threadHead: Object,
});

const message = props.message;
let task = ref(false);
const messageID = props.message.messageID;
const messagesStore = useMessagesStore();

let editTxt = ref({});
editTxt.value[messageID] = markdownToHtml(props.message.messageTxt, props.channel, []);
// console.log('props.message.messageTxt' , props.message.messageTxt);
// console.log('editTxt.value[messageID]' , editTxt.value[messageID]);
const tasking = () => {
  task.value = !task.value;
};

const attach = () => {
  const fileInput = document.getElementById('fileInput_' + messageID);
  if (fileInput) {
    fileInput.click();
  }
  console.log(fileInput);
  // File input要素にchangeイベントリスナーを追加
  // const fileInput = document.getElementById('fileInput_');
  fileInput.addEventListener('change', handleFileInputChange);

}

const fileInfo = ref({});
const handleFileInputChange = (event) => {
	console.log(event);
  const files = event.target.files;
  const newFileInfo = document.createElement('div');
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    const fileContainer = document.createElement('div');
    if (file.type.startsWith('image/')) {
      const image = document.createElement('img');
      image.src = URL.createObjectURL(file);
      image.style.maxWidth = '50px';
      image.style.maxHeight = '50px';
      fileContainer.appendChild(image);
    } else {
      const fileName = document.createTextNode(file.name);
      fileContainer.appendChild(fileName);
    }
    newFileInfo.appendChild(fileContainer);
  }
  console.log(fileInfo.value);
  fileInfo.value[messageID] = newFileInfo.outerHTML;
}

let dm = false;
let clicked = false;
const msgUpsert = async (messageID, delMessage) => {
  if (!messageID && quill.root.innerHTML == '<p><br></p>') {
    return;
  }
  if (clicked) {
    return;
  }
  clicked = true;
  const messageData = delMessage ? '' : htmlToMarkdown(quill.root.innerHTML.replace(/\uFEFF/g, ''));
  const pushTitle = messageID ? 'threadEdit' : 'thread';
  const SecondMsgID = messageID ? 
  	messageID.replace(props.channel.channelID, '') :
  	base62Encode(Math.floor(Date.now() / 1000)) +  generateRandomCode(1);
  const fd = new FormData();
  const alias = props.aliases.find(alias => alias.aliasName === props.channel.myname);
  let userIDs = [];
  let names = [props.channel.myname];
  let backID = '';
  if (props.threadHead) {
    if (props.threadHead.backID) {
      backID = props.threadHead.backID;
    }
    dm = props.threadHead.parentID.includes('@');
    if (dm) {
      let dmNames = props.threadHead.parentID.replace(props.channel.channelID, '').split('@');
      let matchedGroup = props.groups.find(group => dmNames.includes(group.groupName));
      let dmAliasNames = matchedGroup?.aliasNames || [];
      dmNames = dmNames.filter(name => name !== matchedGroup?.groupName);
      names = [...names, ...dmNames, ...dmAliasNames];
    }
    userIDs = userIDsByName(props.aliases, props.threadHead.aliasNames);
  }
  let yets = [];
  if (Array.isArray(props.groups) && !dm) {
    for (const d of props.groups) {
      const atName = `＠＠${d.groupName}・＠＠`;
      if (messageData.includes(atName)) {
        for (const d2 of d.aliasNames) {
          names.push(d2);
		    	if (task.value) {
				    yets.push({
				    	aliasName: d2,
				    	emoji: '☑️'
				    });    		
		    	}
        }
      }
    }
  }
  for (const d of props.aliases) {
    const atName = `＠＠${d.aliasName}・＠＠`;
    if (messageData.includes(atName) && !dm) {
    	if (!dm) {
	      // userIDs.push(d.userID);
	      names.push(d.aliasName);    		
    	}
    	if (task.value) {
		    yets.push({
		    	aliasName: d.aliasName,
		    	emoji: '☑️'
		    });    		
    	}
		}
  }
  const nameUserIDs = userIDsByName(props.aliases, names);
  const fileInput = document.getElementById('fileInput_' + messageID);
  if (fileInput && fileInput.files.length > 10) {
    alert('too many files');
    return;
  }
  if (fileInput && fileInput.files.length > 0) {
    for (const file of fileInput.files) {
      fd.append('files[]', file);
    }
  }
  fd.append('channelID', props.channel.channelID);
  fd.append('updatedBy', props.channel.myname);
  fd.append('userIDs', JSON.stringify([...new Set([...nameUserIDs, ...userIDs])]));
  fd.append('pushTitle', pushTitle);
  const contents = [
  	props.message.parentID,
    SecondMsgID,
    messageData,
    getAliasImg(props),
    [...new Set(names)],
    backID,
    yets
  ];
  fd.append('contents', JSON.stringify(contents));
  fd.append('csrf', localStorage.getItem("csrf"));
  const res = await sendRequest('/ContentsPush/', fd);
  res.csrf && localStorage.setItem('csrf', res.csrf);
  res.pushContents.forEach(content => {
    pushReceive(content);
  });
  quill.root.innerHTML = '';
  fileInfo.value = [];
  task.value = false;
  clicked = false;
}

function getAliasImg(props) {
  let aliasImg = null;
  const myname = props.channel.myname;
  for (let i = 0; i < props.aliases.length; i++) {
    if (props.aliases[i].aliasName === myname) {
      aliasImg = props.aliases[i].aliasImg;
      break;
    }
  }
  return aliasImg;
}

let quill;
onMounted(() => {
  quill = new Quill('#edit_' + messageID, {
    modules: {
      toolbar: '#toolbar_' + messageID,
      mention: {
        allowedChars: /^[A-Za-z\sÅÄÖåäö]*$/,
        mentionDenotationChars: ["@"],
        source: function(searchTerm, renderList, mentionChar) {
          let values;
          if (mentionChar === "@") {
            let aliasForMention = props.aliases
            if (Array.isArray(props.groups)) {
              aliasForMention = props.aliases.concat(
                props.groups.map(group => ({
                  aliasID: group.groupID,
                  aliasName: group.groupName,
                  aliasImg: group.groupImg,
                }))
              );
            }

            values = aliasForMention.map((alias, index) => {
              return {
                id: alias.aliasID,
                value: alias.aliasName,
                icon: alias.aliasImg
              };
            });
          }

          if (searchTerm.length === 0) {
            renderList(values, searchTerm);
          } else {
            const matches = values.filter(item => item.value.toLowerCase().includes(searchTerm.toLowerCase()));
            renderList(matches, searchTerm);
          }
        },
        renderItem: function(item) {
          const mentionWithImage = document.createElement("div");
          if (item.icon.charAt(0) == ',') {
          	const arr = item.icon.split(',');
          	console.log('arr', arr);
          	mentionWithImage.innerHTML = 
          		`<span class="min-icon" style="background-color:${arr[2]}"><span>${arr[1]}</span></span>${item.value}`;
          } else {
          	mentionWithImage.innerHTML = `<img src="${item.icon}" class="min-icon">${item.value}`;
          }
          return mentionWithImage;
        },
        onOpen: function() {
          const quillMentionList = document.getElementById('quill-mention-list');
          const rect = quillMentionList.getBoundingClientRect();
          console.log(rect);
          if (rect.left > 150 && rect.left < 300) {
            quillMentionList.style.left = (- 1 * rect.left) + 'px';
          }
        }
      }
    },
    theme: 'snow'
  });
});

</script>

<style>

.editLeft {
  display: inline-block;
  width: 69%;
}

.editRight {
  display: inline-block;
  width: 30%;
}

.editText .ql-container.ql-snow {
  border: 1px solid #d1d5db;
  border-bottom-width: 0;
}
.editText .ql-editor {
  padding: 4px 0px;
}

.files {
  border-top: none;
  border-right: 1px solid #d1d5db;
  border-bottom: 1px solid #d1d5db;
  border-left: 1px solid #d1d5db;
}
.ql-snow.ql-toolbar {
  padding: 8px 0px;
}
.ql-snow.ql-toolbar .attachment {
  font-size: 12px;
  padding-top: 0px;
}
.ql-snow.ql-toolbar .task {
  background-color: #92a7b54a;
  border-radius: 5px;
}
.ql-mention-list-container {
  background-color: white;
  bottom: 0px;
}

.ql-mention-list {
  display: flex;
  flex-direction: column;
  position: absolute;
  left: -30px;
  width: 300px;
  bottom: 0px;
}

.ql-mention-list-item {
  display: flex;
  align-items: center;
  background-color: white;
}

.ql-mention-list-item-text {
  margin-left: 8px;
}

.ql-mention-list-item-image {
  width: 24px;
  height: 24px;
}

.min-icon {
	width: 26px;
  max-width: 26px;
  height: 26px;
  max-height: 26px;
  border-radius: 4px;
  display: inline-flex;
  vertical-align: middle;
  justify-content: center;
  align-items: center;
}

.mention {
  background-color: #a7cad63d;
  color: blue;
}

</style>
