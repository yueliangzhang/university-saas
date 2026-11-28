import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/utils/request'

export function useLeaveStudentSearch(dialog, form) {
  const matches = ref([])
  const selected = ref(null)
  const loading = ref(false)
  let version = 0

  const options = computed(() => {
    const student = selected.value
    if (!student || matches.value.some(item => item.id === student.id)) return matches.value
    return [student, ...matches.value]
  })

  async function search(keyword = '') {
    const current = ++version
    loading.value = true
    try {
      const result = await request({
        url: '/uni/students',
        method: 'get',
        params: { keyword: keyword.trim(), status: 1, page: 1, pageSize: 20 },
        donNotShowLoading: true
      })
      if (current !== version) return
      if (result?.code !== 0) throw new Error(result?.msg || '学生搜索失败')
      matches.value = result.data?.list || []
    } catch (error) {
      if (current === version) ElMessage.error(error.message || '学生搜索失败')
    } finally {
      if (current === version) loading.value = false
    }
  }

  function open(row) {
    selected.value = row ? { id: row.studentId, name: row.studentName, studentNo: row.studentNo } : null
    matches.value = []
    search()
  }

  watch(() => form.value.studentId, id => {
    const student = matches.value.find(item => item.id === id)
    if (student) selected.value = student
  })

  watch(dialog, visible => {
    if (visible) return
    version++
    loading.value = false
  })

  return { options, loading, search, open }
}
