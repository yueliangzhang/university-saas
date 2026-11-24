import './style/element_visiable.scss'
import 'element-plus/theme-chalk/dark/css-vars.css'
import 'uno.css'
import { createApp } from 'vue'
import ElementPlus from 'element-plus'

import 'element-plus/dist/index.css'
// 引入gin-vue-admin前端初始化相关内容
import './core/gin-vue-admin'
// 引入封装的router
import router from '@/router/index'
import '@/permission'
import run from '@/core/gin-vue-admin.js'
import auth from '@/directive/auth'
import clickOutSide from '@/directive/clickOutSide'
import { store } from '@/pinia'
import App from './App.vue'
import '@/core/error-handel'

const app = createApp(App)

app.config.productionTip = false

const mount = () => {
  app
    .use(run)
    .use(ElementPlus)
    .use(store)
    .use(auth)
    .use(clickOutSide)
    .use(router)
    .mount('#app')
}

if (import.meta.env.DEV) {
  import('vite-check-multiple-dom/client').then(({ setupVueRootValidator }) => {
    setupVueRootValidator(app, { lang: 'zh' })
    mount()
  })
} else {
  mount()
}
export default app

// V17.1 University SaaS design tokens and scoped components
import './styles/university-design-system.css'
