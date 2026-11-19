<template>
  <!-- UNI_ADMIN_V17_6_2 -->
  <div class="uni-admin-v1762-page">
    <header class="uni-admin-v1762-heading">
      <div class="uni-admin-v1762-heading-copy">
        <span class="uni-admin-v1762-eyebrow">UNIVERSITY MANAGEMENT</span>
        <h1>高校租户管理</h1>
        <p>高校租户信息与运营管理</p>
      </div>
      <span class="uni-admin-v1762-heading-mark" aria-hidden="true">校务管理</span>
    </header>
    <div class="uni-admin-v1762-content">

  <div class="p-4">
    <el-card shadow="never">
      <div class="mb-4 flex items-center justify-between">
        <h2 class="text-lg font-semibold">高校管理</h2>
        <el-button type="primary" @click="openCreate">
          新增高校
        </el-button>
      </div>

      <el-form :inline="true" @submit.prevent="loadData">
        <el-form-item label="关键词">
          <el-input
            v-model="keyword"
            placeholder="高校名称或编码"
            clearable
            @keyup.enter="search"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table
        v-loading="loading"
        :data="tableData"
        border
        style="width: 100%"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="code" label="高校编码" min-width="160" />
        <el-table-column prop="name" label="高校名称" min-width="200" />
        <el-table-column prop="plan" label="订阅方案" width="120" />

        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="创建时间" min-width="180">
          <template #default="{ row }">
            {{ formatTime(row.createdAt) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">
              编辑
            </el-button>

            <el-button
              link
              :type="row.status === 1 ? 'danger' : 'success'"
              @click="changeStatus(row)"
            >
              {{ row.status === 1 ? '停用' : '启用' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="mt-4 flex justify-end">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @current-change="loadData"
          @size-change="handleSizeChange"
        />
      </div>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="editing ? '编辑高校' : '新增高校'"
      width="520px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="高校编码" prop="code">
          <el-input
            v-model="form.code"
            :disabled="editing"
            maxlength="64"
            placeholder="例如 NUS"
          />
        </el-form-item>

        <el-form-item label="高校名称" prop="name">
          <el-input
            v-model="form.name"
            maxlength="200"
            placeholder="请输入高校名称"
          />
        </el-form-item>

        <el-form-item label="订阅方案" prop="plan">
          <el-select v-model="form.plan" style="width: 100%">
            <el-option label="基础版" value="basic" />
            <el-option label="专业版" value="pro" />
            <el-option label="企业版" value="enterprise" />
          </el-select>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitForm">
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>

    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createTenant,
  getTenantList,
  updateTenant,
  setTenantStatus
} from '@/api/universityTenant'

const loading = ref(false)
const saving = ref(false)
const tableData = ref([])
const keyword = ref('')
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const dialogVisible = ref(false)
const editing = ref(false)
const formRef = ref(null)

const form = reactive({
  id: 0,
  code: '',
  name: '',
  plan: 'basic'
})

const rules = {
  code: [{ required: true, message: '请输入高校编码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入高校名称', trigger: 'blur' }],
  plan: [{ required: true, message: '请选择订阅方案', trigger: 'change' }]
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await getTenantList({
      page: page.value,
      pageSize: pageSize.value,
      keyword: keyword.value
    })
    if (res.code === 0) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } finally {
    loading.value = false
  }
}

const search = () => {
  page.value = 1
  loadData()
}

const resetSearch = () => {
  keyword.value = ''
  search()
}

const handleSizeChange = () => {
  page.value = 1
  loadData()
}

const openCreate = () => {
  editing.value = false
  Object.assign(form, {
    id: 0,
    code: '',
    name: '',
    plan: 'basic'
  })
  dialogVisible.value = true
}

const openEdit = (row) => {
  editing.value = true
  Object.assign(form, {
    id: row.id,
    code: row.code,
    name: row.name,
    plan: row.plan
  })
  dialogVisible.value = true
}

const submitForm = async () => {
  if (!formRef.value) return

  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  saving.value = true
  try {
    const res = editing.value
      ? await updateTenant({
          id: form.id,
          name: form.name,
          plan: form.plan
        })
      : await createTenant({
          code: form.code,
          name: form.name,
          plan: form.plan
        })

    if (res.code === 0) {
      ElMessage.success(editing.value ? '修改成功' : '创建成功')
      dialogVisible.value = false
      await loadData()
    }
  } finally {
    saving.value = false
  }
}

const changeStatus = async (row) => {
  const nextStatus = row.status === 1 ? 2 : 1
  const action = nextStatus === 1 ? '启用' : '停用'

  try {
    await ElMessageBox.confirm(
      `确定要${action}高校「${row.name}」吗？`,
      '操作确认',
      { type: 'warning' }
    )
  } catch {
    return
  }

  const res = await setTenantStatus({
    id: row.id,
    status: nextStatus
  })

  if (res.code === 0) {
    ElMessage.success(`${action}成功`)
    await loadData()
  }
}

const formatTime = (value) => {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN', {
    hour12: false
  })
}

onMounted(loadData)
</script>
