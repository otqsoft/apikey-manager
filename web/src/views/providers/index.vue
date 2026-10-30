<template>
  <div>
    <!-- 搜索筛选区 -->
    <el-card shadow="never" class="toolbar">
      <el-form :inline="true" :model="query">
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" placeholder="名称 / base_url" clearable style="width: 220px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="query.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="启用" :value="1" />
            <el-option label="停用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" plain :icon="Search" @click="onSearch">查询</el-button>
          <el-button plain :icon="RefreshLeft" @click="onReset">重置</el-button>
          <el-button type="primary" :icon="Plus" @click="openCreate">新增供应商</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 表格 -->
    <el-card shadow="never">
      <el-table v-loading="loading" :data="tableData" border stripe empty-text="暂无供应商数据，点击右上角「新增供应商」录入">
        <el-table-column prop="id" label="ID" width="60" fixed="left" />
        <el-table-column prop="name" label="供应商名称" min-width="140" show-overflow-tooltip fixed="left">
          <template #default="{ row }">
            <a
              v-if="row.platform_url"
              :href="row.platform_url"
              target="_blank"
              rel="noopener noreferrer"
              class="name-link"
            >{{ row.name }}</a>
            <span v-else>{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="协议格式" width="120">
          <template #default="{ row }">
            <el-tag :type="row.api_format === 'openai' ? 'primary' : 'warning'">{{ row.api_format }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Base URL" min-width="260">
          <template #default="{ row }">
            <div class="copy-cell">
              <span class="cell-text" :title="row.base_url">{{ row.base_url }}</span>
              <el-tooltip content="复制 Base URL" placement="top">
                <el-button link type="primary" :icon="CopyDocument" @click="copyText(row.base_url, 'Base URL')" />
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="余额（元）" width="120">
          <template #default="{ row }">¥{{ Number(row.balance || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="max_concurrency" label="最大并发" width="100" align="center" />
        <el-table-column prop="timeout_seconds" label="超时(秒)" width="100" align="center" />
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-switch
              :model-value="row.status === 1"
              :loading="row._switching"
              @change="(val: boolean) => onToggleStatus(row, val)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="200" show-overflow-tooltip />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.page_size"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        class="pager"
        @size-change="loadData"
        @current-change="loadData"
      />
    </el-card>

    <!-- 新增 / 编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑供应商' : '新增供应商'" width="560px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="供应商名称" prop="name">
          <el-input v-model="form.name" placeholder="如 OpenAI 官方" />
        </el-form-item>
        <el-form-item label="协议格式" prop="api_format">
          <el-radio-group v-model="form.api_format">
            <el-radio value="openai">OpenAI 兼容</el-radio>
            <el-radio value="anthropic">Anthropic</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="Base URL" prop="base_url">
          <el-input v-model="form.base_url" placeholder="OpenAI 如 https://api.openai.com/v1" />
        </el-form-item>
        <el-form-item label="平台网址" prop="platform_url">
          <el-input v-model="form.platform_url" placeholder="可选，API 网页后台地址" />
        </el-form-item>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="余额（元）" prop="balance">
              <el-input-number v-model="form.balance" :min="0" :precision="2" :step="10" controls-position="right" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="最大并发" prop="max_concurrency">
              <el-input-number v-model="form.max_concurrency" :min="1" :max="1000" :step="1" controls-position="right" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="超时(秒)" prop="timeout_seconds">
          <el-input-number v-model="form.timeout_seconds" :min="1" :max="600" :step="10" controls-position="right" style="width: 100%" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="statusSwitch" active-text="启用" inactive-text="停用" />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="可选" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CopyDocument, Plus, RefreshLeft, Search } from '@element-plus/icons-vue'
import {
  createProvider,
  deleteProvider,
  getProviderPage,
  updateProvider,
  updateProviderStatus
} from '@/api/providers'
import type { Provider, ProviderForm } from '@/types'
import { copyText } from '@/utils/clipboard'

type Row = Provider & { _switching?: boolean }

const loading = ref(false)
const submitting = ref(false)
const tableData = ref<Row[]>([])
const total = ref(0)

const query = reactive({ page: 1, page_size: 10, keyword: '', status: '' as number | '' })

async function loadData() {
  loading.value = true
  try {
    const res = await getProviderPage({
      page: query.page,
      page_size: query.page_size,
      keyword: query.keyword || undefined,
      status: query.status === '' ? undefined : query.status
    })
    tableData.value = res.list
    total.value = res.total
  } finally {
    loading.value = false
  }
}

function onSearch() {
  query.page = 1
  loadData()
}
function onReset() {
  query.keyword = ''
  query.status = ''
  query.page = 1
  loadData()
}

onMounted(loadData)


// ---------------- 弹窗表单 ----------------
const dialogVisible = ref(false)
const formRef = ref<FormInstance>()
const statusSwitch = ref(true)
const form = reactive<ProviderForm & { id?: number }>({
  id: undefined,
  name: '',
  api_format: 'openai',
  base_url: '',
  platform_url: '',
  balance: 0,
  max_concurrency: 10,
  timeout_seconds: 60,
  status: 1,
  remark: ''
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入供应商名称', trigger: 'blur' }],
  api_format: [{ required: true, message: '请选择协议格式', trigger: 'change' }],
  base_url: [{ required: true, message: '请输入 Base URL', trigger: 'blur' }]
}

function openCreate() {
  dialogVisible.value = true
}

function openEdit(row: Row) {
  Object.assign(form, {
    id: row.id,
    name: row.name,
    api_format: row.api_format,
    base_url: row.base_url,
    platform_url: row.platform_url || '',
    balance: row.balance || 0,
    max_concurrency: row.max_concurrency || 10,
    timeout_seconds: row.timeout_seconds || 60,
    status: row.status,
    remark: row.remark || ''
  })
  statusSwitch.value = row.status === 1
  dialogVisible.value = true
}

async function onSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  form.status = statusSwitch.value ? 1 : 0
  try {
    if (form.id) {
      await updateProvider(form.id, form)
      ElMessage.success('更新成功')
    } else {
      await createProvider(form)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

function resetForm() {
  formRef.value?.resetFields()
  Object.assign(form, {
    id: undefined,
    name: '',
    api_format: 'openai',
    base_url: '',
    platform_url: '',
    balance: 0,
    max_concurrency: 10,
    timeout_seconds: 60,
    status: 1,
    remark: ''
  })
  statusSwitch.value = true
}

async function onToggleStatus(row: Row, val: boolean) {
  row._switching = true
  try {
    await updateProviderStatus(row.id, val ? 1 : 0)
    row.status = val ? 1 : 0
    ElMessage.success(val ? '已启用' : '已停用')
  } finally {
    row._switching = false
  }
}

async function onDelete(row: Row) {
  await ElMessageBox.confirm(`确定删除供应商「${row.name}」吗？若名下存在模型或密钥，后端将拦截删除。`, '删除确认', { type: 'warning' })
  await deleteProvider(row.id)
  ElMessage.success('删除成功')
  loadData()
}
</script>

<style scoped>
.toolbar {
  margin-bottom: 16px;
}
.pager {
  margin-top: 16px;
  justify-content: flex-end;
}
.cell-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}
/* 供应商名称超链接：打开平台网址 */
.name-link {
  color: #2563eb;
  text-decoration: none;
  cursor: pointer;
}
.name-link:hover {
  text-decoration: underline;
}
/* 文本+复制按钮的单元格：按钮恒可见，文本截断省略 */
.copy-cell {
  display: flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
}
.copy-cell .el-button {
  flex-shrink: 0;
}
</style>
