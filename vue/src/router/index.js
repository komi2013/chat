import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/book/:id/',
      component: () => import('../views/Book.vue'),
      props: route => ({id: route.params.id}),
    },
    {
      path: '/bookPattern/:id?/',
      component: () => import('../views/BookPattern.vue'),
      props: route => ({id: route.params.id}),
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
      path: '/menu/:id/:code?/',
      component: () => import('../views/Menu.vue'),
      props: route => ({
        id: route.params.id,
        code: route.params.code
      })
    },
    {
      path: '/menuOrder/:id/:apiKey',
      component: () => import('../views/MenuOrder.vue'),
      props: route => ({
        id: route.params.id,
        apiKey: route.params.apiKey,
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
    {
      path: '/shift/:id/',
      component: () => import('../views/Shift.vue'),
      props: route => ({id: route.params.id}),
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
      component: () => import('../views/Thread.vue'),
      props: route => ({
        channel_id: route.params.channel_id,
        message_id: route.params.message_id,
        backID: route.query.backID
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
      path: '/',
      name: 'top',
      component: () => import('../views/Top.vue')
    },
    {
      path: '/workflow/',
      name: 'workflow',
      component: () => import('../views/Workflow.vue')
    }
  ]
})

export default router
