import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/adSetting/:date?/',
      component: () => import('../views/AdSetting.vue'),
      props: route => ({
        date: route.params.date
      })
    },
    {
      path: '/calendar/:date?/',
      component: () => import('../views/Calendar.vue'),
      props: route => ({
        date: route.params.date
      })
    },
    {
      path: '/calendarEdit/:id?/',
      component: () => import('../views/CalendarEdit.vue'),
      props: route => ({
        id: route.params.id,
        text: route.query.text,
        dates: route.query.dates
      })
    },
    {
      path: '/channel/:id?/',
      component: () => import('../views/Channel.vue'),
      props: route => ({id: route.params.id}),
    },
    {
      path: '/group/:id/',
      component: () => import('../views/Group.vue'),
      props: route => ({
        id: route.params.id
      })
    },
    {
      path: '/people/:id/:name/',
      component: () => import('../views/People.vue'),
      props: route => ({
        id: route.params.id,
        name: route.params.name
      }),
    },
    {
      path: '/profile/:id/',
      component: () => import('../views/Profile.vue'),
      props: route => ({
        id: route.params.id,
        code: route.query.code
      }),
    },
    {
      path: '/reception/:id?/',
      name: 'reception',
      component: () => import('../views/Reception.vue'),
      props: route => ({
        id: route.params.id,
        reception: route.query.reception
      }),
    },
    {
      path: '/receptionBook/:id?/',
      component: () => import('../views/ReceptionBook.vue'),
      props: route => ({id: route.params.id}),
    },
    {
      path: '/receptionMenu/:id/:code?/',
      component: () => import('../views/ReceptionMenu.vue'),
      props: route => ({
        id: route.params.id,
        code: route.params.code
      })
    },
    {
      path: '/receptionMenuOrder/:id/:apiKey',
      component: () => import('../views/ReceptionMenuOrder.vue'),
      props: route => ({
        id: route.params.id,
        apiKey: route.params.apiKey,
      })
    },
    {
      path: '/receptionOpen/:id/:menuID/:passkey/',
      component: () => import('../views/ReceptionOpen.vue'),
      props: route => ({
        id: route.params.id,
        menuID: route.params.menuID,
        passkey: route.params.passkey
      })
    },
    {
      path: '/receptionShift/:id/',
      component: () => import('../views/ReceptionShift.vue'),
      props: route => ({id: route.params.id}),
    },
    {
      path: '/redirect/',
      name: 'redirect',
      component: () => import('../views/Redirect.vue')
    },
    {
      path: '/setting/',
      component: () => import('../views/Setting.vue')
    },
    // {
    //   path: '/sign/',
    //   name: 'sign',
    //   component: () => import('../views/SignView.vue')
    // },
    {
      path: '/threadHead/:parent_id/',
      name: 'threadHead',
      component: () => import('../views/ThreadHead.vue'),
      props: route => ({
        parent_id: route.params.parent_id
      })
    },
    {
      path: '/thread/:channel_id/:parentID/',
      component: () => import('../views/Thread.vue'),
      props: route => ({
        channel_id: route.params.channel_id,
        parentID: route.params.parentID,
        backID: route.query.backID,
        messageID: route.query.messageID
      })
    },
    {
      path: '/ticket/:ticketID?/',
      name: 'ticket',
      component: () => import('../views/Ticket.vue'),
      props: route => ({
        ticketID: route.params.ticketID
      })
    },
    {
      path: '/tickets/',
      component: () => import('../views/Tickets.vue')
    },
    {
      path: '/timestamp/:adminName/:code/',
      component: () => import('../views/Timestamp.vue'),
      props: route => ({
        adminName: route.params.adminName,
        code: route.params.code
      })
    },
    {
      path: '/timestampCode/',
      component: () => import('../views/TimestampCode.vue')
    },
    {
      path: '/timestampReport/:admin/',
      component: () => import('../views/TimestampReport.vue'),
      props: route => ({
        admin: route.params.admin,
        month: route.query.month,
        stamper: route.query.stamper
      })
    },
    {
      path: '/user/',
      component: () => import('../views/User.vue')
    },
    { path: '/html/privacy/', component: () => import('../views/html/Privacy.vue') },
    { path: '/html/rule/', component: () => import('../views/html/Rule.vue') },
    { path: '/', component: () => import('../views/html/Top.vue') },

  ]
})

export default router
