<template>
<div class="uni-admin-v1762-page"><header class="uni-admin-v1762-heading"><div class="uni-admin-v1762-heading-copy"><span class="uni-admin-v1762-eyebrow">UNIVERSITY ADMINISTRATION</span><h1>高校账号管理</h1><p>高校账号及权限管理</p></div></header>
<div class="uni-admin-v1762-content"><el-card shadow="never"><div class="mb-4 flex justify-between"><h2>账号列表</h2><el-button type="primary" @click="open('create')">新增账号</el-button></div>
<el-form inline @submit.prevent="search"><el-form-item label="所属高校"><el-select v-model="tenantId" clearable filterable style="width:220px" placeholder="全部高校"><el-option v-for="t in tenants" :key="t.id" :label="t.name" :value="t.id"/></el-select></el-form-item><el-form-item label="关键词"><el-input v-model="keyword" clearable placeholder="账号或姓名"/></el-form-item><el-button type="primary" native-type="submit">查询</el-button></el-form>
<el-table :data="rows" v-loading="loading" border><el-table-column prop="userId" label="ID" width="70"/><el-table-column prop="username" label="登录账号"/><el-table-column prop="nickName" label="姓名"/><el-table-column label="角色"><template #default="{row}">{{row.authorityId===2002?'请假监管员':'高校管理员'}}</template></el-table-column><el-table-column prop="tenantName" label="所属高校"/><el-table-column prop="phone" label="手机"/><el-table-column prop="email" label="邮箱"/><el-table-column label="状态"><template #default="{row}"><el-tag :type="active(row)?'success':'danger'">{{active(row)?(row.tenantStatus===1?'启用':'高校停用'):'停用'}}</el-tag></template></el-table-column><el-table-column label="操作" width="290"><template #default="{row}"><el-button link type="primary" @click="open('edit',row)">编辑</el-button><el-button link type="primary" @click="open('password',row)">重置密码</el-button><el-button link :type="active(row)?'danger':'success'" @click="status(row)">{{active(row)?'停用':'启用'}}</el-button><el-button link type="danger" :loading="deleting === row.membershipId" :disabled="deleting !== null" @click="removeAccount(row)">删除</el-button></template></el-table-column></el-table>
<el-pagination class="mt-4" v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10,20,50,100]" layout="total,sizes,prev,pager,next" @current-change="load" @size-change="search"/>
</el-card></div>
<el-dialog v-model="visible" :title="{create:'新增账号',edit:'编辑账号',password:'重置密码'}[mode]" width="520px" destroy-on-close><el-form label-width="100px"><template v-if="mode!=='password'"><el-form-item label="所属高校"><el-select v-model="form.tenantId" :disabled="mode==='edit'" filterable><el-option v-for="t in tenants" :key="t.id" :value="t.id" :label="t.name" :disabled="t.status!==1"/></el-select></el-form-item><el-form-item label="角色"><el-select v-model="form.authorityId" :disabled="mode==='edit'"><el-option label="高校管理员" :value="2001"/><el-option label="请假监管员" :value="2002"/></el-select></el-form-item><el-form-item label="登录账号"><el-input v-model="form.username" :disabled="mode==='edit'"/></el-form-item><el-form-item label="姓名"><el-input v-model="form.nickName"/></el-form-item><el-form-item label="辅导员工号"><el-input v-model.trim="form.advisorTeacherNo" clearable maxlength="64" placeholder="填写本校教师工号，留空解除关联"/></el-form-item><el-form-item label="手机号"><el-input v-model="form.phone"/></el-form-item><el-form-item label="邮箱"><el-input v-model="form.email"/></el-form-item></template><el-form-item v-if="mode!=='edit'" label="密码"><el-input v-model="form.password" type="password" show-password placeholder="6至12个字符" autocomplete="new-password"/></el-form-item></el-form><template #footer><el-button @click="visible=false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template></el-dialog></div>
</template>
<script setup>
import {ref,reactive,onMounted,watch} from 'vue'
import {useRoute} from 'vue-router'
import {ElMessage,ElMessageBox} from 'element-plus'
import request from '@/utils/request'
import {getTenantList} from '@/api/universityTenant'
const route=useRoute(),tenantId=ref(Number(route.query.tenantId)||''),keyword=ref(''),page=ref(1),pageSize=ref(10),total=ref(0),rows=ref([]),tenants=ref([]),loading=ref(false),saving=ref(false),visible=ref(false),mode=ref('create'),form=reactive({})
const call=(action,data,method='put')=>request({url:'/uni/admin/'+action,method,data})
const active=r=>r.status===1&&r.userEnabled===1
let seq=0
async function load(){const id=++seq;loading.value=true;try{const r=await call('list',{tenantId:tenantId.value||0,keyword:keyword.value,page:page.value,pageSize:pageSize.value},'post');if(id===seq&&r.code===0){rows.value=r.data.list||[];total.value=r.data.total}}finally{if(id===seq)loading.value=false}}
function search(){page.value=1;return load()}
async function options(){let p=1;const all=[];while(true){const r=await getTenantList({page:p++,pageSize:100,keyword:''});if(r.code!==0)break;all.push(...(r.data.list||[]));if(all.length>=r.data.total||!r.data.list?.length)break}tenants.value=all}
function open(m,r={}){mode.value=m;Object.assign(form,{tenantId:r.tenantId||tenantId.value||'',membershipId:r.membershipId||0,authorityId:r.authorityId||2001,username:r.username||'',nickName:r.nickName||'',phone:r.phone||'',email:r.email||'',advisorTeacherNo:r.advisorTeacherNo||'',password:''});visible.value=true}
async function save(){if(mode.value!=='password'&&(!form.tenantId||!form.username.trim()||!form.nickName.trim()))return ElMessage.warning('请填写必填项');const n=[...form.password].length;if(mode.value!=='edit'&&(n<6||n>12))return ElMessage.warning('密码须为6至12个字符');saving.value=true;try{const payload={...form};if(mode.value==='password')delete payload.advisorTeacherNo;const r=await call(mode.value,payload,mode.value==='create'?'post':'put');if(r.code===0){ElMessage.success('操作成功');visible.value=false;form.password='';await load()}}finally{saving.value=false}}
async function status(r){const next=active(r)?2:1;try{await ElMessageBox.confirm('确认'+(next===1?'启用':'停用')+'该管理员？','操作确认',{type:'warning'})}catch{return}const res=await call('status',{tenantId:r.tenantId,membershipId:r.membershipId,status:next});if(res.code===0){ElMessage.success('操作成功');await load()}}
const deleting = ref(null)
async function removeAccount(row) {
  if (deleting.value !== null) return
  deleting.value = row.membershipId
  try {
    await ElMessageBox.confirm(
      `确定删除「${row.tenantName}」的账号「${row.username}」？该账号将无法登录。`,
      '删除高校账号',
      { type: 'warning', confirmButtonText: '删除' }
    )
  } catch {
    deleting.value = null
    return
  }
  try {
    const result = await call('delete', { tenantId: row.tenantId, membershipId: row.membershipId }, 'delete')
    if (result.code === 0) {
      ElMessage.success('账号已删除')
      if (rows.value.length === 1 && rows.value[0].membershipId === row.membershipId && page.value > 1) {
        page.value--
      }
      await load()
    }
  } finally {
    deleting.value = null
  }
}
watch(()=>route.query.tenantId,v=>{tenantId.value=Number(v)||'';search()})
onMounted(()=>{options();load()})
</script>
