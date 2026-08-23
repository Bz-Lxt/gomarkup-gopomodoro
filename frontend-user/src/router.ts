import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from './lib/api'
import Login from './views/LoginView.vue'
import Board from './views/BoardView.vue'
import Focus from './views/FocusView.vue'
import Burn from './views/BurndownView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: Login },
    { path: '/', component: Board },
    { path: '/focus', component: Focus },
    { path: '/burndown', component: Burn },
  ],
})

router.beforeEach((to) => {
  if (to.path !== '/login' && !getToken()) return '/login'
  if (to.path === '/login' && getToken()) return '/'
})

export default router
