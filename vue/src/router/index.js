import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView
    },
    {
      path: '/about',
      name: 'about',
      component: () => import('../views/AboutView.vue')
    },
    {
      path: '/channel/:id',
      name: 'channel',
      component: () => import('../views/ChannelView.vue'),
      props: route => ({id: route.params.id}),
    },
    {
      path: '/sign',
      name: 'sign',
      component: () => import('../views/SignView.vue')
    },
    {
      path: '/emoji/:id',
      name: 'emoji',
      component: () => import('../views/EmojiView.vue')
    },
    {
      path: '/reply/:id',
      name: 'reply',
      component: () => import('../views/ReplyView.vue')
    }
  ]
})

export default router
