<template>
  <div>
    <el-card shadow="never" class="toolbar">
      <el-form :inline="true" :model="query">
        <el-form-item label="供应商">
          <el-select v-model="query.provider_id" placeholder="全部供应商" clearable style="width: 180px">
            <el-option v-for="p in providerOptions" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" placeholder="模型名称" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">查询</el-button>
          <el-button plain :icon="RefreshLeft" @click="onReset">重置</el-button>
          <el-button type="primary" :icon="Plus" @click="openCreate">新增模型</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <el-table v-loading="loading" :data="tableData" border stripe empty-text="暂无模型数据，点击右上角「新增模型」录入">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="模型名称" min-width="250" show-overflow-tooltip />
        <el-table-column prop="provider_name" label="所属供应商" min-width="140">
          <template #default="{ row }">{{ row.provider_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="命中输入价(元/百万Token)" width="200" align="right">
          <template #default="{ row }">{{ Number(row.input_price_cached ?? 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="未命中输入价(元/百万Token)" width="220" align="right">
          <template #default="{ row }">{{ Number(row.input_price ?? 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="输出价(元/百万Token)" width="200" align="right">
          <template #default="{ row }">{{ Number(row.output_price ?? 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="上下文长度" width="110" align="center">
          <template #default="{ row }">{{ row.context_length ? formatContext(row.context_length) : '—' }}</template>
        </el-table-column>
        <el-table-column label="思考模式" width="86" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.support_thinking" type="success" size="small">支持</el-tag>
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="多模态" width="80" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.support_multimodal" type="primary" size="small">支持</el-tag>
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="86" align="center">
          <template #default="{ row }">
            <el-switch :model-value="row.status === 1" :loading="row._switching" @change="(val: boolean) => onToggleStatus(row, val)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
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
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑模型' : '新增模型'" width="640px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="140px">
        <el-form-item label="模型名称" prop="name">
          <el-input v-model="form.name" placeholder="如 gpt-4o、claude-3-5-sonnet" style="width: 100%" />
        </el-form-item>
        <el-form-item label="所属供应商" prop="provider_id">
          <el-select v-model="form.provider_id" placeholder="选择启用中的供应商" style="width: 100%">
            <el-option v-for="p in providerOptions" :key="p.id" :label="`${p.name}（${p.api_format}）`" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="命中价(元/百万Token)">
              <el-input-number v-model="form.input_price_cached" :min="0" :precision="2" :step="0.1" controls-position="right" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="未命中价(元/百万Token)">
              <el-input-number v-model="form.input_price" :min="0" :precision="2" :step="0.1" controls-position="right" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="输出价(元/百万Token)">
              <el-input-number v-model="form.output_price" :min="0" :precision="2" :step="0.1" controls-position="right" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <div class="price-tip">价格单位：元 / 百万 Token；支持两位小数</div>
        <el-form-item label="上下文长度" prop="context_length">
          <el-input-number v-model="form.context_length" :min="0" :step="1000" controls-position="right" style="width: 100%" />
          <div class="tip-text">录入原始 Token 数，列表按 M（百万 Token）展示</div>
        </el-form-item>
        <el-form-item label="支持思考模式">
          <el-switch v-model="form.support_thinking" />
        </el-form-item>
        <el-form-item label="支持多模态">
          <el-switch v-model="form.support_multimodal" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="statusSwitch" active-text="启用" inactive-text="停用" />
        </el-form-item>
        <el-form-item label="备注">
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
import { Plus, RefreshLeft, Search } from '@element-plus/icons-vue'
import { createModel, deleteModel, getModelPage, updateModel, updateModelStatus } from '@/api/models'
import { getEnabledProviders } from '@/api/providers'
import type { ModelForm, ModelItem, Provider } from '@/types'
import { formatContext } from '@/utils/format'

type Row = ModelItem & { _switching?: boolean }

const loading = ref(false)
const submitting = ref(false)
const tableData = ref<Row[]>([])
const total = ref(0)
const providerOptions = ref<Provider[]>([])

const query = reactive({ page: 1, page_size: 10, keyword: '', provider_id: '' as number | '' })

async function loadData() {
  loading.value = true
  try {
    const res = await getModelPage({
      page: query.page,
      page_size: query.page_size,
      keyword: query.keyword || undefined,
      provider_id: query.provider_id === '' ? undefined : query.provider_id
    })
    tableData.value = res.list
    total.value = res.total
  } finally {
    loading.value = false
  }
}

async function loadProviders() {
  const res = await getEnabledProviders()
  providerOptions.value = res.list
}

function onSearch() {
  query.page = 1
  loadData()
}
function onReset() {
  query.keyword = ''
  query.provider_id = ''
  query.page = 1
  loadData()
}

onMounted(() => {
  loadProviders()
  loadData()
})

// ---------------- 弹窗表单 ----------------
const dialogVisible = ref(false)
const formRef = ref<FormInstance>()
const statusSwitch = ref(true)
const form = reactive<ModelForm & { id?: number }>({
  id: undefined,
  name: '',
  provider_id: undefined,
  input_price_cached: 0,
  input_price: 0,
  output_price: 0,
  context_length: 0,
  support_thinking: false,
  support_multimodal: false,
  status: 1,
  remark: ''
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入模型名称', trigger: 'blur' }],
  provider_id: [{ required: true, message: '请选择所属供应商', trigger: 'change' }]
}

function openCreate() {
  dialogVisible.value = true
}

function openEdit(row: Row) {
  Object.assign(form, {
    id: row.id,
    name: row.name,
    provider_id: row.provider_id,
    input_price_cached: row.input_price_cached || 0,
    input_price: row.input_price || 0,
    output_price: row.output_price || 0,
    context_length: row.context_length || 0,
    support_thinking: !!row.support_thinking,
    support_multimodal: !!row.support_multimodal,
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
      await updateModel(form.id, form)
      ElMessage.success('更新成功')
    } else {
      await createModel(form)
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
    provider_id: undefined,
    input_price_cached: 0,
    input_price: 0,
    output_price: 0,
    context_length: 0,
    support_thinking: false,
    support_multimodal: false,
    status: 1,
    remark: ''
  })
  statusSwitch.value = true
}

async function onToggleStatus(row: Row, val: boolean) {
  row._switching = true
  try {
    await updateModelStatus(row.id, val ? 1 : 0)
    row.status = val ? 1 : 0
    ElMessage.success(val ? '已启用' : '已停用')
  } finally {
    row._switching = false
  }
}

async function onDelete(row: Row) {
  await ElMessageBox.confirm(`确定删除模型「${row.name}」吗？`, '删除确认', { type: 'warning' })
  await deleteModel(row.id)
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
.muted {
  color: #c0c4cc;
}
.price-tip {
  margin: -8px 0 12px 130px;
  font-size: 12px;
  color: #909399;
}
.tip-text {
  margin-left: 10px;
  font-size: 12px;
  color: #909399;
}
</style>
