import { createRouter, createWebHistory } from 'vue-router'
// import { useMessagesStore } from '../stores/messages.js';

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/channelAdd/',
      name: 'channelAdd',
      component: () => import('../views/ChannelAdd.vue')
    },
    {
      path: '/channelInfo/:id?/',
      name: 'channelInfo',
      component: () => import('../views/ChannelInfo.vue'),
      props: route => ({id: route.params.id}),
    },
    {
      path: '/channel/:id/',
      name: 'channel',
      component: () => import('../views/ChannelView.vue'),
      props: route => ({id: route.params.id}),
    },
    {
      path: '/communityJoin/:code/',
      name: 'communityJoin',
      component: () => import('../views/CommunityJoin.vue'),
      props: route => ({
        code: route.params.code
      }),
    },
    {
      path: '/mypage/',
      name: 'mypage',
      component: () => import('../views/mypage.vue')
    },
    {
      path: '/redirect/',
      name: 'redirect',
      component: () => import('../views/Redirect.vue')
    },
    {
      path: '/sign/',
      name: 'sign',
      component: () => import('../views/SignView.vue')
    },
    {
      path: '/threadHead/:parent_id/',
      name: 'threadHead',
      component: () => import('../views/ThreadHead.vue'),
      props: route => ({
        parent_id: route.params.parent_id
      })
    },
    {
      path: '/thread/:channel_id/:message_id/',
      name: 'thread',
      component: () => import('../views/ThreadView.vue'),
      props: route => ({
        channel_id: route.params.channel_id,
        message_id: route.params.message_id
      })
    },
    {
      path: '/',
      name: 'top',
      component: () => import('../views/Top.vue')
    }
  ]
})

// router.beforeEach((to, from, next) => {
//   console.log('Leaving route:', from.path);
//   const messagesStore = useMessagesStore();
//   messagesStore.deleteAll();
//   console.log('messagesStore', messagesStore);
//   next(); // 次のナビゲーションステップを実行
// })


export default router
