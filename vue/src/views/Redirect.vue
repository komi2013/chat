<script setup>

switch (localStorage.getItem("TO")) {
  case 'dm':
    toDirectMessage();
    break;
  case null:
  case undefined:
    console.log('localStorage.TO is not set');
    location.href = '/';
    break;
  default:
    location.href = localStorage.TO;
}

function toDirectMessage() {
// /redirect/?to=dm&channelID=65acf63d8309d8da55154dea&toWhom=sei2
  let channel;
  try {
    getIDB('channel', getParam('channelID'))
      .then((channel) => {
        console.log(channel.aliasName);
        const toWhom = getParam('toWhom');
        const aliasName = channel.aliasName;
        const channelID = channel.channelID;
        const parentID = toWhom < aliasName ? toWhom + '@' +aliasName : aliasName + '@' + toWhom;
        const uri = `/thread/${channelID}/${parentID}/`;
        console.log(uri);
        location.href = uri;
      });
  } catch (error) {
    console.error(error);
    return;
  }
}


</script>

<template>

</template>

<style>

</style>

