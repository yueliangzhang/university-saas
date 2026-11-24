<template>
  <div style="display:flex;gap:8px;flex-wrap:wrap">
    <el-button :loading="busy" @click="download('template')">下载模板</el-button>
    <el-button :disabled="busy" @click="open">批量导入</el-button>
    <el-button :loading="busy" @click="download('export')">导出 Excel</el-button>
  </div>
  <el-dialog v-model="visible" title="Excel 批量导入" width="min(760px,95vw)" append-to-body :close-on-click-modal="!busy" :close-on-press-escape="!busy" :show-close="!busy">
    <p>请使用最新模板，单次最多 5000 条，文件不超过 10 MB。重复编号不覆盖，全部校验通过后才能导入。</p>
    <p>{{ kind==='students'?'学号、姓名、班级、辅导员必填。':'工号、姓名、所属学院必填。' }}</p>
    <input type="file" accept=".xlsx" :disabled="busy" @change="choose" />
    <el-button :disabled="!file||busy" :loading="busy" @click="submit(false)">校验并预览</el-button>
    <template v-if="result">
      <p>通过 {{result.valid}} 条，失败 {{result.errors.length}} 条</p>
      <el-table v-if="result.errors.length" :data="result.errors" max-height="300"><el-table-column prop="row" label="行号" width="80"/><el-table-column prop="number" label="编号"/><el-table-column prop="message" label="错误原因"/></el-table>
      <el-button v-if="result.errors.length" @click="report">下载错误报告 (CSV)</el-button>
      <el-table v-else :data="(result.preview||[]).slice(0,100)" max-height="300"><el-table-column label="编号"><template #default="{row}">{{row[0]}}</template></el-table-column><el-table-column label="姓名"><template #default="{row}">{{row[1]}}</template></el-table-column></el-table>
      <p v-if="result.valid>100">仅展示前100条通过记录。</p>
    </template>
    <template #footer><el-button :disabled="busy" @click="visible=false">关闭</el-button><el-button type="primary" :loading="busy" :disabled="!result||!result.valid||result.errors.length>0||busy" @click="submit(true)">确认导入</el-button></template>
  </el-dialog>
</template>
<script setup>
import {ref} from 'vue'
import {ElMessage} from 'element-plus'
import request from '@/utils/request'
const props=defineProps({kind:{type:String,required:true},filters:{type:Object,default:()=>({})}})
const emit=defineEmits(['imported'])
const visible=ref(false),busy=ref(false),file=ref(null),result=ref(null)
const open=()=>{file.value=null;result.value=null;visible.value=true}
const choose=e=>{file.value=e.target.files?.[0]||null;result.value=null;if(file.value&&(!/\.xlsx$/i.test(file.value.name)||file.value.size>10*1024*1024)){file.value=null;e.target.value='';ElMessage.warning('请选择10MB以内的xlsx文件')}}
const saveBlob=(blob,name)=>{const url=URL.createObjectURL(blob);const a=document.createElement('a');a.href=url;a.download=name;a.click();setTimeout(()=>URL.revokeObjectURL(url),1000)}
const download=async action=>{busy.value=true;try{const r=await request({url:`/uni/${props.kind}/${action}`,method:'get',params:action==='export'?props.filters:{},responseType:'blob'});if(!r.data?.type?.includes('spreadsheetml'))throw Error('下载失败');saveBlob(r.data,`${props.kind}-${action}.xlsx`)}catch(e){ElMessage.error(e.message||'下载失败')}finally{busy.value=false}}
const submit=async commit=>{if(!file.value)return;busy.value=true;try{const data=new FormData();data.append('file',file.value);const r=await request({url:`/uni/${props.kind}/import/${commit?'confirm':'preview'}`,method:'post',data,headers:{'Content-Type':'multipart/form-data'}});if(r.code!==0)throw Error(r.msg);result.value=r.data;if(r.data.imported){ElMessage.success(`成功导入 ${r.data.valid} 条`);visible.value=false;emit('imported')}}catch(e){result.value=null;ElMessage.error(e.message||'导入失败')}finally{busy.value=false}}
const report=()=>{const esc=v=>'"'+String(v??'').replace(/^[=+\-@\t\r]/,"'$&").replaceAll('"','""')+'"';saveBlob(new Blob(['\ufeff行号,编号,错误原因\n'+result.value.errors.map(r=>[r.row,r.number,r.message].map(esc).join(',')).join('\n')],{type:'text/csv;charset=utf-8'}),'import-errors.csv')}
</script>
