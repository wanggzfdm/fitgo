import { createRouter, createWebHistory } from 'vue-router'
import Home from './App.vue'
import RenderMd from './components/RenderMd.vue'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: Home
  },
  {
    path: '/render-md',
    name: 'RenderMd',
    component: RenderMd
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router