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
      path: '/rounds/new',
      name: 'create-round',
      component: () => import('@/views/CreateRoundView.vue'),
      meta: { requiresAuth: true, adminOnly: true }
    },
    {
      path: '/my-feedback',
      name: 'my-feedback',
      component: () => import('@/views/MyFeedbackView.vue'),
      meta: { requiresAuth: true }
    }
  ]
})

router.beforeEach((to, from, next) => {
  const auth = useAuthStore()

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    next('/login')
  } else if (to.meta.guestOnly && auth.isAuthenticated) {
    next('/')
  } else if (to.meta.adminOnly && !auth.isAdmin) {
    next('/')
  } else {
    next()
  }
})

export default router
