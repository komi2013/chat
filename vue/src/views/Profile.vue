<script setup>
import { ref, computed, onBeforeMount, onMounted } from 'vue';
import DrawerColumn from '@/components/DrawerColumn.vue';
import PeopleImg from '@/components/PeopleImg.vue';
import { fetchChannel, fetchAliases, fetchGroups, userIDsByName } from '@/my/channelFunc';
import { getRandomEmoji, getRandomColor } from '@/my/emoji';

const props = defineProps({
  id: String,
  name: String,
  code: String,
  toAliasName: String
})

const channel = ref(null);
const alias = ref(null);
let aliases;
// const groups = ref([]);
const fetched = ref(false);
const isEditable = ref(false);
const aliasName = ref(props.name);
const aliasImg = ref(',' + getRandomEmoji() + ',' + getRandomColor());

// console.log(aliasImg.value);

onBeforeMount(async () => {
  if (!props.code) {
    channel.value = await fetchChannel(props.id);
    aliases = await fetchAliases(props.id);
    if (!props.name) {
      aliasName.value = channel.value.myname;
      isEditable.value = true;
    }
    alias.value = aliases.find(
      (item) => item.aliasName === aliasName.value
    );
    console.log(aliases);
    if (alias.value) {
      aliasImg.value = alias.value.aliasImg;
    }
  }
  fetched.value = true;
});

async function aliasEdit() {
  if (!confirm("実行▶️")) {
    return;
  }
  if (props.code) {
    join();
  } else {
    editAlias();
  }
}

function editAlias() {
  const fd = new FormData()
  fd.append('channelID', props.id);
  fd.append('updatedBy', channel.value.myname);
  fd.append('pushTitle', 'alias');
  fd.append('userIDs', JSON.stringify(userIDsByName(aliases)));
  const contents = [
    aliasImg.value,
    alias.value.userID
  ];
  fd.append('contents', JSON.stringify(contents));
  sendRequest('/ContentsPush/', fd);
}

async function join () {
  const fd = new FormData()
  fd.append('channelID', props.id);
  fd.append('code', props.code);
  fd.append('myname', aliasName.value);
  fd.append('myimg', aliasImg.value);
  await sendRequest('/ChannelJoin/', fd);
  location.href = '/profile/' + props.id + '/';
}

</script>

<template>
<DrawerColumn />
<div id="content" v-if="fetched">
<br><br>
  <div class="join">
    <template v-if="!isEditable">
      <img v-if="aliasImg && aliasImg.charAt(0) != ','" 
        :src="aliasImg" class="new-alias-img">
      <span v-if="aliasImg && aliasImg.charAt(0) == ','"
        class="new-alias-img" 
        :style="'background-color:' + aliasImg.split(',')[2] ">
        {{aliasImg.split(',')[1]}}</span>
    </template>

    <input v-if="props.code" type="text" v-model="aliasName" placeholder="このチャネルのニックネーム" class="aliasName">
    <span v-if="!props.code" class="aliasName">{{aliasName}}</span>
    <PeopleImg v-if="isEditable" v-model="aliasImg" />

    <div v-if="!isEditable && alias" class="display-mode">
      <p>{{ alias.bio }}</p>
    </div>
    <textarea 
      v-if="isEditable && alias"
      class="edit-mode" 
      v-model="alias.bio" 
      placeholder="自己紹介を入力してください">
    </textarea>

    <button @click="aliasEdit">▶️</button>
  </div>
</div>
<div v-if="!fetched"> Loading... or Something Went </div>
</template>

<style>

.join {
  width: 100%;
  text-align: center;
}

.aliasName {
  padding: 8px;
  margin: 10px;
  width: 200px;
}

button {
  padding: 8px;
  margin: 10px;
  width: 200px;
  font-size: 16px;
  background-color: #007bff;
  color: #fff;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

button:hover {
  background-color: #0056b3;
}

.display-mode {
  background-color: #f9f9f9;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 5px;
  cursor: pointer;
  white-space: pre-wrap; /* 改行を反映 */
}

.edit-mode {
  width: 94%;
  min-height: 100px;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 5px;
  font-size: 14px;
  font-family: Arial, sans-serif;
}

</style>

