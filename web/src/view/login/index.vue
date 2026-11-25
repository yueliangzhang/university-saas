<template>
  <div id="userLayout" class="w-full h-full relative">
    <div
      class="rounded-lg flex items-center justify-evenly w-full h-full md:w-screen md:h-screen md:bg-[#194bfb] bg-white"
    >
      <div class="md:w-3/5 w-10/12 h-full flex items-center justify-evenly">
        <div
          class="oblique h-[130%] w-3/5 bg-white dark:bg-slate-900 transform -rotate-12 absolute -ml-52"
        />
        <!-- 分割斜块 -->
        <div
          class="z-[999] pt-12 pb-10 md:w-96 w-full rounded-lg flex flex-col justify-between box-border"
        >
          <div>
            <div class="flex items-center justify-center">
              <Logo :size="6" />
            </div>
            <div class="mb-9">
              <p class="text-center text-4xl font-bold">
                {{ $GIN_VUE_ADMIN.appName }}
              </p>
              <p class="text-center text-sm font-normal text-gray-500 mt-2.5">
                {{ platformMode ? '平台管理登录' : schoolName || '高校统一登录' }}
              </p>
            </div>
            <el-form
              ref="loginForm"
              :model="loginFormData"
              :disabled="submitting"
              :rules="rules"
              :validate-on-rule-change="false"
              @keyup.enter="submitForm()"
            >
              <el-form-item prop="username" class="mb-6">
                <el-input
                  v-model="loginFormData.username"
                  size="large"
                  :placeholder="platformMode ? '请输入用户名' : '用户名、学号或教师工号'"
                  autocomplete="username"
                  suffix-icon="user"
                />
              </el-form-item>
              <el-form-item prop="password" class="mb-6">
                <el-input
                  v-model="loginFormData.password"
                  show-password
                  size="large"
                  type="password"
                  placeholder="请输入密码"
                  autocomplete="current-password"
                />
              </el-form-item>
              <el-form-item v-if="!platformMode" prop="tenantCode" class="mb-6">
                <el-input
                  v-model.trim="loginFormData.tenantCode"
                  size="large"
                  maxlength="64"
                  placeholder="高校编号"
                  aria-label="高校编号"
                />
              </el-form-item>
              <el-form-item
                v-if="platformMode && loginFormData.openCaptcha"
                prop="captcha"
                class="mb-6"
              >
                <div class="flex w-full justify-between">
                  <el-input
                    v-model="loginFormData.captcha"
                    placeholder="请输入验证码"
                    size="large"
                    class="flex-1 mr-5"
                  />
                  <div class="w-1/3 h-11 bg-[#c3d4f2] rounded">
                    <img
                      v-if="picPath"
                      class="w-full h-full"
                      :src="picPath"
                      alt="请输入验证码"
                      @click="loginVerify()"
                    />
                  </div>
                </div>
              </el-form-item>
              <el-form-item class="mb-6">
                <el-button
                  class="shadow shadow-active h-11 w-full"
                  type="primary"
                  size="large"
                  :loading="submitting"
                  @click="submitForm()"
                  >登 录</el-button
                >
              </el-form-item>
              <el-form-item v-if="matchingRoles.length">
                <span class="w-full">请选择本次登录的身份</span>
                <el-button
                  v-for="role in matchingRoles"
                  :key="role"
                  :disabled="submitting"
                  @click="submitForm(role)"
                >{{ portalRoles[role].label }}</el-button>
              </el-form-item>
              <el-form-item v-if="isDev" class="mb-6">
                <el-button
                  class="shadow shadow-active h-11 w-full"
                  type="primary"
                  size="large"
                  @click="checkInit"
                  >前往初始化</el-button
                >
              </el-form-item>
            </el-form>
            <router-link v-if="platformMode" :to="{ name: 'Login' }" class="text-sm text-gray-500">
              返回高校统一登录
            </router-link>
          </div>
        </div>
      </div>
      <div class="hidden md:block w-1/2 h-full float-right bg-[#194bfb]">
        <img
          class="h-full"
          src="@/assets/login_right_banner.jpg"
          alt="banner"
        />
      </div>
    </div>

  </div>
</template>

<script setup>
  import { captcha } from '@/api/user'
  import { checkDB } from '@/api/initdb'
  import { computed, reactive, ref, watch } from 'vue'
  import axios from 'axios'
  import { ElMessage } from 'element-plus'
  import { useRouter } from 'vue-router'
  import { useUserStore } from '@/pinia/modules/user'
  import Logo from '@/components/logo/index.vue'
  import { isDev } from '@/utils/env.js'

  defineOptions({
    name: 'Login'
  })

  const router = useRouter()
  const platformMode = computed(() => router.currentRoute.value.name === 'PlatformLogin')
  const captchaRequiredLength = ref(6)
  const checkCaptcha = (rule, value, callback) => {
    if (!platformMode.value || !loginFormData.openCaptcha) {
      return callback()
    }
    const sanitizedValue = (value || '').replace(/\s+/g, '')
    if (!sanitizedValue) {
      return callback(new Error('请输入验证码'))
    }
    if (!/^\d+$/.test(sanitizedValue)) {
      return callback(new Error('验证码须为数字'))
    }
    if (sanitizedValue.length < captchaRequiredLength.value) {
      return callback(
        new Error(`请输入至少${captchaRequiredLength.value}位数字验证码`)
      )
    }
    if (sanitizedValue !== value) {
      loginFormData.captcha = sanitizedValue
    }
    callback()
  }

  // 获取验证码
  const loginVerify = async () => {
    const ele = await captcha()
    captchaRequiredLength.value = Number(ele.data?.captchaLength) || 0
    picPath.value = ele.data?.picPath
    loginFormData.captcha = ''
    loginFormData.captchaId = ele.data?.captchaId
    loginFormData.openCaptcha = ele.data?.openCaptcha
  }

  // 登录相关操作
  const loginForm = ref(null)
  const picPath = ref('')
  const loginFormData = reactive({
    username: '',
    tenantCode: '',
    password: '',
    captcha: '',
    captchaId: '',
    openCaptcha: false
  })
  const rules = reactive({
    tenantCode: [{ required: true, whitespace: true, message: '请输入高校编号', trigger: 'blur' }],
    username: [{ required: true, whitespace: true, message: '请输入账号', trigger: 'blur' }],
    password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
    captcha: [{ validator: checkCaptcha, trigger: 'blur' }]
  })

  const userStore = useUserStore()
  const submitting = ref(false)
  const matchingRoles = ref([])
  const portalRoles = {
    student: { label: '学生', route: 'studentPortal', sessionKey: 'uni_student_token' },
    teacher: { label: '教师', route: 'teacherPortal', sessionKey: 'uni_teacher_token' },
    tenant_admin: { label: '高校管理员' },
    leave_monitor: { label: '请假监督员' }
  }
  const portalClient = axios.create({
    baseURL: import.meta.env.VITE_BASE_API || '/api',
    timeout: 15000
  })

  watch(
    () => [loginFormData.username, loginFormData.tenantCode, loginFormData.password],
    () => { matchingRoles.value = [] }
  )
  watch(platformMode, active => {
    if (active) loginVerify().catch(() => {})
  }, { immediate: true })

  const schoolName = ref('')
  watch(() => [platformMode.value, loginFormData.tenantCode], ([platform, code], _, onCleanup) => {
    schoolName.value = ''
    if (platform || !code) return
    const request = new AbortController()
    const timer = setTimeout(async () => {
      try {
        const { data } = await portalClient.get('/uni/portal/school', {
          params: { code }, signal: request.signal
        })
        if (!request.signal.aborted && data.code === 0) schoolName.value = data.data?.name || ''
      } catch {
        // The school lookup is optional for signing in.
      }
    }, 350)
    onCleanup(() => {
      clearTimeout(timer)
      request.abort()
    })
  })

  const submitForm = async role => {
    if (submitting.value || (matchingRoles.value.length && !role)) return
    submitting.value = true
    try {
      if (!await loginForm.value.validate().catch(() => false)) return
      if (platformMode.value) {
        const ok = await userStore.LoginIn({
          ...loginFormData,
          username: loginFormData.username.trim()
        })
        if (ok) sessionStorage.setItem('platformLogin', 'true')
        else await loginVerify()
        return
      }

      const payload = {
        tenantCode: loginFormData.tenantCode,
        account: loginFormData.username.trim(),
        password: loginFormData.password,
        role
      }
      const { data: result } = await portalClient.post('/uni/portal/login', payload)
      if (result.code !== 0) throw new Error(result.msg || '登录失败')
      if (result.data.roles) {
        if (payload.tenantCode === loginFormData.tenantCode &&
            payload.account === loginFormData.username.trim() &&
            payload.password === loginFormData.password) {
          matchingRoles.value = result.data.roles
        }
        return
      }
      if (result.data.user) {
        sessionStorage.removeItem('platformLogin')
        if (!await userStore.LoginIn(undefined, result)) {
          throw new Error('登录成功，但工作台加载失败')
        }
        return
      }
      const destination = portalRoles[result.data.role]
      sessionStorage.setItem(destination.sessionKey, result.data.token)
      loginFormData.password = ''
      await router.replace({ name: destination.route })
    } catch (error) {
      ElMessage.error(error.response?.data?.msg || error.message || '登录失败')
      if (platformMode.value) await loginVerify().catch(() => {})
    } finally {
      submitting.value = false
    }
  }

  // 跳转初始化
  const checkInit = async () => {
    const res = await checkDB()
    if (res.code === 0) {
      if (res.data?.needInit) {
        userStore.NeedInit()
        await router.push({ name: 'Init' })
      } else {
        ElMessage({
          type: 'info',
          message: '已配置数据库信息，无法初始化'
        })
      }
    }
  }
</script>
