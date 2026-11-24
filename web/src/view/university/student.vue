<template>
  <!-- UNI_STUDENT_V17_6_9 -->
  <main class="uni-student-page">
    <div class="uni-student-heading">
      <div>
        <h1>学生管理</h1>
        <p>统一管理学生档案、请假审批与登录账号</p>
      </div>
      <div v-if="activeTab === 'profile'" style="display:flex;gap:8px;flex-wrap:wrap"><el-button v-if="activeTab === 'profile'" type="primary" @click="openCreate">+ 新增学生</el-button><ExcelImport kind="students" :filters="{keyword, classId:filterClass||undefined, advisorId:filterAdvisor??undefined}" @imported="search" /></div>
    </div>

    <nav class="uni-student-tabs" aria-label="学生管理功能">
      <button type="button" :class="{ active: activeTab === 'profile' }" @click="activeTab = 'profile'">学生档案</button>
      <button type="button" :class="{ active: activeTab === 'leave' }" @click="activeTab = 'leave'">请假审批</button>
      <button type="button" :class="{ active: activeTab === 'account' }" @click="activeTab = 'account'">登录账号</button>
    </nav>
    <section v-show="activeTab === 'profile'" class="uni-student-panel">
      <div class="uni-student-panel-title">
        <div class="uni-student-panel-name">学生档案 <span class="uni-student-count">共 {{ total }} 人</span></div>
      </div>
      <el-form class="uni-student-filter" inline @submit.prevent="search">
        <el-form-item label="姓名 / 学号">
          <el-input v-model="keyword" clearable placeholder="搜索姓名或学号" style="width:220px" @keyup.enter="search" />
        </el-form-item>
        <el-form-item label="所属班级">
          <el-select v-model="filterClass" clearable filterable placeholder="全部班级" style="width:200px" @change="search">
            <el-option v-for="x in classes" :key="x.id" :label="x.name" :value="x.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="辅导员"><el-select v-model="filterAdvisor" clearable filterable placeholder="全部辅导员" style="width:190px" @change="search"><el-option label="未分配" :value="0"/><el-option v-for="t in teachers" :key="t.id" :label="`${t.name}（${t.teacherNo}）`" :value="t.id"/></el-select></el-form-item>
        <el-form-item class="uni-student-filter-actions">
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="reset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="uni-student-table-scroll">
        <el-table v-loading="loading" :data="rows" stripe style="width:100%" empty-text="暂无符合条件的学生">
          <el-table-column prop="studentNo" label="学号" min-width="130">
            <template #default="{row}"><span class="uni-student-number">{{ row.studentNo }}</span></template>
          </el-table-column>
          <el-table-column prop="name" label="姓名" min-width="150">
            <template #default="{row}"><span class="uni-student-name">{{ row.name }}</span></template>
          </el-table-column>
          <el-table-column label="性别" width="85">
            <template #default="{row}">{{ ['未填写','男','女'][row.gender] || '未知' }}</template>
          </el-table-column>
          <el-table-column label="所属班级" min-width="180" show-overflow-tooltip>
            <template #default="{row}">{{ className(row.classId) }}</template>
          </el-table-column>
          <el-table-column label="所属辅导员" min-width="135"><template #default="{row}">{{row.advisorName || '未分配'}}</template></el-table-column>
          <el-table-column prop="enrollmentYear" label="入学年份" width="110" />
          <el-table-column prop="phone" label="联系电话" min-width="140" show-overflow-tooltip>
            <template #default="{row}">{{ row.phone || '—' }}</template>
          </el-table-column>
          <el-table-column label="学籍状态" width="110">
            <template #default="{row}"><el-tag :type="row.status === 1 ? 'success' : 'info'" effect="light" size="small">{{ row.status === 1 ? '在读' : '停用' }}</el-tag></template>
          </el-table-column>
          <el-table-column label="操作" width="142" fixed="right" align="right">
            <template #default="{row}">
              <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
              <el-button link type="danger" @click="remove(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <div class="uni-student-pagination">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10,20,50,100]" layout="total, sizes, prev, pager, next" @current-change="load" @size-change="search" />
      </div>
    </section>

    <section v-if="activeTab === 'leave'" class="uni-student-panel uni-student-subpanel"><StudentLeavePanel /></section>
    <section v-if="activeTab === 'account'" class="uni-student-panel uni-student-subpanel"><StudentAccountPanel /></section>

    <el-dialog v-model="dialog" :title="editing ? '编辑学生' : '新增学生'" width="560px" append-to-body class="uni-student-dialog">
      <el-form :model="form" label-width="95px">
        <el-form-item label="学号" required><el-input v-model="form.studentNo" maxlength="64" /></el-form-item>
        <el-form-item label="姓名" required><el-input v-model="form.name" maxlength="191" /></el-form-item>
        <el-form-item label="性别"><el-select v-model="form.gender" style="width:100%"><el-option label="未填写" :value="0" /><el-option label="男" :value="1" /><el-option label="女" :value="2" /></el-select></el-form-item>
        <el-form-item label="班级" required><el-select v-model="form.classId" filterable placeholder="选择班级" style="width:100%"><el-option v-for="x in activeClasses" :key="x.id" :label="x.name" :value="x.id" /></el-select></el-form-item>
        <el-form-item label="辅导员"><el-select v-model="advisorChoice" filterable style="width:100%"><el-option v-if="!editing" label="按班级自动分配" value="auto"/><el-option label="暂不分配" value="none"/><el-option v-for="t in teachers" :key="t.id" :label="`${t.name}（${t.teacherNo}）${t.status!==1?' · 已停用':''}`" :value="t.id" :disabled="t.status!==1"/></el-select></el-form-item>
        <el-form-item label="入学年份" required><el-input-number v-model="form.enrollmentYear" :min="1900" :max="2200" /></el-form-item>
        <el-form-item label="电话"><el-input v-model="form.phone" maxlength="32" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="form.email" maxlength="191" /></el-form-item>
        <el-form-item label="状态"><el-radio-group v-model="form.status"><el-radio :value="1">在读</el-radio><el-radio :value="2">停用</el-radio></el-radio-group></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
    </el-dialog>
  </main>
</template>
<script setup>
import ExcelImport from './ExcelImport.vue' 
import { ref, computed, onMounted } from 'vue'
import StudentLeavePanel from './studentLeavePanel.vue'
import StudentAccountPanel from './studentAccountPanel.vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
const activeTab=ref('profile')
const teachers=ref([]),filterAdvisor=ref(null),advisorChoice=ref('auto')
const loadTeachers=async()=>{try{const all=[];for(let p=1;;p++){const r=await request({url:'/uni/teachers',method:'get',params:{page:p,pageSize:100}});if(r?.code!==0)throw Error(r?.msg||'教师加载失败');const chunk=r.data?.list||[];all.push(...chunk);if(!chunk.length||all.length>=r.data.total)break}teachers.value=all}catch(e){ElMessage.error(e.message)}}
const rows=ref([]),classes=ref([]),majors=ref([]),colleges=ref([])
const keyword=ref(''),filterClass=ref(null),page=ref(1),pageSize=ref(20),total=ref(0)
const loading=ref(false),saving=ref(false),dialog=ref(false),editing=ref(false),form=ref({})
const activeClasses=computed(()=>classes.value.filter(x=>x.status===1 && majors.value.some(m=>m.id===x.majorId && m.status===1 && colleges.value.some(c=>c.id===m.collegeId && c.status===1))))
const className=id=>classes.value.find(x=>x.id===id)?.name||`班级 #${id}`
const api=(method='get',data,params)=>request({url:'/uni/students',method,data,params})
const load=async()=>{loading.value=true;try{const r=await api('get',undefined,{page:page.value,pageSize:pageSize.value,keyword:keyword.value,classId:filterClass.value||undefined,advisorId:filterAdvisor.value ?? undefined});if(r?.code!==0)throw Error(r?.msg||'加载失败');rows.value=r.data?.list||[];total.value=r.data?.total||0}catch(e){ElMessage.error(e.message)}finally{loading.value=false}}
const loadOptions=async()=>{try{for(const [kind,target] of [['colleges',colleges],['majors',majors],['classes',classes]]){const r=await request({url:`/uni/org/${kind}`,method:'get'});if(r?.code!==0)throw Error(r?.msg||'组织架构加载失败');target.value=r.data?.list||[]}}catch(e){ElMessage.error(e.message)}}
const search=()=>{page.value=1;load()}
const reset=()=>{keyword.value='';filterClass.value=null;filterAdvisor.value=null;search()}
const openCreate=()=>{editing.value=false;advisorChoice.value='auto';loadTeachers();form.value={studentNo:'',name:'',gender:0,classId:null,enrollmentYear:new Date().getFullYear(),phone:'',email:'',status:1};dialog.value=true}
const openEdit=row=>{editing.value=true;advisorChoice.value=row.advisorId ?? 'none';loadTeachers();form.value={...row};dialog.value=true}
const save=async()=>{if(!form.value.studentNo?.trim()||!form.value.name?.trim()||!form.value.classId){ElMessage.warning('请填写学号、姓名并选择班级');return}saving.value=true;try{const payload={...form.value};delete payload.advisorId;if(advisorChoice.value!=='auto'&&(!editing.value||advisorChoice.value!==(form.value.advisorId??'none')))payload.advisorId=advisorChoice.value==='none'?null:advisorChoice.value;const r=await api(editing.value?'put':'post',payload);if(r?.code!==0)throw Error(r?.msg||'保存失败');ElMessage.success('保存成功');dialog.value=false;await load()}catch(e){ElMessage.error(e.message)}finally{saving.value=false}}
const remove=async row=>{try{await ElMessageBox.confirm(`确定删除学生「${row.name}」？`,'删除确认',{type:'warning'});const r=await api('delete',undefined,{id:row.id});if(r?.code!==0)throw Error(r?.msg||'删除失败');ElMessage.success('删除成功');await load()}catch(e){if(e!=='cancel')ElMessage.error(e.message||'删除失败')}}
onMounted(async()=>{await Promise.all([loadOptions(),loadTeachers()]);await load()})
</script>
<style scoped>
.uni-student-tabs { display:flex; align-items:center; gap:5px; border-bottom:1px solid #e7edf5; margin:0 0 16px; padding:0 2px; overflow-x:auto; }
.uni-student-tabs button { appearance:none; cursor:pointer; white-space:nowrap; background:transparent; border:0; border-bottom:2px solid transparent; padding:12px 20px; font-size:14px; color:#65748b; font-weight:550; }
.uni-student-tabs button:hover { color:#2563eb; }
.uni-student-tabs button.active { color:#2563eb; border-bottom-color:#2563eb; font-weight:650; }
.uni-student-subpanel { overflow:visible; }
@media(max-width:768px) { .uni-student-tabs button { padding:11px 14px; } }

.uni-student-page { padding: 18px 22px 28px; min-width: 0; color: #25344d; }
.uni-student-heading { display:flex; align-items:center; justify-content:space-between; gap:16px; margin-bottom:18px; }
.uni-student-heading h1 { margin:0; font-size:22px; line-height:1.4; font-weight:650; color:#1d2b43; }
.uni-student-heading p { margin:5px 0 0; color:#7a879a; font-size:13px; }
.uni-student-panel { border:1px solid #e6ebf3; background:#fff; border-radius:12px; overflow:hidden; box-shadow:0 3px 14px rgba(27,47,84,.025); }
.uni-student-panel-title { padding:17px 20px; border-bottom:1px solid #edf0f5; }
.uni-student-panel-name { font-size:15px; font-weight:650; }
.uni-student-count { margin-left:10px; font-size:12px; font-weight:400; color:#8190a4; }
.uni-student-filter { display:flex; align-items:center; flex-wrap:wrap; gap:0 14px; padding:18px 20px 2px; }
.uni-student-filter :deep(.el-form-item) { margin-right:0; margin-bottom:16px; }
.uni-student-filter :deep(.el-form-item__label) { color:#56657b; font-weight:500; }
.uni-student-filter :deep(.el-input__wrapper), .uni-student-filter :deep(.el-select__wrapper) { border-radius:7px; }
.uni-student-filter-actions { margin-left:auto; }
.uni-student-page :deep(.el-button) { border-radius:7px; }
.uni-student-page :deep(.el-button--primary) { --el-button-bg-color:#2563eb; --el-button-border-color:#2563eb; --el-button-hover-bg-color:#1d4ed8; --el-button-hover-border-color:#1d4ed8; }
.uni-student-table-scroll { padding:0 20px; min-width:0; }
.uni-student-page :deep(.el-table) { --el-table-header-bg-color:#f7f9fc; --el-table-border-color:#edf1f6; --el-table-row-hover-bg-color:#f7faff; font-size:13px; }
.uni-student-page :deep(.el-table th.el-table__cell) { color:#617087; font-weight:600; height:46px; }
.uni-student-page :deep(.el-table td.el-table__cell) { height:48px; color:#334155; }
.uni-student-name { font-weight:600; color:#25344d; }
.uni-student-number { font-variant-numeric:tabular-nums; color:#475569; }
.uni-student-page :deep(.el-tag) { border-radius:5px; }
.uni-student-pagination { display:flex; justify-content:flex-end; padding:17px 20px 20px; }
@media (max-width: 768px) {
  .uni-student-page { padding:12px; }
  .uni-student-heading { align-items:flex-start; }
  .uni-student-heading h1 { font-size:19px; }
  .uni-student-panel-title { padding:14px; }
  .uni-student-filter { padding:14px 14px 0; display:block; }
  .uni-student-filter :deep(.el-form-item) { display:flex; margin-bottom:12px; }
  .uni-student-filter :deep(.el-form-item__content) { flex:1; min-width:0; }
  .uni-student-filter :deep(.el-input), .uni-student-filter :deep(.el-select) { width:100% !important; }
  .uni-student-filter-actions { margin-left:0; }
  .uni-student-table-scroll { padding:0 10px; }
  .uni-student-pagination { padding:14px 10px; overflow-x:auto; justify-content:flex-start; }
  .uni-student-pagination :deep(.el-pagination) { flex-wrap:nowrap; }
}
</style>
