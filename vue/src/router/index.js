import { createRouter, createWebHistory } from 'vue-router'

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
      component: () => import('../views/Channel.vue'),
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
      path: '/groupAlias/:id/:groupAliasName?/',
      name: 'groupAlias',
      component: () => import('../views/GroupAlias.vue'),
      props: route => ({
        id: route.params.id,
        groupAliasName: route.params.groupAliasName || null
      })
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
      path: '/ticket/:ticketID/',
      name: 'ticket',
      component: () => import('../views/Ticket.vue'),
      props: route => ({
        ticketID: route.params.ticketID
      })
    },
    {
      path: '/timestamp/:name/:code/',
      name: 'timestamp',
      component: () => import('../views/Timestamp.vue'),
      props: route => ({
        name: route.params.name,
        code: route.params.code
      })
    },
    {
      path: '/timestampCodeIssue/',
      name: 'timestampCodeIssue',
      component: () => import('../views/TimestampCodeIssue.vue')
    },
    {
      path: '/timestampReport/:admin/:month/:stamper?/',
      name: 'timestampReport',
      component: () => import('../views/TimestampReport.vue'),
      props: route => ({
        admin: route.params.admin,
        month: route.params.month,
        stamper: route.params.stamper
      })
    },
    {
      path: '/',
      name: 'top',
      component: () => import('../views/Top.vue')
    },
    {
      path: '/workflow/:ticketID/',
      name: 'workflow',
      component: () => import('../views/Workflow.vue'),
      props: route => ({
        ticketID: route.params.ticketID
      })
    }
  ]
})

export default router
