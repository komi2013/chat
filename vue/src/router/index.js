import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/channelInfo/:id/',
      name: 'channelInfo',
      component: () => import('../views/ChannelInfoView.vue'),
      props: route => ({id: route.params.id}),
    },
    {
      path: '/channel/:id/',
      name: 'channel',
      component: () => import('../views/ChannelView.vue'),
      props: route => ({id: route.params.id}),
    },
    {
      path: '/sign/',
      name: 'sign',
      component: () => import('../views/SignView.vue')
    },
    {
      path: '/emoji/:id/',
      name: 'emoji',
      component: () => import('../views/EmojiView.vue')
    },
    {
      path: '/addChannel/',
      name: 'addChannel',
      component: () => import('../views/AddChannel.vue')
    },
    {
      path: '/thread/:message_id/',
      name: 'reply',
      component: () => import('../views/ReplyView.vue')
    }
  ]
})

export default router
