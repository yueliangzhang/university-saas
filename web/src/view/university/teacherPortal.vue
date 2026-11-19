<template>
  <div class="teacher-portal uni-ui">
    <div class="topbar"><div class="brand">🎓 高校教师服务平台</div><el-button v-if="token" plain @click="logout">退出登录</el-button></div>
    <div v-if="!token" class="login-wrap"><el-card shadow="never"><h2>教师登录</h2><p class="muted">使用高校 ID、教师工号和管理员分配的密码</p><el-form label-position="top" @submit.prevent="login"><el-form-item label="高校 ID"><el-input v-model="loginForm.tenantId" /></el-form-item><el-form-item label="教师工号"><el-input v-model="loginForm.teacherNo" /></el-form-item><el-form-item label="密码"><el-input v-model="loginForm.password" type="password" show-password autocomplete="current-password" /></el-form-item><el-button native-type="submit" type="primary" :loading="busy" style="width:100%">登录教师门户</el-button></el-form></el-card></div>
    <main v-else class="main"><div class="welcome"><div><h2>你好，{{me?.name || '老师'}}</h2><p class="muted">工号：{{me?.teacherNo}} · {{me?.title || '教师'}} · 所属学院 ID：{{me?.collegeId}}</p></div><el-button type="primary" @click="refresh">刷新数据</el-button></div>
      <el-row :gutter="16" class="stats"><el-col :xs="12" :sm="8"><el-card shadow="never"><div class="muted">负责班级</div><h2>{{classes.length}}</h2></el-card></el-col><el-col :xs="12" :sm="8"><el-card shadow="never"><div class="muted">待审批申请（当前筛选）</div><h2>{{status==='pending'?total:'—'}}</h2></el-card></el-col><el-col :xs="24" :sm="8"><el-card shadow="never"><div class="muted">请假记录（当前筛选）</div><h2>{{total}}</h2></el-card></el-col></el-row>
      <el-card shadow="never" class="section"><template #header><b>我的负责班级</b></template><el-empty v-if="!classes.length" description="管理员尚未分配班级" :image-size="72"/><el-tag v-for="cl in classes" :key="cl.id" style="margin:4px">{{cl.name}}（{{cl.code}}）</el-tag></el-card>
      <el-card shadow="never" class="section"><template #header><div class="section-header"><b>学生请假审批</b><el-select v-model="status" style="width:140px" @change="search"><el-option label="待审批" value="pending"/><el-option label="全部记录" value=""/><el-option label="已通过" value="approved"/><el-option label="已驳回" value="rejected"/></el-select></div></template>
        <el-table v-loading="loading" :data="rows" style="width:100%"><el-table-column prop="studentName" label="学生" min-width="95"/><el-table-column prop="studentNo" label="学号" min-width="105"/><el-table-column label="类型" width="75"><template #default="{row}">{{types[row.leaveType]||row.leaveType}}</template></el-table-column><el-table-column label="时间" min-width="180"><template #default="{row}">{{fmt(row.startAt)}} 至 {{fmt(row.endAt)}}</template></el-table-column><el-table-column prop="reason" label="请假原因" min-width="140" show-overflow-tooltip/><el-table-column label="状态" width="95"><template #default="{row}"><el-tag :type="colors[row.status]||'info'">{{statuses[row.status]||row.status}}</el-tag></template></el-table-column><el-table-column prop="reviewNote" label="审批备注" min-width="120" show-overflow-tooltip/><el-table-column label="操作" width="140" fixed="right"><template #default="{row}"><template v-if="row.status==='pending'"><el-button type="success" link @click="review(row,'approved')">通过</el-button><el-button type="danger" link @click="review(row,'rejected')">驳回</el-button></template></template></el-table-column></el-table>
        <el-pagination class="pager" v-model:current-page="page" :page-size="10" :total="total" layout="total,prev,pager,next" @current-change="loadLeaves" />
      </el-card>
      <el-card shadow="never" class="section"><template #header><b>账号安全</b></template><el-button @click="changeDialog=true">修改密码</el-button></el-card>
    </main>
    <el-dialog v-model="changeDialog" title="修改登录密码" width="min(460px,94vw)"><el-form label-position="top"><el-form-item label="旧密码"><el-input v-model="oldPassword" type="password" show-password/></el-form-item><el-form-item label="新密码（至少12位）"><el-input v-model="newPassword" type="password" show-password/></el-form-item></el-form><template #footer><el-button @click="changeDialog=false">取消</el-button><el-button type="primary" :loading="busy" @click="changePassword">确认修改</el-button></template></el-dialog>
  </div>
</template>
<script setup>
import {ref,onMounted} from 'vue'
import {ElMessage,ElMessageBox} from 'element-plus'
import axios from 'axios'
const base=(import.meta.env.VITE_BASE_API||'/api').replace(/\/$/,'')
const client=axios.create({baseURL:base,timeout:15000})
const token=ref(sessionStorage.getItem('uni_teacher_token')||'')
const loginForm=ref({tenantId:'',teacherNo:'',password:''}),me=ref(null),classes=ref([]),rows=ref([]),total=ref(0),page=ref(1),status=ref('pending'),loading=ref(false),busy=ref(false)
const changeDialog=ref(false),oldPassword=ref(''),newPassword=ref('')
const types={personal:'事假',sick:'病假',other:'其他'},statuses={pending:'待审批',approved:'已通过',rejected:'已驳回',cancelled:'已撤销'},colors={pending:'warning',approved:'success',rejected:'danger',cancelled:'info'}
const fmt=s=>s?String(s).replace('T',' ').slice(0,16):''
const error=e=>e.response?.data?.msg||e.message||'操作失败'
const call=async(url,method='get',data,params)=>{const r=await client({url:`/uni/teacher-portal${url}`,method,data,params,headers:token.value?{Authorization:`Bearer ${token.value}`}:{}});if(r.data?.code!==0)throw Error(r.data?.msg||'请求失败');return r.data.data}
const clear=()=>{token.value='';sessionStorage.removeItem('uni_teacher_token');me.value=null;classes.value=[];rows.value=[];total.value=0}
const loadLeaves=async()=>{if(!token.value)return;loading.value=true;try{const d=await call('/leaves','get',undefined,{page:page.value,pageSize:10,status:status.value});rows.value=d?.list||[];total.value=d?.total||0}catch(e){ElMessage.error(error(e))}finally{loading.value=false}}
const refresh=async()=>{if(!token.value)return;try{me.value=await call('/me');const d=await call('/classes');classes.value=d?.list||[];await loadLeaves()}catch(e){ElMessage.error(error(e));if(e.response?.status===401)clear()}}
const login=async()=>{busy.value=true;try{const d=await call('/login','post',{tenantId:Number(loginForm.value.tenantId),teacherNo:loginForm.value.teacherNo,password:loginForm.value.password});token.value=d.token;sessionStorage.setItem('uni_teacher_token',d.token);loginForm.value.password='';await refresh()}catch(e){ElMessage.error(error(e))}finally{busy.value=false}}
const logout=async()=>{try{await call('/logout','post',{})}catch(e){}finally{clear()}}
const search=()=>{page.value=1;loadLeaves()}
const review=async(row,action)=>{try{const result=await ElMessageBox.prompt(action==='rejected'?'请填写驳回原因':'审批备注（可留空）',action==='rejected'?'驳回申请':'通过申请',{inputValidator:v=>action!=='rejected'||!!v?.trim(),inputErrorMessage:'驳回原因不能为空'});await call('/leaves/review','post',{id:row.id,action,note:result.value||''});ElMessage.success('审批完成');await loadLeaves()}catch(e){if(e!=='cancel'&&e!=='close')ElMessage.error(error(e))}}
const changePassword=async()=>{busy.value=true;try{await call('/change-password','post',{oldPassword:oldPassword.value,newPassword:newPassword.value});ElMessage.success('密码已修改，请重新登录');changeDialog.value=false;oldPassword.value='';newPassword.value='';clear()}catch(e){ElMessage.error(error(e))}finally{busy.value=false}}
onMounted(()=>{if(token.value)refresh()})
</script>
<style scoped>.teacher-portal{min-height:100vh;background:#f4f7fb;color:#1e293b}.topbar{background:#fff;padding:18px max(24px,calc((100vw - 1200px)/2));display:flex;justify-content:space-between;align-items:center;border-bottom:1px solid #e2e8f0}.brand{font-size:19px;font-weight:700;color:#1d4ed8}.login-wrap{max-width:430px;margin:90px auto;padding:0 16px}.login-wrap h2{margin:0 0 8px}.muted{color:#64748b;font-size:13px}.main{max-width:1200px;margin:0 auto;padding:28px 18px}.welcome,.section-header{display:flex;justify-content:space-between;align-items:center;gap:12px}.welcome h2{margin:0}.stats{margin:20px 0}.stats h2{margin:8px 0 0;font-size:30px}.section{margin-bottom:18px}.pager{margin-top:16px;display:flex;justify-content:flex-end}@media(max-width:600px){.welcome{align-items:flex-start;flex-direction:column}.main{padding:14px 10px}.topbar{padding:14px}}

/* Legacy scoped scroll rules replaced by portal-specific global rules below. */

</style>

<style scoped>
/* V17.3 TEACHER PORTAL VISUAL UPGRADE */
.teacher-portal.uni-ui { --uni-primary:#2563eb; --uni-border:#e5eaf2; --uni-muted:#64748b; min-height:100vh; background:radial-gradient(circle at 90% 0%,#eaf2ff 0%,transparent 32%),#f5f7fb; color:#17243b; }
.teacher-portal.uni-ui .topbar { min-height:72px; padding-top:14px; padding-bottom:14px; background:rgba(255,255,255,.96); border-bottom:1px solid var(--uni-border); box-shadow:0 2px 16px rgba(15,23,42,.035); }
.teacher-portal.uni-ui .brand { font-size:19px; font-weight:750; letter-spacing:-.4px; color:#1d4ed8; }
.teacher-portal.uni-ui .main { max-width:1280px; padding-top:32px; padding-bottom:56px; }
.teacher-portal.uni-ui .login-wrap { max-width:470px; margin:clamp(45px,10vh,110px) auto; }
.teacher-portal.uni-ui .login-wrap :deep(.el-card) { border:1px solid var(--uni-border); border-radius:22px; overflow:hidden; box-shadow:0 22px 65px rgba(23,54,112,.10); }
.teacher-portal.uni-ui .login-wrap :deep(.el-card__body) { padding:38px 36px 40px; }
.teacher-portal.uni-ui .login-wrap h2 { font-size:27px; font-weight:760; letter-spacing:-.7px; }
.teacher-portal.uni-ui .login-wrap .muted { margin-bottom:26px; line-height:1.7; }
.teacher-portal.uni-ui .login-wrap :deep(.el-form-item__label) { font-weight:600; color:#334155; }
.teacher-portal.uni-ui :deep(.el-input__wrapper), .teacher-portal.uni-ui :deep(.el-select__wrapper) { min-height:40px; border-radius:10px; }
.teacher-portal.uni-ui :deep(.el-button) { border-radius:10px; font-weight:550; }
.teacher-portal.uni-ui :deep(.el-button--primary) { --el-button-bg-color:#2563eb; --el-button-border-color:#2563eb; --el-button-hover-bg-color:#1d4ed8; --el-button-hover-border-color:#1d4ed8; }
.teacher-portal.uni-ui .welcome { position:relative; overflow:hidden; background:linear-gradient(110deg,#183b9b,#2563eb 65%,#4589f9); color:#fff; padding:30px 34px; border-radius:20px; box-shadow:0 14px 35px rgba(37,99,235,.15); }
.teacher-portal.uni-ui .welcome::after { content:''; position:absolute; width:220px; height:220px; right:-90px; top:-120px; border-radius:50%; background:rgba(255,255,255,.08); pointer-events:none; }
.teacher-portal.uni-ui .welcome h2 { font-size:27px; font-weight:750; letter-spacing:-.5px; }
.teacher-portal.uni-ui .welcome .muted { color:rgba(255,255,255,.85); line-height:1.7; }
.teacher-portal.uni-ui .welcome :deep(.el-button--primary) { --el-button-bg-color:#fff; --el-button-border-color:#fff; --el-button-text-color:#1d4ed8; --el-button-hover-bg-color:#eff6ff; --el-button-hover-border-color:#eff6ff; --el-button-hover-text-color:#1d4ed8; }
.teacher-portal.uni-ui .stats { margin:22px 0 8px; }
.teacher-portal.uni-ui .stats :deep(.el-col) { margin-bottom:16px; }
.teacher-portal.uni-ui .stats :deep(.el-card) { height:100%; min-height:122px; border:1px solid var(--uni-border); border-radius:17px; box-shadow:0 5px 20px rgba(15,23,42,.035); transition:transform .2s ease,box-shadow .2s ease; }
.teacher-portal.uni-ui .stats :deep(.el-card__body) { padding:24px; }
.teacher-portal.uni-ui .stats h2 { color:#1d4ed8; font-size:34px; font-weight:780; letter-spacing:-1px; }
.teacher-portal.uni-ui .muted { color:var(--uni-muted); }
.teacher-portal.uni-ui .section { margin-bottom:20px; border:1px solid var(--uni-border); border-radius:17px; box-shadow:0 5px 20px rgba(15,23,42,.035); }
.teacher-portal.uni-ui .section :deep(.el-card__header) { padding:19px 23px; border-bottom:1px solid #edf1f7; font-size:16px; color:#1e293b; }
.teacher-portal.uni-ui .section :deep(.el-card__body) { padding:22px 23px; }
.teacher-portal.uni-ui .section-header { flex-wrap:wrap; }
.teacher-portal.uni-ui :deep(.el-table) { --el-table-header-bg-color:#f8fafc; --el-table-row-hover-bg-color:#f2f7ff; --el-table-border-color:#edf1f7; border-radius:10px; }
.teacher-portal.uni-ui :deep(.el-table th.el-table__cell) { color:#475569; font-weight:650; }
.teacher-portal.uni-ui :deep(.el-tag) { border-radius:7px; }
.teacher-portal.uni-ui .pager { flex-wrap:wrap; gap:6px; }
.teacher-portal.uni-ui :deep(.el-button:focus-visible), .teacher-portal.uni-ui :deep(.el-input__wrapper:focus-within) { outline:2px solid #93c5fd; outline-offset:2px; }
@media (hover:hover) { .teacher-portal.uni-ui .stats :deep(.el-card:hover) { transform:translateY(-2px); box-shadow:0 12px 30px rgba(15,23,42,.075); } }
@media (max-width:640px) {
 .teacher-portal.uni-ui .topbar { min-height:62px; padding:12px 16px; gap:8px; }
 .teacher-portal.uni-ui .brand { font-size:16px; }
 .teacher-portal.uni-ui .main { padding:14px 12px 40px; }
 .teacher-portal.uni-ui .welcome { padding:23px 20px; border-radius:16px; }
 .teacher-portal.uni-ui .welcome h2 { font-size:23px; }
 .teacher-portal.uni-ui .stats :deep(.el-card__body) { padding:17px; }
 .teacher-portal.uni-ui .stats h2 { font-size:28px; }
 .teacher-portal.uni-ui .section { border-radius:14px; }
 .teacher-portal.uni-ui .section :deep(.el-card__header), .teacher-portal.uni-ui .section :deep(.el-card__body) { padding:16px; }
 .teacher-portal.uni-ui .login-wrap { margin:45px auto; padding:0 14px; }
 .teacher-portal.uni-ui .login-wrap :deep(.el-card__body) { padding:30px 23px; }
 .teacher-portal.uni-ui .pager { justify-content:center; }
}
@media (prefers-reduced-motion:reduce) { .teacher-portal.uni-ui .stats :deep(.el-card) { transition:none; transform:none; } }
</style>


<style>
/* V17.7.1 TEACHER PORTAL SCROLL FIX */
/* Only while teacher portal is mounted. Restore document scrolling when outer app styles lock it. */
html:has(.teacher-portal),
body:has(.teacher-portal) {
  height: auto !important;
  min-height: 100% !important;
  overflow-y: auto !important;
}
body:has(.teacher-portal) #app {
  height: auto !important;
  min-height: 100vh !important;
  max-height: none !important;
  overflow-y: visible !important;
}
body:has(.teacher-portal) .teacher-portal {
  height: auto !important;
  min-height: 100vh;
  max-height: none !important;
  overflow: visible !important;
}
</style>
