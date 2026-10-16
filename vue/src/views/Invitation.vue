<script setup>
import { computed } from 'vue'

const props = defineProps({
  id: String,
  code: String
})

const isValid = value => /^[A-Za-z0-9_-]{1,128}$/.test(value || '')
const validInvitation = computed(() => isValid(props.id) && isValid(props.code))

// This package name matches the Android app declared in public/.well-known/assetlinks.json.
const packageName = 'com.chat.android'
const appUrl = computed(() => {
  if (!validInvitation.value) return ''
  const url = new URL(`chat://profile/${props.id}`)
  url.searchParams.set('code', props.code)
  return url.toString()
})
const installUrl = computed(() => {
  if (!validInvitation.value) return ''

  const referrer = new URLSearchParams({ m: props.id, code: props.code })
  const playUrl = new URL('https://play.google.com/store/apps/details')
  playUrl.searchParams.set('id', packageName)
  playUrl.searchParams.set('referrer', referrer.toString())
  return playUrl.toString()
})

document.title = 'Join Chat'
</script>

<template>
  <main class="invitation-page">
    <h1>Continue with Chat</h1>
    <p>Install the Android app to continue your invitation.</p>

    <template v-if="validInvitation">
      <p>Invitation: {{ id }}</p>
      <div class="app-open">
        <a :href="appUrl">📱 アプリで開く</a>
      </div>
      <a class="install-link" :href="installUrl">Install from Google Play</a>
    </template>
    <p v-else class="error" role="alert">
      Invalid or missing invitation information.
    </p>
  </main>
</template>

<style scoped>
.invitation-page {
  box-sizing: border-box;
  max-width: 36rem;
  margin: 3rem auto;
  padding: 2rem;
  color: #222;
  font-family: sans-serif;
  line-height: 1.5;
}

.install-link {
  display: inline-block;
  margin-top: 0.75rem;
  padding: 0.75rem 1rem;
  border-radius: 0.4rem;
  background: #087f5b;
  color: #fff;
  text-decoration: none;
}

.app-open {
  margin: 8px 0;
}

.app-open a {
  display: inline-block;
  padding: 8px 16px;
  border: 1px solid #888;
  border-radius: 6px;
}

.error {
  color: #b42318;
}
</style>
