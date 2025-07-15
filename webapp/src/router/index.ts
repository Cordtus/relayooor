import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'dashboard',
    component: () => import('@/views/Dashboard.vue')
  },
  {
    path: '/monitoring',
    name: 'monitoring',
    component: () => import('@/views/Monitoring.vue')
  },
  {
    path: '/packet-clearing',
    name: 'packet-clearing',
    component: () => import('@/views/PacketClearing.vue')
  },
  {
    path: '/analytics',
    name: 'analytics',
    component: () => import('@/views/Analytics.vue')
  },
  {
    path: '/settings',
    name: 'settings',
    component: () => import('@/views/Settings.vue')
  }
]

export const router = createRouter({
  history: createWebHistory(),
  routes
})