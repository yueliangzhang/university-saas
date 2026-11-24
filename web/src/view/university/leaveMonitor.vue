<template>
  <main class="monitor">
    <header class="heading"><div><h1>学生请假监管</h1><p>全校学生请假记录查询与实时监管</p></div><el-tag type="info" effect="plain">只读权限</el-tag></header>
    <section class="metrics">
      <div v-for="(label,key) in labels" :key="key" class="metric" :class="key">
        <div class="metric-top"><span>{{label}}</span><span class="symbol">{{symbols[key]}}</span></div>
        <strong>{{stats[key] ?? 0}} <small>条</small></strong><p>{{notes[key]}}</p>
      </div>
    </section>
    <section class="panel">
      <div class="toolbar"><h2>请假记录</h2><el-button :loading="exporting" @click="download">导出 Excel</el-button></div>
      <div class="quick"><el-radio-group v-model="quick" @change="applyQuick"><el-radio-button v-for="v in ['全部','今日','本周','本月','正在请假']" :key="v" :label="v">{{v}}</el-radio-button></el-radio-group></div>
      <el-form label-position="top" @submit.prevent="search">
        <div class="filters">
          <el-form-item v-for="(label,key) in optionLabels" :key="key" :label="label"><el-select v-model="filters[key]" clearable filterable placeholder="全部" @change="changed(key)"><el-option v-for="o in choices(key)" :key="o.id" :label="o.name" :value="o.id"/></el-select></el-form-item>
          <el-form-item label="审批状态"><el-select v-model="filters.status" clearable placeholder="全部状态"><el-option v-for="(v,k) in statuses" :key="k" :label="v" :value="k"/></el-select></el-form-item>
          <el-form-item label="请假类型"><el-select v-model="filters.leaveType" clearable placeholder="全部类型"><el-option v-for="(v,k) in types" :key="k" :label="v" :value="k"/></el-select></el-form-item>
          <el-form-item label="请假日期" class="wide"><el-date-picker v-model="dates" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始日期" end-placeholder="结束日期" @change="quick=''"/></el-form-item>
          <el-form-item label="关键词" class="wide"><el-input v-model="filters.keyword" placeholder="姓名、学号、辅导员" clearable/></el-form-item>
          <div class="actions wide"><el-checkbox v-model="current" @change="quick=''">正在请假</el-checkbox><el-button type="primary" native-type="submit" :loading="loading">查询</el-button><el-button @click="reset">重置</el-button></div>
        </div>
      </el-form>
      <el-table :data="rows" v-loading="loading" stripe size="small" empty-text="暂无符合条件的请假记录" style="width:100%">
        <el-table-column label="学生信息" min-width="115"><template #default="{row}"><div class="primary">{{row.name}}</div><div class="secondary">{{row.studentNo}}</div></template></el-table-column>
        <el-table-column label="学院 / 专业" min-width="115"><template #default="{row}"><div>{{row.college}}</div><div class="secondary">{{row.major}}</div></template></el-table-column>
        <el-table-column label="班级 / 辅导员" min-width="125"><template #default="{row}"><div>{{row.className}}</div><div class="secondary">{{row.advisor || '未分配'}}</div></template></el-table-column>
        <el-table-column label="类型" width="70"><template #default="{row}">{{types[row.leaveType] || row.leaveType}}</template></el-table-column>
        <el-table-column label="状态" width="92"><template #default="{row}"><el-tag :type="tagTypes[row.status] || 'info'" size="small" effect="light">{{statuses[row.status] || row.status}}</el-tag></template></el-table-column>
        <el-table-column v-for="field in ['startAt','endAt']" :key="field" :label="field==='startAt'?'开始时间':'结束时间'" min-width="112"><template #default="{row}"><div class="date">{{fmt(row[field]).slice(0,10)}}</div><div class="secondary">{{fmt(row[field]).slice(11)}}</div></template></el-table-column>
        <el-table-column label="时长" min-width="85"><template #default="{row}">{{duration(row)}}</template></el-table-column>
        <el-table-column label="操作" width="65"><template #default="{row}"><el-button link type="primary" @click="detail(row)">详情</el-button></template></el-table-column>
      </el-table>
      <div class="pagination"><el-pagination v-model:current-page="page" v-model:page-size="size" :total="total" :page-sizes="[20,50,100]" layout="total,sizes,prev,pager,next" @current-change="load" @size-change="search"/></div>
    </section>
    <el-drawer v-model="visible" title="请假详情" size="min(520px, 100vw)"><template v-if="selected"><h3>{{selected.name}} · {{selected.studentNo}}</h3><el-descriptions :column="1" border><el-descriptions-item v-for="(label,key) in columns" :key="key" :label="label">{{selected[key] || '—'}}</el-descriptions-item><el-descriptions-item label="请假类型">{{types[selected.leaveType]}}</el-descriptions-item><el-descriptions-item label="请假原因">{{selected.reason}}</el-descriptions-item><el-descriptions-item label="开始时间">{{fmt(selected.startAt)}}</el-descriptions-item><el-descriptions-item label="结束时间">{{fmt(selected.endAt)}}</el-descriptions-item><el-descriptions-item label="请假时长">{{duration(selected)}}</el-descriptions-item><el-descriptions-item label="申请时间">{{fmt(selected.createdAt)}}</el-descriptions-item><el-descriptions-item label="审批状态"><el-tag :type="tagTypes[selected.status] || 'info'">{{statuses[selected.status]}}</el-tag></el-descriptions-item><el-descriptions-item label="审批备注">{{selected.reviewNote || '—'}}</el-descriptions-item><el-descriptions-item label="审批时间">{{fmt(selected.reviewedAt)}}</el-descriptions-item></el-descriptions></template></el-drawer>
  </main>
</template>
<script setup>
import {ref,reactive,onMounted} from 'vue'
import request from '@/utils/request'
import {ElMessage} from 'element-plus'
const labels={today:'今日已批准请假记录',current:'正在请假记录',pending:'待审批记录',total:'累计记录'}, columns={studentNo:'学号',name:'姓名',college:'学院',major:'专业',className:'班级',advisor:'辅导员'}, statuses={pending:'待审批',approved:'已批准',rejected:'已驳回',cancelled:'已取消'},types={sick:'病假',personal:'事假',other:'其他'},optionLabels={collegeId:'学院',majorId:'专业',classId:'班级',advisorId:'辅导员'}
const filters=reactive({}),dates=ref([]),current=ref(false),page=ref(1),size=ref(20),total=ref(0),rows=ref([]),stats=ref({}),options=ref({}),loading=ref(false),visible=ref(false),selected=ref(null)
const get=(path,params={},extra={})=>request({url:'/uni/monitor/'+path,method:'get',params,...extra})
const choices=k=>{const a=options.value[{collegeId:'colleges',majorId:'majors',classId:'classes',advisorId:'advisors'}[k]]||[];return a.filter(o=>k==='majorId'&&filters.collegeId?o.college_id===filters.collegeId:k==='classId'&&filters.majorId?o.major_id===filters.majorId:k==='classId'&&filters.collegeId?(options.value.majors||[]).some(m=>m.id===o.major_id&&m.college_id===filters.collegeId):true)}
function changed(k){if(k==='collegeId'){delete filters.majorId;delete filters.classId}if(k==='majorId')delete filters.classId}
const params=()=>({...Object.fromEntries(Object.entries(filters).filter(([,v])=>v!==''&&v!=null)),start:dates.value?.[0],end:dates.value?.[1],current:current.value?'1':undefined})
let seq=0;async function load(){const id=++seq;loading.value=true;try{const r=await get('leaves',{...params(),page:page.value,pageSize:size.value});if(id===seq&&r.code===0){rows.value=r.data.list;total.value=r.data.total}}finally{if(id===seq)loading.value=false}}
async function search(){page.value=1;await load();const r=await get('dashboard',params());if(r.code===0)stats.value=r.data}
function reset(){quick.value="全部";Object.keys(filters).forEach(k=>delete filters[k]);dates.value=[];current.value=false;search()}
const fmt=v=>v?String(v).replace('T',' ').slice(0,19):'—'
async function detail(row){const r=await get('detail',{id:row.id});if(r.code===0){selected.value=r.data;visible.value=true}}
async function download(){if(exporting.value)return;exporting.value=true;try{const r=await get('export',params(),{responseType:'blob'});const b=r.data;if(!(b instanceof Blob)||!b.type.includes('spreadsheet')){ElMessage.error('导出失败');return}const u=URL.createObjectURL(b),a=document.createElement('a');a.href=u;a.download='leave-monitor.xlsx';a.click();setTimeout(()=>URL.revokeObjectURL(u),1000)}finally{exporting.value=false}}
const quick=ref('全部'), exporting=ref(false)
const symbols={today:'▣',current:'◷',pending:'≡',total:'▤'}
const notes={today:'今日涉及的已批准记录',current:'当前处于请假期间',pending:'等待审批的申请',total:'累计请假记录'}
const tagTypes={approved:'success',pending:'warning',rejected:'danger',cancelled:'info'}
const localDate=d=>`${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`
function applyQuick(v){
  current.value=v==='正在请假'; dates.value=[]
  const a=new Date(),b=new Date(a)
  if(v==='今日') dates.value=[localDate(a),localDate(b)]
  if(v==='本周'){a.setDate(a.getDate()-(a.getDay()+6)%7);b.setTime(a.getTime());b.setDate(b.getDate()+6);dates.value=[localDate(a),localDate(b)]}
  if(v==='本月'){a.setDate(1);b.setMonth(b.getMonth()+1,0);dates.value=[localDate(a),localDate(b)]}
  search()
}
function duration(row){const a=Date.parse(String(row.startAt||'').replace(' ','T')),b=Date.parse(String(row.endAt||'').replace(' ','T'));if(!Number.isFinite(a)||!Number.isFinite(b)||b<a)return '—';const m=Math.round((b-a)/60000);return [Math.floor(m/1440)?Math.floor(m/1440)+'天':'',Math.floor(m%1440/60)?Math.floor(m%1440/60)+'小时':'',m%60?m%60+'分':''].filter(Boolean).join(' ')||'0分钟'}
onMounted(async()=>{const r=await get('options');if(r.code===0)options.value=r.data;await search()})
</script>

<style scoped>
.monitor{padding:24px;min-width:0;color:var(--el-text-color-primary)}
.heading,.toolbar,.metric-top{display:flex;align-items:center;justify-content:space-between;gap:12px}.heading{margin-bottom:20px}.heading h1{font-size:24px;font-weight:700;margin:0 0 6px}.heading p,.metric p{margin:0;color:var(--el-text-color-secondary);font-size:13px}.metrics{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:16px;margin-bottom:20px}.metric,.panel{background:var(--el-bg-color);border:1px solid var(--el-border-color-lighter);border-radius:12px}.metric{padding:18px;--accent:#2563eb}.metric.current{--accent:#059669}.metric.pending{--accent:#d97706}.metric.total{--accent:#7c3aed}.metric-top{font-size:14px;color:var(--el-text-color-regular)}.symbol{font-size:24px;color:var(--accent);background:var(--el-fill-color-light);width:38px;height:38px;display:grid;place-items:center;border-radius:9px}.metric strong{display:block;font-size:32px;line-height:1.4;color:var(--accent);font-variant-numeric:tabular-nums}.metric small{font-size:12px;font-weight:400;color:var(--el-text-color-secondary)}.panel{padding:20px;min-width:0}.toolbar h2{font-size:18px;font-weight:600;margin:0}.quick{margin:16px 0}.filters{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:0 16px}.filters :deep(.el-select),.filters :deep(.el-date-editor){width:100%}.filters :deep(.el-form-item){margin-bottom:14px}.wide{grid-column:span 2}.actions{display:flex;align-items:center;justify-content:flex-end;gap:8px;flex-wrap:wrap;padding-bottom:12px}.actions :deep(.el-button){margin-left:0}.primary{font-weight:600}.secondary{font-size:12px;color:var(--el-text-color-secondary)}.date{font-variant-numeric:tabular-nums;white-space:nowrap}.panel :deep(.el-table .cell){line-height:19px;overflow-wrap:anywhere}.panel :deep(.el-table th.el-table__cell){background:var(--el-fill-color-light)}.panel :deep(.el-table td.el-table__cell){padding:9px 0}.pagination{display:flex;justify-content:flex-end;overflow:auto;margin-top:18px}@media(max-width:1100px){.metrics,.filters{grid-template-columns:repeat(2,minmax(0,1fr))}}@media(max-width:600px){.monitor,.panel{padding:12px}.metrics,.filters{grid-template-columns:1fr}.wide{grid-column:span 1}.heading h1{font-size:20px}.quick{overflow:auto}.quick :deep(.el-radio-group){flex-wrap:nowrap}.pagination{justify-content:flex-start}}
</style>
