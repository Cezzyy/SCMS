import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/authStore'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/login'
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
    },
    {
      path: '/home',
      name: 'home',
      component: () => import('../views/HomeView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/admin',
      name: 'admin',
      component: () => import('../views/AdminView.vue'),
      meta: { requiresAuth: true, requiresAdmin: false }
    }
  ],
})

// Navigation guard
router.beforeEach((to, from, next) => {
  // Check if the route requires authentication
  if (to.matched.some(record => record.meta.requiresAuth)) {
    const authStore = useAuthStore()

    // If not authenticated, redirect to login
    if (!authStore.isAuthenticated) {
      next({ name: 'login' })
      return
    }

    // If requires admin role, check user role
    if (to.matched.some(record => record.meta.requiresAdmin)) {
      if (authStore.userRole !== 'admin') {
        next({ name: 'home' })
        return
      }
    }
  }

  // Proceed to the route
  next()
})

export default router
