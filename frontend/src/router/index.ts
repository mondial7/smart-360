import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'dashboard',
      component: () => import('@/views/DashboardView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { guestOnly: true }
    },
    {
      path: '/auth/callback',
      name: 'auth-callback',
      component: () => import('@/views/AuthCallback.vue')
    },
    {
      path: '/team',
      name: 'team',
      component: () => import('@/views/TeamView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/rounds',
      name: 'rounds',
      component: () => import('@/views/RoundsView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/rounds/:id',
      name: 'round-details',
      component: () => import('@/views/RoundDetailsView.vue'),
      meta: { requiresAuth: true, adminOnly: true }
    },
    {
      path: '/rounds/new',
      name: 'create-round',
      component: () => import('@/views/CreateRoundView.vue'),
      meta: { requiresAuth: true, adminOnly: true }
    },
    {
      path: '/rounds/:id/submit',
      name: 'submit-feedback',
      component: () => import('@/views/SubmitFeedbackView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/rounds/:id/submission',
      name: 'view-submission',
      component: () => import('@/views/SubmissionView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/rounds/:roundId/consolidation',
      name: 'consolidation',
      component: () => import('@/views/ConsolidationView.vue'),
      meta: { requiresAuth: true, adminOnly: true }
    },
    {
      path: '/audit-logs',
      name: 'audit-logs',
      component: () => import('@/views/AuditLogView.vue'),
      meta: { requiresAuth: true, adminOnly: true }
    },
    {
      path: '/analytics',
      name: 'analytics',
      component: () => import('@/views/AnalyticsView.vue'),
      meta: { requiresAuth: true, adminOnly: true }
    },
    {
      path: '/my-feedback',
      name: 'my-feedback',
      component: () => import('@/views/MyFeedbackView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/teams',
      name: 'teams',
      component: () => import('@/views/TeamsManagementView.vue'),
      meta: { requiresAuth: true, adminOnly: true }
    },
    {
      path: '/teams/new',
      name: 'create-team',
      component: () => import('@/views/CreateTeamView.vue'),
      meta: { requiresAuth: true, adminOnly: true }
    },
    {
      path: '/teams/:id',
      name: 'team-details',
      component: () => import('@/views/TeamDetailsView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/teams/:teamId/create-round',
      name: 'create-team-round',
      component: () => import('@/views/CreateTeamRoundView.vue'),
      meta: { requiresAuth: true, teamAdminOrAdmin: true }
    },
  ]
})

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    next('/login')
  } else if (to.meta.guestOnly && auth.isAuthenticated) {
    next('/')
  } else if (to.meta.adminOnly && !auth.isAdmin) {
    next('/')
  } else if (to.meta.teamAdminOrAdmin && !(auth.isAdmin || auth.user?.role === 'team_admin')) {
    next('/')
  } else {
    next()
  }
})

export default router
