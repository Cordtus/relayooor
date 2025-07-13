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
    path: '/channels',
    name: 'channels',
    component: () => import('@/views/Channels.vue')
  },
  {
    path: '/relayers',
    name: 'relayers',
    component: () => import('@/views/Relayers.vue')
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
  }
]

export const router = createRouter({
  history: createWebHistory(),
  routes
})