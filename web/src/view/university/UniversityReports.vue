<template>
  <section class="reports">
    <header class="report-heading">
      <div><h2>数据统计与报表中心</h2><p>按时间和组织范围查看历史请假情况。</p></div>
      <div class="report-actions">
        <el-button :disabled="!report || loading" @click="printReport">打印报表</el-button>
        <el-button :disabled="!report || loading" :loading="exporting" @click="exportReport">导出 Excel</el-button>
      </div>
    </header>

    <el-radio-group v-model="preset" aria-label="报表周期" @change="applyPreset">
      <el-radio-button v-for="(label, key) in periods" :key="key" :value="key">{{ label }}</el-radio-button>
    </el-radio-group>
    <form class="report-filters" @submit.prevent="load">
      <label>开始日期 <input v-model="from" type="date" required @input="markCustom" /></label>
      <label>结束日期 <input v-model="to" type="date" required @input="markCustom" /></label>
      <el-select v-model="college" clearable filterable placeholder="全部学院" aria-label="学院" @change="major = null; cls = null">
        <el-option v-for="item in options.colleges" :key="item.id" :label="item.name" :value="item.id" />
      </el-select>
      <el-select v-model="major" clearable filterable placeholder="全部专业" aria-label="专业" @change="cls = null">
        <el-option v-for="item in majors" :key="item.id" :label="item.name" :value="item.id" />
      </el-select>
      <el-select v-model="cls" clearable filterable placeholder="全部班级" aria-label="班级">
        <el-option v-for="item in classes" :key="item.id" :label="item.name" :value="item.id" />
      </el-select>
      <el-select v-model="advisor" clearable filterable placeholder="全部辅导员" aria-label="辅导员">
        <el-option v-for="item in options.advisors" :key="item.id" :label="item.name" :value="item.id" />
      </el-select>
      <el-select v-model="leaveType" clearable placeholder="全部请假类型" aria-label="请假类型">
        <el-option v-for="(label, key) in types" :key="key" :label="label" :value="key" />
      </el-select>
      <el-select v-model="dimension" aria-label="统计维度">
        <el-option v-for="(label, key) in dimensions" :key="key" :label="label" :value="key" />
      </el-select>
      <el-input v-if="dimension === 'student'" v-model="keyword" clearable placeholder="姓名或学号" aria-label="姓名或学号" />
      <el-button native-type="submit" type="primary" :loading="loading">生成报表</el-button>
    </form>
    <el-alert v-if="error" :title="error" type="error" :closable="false" />
    <el-skeleton v-if="loading && !report" :rows="6" animated />
    <template v-if="report">
      <p class="report-period">{{ report.schoolName }} · {{ report.from }} 至 {{ report.to }} · {{ dimensions[report.dimension] }}</p>
      <section class="report-metrics" aria-label="报表汇总">
        <article v-for="metric in metrics" :key="metric[0]"><span>{{ metric[0] }}</span><strong>{{ metric[1] }}</strong></article>
      </section>
      <section class="report-panel">
        <h3>每日已批准请假人数与请假率</h3>
        <v-chart class="report-chart" :option="chart" autoresize />
        <ul class="report-types"><li v-for="(label, key) in types" :key="key">{{ label }} {{ report.types[key] ?? 0 }} 人次</li></ul>
      </section>
      <section class="report-panel">
        <h3>{{ dimensions[report.dimension] }}</h3>
        <p class="report-note">批准涉及比例按本期已批准学生去重人数除以当前分组学生数计算。</p>
        <el-table :data="pageRows" row-key="id" empty-text="无符合条件的学生" max-height="620">
          <el-table-column v-for="field in columns" :key="field[0]" :prop="field[0]" :label="field[1]" :min-width="field[0] === 'name' ? 170 : 118">
            <template #default="{ row }">{{ displayCell(row, field[0]) }}</template>
          </el-table-column>
        </el-table>
        <el-pagination v-if="report.groups.length > pageSize" v-model:current-page="page" :total="report.groups.length"
          :page-size="pageSize" :pager-count="5" layout="total, prev, pager, next" class="report-pagination" />
      </section>
    </template>
  </section>
</template>

<script setup>
import { computed, defineAsyncComponent, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/utils/request'

const VChart = defineAsyncComponent(() => import('./UniversityChart.vue'))

const props = defineProps({ options: { type: Object, required: true } })
const periods = { day: '日报', week: '周报', month: '月报', term: '学期报表', custom: '自定义报表' }
const dimensions = { college: '学院统计', major: '专业统计', class: '班级统计', advisor: '辅导员统计', student: '学生个人统计' }
const types = { sick: '病假', personal: '事假', other: '其他' }
const columns = [
  ['name', '分组'], ['roster', '学生数'], ['requests', '请假人次'], ['students', '涉及学生'],
  ['approved', '已批准学生'], ['reviewed', '已处理申请'], ['approvedMinutes', '已批准时长（小时）'], ['rate', '批准涉及比例']
]
const preset = ref('month')
const from = ref('')
const to = ref('')
const college = ref(null)
const major = ref(null)
const cls = ref(null)
const advisor = ref(null)
const leaveType = ref('')
const dimension = ref('college')
const keyword = ref('')
const report = ref(null)
const loading = ref(false)
const exporting = ref(false)
const error = ref('')
const page = ref(1)
const pageSize = 20
let appliedParams
const majors = computed(() => props.options.majors.filter(item => !college.value || String(item.college_id) === String(college.value)))
const classes = computed(() => props.options.classes.filter(item => major.value
  ? String(item.major_id) === String(major.value) : majors.value.some(parent => String(parent.id) === String(item.major_id))))
const pageRows = computed(() => report.value?.groups.slice((page.value - 1) * pageSize, page.value * pageSize) || [])
const metrics = computed(() => [
  ['请假人次', report.value.requests], ['涉及学生', report.value.students],
  ['日均已批准请假人数', report.value.averageDaily.toFixed(1)],
  ['已批准请假时长（小时）', (report.value.approvedMinutes / 60).toFixed(1)]
])
const chart = computed(() => ({
  tooltip: { trigger: 'axis' }, legend: { bottom: 0 }, grid: { left: 40, right: 40, bottom: 54, top: 22 },
  xAxis: { type: 'category', data: report.value.trend.map(day => day.date) },
  yAxis: [{ type: 'value', minInterval: 1 }, { type: 'value', axisLabel: { formatter: '{value}%' } }],
  series: [
    { name: '全部', type: 'line', showSymbol: report.value.trend.length < 30, data: report.value.trend.map(day => day.count) },
    ...Object.entries(types).map(([key, label]) => ({
      name: label, type: 'line', showSymbol: false, data: report.value.trend.map(day => day.types[key] ?? 0)
    })),
    { name: '请假率', type: 'line', showSymbol: false, yAxisIndex: 1, data: report.value.trend.map(day => Number(day.rate.toFixed(2))) }
  ]
}))
const dateText = date => [date.getFullYear(), date.getMonth() + 1, date.getDate()].map((value, index) =>
  String(value).padStart(index ? 2 : 4, '0')).join('-')
const applyPreset = value => {
  if (value === 'custom') return
  const now = new Date()
  const start = new Date(now)
  if (value === 'week') start.setDate(start.getDate() - (start.getDay() + 6) % 7)
  if (value === 'month') start.setDate(1)
  if (value === 'term') {
    const month = now.getMonth()
    start.setDate(1)
    if (month === 0) start.setFullYear(now.getFullYear() - 1, 8, 1)
    else start.setMonth(month < 8 ? 1 : 8)
  }
  from.value = dateText(start)
  to.value = dateText(now)
}
const markCustom = () => { preset.value = 'custom' }
const displayCell = (row, key) => {
  if (key === 'rate') return row.rate.toFixed(1) + '%'
  if (key === 'approvedMinutes') return (row.approvedMinutes / 60).toFixed(1)
  return row[key]
}
const load = async () => {
  if (loading.value) return
  if (!from.value || !to.value || from.value > to.value) {
    error.value = '请选择有效的日期范围'
    return
  }
  loading.value = true
  error.value = ''
  const params = {
    mode: 'report', from: from.value, to: to.value, dimension: dimension.value,
    collegeId: college.value || undefined, majorId: major.value || undefined,
    classId: cls.value || undefined, advisorId: advisor.value || undefined,
    leaveType: leaveType.value || undefined,
    keyword: dimension.value === 'student' ? keyword.value.trim() || undefined : undefined
  }
  try {
    const response = await request({ url: '/uni/monitor/analytics', params })
    if (response.code !== 0) throw new Error(response.msg)
    report.value = response.data
    appliedParams = params
    page.value = 1
  } catch (cause) {
    error.value = cause.response?.data?.msg || cause.message || '生成报表失败'
  } finally {
    loading.value = false
  }
}
const exportReport = async () => {
  exporting.value = true
  try {
    const result = await request({ url: '/uni/monitor/analytics', params: { ...appliedParams, format: 'xlsx' }, responseType: 'blob' })
    if (!result.data?.type?.includes('spreadsheetml')) throw new Error('报表导出失败')
    const url = URL.createObjectURL(result.data)
    const link = document.createElement('a')
    link.href = url
    link.download = `请假报表-${report.value.from}-${report.value.to}.xlsx`
    link.click()
    setTimeout(() => URL.revokeObjectURL(url), 1000)
  } catch {
    ElMessage.error('报表导出失败，请重试')
  } finally {
    exporting.value = false
  }
}
const printReport = () => {
  const printWindow = window.open('', '_blank', 'popup')
  if (!printWindow) {
    ElMessage.warning('请允许浏览器打开打印窗口')
    return
  }
  printWindow.opener = null
  const doc = printWindow.document
  const add = (parent, tag, text) => {
    const node = doc.createElement(tag)
    if (text != null) node.textContent = String(text)
    parent.appendChild(node)
    return node
  }
  doc.title = '高校请假报表'
  add(doc.head, 'style', 'body{font:14px sans-serif;padding:20px}table{border-collapse:collapse;width:100%;margin-bottom:20px}th,td{border:1px solid #ccc;padding:6px;text-align:left}thead{display:table-header-group}tr{break-inside:avoid}')
  add(doc.body, 'h1', '高校请假报表')
  add(doc.body, 'p', `${report.value.schoolName} · ${report.value.from} 至 ${report.value.to}`)
  add(doc.body, 'p', metrics.value.map(([label, value]) => `${label} ${value}`).join(' · '))
  add(doc.body, 'p', Object.entries(types).map(([key, label]) => `${label} ${report.value.types[key] ?? 0} 人次`).join(' · '))
  const table = (title, headers, records) => {
    add(doc.body, 'h2', title)
    const grid = add(doc.body, 'table')
    const head = add(add(grid, 'thead'), 'tr')
    headers.forEach(label => add(head, 'th', label))
    const body = add(grid, 'tbody')
    records.forEach(values => {
      const row = add(body, 'tr')
      values.forEach(value => add(row, 'td', value))
    })
  }
  table(dimensions[report.value.dimension], columns.map(field => field[1]), report.value.groups.map(row => columns.map(field => displayCell(row, field[0]))))
  table('每日已批准请假人数', ['日期', '人数', '请假率', ...Object.values(types)],
    report.value.trend.map(day => [day.date, day.count, day.rate.toFixed(1) + '%', ...Object.keys(types).map(key => day.types[key] ?? 0)]))
  printWindow.addEventListener('afterprint', () => printWindow.close(), { once: true })
  printWindow.focus()
  printWindow.print()
}
onMounted(() => { applyPreset(preset.value); load() })
</script>

<style scoped>
.reports { display: grid; gap: 16px; min-width: 0; }
.report-heading, .report-actions, .report-filters { display: flex; flex-wrap: wrap; align-items: center; gap: 10px; }
.report-heading { justify-content: space-between; }
.report-heading h2, .report-panel h3 { margin: 0; }
.report-heading p, .report-period, .report-note, .report-metrics span { color: var(--el-text-color-secondary); }
.report-filters, .report-panel, .report-metrics article { background: var(--el-bg-color); border: 1px solid var(--el-border-color-lighter); padding: 16px; border-radius: 12px; min-width: 0; }
.report-filters .el-select, .report-filters .el-input { width: 150px; }
.report-filters label { display: flex; align-items: center; gap: 6px; }
.report-filters input { background: var(--el-bg-color); color: var(--el-text-color-primary); border: 1px solid var(--el-border-color); border-radius: 6px; padding: 5px; }
.report-metrics { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 12px; }
.report-metrics strong { display: block; margin-top: 8px; font-size: 26px; color: var(--el-color-primary); }
.report-chart { height: 260px; }
.report-types { display: flex; flex-wrap: wrap; gap: 16px; padding: 0; list-style: none; }
.report-pagination { padding-top: 12px; justify-content: flex-end; flex-wrap: wrap; }
</style>
