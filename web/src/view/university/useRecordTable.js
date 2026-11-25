import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

export function useRecordTable(deleteRecords, reload) {
  const selectedRows = ref([])
  const deleting = ref(false)
  const sort = ref({})

  const setSort = ({ prop, order }) => {
    sort.value = order
      ? { sortBy: prop, sortOrder: order === 'ascending' ? 'asc' : 'desc' }
      : {}
  }

  const removeSelected = async () => {
    const ids = selectedRows.value.map(row => row.id)
    if (!ids.length || deleting.value) return

    deleting.value = true
    try {
      await ElMessageBox.confirm(
        `确定删除选中的 ${ids.length} 条记录？`,
        '确认批量删除',
        { type: 'warning' }
      )
      const result = await deleteRecords(ids.join(','))
      if (result?.code !== 0) throw new Error(result?.msg || '批量删除失败')
      selectedRows.value = []
      ElMessage.success(`已删除 ${ids.length} 条记录`)
      await reload()
    } catch (error) {
      if (error !== 'cancel' && error !== 'close' && !error.isAxiosError) {
        ElMessage.error(error.message || '批量删除失败')
      }
    } finally {
      deleting.value = false
    }
  }

  return { selectedRows, deleting, removeSelected, sort, setSort }
}
