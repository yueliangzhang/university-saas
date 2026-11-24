<template>
  <span class="headerAvatar">
    <template v-if="picType === 'avatar'">
      <el-avatar
        v-if="avatar && !avatarFailed"
        :key="avatar"
        :size="size"
        :src="avatar"
        alt="用户头像"
        @error="avatarFailed = true"
      />
      <el-avatar v-else :size="size" :icon="UserFilled" class="default-avatar" aria-label="默认头像" />
    </template>
    <template v-if="picType === 'img'">
      <img :key="avatar" :src="avatar || noAvatarPng" class="avatar" alt="用户头像" @error.once="useDefaultImage" />
    </template>
    <template v-if="picType === 'file'">
      <el-image
        :src="file"
        class="file"
        :preview-src-list="previewSrcList"
        :preview-teleported="true"
      />
    </template>
  </span>
</template>

<script setup>
  import noAvatarPng from '@/assets/noBody.png'
  import { UserFilled } from '@element-plus/icons-vue'
  import { useUserStore } from '@/pinia/modules/user'
  import { getUrl } from '@/utils/image'
  import { computed, ref, watch } from 'vue'

  defineOptions({
    name: 'CustomPic'
  })

  const props = defineProps({
    picType: {
      type: String,
      required: false,
      default: 'avatar'
    },
    picSrc: {
      type: String,
      required: false,
      default: undefined
    },
    size: {
      type: Number,
      default: 30
    },
    preview: {
      type: Boolean,
      default: false
    }
  })

  const userStore = useUserStore()
  const avatarFailed = ref(false)
  const avatar = computed(() => {
    const src = props.picSrc === undefined ? userStore.userInfo.headerImg : props.picSrc
    return getUrl(src?.trim() || '') || undefined
  })
  watch(avatar, () => { avatarFailed.value = false })
  const useDefaultImage = (event) => { event.currentTarget.src = noAvatarPng }
  const file = computed(() => getUrl(props.picSrc))
  const previewSrcList = computed(() => (props.preview ? [file.value] : []))
</script>

<style scoped>
  .headerAvatar {
    display: flex;
    justify-content: center;
    align-items: center;
    margin-right: 8px;
  }
  .default-avatar {
    background: #e8efff;
    color: #3158a6;
  }
  .file {
    width: 80px;
    height: 80px;
    position: relative;
  }
</style>
