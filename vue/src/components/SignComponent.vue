<template>
  <div class="sign-box">
    <h1>Sign Component Mounted!</h1>
    <p>これは signTmp.tmpl からマウントされています !!!</p>
  </div>
</template>

<script setup>
// ここにロジック

function testComp () {
  console.log('mamam')
}
</script>

<style scoped>
.sign-box {
  padding: 16px;
  border: 2px solid #42b983;
  background: #f0fdf4;
}
</style>
