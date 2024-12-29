<script setup>
import { ref, computed, onBeforeMount } from 'vue'
import DrawerColumn from '../components/DrawerColumn.vue'
import EditBox from '../components/EditBox.vue'
import Messages from '../components/Messages.vue'
import { useThreadHeadsStore } from '../stores/threadHeads.js';
import { useChannelsStore } from '../stores/channels.js';
import { isEmojiOpen, selectedMessageId, openEmoji, closeEmoji, selectEmoji, calcEmoji, emojiPath } from '../my/emoji.js';
import { isOtherOpen, otherMessageId, openOther, closeOther, selectOther, activeEdit, adjustHeight, textareaRefs } from '../my/toggle.js';

const props = defineProps({
  id: '',
})

const channel = ref('');
const threadHeads = ref('');

async function fetchData() {
  try {
    const data = await getIDB('channel', props.id);
    channel.value = data;
  } catch (error) {
    channel.value = null;
  }
}

async function fetchThreadHead() {
  try {
    const data = await getIDBs('threadHead', 'channelIDIndex', props.id);
    threadHeads.value = data;
  } catch (error) {
    threadHeads.value = null;
  }
}

const newThreadURL = '/thread/' + props.id + '/' + generateRandomCode(3) + '/';
onBeforeMount(async () => {
  await fetchData();
  await fetchThreadHead();
});


</script>

<template>
<DrawerColumn />
<div id="content">
<div class="headTitle">
  <div>
    <a :href="'/channelInfo/' + channel.channelID"> {{ channel.channelName }} </a>
  </div>
  <div style="line-height: 50px;">
    &nbsp;
  </div>
</div>
<p><a :href="newThreadURL"> <button> + </button></a></p>
<template v-for="d in threadHeads" >
  <p>
    <a :href="'/thread/' + d.parentID.slice(0, 4) + '/' + d.parentID.slice(4) + '/'">
      {{d.title}}
    </a>
  </p>
</template>
</div>
</template>

<style>

#content {
  height: 100%;
  overflow-y: auto;
}

@media screen and (min-width : 701px) { 
  .headTitle {
    margin-left: 50px;
    display: flex;
  }
}

@media screen and (max-width : 700px) {
  .headTitle {
    margin-left: 50px;
    display: flex;
  }
  .headTitle div {
    display: table-cell;
  }
}
</style>

