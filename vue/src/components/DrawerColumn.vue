<script setup>
import { ref, computed, onBeforeMount } from 'vue'
import { useThreadHeadsStore } from '../stores/threadHeads.js';
import { useChannelsStore } from '../stores/channels.js';
import { getIDB, getIDBs, getAllIDBs, upsertData } from '../my/indexDB.js';

const threadHeadsStore = useThreadHeadsStore()
const threadHeads = computed(() => {
  return threadHeadsStore.threadHeads
})

const channelsStore = useChannelsStore()
const channels = computed(() => {
  return channelsStore.channels
})

// const fetchThreadHeads = () => {
//   return new Promise((resolve, reject) => {
//     getAllIDBs('threadHead')
//       .then((data) => {
//         data.forEach(d => {
//           threadHeadsStore.insert(d);
//         });
//         resolve();
//       })
//       .catch((error) => {
//         reject(error);
//       });
//   });
// };


const fetchMention = () => {
  return new Promise((resolve, reject) => {
    getIDBs('threadHead', 'displayStatusIndex', 2)
      .then((data) => {
        const latest = data.reverse();
        latest.forEach(d => {
          threadHeadsStore.insert(d);
        });
        addFaviconBadge();
        resolve();
      })
      .catch((error) => {
        console.error(error);
        resolve();
      });
  });
};

const fetchUnread = () => {
  return new Promise((resolve, reject) => {
    getIDBs('threadHead', 'displayStatusIndex', 1)
      .then((data) => {
        const latest = data.reverse();
        latest.forEach(d => {
          threadHeadsStore.insert(d);
        });
        resolve();
      })
      .catch((error) => {
        reject(error);
      });
  });
};

const fetchRead = () => {
  return new Promise((resolve, reject) => {
    getIDBs('threadHead', 'displayStatusIndex', 0)
      .then((data) => {
        const latest = data.reverse();
        latest.forEach(d => {
          threadHeadsStore.insert(d);
        });
        resolve();
      })
      .catch((error) => {
        reject(error);
      });
  });
};

const fetchMute = () => {
  return new Promise((resolve, reject) => {
    getIDBs('threadHead', 'displayStatusIndex', 3)
      .then((data) => {
        const latest = data.reverse();
        latest.forEach(d => {
          threadHeadsStore.insert(d);
        });
        resolve();
      })
      .catch((error) => {
        reject(error);
      });
  });
};

const mentionChannel = () => {
  return new Promise((resolve, reject) => {
    getIDBs('channel', 'displayStatusIndex', 2)
      .then((data) => {
        const latest = data.reverse();
        latest.forEach(d => {
          channelsStore.insert(d);
        });
        resolve();
      })
      .catch((error) => {
        reject(error);
      });
  });
};

const unreadChannel = () => {
  return new Promise((resolve, reject) => {
    getIDBs('channel', 'displayStatusIndex', 1)
      .then((data) => {
        const latest = data.reverse();
        latest.forEach(d => {
          channelsStore.insert(d);
        });
        resolve();
      })
      .catch((error) => {
        reject(error);
      });
  });
};

const readChannel = () => {
  return new Promise((resolve, reject) => {
    getIDBs('channel', 'displayStatusIndex', 0)
      .then((data) => {
        const latest = data.reverse();
        latest.forEach(d => {
          channelsStore.insert(d);
        });
        resolve();
      })
      .catch((error) => {
        reject(error);
      });
  });
};

const muteChannel = () => {
  return new Promise((resolve, reject) => {
    getIDBs('channel', 'displayStatusIndex', 3)
      .then((data) => {
        const latest = data.reverse();
        latest.forEach(d => {
          channelsStore.insert(d);
        });
        resolve();
      })
      .catch((error) => {
        reject(error);
      });
  });
};

// const fetchChannels = () => {
//   return new Promise((resolve, reject) => {
//     getAllIDBs('channel')
//       .then((data) => {
//         data.forEach(d => {
//           channelsStore.insert(d);
//         });
//         resolve();
//       })
//       .catch((error) => {
//         reject(error);
//       });
//   });
// };

function getStatusClass(status) {
  return {
    'read': status === 0,
    'unread': status === 1,
    'mention': status === 2,
    'mute': status === 3
  };
}

const addFaviconBadge = () => {
  const favicon = document.querySelector('link[rel="icon"]') || document.querySelector('link[rel="shortcut icon"]');
  favicon.href = '/me.jpg';
};


onBeforeMount(async () => {
  // await messagesStore.deleteAll();

  await fetchMention();
  await fetchUnread();
  await fetchRead();
  await fetchMute();
  await mentionChannel();
  await unreadChannel();
  await readChannel();
  await muteChannel();

  // await fetchMessageData();

  // parentID.value = props.message_id;
  // console.log('threadHead', threadHead.value);
  // const content = document.getElementById('content');
  // content.scrollTop = content.scrollHeight;
  // window.scrollTo(0, content.scrollHeight);
});


</script>

<template>
  <div id="drawer_column">
    <label for="drawer_check" class="pc_disp_none for_drawer">≡</label>
    <input id="drawer_check" type="checkbox" class="pulling pc_disp_none">
    <table id="drawer">
      <tr><td><RouterLink to="/" >Home</RouterLink></td></tr>
      <tr><td><RouterLink to="/threads/" >threads</RouterLink></td></tr>

      <tr v-for="d in threadHeads">
        <td class="channel_menu" :class="getStatusClass(d.displayStatus)">
          <a :href="'/thread/' + d.channelID + '/' + d.parentID + '/'">{{ d.title }}</a>
        </td>
      </tr>

      <tr><td><RouterLink to="/sign/" >Sign</RouterLink></td></tr>

      <tr v-for="d in channels">
        <td class="channel_menu">
          <a :href="'/channel/' + d.channelID">{{ d.channelName }}</a>
        </td>
      </tr>
      <tr><td><RouterLink to="/addChannel/" ><span> + </span><span>チャネルを追加</span></RouterLink></td></tr>
      <tr><td><RouterLink to="/sign/" >Sign</RouterLink></td></tr>
    </table>
  </div>
</template>

<style scoped>

#drawer td {
  background-color: #EEEEEE;
}
/*#drawer td a {
  display: inline-block;
  width: 100%;
}
*/

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
  opacity: 0.5;
}

.favicon-badge {
  background-color: red;
  color: white;
  padding: 3px 5px;
  border-radius: 50%;
}

@media screen and (min-width : 901px) {
  #drawer {
    margin-top : -1px;
    background-color: white;
  }
}

@media screen and (max-width : 900px) {
  #drawer {
    width: 80%;
    overflow: scroll;
    position: absolute;
    z-index: 10;
    margin: 0;
    background-color: white;
    left: -100%;
    top : 51px;
    float: left;
  }
  .pulling {
    position: absolute;
    top: 0px;
    height: 50px;
    width: 50px;
    opacity: 0;
    z-index: 10;
  }
  .pulling:checked ~ #drawer{
    left: 0%;
  }
  .for_drawer {
    position: absolute;
    font-size: 40px;
    top: -10px;
    width: 50px;
    text-align: center;
  }
}
</style>
