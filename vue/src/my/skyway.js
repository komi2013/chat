// src/webrtc/skyway.js
import {
  SkyWayContext,
  SkyWayRoom,
  SkyWayStreamFactory
} from '@skyway-sdk/room';

// export async function createSkywaySession({ apiKey, roomName }) {
//   // SkyWay 初期化
//   const context = await SkyWayContext.Create(apiKey);

//   // SkyWay Room 作成（なければ作成、あれば参加）
//   const room = await SkyWayRoom.FindOrCreate(context, {
//     type: 'p2p',   // DM の場合 p2p でOK（最大2人）
//     name: roomName,
//   });

//   return room;
// }

export async function createSkywaySession({ token, roomName }) {
  const context = await SkyWayContext.Create(token);

  const room = await SkyWayRoom.FindOrCreate(context, {
    type: 'p2p',
    name: roomName,
  });

  return room;
}


// export async function startSkywayCall({ token, roomName, localVideoEl, remoteVideoEl, onStatus }) {
//   console.log("STEP3 token received =", token);

//   try {
//     console.log("STEP4 token before SkyWayContext.Create =", token);

//     console.log("TOKEN BEFORE CREATE:", token);
//     console.log("TOKEN LENGTH:", token.length);
//     console.log("TOKEN DOT COUNT:", (token.match(/\./g) || []).length);

//     const context = await SkyWayContext.Create(token);

//     onStatus("context created");

//     const room = await SkyWayRoom.FindOrCreate(context, {
//       name: roomName,
//     });
//     onStatus("room created");

//     const member = await room.join();
//     onStatus("joined");

//     const { audio, video } = await SkyWayStreamFactory.createMicrophoneAudioAndCameraStream();
//     video.attach(localVideoEl);

//     await member.publish(audio);
//     await member.publish(video);
//     onStatus("published");

//     room.onStreamPublished.add((e) => {
//       member.subscribe(e.publication.id).then((sub) => {
//         const stream = sub.stream;
//         stream.attach(remoteVideoEl);
//         onStatus("connected");
//       });
//     });

//     return {
//       end() {
//         member.leave();
//         room.dispose();
//         context.dispose();
//       },
//     };

//   } catch (err) {
//     console.error("STEP5 SkyWayContext.Create error =", err);
//     onStatus("error");
//     throw err;
//   }
// }

export async function startSkywayCall({
  token,
  roomName,
  localVideoEl,
  remoteVideoEl,
  onStatus
  }) {
  console.log("STEP3 token received =", token);

  try {
    console.log("STEP4 token before SkyWayContext.Create =", token);
    console.log("TOKEN LENGTH:", token.length);
    console.log("TOKEN DOT COUNT:", (token.match(/\./g) || []).length);

    // const context = await SkyWayContext.Create(token);

    const context = await SkyWayContext.Create(token, {
      log: { level: 'warn', format: 'object' }
    });


    onStatus("context created");

    const room = await SkyWayRoom.FindOrCreate(context, { name: roomName });
    onStatus("room created");

    const member = await room.join();
    onStatus("joined");

    const { audio, video } =
      await SkyWayStreamFactory.createMicrophoneAudioAndCameraStream();

    video.attach(localVideoEl);

    await member.publish(audio);
    await member.publish(video);

    onStatus("published");

    room.onStreamPublished.add((e) => {
      member.subscribe(e.publication.id).then((sub) => {
        const stream = sub.stream;
        stream.attach(remoteVideoEl);
        onStatus("connected");
      });
    });

    return {
      end() {
        member.leave();
        room.dispose();
        context.dispose();
      },
    };

  } catch (err) {
    console.error("STEP5 SkyWayContext.Create error =", err);
    onStatus("error");
    throw err;
  }
}
