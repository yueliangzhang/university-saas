import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    redirect: '/login'
  },
  {
    path: '/init',
    name: 'Init',
    component: () => import('@/view/init/index.vue')
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/view/login/index.vue')
  },
  {
    path: '/platform/login',
    name: 'PlatformLogin',
    component: () => import('@/view/login/index.vue')
  },
  {
    path: '/scanUpload',
    name: 'ScanUpload',
    meta: {
      title: '扫码上传',
      client: true
    },
    component: () => import('@/view/example/upload/scanUpload.vue')
  },
  {
    path: '/:catchAll(.*)',
    meta: {
      closeTab: true
    },
    component: () => import('@/view/error/index.vue')
  },
  { path: '/teacher/login', redirect: '/teacher' },
  { path: '/student/login', redirect: '/student' },
  {
    path: '/teacher',
    name: 'teacherPortal',
    component: () => import('@/view/university/teacherPortal.vue'),
    meta: { title: '教师门户', client: true }
  },
  {
    path: '/student',
    name: 'studentPortal',
    component: () => import('@/view/university/studentPortal.vue'),
    meta: { title: '学生门户', client: true }
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router
