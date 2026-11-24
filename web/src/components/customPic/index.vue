<template>
  <span class="headerAvatar">
    <template v-if="picType === 'avatar'">
      <el-avatar :key="avatar" :size="size" :src="avatar" :icon="UserFilled" alt="用户头像" />
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
  import { computed } from 'vue'

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
  const avatar = computed(() => {
    const src = props.picSrc === undefined ? userStore.userInfo.headerImg : props.picSrc
    return getUrl(src?.trim() || '') || undefined
  })
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
  .file {
    width: 80px;
    height: 80px;
    position: relative;
  }
</style>
