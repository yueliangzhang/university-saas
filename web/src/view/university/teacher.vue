<template>
  <div class="teacher-page">
    <el-card shadow="never" class="hero">
      <div class="hero-inner"><div><h2>教师管理</h2><p>维护本校教师档案、学院归属与在职状态</p></div><el-button type="primary" @click="openCreate">新增教师</el-button></div>
    </el-card>
    <el-card shadow="never" class="content">
      <el-form inline @submit.prevent="search">
        <el-form-item label="姓名 / 工号"><el-input v-model="keyword" clearable placeholder="输入姓名或工号" @keyup.enter="search" /></el-form-item>
        <el-form-item label="学院"><el-select v-model="filterCollege" clearable placeholder="全部学院" style="width:190px" @change="search"><el-option v-for="c in colleges" :key="c.id" :label="c.name" :value="c.id" /></el-select></el-form-item>
        <el-form-item><el-button type="primary" @click="search">查询</el-button><el-button @click="reset">重置</el-button></el-form-item>
      </el-form>
      <el-table v-loading="loading" :data="rows" stripe style="width:100%">
        <el-table-column prop="teacherNo" label="教师工号" min-width="125" />
        <el-table-column prop="name" label="姓名" min-width="110" />
        <el-table-column label="所属学院" min-width="150"><template #default="{row}">{{collegeName(row.collegeId)}}</template></el-table-column>
        <el-table-column prop="title" label="职称" min-width="100" />
        <el-table-column prop="phone" label="电话" min-width="130" />
        <el-table-column prop="email" label="邮箱" min-width="160" show-overflow-tooltip />
        <el-table-column label="状态" width="100"><template #default="{row}"><el-tag :type="row.status===1?'success':'info'">{{row.status===1?'在职':'停用'}}</el-tag></template></el-table-column>
        <el-table-column label="操作" width="150" fixed="right"><template #default="{row}"><el-button link type="primary" @click="openEdit(row)">编辑</el-button><el-button link type="danger" @click="remove(row)">删除</el-button></template></el-table-column>
      </el-table>
      <div class="pager"><el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10,20,50,100]" layout="total, sizes, prev, pager, next" @current-change="load" @size-change="search" /></div>
    </el-card>
    <el-dialog v-model="dialog" :title="editing?'编辑教师':'新增教师'" width="min(560px,95vw)" append-to-body>
      <el-form label-width="100px" :model="form">
        <el-form-item label="教师工号" required><el-input v-model="form.teacherNo" maxlength="64" /></el-form-item>
        <el-form-item label="姓名" required><el-input v-model="form.name" maxlength="191" /></el-form-item>
        <el-form-item label="所属学院" required><el-select v-model="form.collegeId" filterable placeholder="请选择学院" style="width:100%"><el-option v-for="c in activeColleges" :key="c.id" :label="c.name" :value="c.id" /></el-select></el-form-item>
        <el-form-item label="职称"><el-input v-model="form.title" maxlength="100" placeholder="如：讲师、副教授" /></el-form-item>
        <el-form-item label="性别"><el-select v-model="form.gender" style="width:100%"><el-option label="未填写" :value="0" /><el-option label="男" :value="1" /><el-option label="女" :value="2" /></el-select></el-form-item>
        <el-form-item label="电话"><el-input v-model="form.phone" maxlength="32" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="form.email" maxlength="191" /></el-form-item>
        <el-form-item label="状态"><el-radio-group v-model="form.status"><el-radio :value="1">在职</el-radio><el-radio :value="2">停用</el-radio></el-radio-group></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
    </el-dialog>
  </div>
</template>
<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
const rows=ref([]),colleges=ref([]),keyword=ref(''),filterCollege=ref(null),page=ref(1),pageSize=ref(20),total=ref(0)
const loading=ref(false),saving=ref(false),dialog=ref(false),editing=ref(false),form=ref({})
const activeColleges=computed(()=>colleges.value.filter(c=>c.status===1))
const collegeName=id=>colleges.value.find(c=>c.id===id)?.name||`学院 #${id}`
const api=(method='get',data,params)=>request({url:'/uni/teachers',method,data,params})
const load=async()=>{loading.value=true;try{const r=await api('get',undefined,{page:page.value,pageSize:pageSize.value,keyword:keyword.value,collegeId:filterCollege.value||undefined});if(r?.code!==0)throw Error(r?.msg||'加载失败');rows.value=r.data?.list||[];total.value=r.data?.total||0}catch(e){ElMessage.error(e.message||'加载失败')}finally{loading.value=false}}
const loadColleges=async()=>{try{const r=await request({url:'/uni/org/colleges',method:'get'});if(r?.code!==0)throw Error(r?.msg||'学院加载失败');colleges.value=r.data?.list||[]}catch(e){ElMessage.error(e.message||'学院加载失败')}}
const search=()=>{page.value=1;load()}
const reset=()=>{keyword.value='';filterCollege.value=null;search()}
const openCreate=()=>{editing.value=false;form.value={teacherNo:'',name:'',gender:0,phone:'',email:'',collegeId:null,title:'',status:1};dialog.value=true}
const openEdit=row=>{editing.value=true;form.value={...row};dialog.value=true}
const save=async()=>{const f=form.value;if(!f.teacherNo?.trim()||!f.name?.trim()||!f.collegeId){ElMessage.warning('请填写工号、姓名和学院');return}saving.value=true;try{const r=await api(editing.value?'put':'post',f);if(r?.code!==0)throw Error(r?.msg||'保存失败');ElMessage.success('保存成功');dialog.value=false;await load()}catch(e){ElMessage.error(e.message||'保存失败')}finally{saving.value=false}}
const remove=async row=>{try{await ElMessageBox.confirm(`确定删除教师「${row.name}」？`,'删除确认',{type:'warning'});const r=await api('delete',undefined,{id:row.id});if(r?.code!==0)throw Error(r?.msg||'删除失败');ElMessage.success('已删除');await load()}catch(e){if(e!=='cancel'&&e!=='close')ElMessage.error(e.message||'删除失败')}}
onMounted(async()=>{await loadColleges();await load()})
</script>
<style scoped>.teacher-page{padding:20px;display:grid;gap:18px}.hero{background:linear-gradient(115deg,#eff6ff,#fff)}.hero-inner{display:flex;justify-content:space-between;align-items:center;gap:12px}.hero h2{margin:0;font-size:23px}.hero p{margin:8px 0 0;color:#64748b}.pager{display:flex;justify-content:flex-end;margin-top:18px}@media(max-width:600px){.teacher-page{padding:10px}.hero-inner{align-items:flex-start;flex-direction:column}.pager{overflow-x:auto}}
</style>
