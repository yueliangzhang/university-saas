<template>
  <div class="p-6">
    <el-card>
      <template #header>高校组织架构</template>
      <el-tabs v-model="tab" @tab-change="load">
        <el-tab-pane label="学院" name="colleges" />
        <el-tab-pane label="专业" name="majors" />
        <el-tab-pane label="班级" name="classes" />
      </el-tabs>
      <el-alert v-if="error" :title="error" type="error" :closable="false" class="mb-3" />
      <el-button type="primary" @click="openCreate">新增{{ labels[tab] }}</el-button>
      <el-table v-loading="loading" :data="rows" border class="mt-4">
        <el-table-column prop="id" label="ID" width="90" />
        <el-table-column prop="code" label="编码" />
        <el-table-column prop="name" label="名称" />
        <el-table-column v-if="tab==='majors'" prop="collegeId" label="学院 ID" />
        <el-table-column v-if="tab==='classes'" prop="majorId" label="专业 ID" />
        <el-table-column v-if="tab==='classes'" prop="enrollmentYear" label="入学年份" />
        <el-table-column label="状态" width="100"><template #default="{row}">{{ row.status===1?'启用':'停用' }}</template></el-table-column>
        <el-table-column label="操作" width="170"><template #default="{row}">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template></el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="dialog" :title="(editing?'编辑':'新增')+labels[tab]" width="520px">
      <el-form label-width="100px">
        <el-form-item label="编码"><el-input v-model="form.code" maxlength="64" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" maxlength="191" /></el-form-item>
        <el-form-item v-if="tab==='majors'" label="所属学院">
          <el-select v-model="form.collegeId" placeholder="选择学院" style="width:100%">
            <el-option v-for="x in colleges" :key="x.id" :label="`${x.name} (${x.code})`" :value="x.id" :disabled="x.status!==1" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="tab==='classes'" label="所属专业">
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
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
const labels={colleges:'学院',majors:'专业',classes:'班级'}
const tab=ref('colleges'),rows=ref([]),colleges=ref([]),majors=ref([])
const loading=ref(false),saving=ref(false),error=ref(''),dialog=ref(false),editing=ref(false)
const form=ref({})
const api=(kind,method='get',data,params)=>request({url:`/uni/org/${kind}`,method,data,params})
const fetchList=async kind=>{const r=await api(kind);if(r?.code!==0)throw new Error(r?.msg||'加载失败');return r.data?.list||[]}
const load=async()=>{loading.value=true;error.value='';try{rows.value=await fetchList(tab.value)}catch(e){error.value=e.message}finally{loading.value=false}}
const openCreate=async()=>{editing.value=false;form.value={code:'',name:'',status:1,collegeId:null,majorId:null,enrollmentYear:new Date().getFullYear()};await openDialog()}
const openEdit=async row=>{editing.value=true;form.value={...row};await openDialog()}
const openDialog=async()=>{try{if(tab.value==='majors')colleges.value=await fetchList('colleges');if(tab.value==='classes')majors.value=await fetchList('majors');dialog.value=true}catch(e){ElMessage.error(e.message)}}
const save=async()=>{if(!form.value.code?.trim()||!form.value.name?.trim()){ElMessage.warning('请填写编码和名称');return}
 saving.value=true;try{const r=await api(tab.value,editing.value?'put':'post',form.value);if(r?.code!==0)throw new Error(r?.msg||'保存失败');ElMessage.success('保存成功');dialog.value=false;await load()}catch(e){ElMessage.error(e.message)}finally{saving.value=false}}
const remove=async row=>{try{await ElMessageBox.confirm(`确定删除「${row.name}」？关联的下级记录必须先删除。`,'确认删除',{type:'warning'});const r=await api(tab.value,'delete',undefined,{id:row.id});if(r?.code!==0)throw new Error(r?.msg||'删除失败');ElMessage.success('删除成功');await load()}catch(e){if(e!=='cancel')ElMessage.error(e.message||'操作失败')}}
onMounted(load)
</script>
