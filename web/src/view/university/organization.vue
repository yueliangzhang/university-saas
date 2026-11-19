<template>
  <div class="uni-org-v1764">
    <header class="org-heading">
      <div class="org-heading-copy">
        <div class="org-eyebrow">UNIVERSITY / ORGANIZATION</div>
        <h1>高校组织架构</h1>
        <p>统一管理学院、专业与班级，建立清晰的校园组织体系。</p>
      </div>
      <div class="org-heading-symbol" aria-hidden="true">组织管理</div>
    </header>

    <section class="org-stats" aria-label="组织架构概览">
      <div v-for="item in [
        {key:'colleges',title:'学院总数',unit:'个',symbol:'院'},
        {key:'majors',title:'专业总数',unit:'个',symbol:'专'},
        {key:'classes',title:'班级总数',unit:'个',symbol:'班'}
      ]" :key="item.key" class="org-stat">
        <div class="org-stat-top"><span class="org-stat-icon">{{ item.symbol }}</span><span class="org-stat-label">{{ item.title }}</span></div>
        <div class="org-stat-value">{{ overview[item.key] ?? '—' }}<small v-if="overview[item.key]!==null">{{ item.unit }}</small></div>
        <div class="org-stat-hint">{{ overview[item.key]===null ? '暂无统计数据' : '来自学校组织数据' }}</div>
      </div>
    </section>

    <section class="org-panel">
      <div class="org-panel-head">
        <div><h2>组织信息管理</h2><p>切换分类，查看与维护学校组织数据</p></div>
        <el-button type="primary" class="org-create" @click="openCreate">＋ 新增{{ labels[tab] }}</el-button>
      </div>
      <div class="org-tabs" role="tablist" aria-label="组织分类">
        <button v-for="item in [
          {key:'colleges',label:'学院',symbol:'院'},
          {key:'majors',label:'专业',symbol:'专'},
          {key:'classes',label:'班级',symbol:'班'}
        ]" :key="item.key" type="button" class="org-tab" :class="{active:tab===item.key}"
          :aria-selected="tab===item.key" role="tab" @click="tab=item.key;changeTab()">
          <span class="org-tab-symbol">{{ item.symbol }}</span>{{ item.label }}
          <span class="org-tab-count">{{ overview[item.key] ?? '—' }}</span>
        </button>
      </div>
      <div class="org-toolbar">
        <div class="org-filters">
          <el-input v-model="keyword" clearable placeholder="搜索名称或编码" class="org-search" />
          <el-select v-model="statusFilter" placeholder="全部状态" class="org-status-filter">
            <el-option label="全部状态" value="" />
            <el-option label="启用" value="1" />
            <el-option label="停用" value="2" />
          </el-select>
        </div>
        <div class="org-toolbar-right">
          <span>共 {{ filteredRows.length }} 条{{ labels[tab] }}记录</span>
          <el-button :loading="loading" @click="load();refreshOverview()">刷新</el-button>
        </div>
      </div>
      <el-alert v-if="error" :title="error" type="error" :closable="false" class="org-error" />
      <el-table v-loading="loading" :data="filteredRows" class="org-table" row-key="id" empty-text="暂无符合条件的组织数据">
        <el-table-column prop="code" label="组织编码" min-width="130">
          <template #default="{row}"><span class="org-code">{{ row.code }}</span></template>
        </el-table-column>
        <el-table-column prop="name" label="名称" min-width="190">
          <template #default="{row}"><strong class="org-name">{{ row.name }}</strong></template>
        </el-table-column>
        <el-table-column v-if="tab==='majors'" label="所属学院" min-width="150">
          <template #default="{row}">{{ collegeLabel(row.collegeId) }}</template>
        </el-table-column>
        <el-table-column v-if="tab==='classes'" label="所属专业" min-width="150">
          <template #default="{row}">{{ majorLabel(row.majorId) }}</template>
        </el-table-column>
        <el-table-column v-if="tab==='classes'" prop="enrollmentYear" label="入学年份" min-width="110" />
        <el-table-column label="状态" width="115">
          <template #default="{row}"><el-tag :type="row.status===1?'success':'info'" effect="light" round>{{ row.status===1?'启用':'停用' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{row}">
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="org-table-footer">显示 {{ filteredRows.length }} 条记录 · {{ labels[tab] }}管理</div>
    </section>

    <el-dialog v-model="dialog" :title="(editing?'编辑':'新增')+labels[tab]" width="min(520px, 94vw)" append-to-body>
      <el-form :model="form" label-width="100px" class="org-form">
        <el-form-item label="编码" required><el-input v-model="form.code" maxlength="64" placeholder="请输入编码" /></el-form-item>
        <el-form-item label="名称" required><el-input v-model="form.name" maxlength="191" placeholder="请输入名称" /></el-form-item>
        <el-form-item v-if="tab==='majors'" label="所属学院" required>
          <el-select v-model="form.collegeId" placeholder="选择学院" style="width:100%">
            <el-option v-for="x in colleges" :key="x.id" :label="`${x.name} (${x.code})`" :value="x.id" :disabled="x.status!==1" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="tab==='classes'" label="所属专业" required>
          <el-select v-model="form.majorId" placeholder="选择专业" style="width:100%">
            <el-option v-for="x in majors" :key="x.id" :label="`${x.name} (${x.code})`" :value="x.id" :disabled="x.status!==1" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="tab==='classes'" label="入学年份"><el-input-number v-model="form.enrollmentYear" :min="1900" :max="2200" /></el-form-item>
        <el-form-item label="状态"><el-radio-group v-model="form.status"><el-radio :value="1">启用</el-radio><el-radio :value="2">停用</el-radio></el-radio-group></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
    </el-dialog>
  </div>
</template>
<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
const labels={colleges:'学院',majors:'专业',classes:'班级'}
const tab=ref('colleges'),rows=ref([]),colleges=ref([]),majors=ref([])
const loading=ref(false),saving=ref(false),error=ref(''),dialog=ref(false),editing=ref(false)
const form=ref({})
const keyword=ref('')
const statusFilter=ref('')
const overview=ref({colleges:null,majors:null,classes:null})
const collegeNames=ref({})
const majorNames=ref({})
const filteredRows=computed(()=>rows.value.filter(row=>{
  const q=keyword.value.trim().toLowerCase()
  return (!q || String(row.name||'').toLowerCase().includes(q) || String(row.code||'').toLowerCase().includes(q))
    && (statusFilter.value==='' || row.status===Number(statusFilter.value))
}))
const collegeLabel=id=>collegeNames.value[id] || `学院 #${id}`
const majorLabel=id=>majorNames.value[id] || `专业 #${id}`
const refreshOverview=async()=>{
  const kinds=['colleges','majors','classes']
  const result=await Promise.allSettled(kinds.map(fetchList))
  result.forEach((r,i)=>{
    overview.value[kinds[i]]=r.status==='fulfilled'?r.value.length:null
    if(r.status==='fulfilled'&&kinds[i]==='colleges')collegeNames.value=Object.fromEntries(r.value.map(x=>[x.id,x.name]))
    if(r.status==='fulfilled'&&kinds[i]==='majors')majorNames.value=Object.fromEntries(r.value.map(x=>[x.id,x.name]))
  })
}
const changeTab=()=>{keyword.value='';statusFilter.value='';load()}

const api=(kind,method='get',data,params)=>request({url:`/uni/org/${kind}`,method,data,params})
const fetchList=async kind=>{const r=await api(kind);if(r?.code!==0)throw new Error(r?.msg||'加载失败');return r.data?.list||[]}
const load=async()=>{loading.value=true;error.value='';try{rows.value=await fetchList(tab.value)}catch(e){error.value=e.message}finally{loading.value=false}}
const openCreate=async()=>{editing.value=false;form.value={code:'',name:'',status:1,collegeId:null,majorId:null,enrollmentYear:new Date().getFullYear()};await openDialog()}
const openEdit=async row=>{editing.value=true;form.value={...row};await openDialog()}
const openDialog=async()=>{try{if(tab.value==='majors')colleges.value=await fetchList('colleges');if(tab.value==='classes')majors.value=await fetchList('majors');dialog.value=true}catch(e){ElMessage.error(e.message)}}
const save=async()=>{if(!form.value.code?.trim()||!form.value.name?.trim()){ElMessage.warning('请填写编码和名称');return}
 saving.value=true;try{const r=await api(tab.value,editing.value?'put':'post',form.value);if(r?.code!==0)throw new Error(r?.msg||'保存失败');ElMessage.success('保存成功');dialog.value=false;await load()}catch(e){ElMessage.error(e.message)}finally{saving.value=false}}
const remove=async row=>{try{await ElMessageBox.confirm(`确定删除「${row.name}」？关联的下级记录必须先删除。`,'确认删除',{type:'warning'});const r=await api(tab.value,'delete',undefined,{id:row.id});if(r?.code!==0)throw new Error(r?.msg||'删除失败');ElMessage.success('删除成功');await load();await refreshOverview()}catch(e){if(e!=='cancel')ElMessage.error(e.message||'操作失败')}}
onMounted(async()=>{await Promise.all([load(),refreshOverview()])})
</script>
<style scoped>
.uni-org-v1764{padding:20px 24px 32px;min-width:0;color:#1d2c45}
.org-heading{position:relative;display:flex;justify-content:space-between;align-items:center;gap:20px;padding:28px 30px;border:1px solid #dce7ff;border-radius:18px;background:linear-gradient(110deg,#fff 15%,#eef3ff 100%);overflow:hidden}
.org-eyebrow{font-size:11px;letter-spacing:1.6px;font-weight:800;color:#4269c9;margin-bottom:12px}
.org-heading h1{font-size:25px;line-height:1.3;margin:0;font-weight:750;color:#192d4f}
.org-heading p{margin:9px 0 0;color:#6a7a92;font-size:13px}
.org-heading-symbol{font-size:36px;font-weight:800;letter-spacing:4px;color:#d9e4fb;white-space:nowrap}
.org-stats{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:16px;margin:20px 0}
.org-stat{background:#fff;border:1px solid #e6ebf4;border-radius:16px;padding:21px 23px;box-shadow:0 6px 22px rgba(34,61,107,.035)}
.org-stat-top{display:flex;align-items:center;gap:11px}
.org-stat-icon{display:grid;place-items:center;width:34px;height:34px;border-radius:10px;background:#eaf1ff;color:#315fda;font-size:13px;font-weight:700}
.org-stat-label{color:#697a94;font-size:13px;font-weight:600}
.org-stat-value{font-size:29px;line-height:1.35;font-weight:800;color:#1d2c45;margin:14px 0 4px;font-variant-numeric:tabular-nums}
.org-stat-value small{font-size:13px;color:#8190a4;margin-left:5px;font-weight:500}
.org-stat-hint{font-size:12px;color:#97a3b6}
.org-panel{background:#fff;border:1px solid #e5ebf5;border-radius:17px;box-shadow:0 7px 28px rgba(34,61,107,.035);overflow:hidden}
.org-panel-head{display:flex;align-items:center;justify-content:space-between;gap:12px;padding:23px 24px 18px}
.org-panel-head h2{font-size:17px;font-weight:750;margin:0;color:#1d2c45}
.org-panel-head p{margin:7px 0 0;font-size:12px;color:#8490a4}
.org-tabs{display:flex;gap:8px;padding:0 24px 18px;border-bottom:1px solid #eef1f6;flex-wrap:wrap}
.org-tab{display:inline-flex;align-items:center;gap:8px;border:1px solid #e4eaf5;background:#f8faff;color:#5f6f87;border-radius:10px;padding:10px 14px;font-size:13px;font-weight:650;cursor:pointer}
.org-tab.active{color:#2759db;background:#eaf1ff;border-color:#bcd0ff}
.org-tab-symbol{font-size:11px;font-weight:700}
.org-tab-count{background:#fff;border-radius:6px;padding:1px 6px;font-size:11px;color:#6c7c94}
.org-toolbar{display:flex;align-items:center;justify-content:space-between;gap:12px;padding:18px 24px;flex-wrap:wrap}
.org-filters{display:flex;align-items:center;gap:10px;flex-wrap:wrap}
.org-search{width:255px}.org-status-filter{width:142px}
.org-toolbar-right{display:flex;align-items:center;gap:12px;font-size:12px;color:#8490a4}
.org-error{margin:0 24px 12px;width:calc(100% - 48px)}
.org-table{width:calc(100% - 48px);margin:0 24px}
.org-code{font-family:ui-monospace,SFMono-Regular,Menlo,monospace;background:#f2f5fa;color:#46607e;padding:5px 9px;border-radius:6px;font-size:12px}
.org-name{font-size:13px;color:#273a55}
.org-table-footer{padding:17px 24px;color:#8794a8;font-size:12px}
:deep(.org-table th.el-table__cell){background:#f7f9fd!important;color:#5b6d87;font-weight:700;height:48px}
:deep(.org-table td.el-table__cell){height:54px;border-bottom-color:#eef1f6}
:deep(.org-table .el-table__inner-wrapper::before){display:none}
:deep(.el-button){border-radius:9px}
:deep(.el-input__wrapper),:deep(.el-select__wrapper){border-radius:9px}
@media(max-width:760px){.uni-org-v1764{padding:12px}.org-heading{padding:20px}.org-heading-symbol{display:none}.org-stats{gap:9px}.org-stat{padding:14px}.org-stat-value{font-size:23px}.org-panel-head,.org-toolbar{padding:16px}.org-tabs{padding:0 16px 14px}.org-table{margin:0 12px;width:calc(100% - 24px)}.org-search{width:min(100%,240px)}}
@media(max-width:480px){.org-stats{grid-template-columns:1fr}.org-panel-head{align-items:flex-start;flex-direction:column}.org-create{align-self:flex-end}.org-toolbar-right{width:100%;justify-content:space-between}}
</style>
