<script setup>
import { ref, computed, onMounted } from 'vue'
import { useBookmarksStore } from '@/stores/bookmarks.js';
import { useThreadHeadsStore } from '@/stores/threadHeads.js';
import Advertisement from '@/components/Advertisement.vue';

const threadHeadsStore = useThreadHeadsStore()
const threadHeads = computed(() => {
  return threadHeadsStore.threadHeads
})

const fetchAllThreadHeads = async () => {
  const th = await getAllIDBs('threadHead');
  if (Array.isArray(th) && th.length > 0) {
    const data = th.filter(d => d.channelID === localStorage.getItem('channelID'));
    if (data.length > 0) {
      const sorted = [
        ...data.filter(d => d.displayStatus === 2), // mention
        ...data.filter(d => d.displayStatus === 1), // unread
        ...data.filter(d => d.displayStatus === 0), // read
        ...data.filter(d => d.displayStatus === 3)  // mute
      ];

      sorted.forEach(d => {
        if (!d.backID || (d.joinNames && d.joinNames.includes(localStorage.getItem('myname')) && d.displayStatus === 1)) {
          threadHeadsStore.insert(d)
        }
      });
    }
  }
};

const bookmarksStore = useBookmarksStore()
const bookmarks = computed(() => {
  return bookmarksStore.bookmarks
})

const fetchBookmarks = async () => {
  const data = await getAllIDBs('bookmark');
  data.forEach(d => {
    bookmarksStore.insert(d);
  });
};

function getStatusClass(status) {
  return {
    'read': status === 0,
    'unread': status === 1,
    'mention': status === 2,
    'mute': status === 3
  };
}

function goBookmarkThread (bookmark) {
  const channelID = bookmark.channelID;
  const msgSecondID = bookmark.messageID
  // const msgSecondID = bookmark.messageID.replace(channelID, '');
  // const backID = bookmark.backID.replace(channelID, '');
  location.href = '/thread/' + channelID + '/' + msgSecondID + '/?backID=' + bookmark.backID
  // :href="'/thread/' + d.channelID + '/' + paramMsg(d) + '/?backID=' + paramParent(d)"
}

onMounted(async () => {
  await fetchAllThreadHeads()
  await fetchBookmarks()
});


</script>

<template>
  <div id="drawer_column">
    <label for="drawer_check" class="pc_disp_none for_drawer">≡</label>
    <input id="drawer_check" type="checkbox" class="pulling pc_disp_none">
    <table id="drawer">
      <tr><td><a href="/" > ホーム </a></td></tr>
      <tr><td style="text-align: center;"> <Advertisement /> </td></tr>
      <tr><td><a href="/channel/" > 組織・チャネル設定 </a></td></tr>
      <tr><td>スレッド</td></tr>
      <tr v-for="d in threadHeads">
        <td :class="getStatusClass(d.displayStatus)">
          <span v-if="d.backID">💬</span><span v-if="!d.backID">&nbsp;</span>
          <a :href="'/thread/' + d.channelID + '/' + d.parentID + '/'">{{ d.title }}</a>
        </td>
      </tr>
    <template v-if="bookmarks.length > 0">
      <tr><td> 🔖 ブックマーク </td></tr>
      <tr v-for="d in bookmarks">
        <td :class="getStatusClass(d.displayStatus)">
          &nbsp;<a @click="goBookmarkThread(d)">
            {{ d.title }}</a>
        </td>
      </tr>
    </template>
    </table>
  </div>
</template>

<style scoped>

#drawer td {
  background-color: #EEEEEE;
}

.mention a {
  color: red;
  font-weight: bold;
}

.unread a {
  color: blueviolet;
}

.read a {
  color: blue;
  opacity: 0.8;
}

.mute a {
  color: blue;
  opacity: 0.3;
}

.favicon-badge {
  background-color: red;
  color: white;
  padding: 3px 5px;
  border-radius: 50%;
}

</style>
