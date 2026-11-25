<template>
<main class="dash university-workspace" v-loading="loading && !data">
<header class="workspace-heading"><div><h1>{{ admin ? '高校数据工作台' : '请假数据看板' }}</h1><p>{{ data?.tenantName }} · 数据统计与请假分析</p></div>
  <div class="workspace-actions"><el-button :loading="loading" @click="refresh">刷新数据</el-button>
  <el-button @click="resetFilters">清除筛选</el-button></div></header>
<el-radio-group v-model="currentView" aria-label="统计页面">
  <el-radio-button value="dashboard">数据工作台</el-radio-button>
  <el-radio-button value="reports">统计报表</el-radio-button>
</el-radio-group>
<UniversityReports v-if="currentView === 'reports'" :options="options" />
<template v-else>
<el-alert v-if="error" :title="error" type="error" :closable="false" show-icon/>
<section v-if="admin && data?.overview" class="cards"><div v-for="(label,key) in overviewLabels" :key="key" class="tile"><span>{{label}}</span><strong>{{data.overview[key]}}</strong></div></section>
<section class="panel filters"><el-select v-model="college" clearable placeholder="全部学院" @change="major=''; cls=''; refresh()"><el-option v-for="x in options.colleges" :key="x.id" :label="x.name" :value="x.id"/></el-select><el-select v-model="major" clearable placeholder="全部专业" @change="cls=''; refresh()"><el-option v-for="x in majors" :key="x.id" :label="x.name" :value="x.id"/></el-select><el-select v-model="cls" clearable placeholder="全部班级" @change="refresh"><el-option v-for="x in classes" :key="x.id" :label="x.name" :value="x.id"/></el-select><el-radio-group v-model="days" @change="refresh"><el-radio-button :value="7">近7天</el-radio-button><el-radio-button :value="30">近30天</el-radio-button><el-radio-button :value="90">近90天</el-radio-button></el-radio-group></section>
<p class="muted">基础总览为全校数据；请假指标按组织筛选。趋势及分布按所选周期，今日指标及待审批不受周期影响。更新：{{data?.asOf || '—'}}（数据库时间）</p>
<section class="cards"><div v-for="(label,key) in labels" :key="key" class="tile"><span>{{label}}</span><strong>{{data ? (key==='rate' ? data.rate.toFixed(2)+'%' : data[key]) : '—'}}</strong></div></section>
<section v-if="data" class="panel" v-loading="loading"><h3>每日已批准请假人数</h3><v-chart class="chart" :option="trend" autoresize/></section>
<section v-if="data" class="pair"><div class="panel"><h3>请假类型分布（记录）</h3><v-chart v-if="data.types.length" class="chart" :option="pie(data.types)" autoresize/><el-empty v-else description="暂无记录"/></div><div class="panel"><h3>审批状态分布（记录）</h3><v-chart v-if="data.statuses.length" class="chart" :option="pie(data.statuses)" autoresize/><el-empty v-else description="暂无记录"/></div></section>
<section v-if="data" class="panel" v-for="(label,key) in rankings" :key="key"><h3>{{label}}（今日 · 前20）</h3><el-table :data="data[key]" stripe><el-table-column prop="name" label="名称" min-width="150"/><el-table-column prop="total" label="学生数"/><el-table-column prop="count" label="请假人数"/><el-table-column label="请假率"><template #default="{row}">{{row.total ? (row.count/row.total*100).toFixed(2) : '0.00'}}%</template></el-table-column><el-table-column prop="pending" label="待审批"/></el-table></section>
<section v-if="data" class="panel"><h3>最近请假申请（所选周期）</h3><el-table :data="data.recent" stripe><el-table-column prop="name" label="学生"/><el-table-column prop="className" label="班级"/><el-table-column label="类型"><template #default="{row}">{{names[row.leaveType] || row.leaveType}}</template></el-table-column><el-table-column label="状态"><template #default="{row}">{{names[row.status] || row.status}}</template></el-table-column><el-table-column prop="createdAt" label="申请时间" min-width="180"/></el-table></section>
<section class="panel"><h3>快捷入口</h3><el-button v-for="x in links" :key="x[0]" @click="router.push({name:x[0]})">{{x[1]}}</el-button></section>
</template>
</main>
</template>
<script setup>
import {defineAsyncComponent,ref,computed,onMounted,onUnmounted,onActivated,onDeactivated} from 'vue'
import { useSessionStorage, StorageSerializers } from '@vueuse/core'
import {useRouter} from 'vue-router'
import request from '@/utils/request'
import VChart from 'vue-echarts'
import {use} from 'echarts/core'
import {CanvasRenderer} from 'echarts/renderers'
import {LineChart,PieChart} from 'echarts/charts'
import {GridComponent,TooltipComponent,LegendComponent} from 'echarts/components'
use([CanvasRenderer,LineChart,PieChart,GridComponent,TooltipComponent,LegendComponent])
const props=defineProps({admin:Boolean})
const UniversityReports = defineAsyncComponent(() => import('./UniversityReports.vue'))
const currentView = ref('dashboard')
const router=useRouter(), data=ref(null), loading=ref(false), error=ref(''), days=useSessionStorage('university-dashboard-days', 7), college=useSessionStorage('university-dashboard-college', '', { serializer: StorageSerializers.object }), major=useSessionStorage('university-dashboard-major', '', { serializer: StorageSerializers.object }), cls=useSessionStorage('university-dashboard-class', '', { serializer: StorageSerializers.object }), options=ref({colleges:[],majors:[],classes:[],advisors:[]})
const majors=computed(()=>options.value.majors.filter(x=>!college.value||String(x.college_id)===String(college.value)))
const classes=computed(()=>options.value.classes.filter(x=>major.value?String(x.major_id)===String(major.value):majors.value.some(m=>String(m.id)===String(x.major_id))))
const overviewLabels={students:'学生总数',teachers:'教师总数',colleges:'学院总数',majors:'专业总数',classes:'班级总数'},labels={today:'今日请假人数',current:'正在请假人数',pending:'待审批申请',rate:'今日请假率'},rankings={collegeRank:'学院请假排行',classRank:'班级请假排行',advisorRank:'辅导员统计'},names={sick:'病假',personal:'事假',other:'其他',approved:'已批准',pending:'待审批',rejected:'已驳回',cancelled:'已取消'}
const links=computed(()=>props.admin?[['universityStudent','学生管理'],['universityTeacher','教师管理'],['universityOrganization','组织架构'],['universityLeave','请假管理']]:[['universityLeaveMonitor','请假监管']])
const trend=computed(()=>({tooltip:{trigger:'axis'},grid:{left:45,right:20,bottom:35,top:20},xAxis:{type:'category',data:data.value?.trend.map(x=>x.name.slice(5))||[]},yAxis:{type:'value',minInterval:1},series:[{type:'line',smooth:true,areaStyle:{opacity:.12},data:data.value?.trend.map(x=>x.count)||[]}]}))
const pie=rows=>({tooltip:{trigger:'item'},legend:{bottom:0},series:[{type:'pie',radius:['40%','68%'],center:['50%','43%'],data:rows.map(x=>({name:names[x.name]||x.name,value:x.count}))}]})
let seq=0, timer, active=true
async function refresh(){const id=++seq;loading.value=true;error.value='';try{const params={days:days.value};if(college.value)params.collegeId=college.value;if(major.value)params.majorId=major.value;if(cls.value)params.classId=cls.value;const r=await request({url:'/uni/monitor/analytics',method:'get',params});if(id!==seq||!active)return;if(r.code!==0)throw Error(r.msg);data.value=r.data}catch(e){if(id===seq&&active)error.value='统计刷新失败，请重试（保留上次数据）'}finally{if(id===seq)loading.value=false}}
function resetFilters() {
  college.value = ''
  major.value = ''
  cls.value = ''
  days.value = 7
  refresh()
}
function start(){active=true;if(!timer)timer=setInterval(()=>{if(!loading.value && currentView.value === 'dashboard')refresh()},60000)}
function stop(){active=false;++seq;clearInterval(timer);timer=null;loading.value=false}
onMounted(async()=>{start();try{const r=await request({url:'/uni/monitor/options',method:'get'});if(r.code===0)options.value=r.data}catch(e){error.value='组织筛选加载失败'}if(active)refresh()});onActivated(()=>{start();if(data.value)refresh()});onDeactivated(stop);onUnmounted(stop)
</script>
<style scoped>
.dash{padding:24px;display:flex;flex-direction:column;gap:16px;color:var(--el-text-color-primary);min-width:0}.dash h1{font-size:24px;margin:0}.dash h3{font-size:16px;margin:0 0 14px}.dash p{margin:6px 0;color:var(--el-text-color-secondary);font-size:13px}.cards{display:grid;grid-template-columns:repeat(auto-fit,minmax(150px,1fr));gap:14px}.tile,.panel{background:var(--el-bg-color);border:1px solid var(--el-border-color-lighter);border-radius:12px;padding:18px;min-width:0}.tile span{color:var(--el-text-color-secondary);font-size:13px}.tile strong{display:block;font-size:30px;margin-top:12px;color:var(--el-color-primary)}.tile:nth-child(2) strong{color:var(--el-color-success)}.tile:nth-child(3) strong{color:var(--el-color-warning)}.pair{display:grid;grid-template-columns:1fr 1fr;gap:16px}.chart{height:280px}.filters{display:flex;gap:12px;flex-wrap:wrap}.filters .el-select{width:170px}@media(max-width:900px){.pair{grid-template-columns:1fr}.dash{padding:12px}}
</style>

<style scoped src="./adminWorkspace.css"></style>
