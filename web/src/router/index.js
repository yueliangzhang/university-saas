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
  { path: '/teacher/login', name: 'teacherPortalLogin', component: () => import('@/view/university/teacherPortal.vue'), meta: { client: true } },
    {
  path: '/student/login',
  name: 'studentLogin',
  component: () => import('@/view/university/studentPortal.vue'),
  meta: {
    title: '学生登录',
    client: true
  }
}
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router
