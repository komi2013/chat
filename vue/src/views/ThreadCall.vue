<template>
  <div class="call-wrapper">
    <h2>通話中 - {{ channelName }}</h2>

    <div class="video-container">
      <video ref="localVideo" autoplay playsinline muted style="display: none"></video>
      <video ref="remoteVideo" autoplay playsinline></video>
    </div>

    <button class="leave-btn" @click="leaveCall">切断</button>
  </div>
</template>

<script setup>
import {
  SkyWayContext,
  SkyWayRoom,
  SkyWayStreamFactory
} from '@skyway-sdk/room';

// import { appId, secret } from '../../env';
import { ref, onMounted, onBeforeUnmount } from 'vue';
// import { useRouter } from 'vue-router';

const props = defineProps({
  channelName: {
    type: String,
    required: true
  }
});

async function findToken() {
  const fd = new FormData()
  fd.append('csrf', localStorage.getItem('csrf'))
  const res = await sendRequest('/WebRTCTokenGet/', fd)
  if (!res.csrf) {
    errorMessage.value = res
    return
  }
  localStorage.setItem('csrf', res.csrf);
  if (Array.isArray(res.pushContents)) {
    for (const content of res.pushContents) {
      await pushReceive(content)
    }
  }
  if (res.error) { errorMessage.value = res.error }
  // fetched.value = true
  return res.token
}

// const router = useRouter();

// DOM refs
const localVideo = ref(null);
const remoteVideo = ref(null);

// State
let room = null;
let member = null;

async function startCall(token) {
  const { audio, video } =
    await SkyWayStreamFactory.createMicrophoneAudioAndCameraStream();

  video.attach(localVideo.value);
  const context = await SkyWayContext.Create(token, {
    log: { level: "warn", format: "object" }
  });
  room = await SkyWayRoom.FindOrCreate(context, {
    name: props.channelName,
  });

  member = await room.join();
  await member.publish(audio, { type: "sfu" });
  await member.publish(video, {
    type: "sfu",
    encodings: [
      { id: "low", maxBitrate: 10000 },
      { id: "high", maxBitrate: 800000 }
    ]
  });

  // Remote video
  // member.onPublicationSubscribed.add(async ({ stream }) => {
  //   if (stream.contentType !== "video") return;
  //   stream.attach(remoteVideo.value);
  // });

  member.onPublicationSubscribed.add(async ({ stream }) => {
    // attach audio + video both to remoteVideo
    stream.attach(remoteVideo.value);

    // make sure audio plays
    remoteVideo.value.muted = false;
    remoteVideo.value.volume = 1.0;

    try {
      await remoteVideo.value.play();
    } catch (e) {
      console.warn("Autoplay blocked:", e);
    }
  });

  const subscribe = async (pub) => {
    if (pub.publisher.id === member.id) return;
    await member.subscribe(pub.id);
  };

  room.publications.forEach(subscribe);
  room.onStreamPublished.add((e) => subscribe(e.publication));
}

async function leaveCall() {
  if (member) await member.leave()
  if (room) await room.dispose()
  // router.push("/");
  location.href = '/'
}

// Lifecycle
onMounted(async () => {
  const token = await findToken()
  startCall(token)
});

onBeforeUnmount(() => {
  leaveCall();
});
</script>

<style scoped>
.call-wrapper {
  text-align: center;
}

.video-container {
  display: flex;
  justify-content: center;
  gap: 20px;
}

video {
  width: 100%;
  background: #000;
}

.leave-btn {
  margin-top: 20px;
  padding: 10px 20px;
  background: red;
  color: white;
  border-radius: 6px;
}
</style>
