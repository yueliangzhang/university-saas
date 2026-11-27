<template>
  <div
    class="uni-admin-header flex justify-between fixed top-0 left-0 right-0 z-10 h-16 bg-white text-slate-700 dark:text-slate-300 dark:bg-slate-900 shadow dark:shadow-gray-700 items-center px-2"
  >
    <div class="flex items-center cursor-pointer flex-1">
      <div
        class="flex items-center justify-center cursor-pointer"
        :class="isMobile ? '' : 'min-w-48'"
        @click="router.push({ path: '/' })"
      >
        <Logo />
        <div
          v-if="!isMobile"
          class="inline-flex font-bold text-2xl ml-2"
          :class="
            (config.side_mode === 'head' ||
              config.side_mode === 'combination') &&
            'min-w-fit'
          "
        >
          {{ $GIN_VUE_ADMIN.appName }}
        </div>
      </div>

      <el-breadcrumb
        v-show="!isMobile"
        v-if="config.side_mode !== 'head' && config.side_mode !== 'combination'"
        class="ml-4"
      >
        <el-breadcrumb-item
          v-for="item in matched.slice(1, matched.length)"
          :key="item.path"
        >
          {{ fmtTitle(item.meta.title, route) }}
        </el-breadcrumb-item>
      </el-breadcrumb>
      <gva-aside
        v-if="config.side_mode === 'head' && !isMobile"
        class="flex-1"
      />
      <gva-aside
        v-if="config.side_mode === 'combination' && !isMobile"
        mode="head"
        class="flex-1"
      />
    </div>

    <div class="ml-2 flex shrink-0 items-center gap-2">
      <el-dropdown>
        <div class="flex justify-center items-center h-full w-full">
          <span
            class="cursor-pointer flex justify-center items-center text-black dark:text-gray-100"
          >
            <CustomPic />
            <span v-show="!isMobile" class="w-16">{{
              userStore.userInfo.nickName
            }}</span>
            <el-icon>
              <arrow-down />
            </el-icon>
          </span>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item>
              <span class="font-bold">
                当前角色：{{ userStore.userInfo.authority.authorityName }}
              </span>
            </el-dropdown-item>
            <template v-if="userStore.userInfo.authorities">
              <el-dropdown-item
                v-for="item in userStore.userInfo.authorities.filter(
                  (i) => i.authorityId !== userStore.userInfo.authorityId
                )"
                :key="item.authorityId"
                @click="changeUserAuth(item.authorityId)"
              >
                <span> 切换为：{{ item.authorityName }} </span>
              </el-dropdown-item>
            </template>
            <el-dropdown-item v-if="advisorAvailable" @click="router.push({ name: 'advisorWorkspace' })">
              切换为：辅导员
            </el-dropdown-item>
            <el-dropdown-item icon="reading-lamp" @click="userStore.LoginOut">
              登 出
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<script setup>
  import CustomPic from '@/components/customPic/index.vue'
  import { useUserStore } from '@/pinia/modules/user'
  import { useRoute, useRouter } from 'vue-router'
  import { useAppStore } from '@/pinia'
  import { storeToRefs } from 'pinia'
  import { computed, ref, watch } from 'vue'
  import request from '@/utils/request'
  import { setUserAuthority } from '@/api/user'
  import { fmtTitle } from '@/utils/fmtRouterTitle'
  import gvaAside from '@/view/layout/aside/index.vue'
  import Logo from '@/components/logo/index.vue'

  const userStore = useUserStore()
  const router = useRouter()
  const route = useRoute()
  const appStore = useAppStore()
  const { device, config } = storeToRefs(appStore)
  const isMobile = computed(() => {
    return device.value === 'mobile'
  })
  const matched = computed(() => route.meta.matched)

  const advisorAvailable = ref(false)
  watch(
    () => [userStore.userInfo.ID, userStore.userInfo.authorityId],
    async ([userID, authorityID], _, onCleanup) => {
      advisorAvailable.value = false
      if (!userID || ![2001, 2002].includes(authorityID)) return
      let current = true
      onCleanup(() => { current = false })
      try {
        const result = await request({ url: '/uni/advisor-workspace/status', donNotShowLoading: true })
        if (current) advisorAvailable.value = result.data?.available === true
      } catch {
        // The existing menu remains available if the status request fails.
      }
    },
    { immediate: true }
  )

  const changeUserAuth = async (id) => {
    const res = await setUserAuthority({
      authorityId: id
    })
    if (res.code === 0) {
      window.sessionStorage.setItem('needCloseAll', 'true')
      window.sessionStorage.setItem('needToHome', 'true')
      window.location.reload()
    }
  }
</script>

<style scoped lang="scss"></style>
