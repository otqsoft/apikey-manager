<template>
  <div>
    <el-card shadow="never" class="toolbar">
      <el-form :inline="true" :model="query">
        <el-form-item label="供应商">
          <el-select v-model="query.provider_id" placeholder="全部供应商" clearable style="width: 180px" @change="onSearch">
            <el-option v-for="p in providerOptions" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" placeholder="名称 / 掩码 key" clearable style="width: 220px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" plain :icon="Search" @click="onSearch">查询</el-button>
          <el-button plain :icon="RefreshLeft" @click="onReset">重置</el-button>
          <el-button type="primary" :icon="Plus" @click="openCreate">新增 Key</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <el-table v-loading="loading" :data="tableData" border stripe empty-text="暂无 Key 数据，点击右上角「新增 Key」录入">
        <el-table-column prop="id" label="ID" width="60" fixed="left" />
        <el-table-column prop="name" label="名称" min-width="100" fixed="left" show-overflow-tooltip />
        <el-table-column label="API-KEY" min-width="180" fixed="left" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="masked-key">{{ row.api_key || '******' }}</span>
            <el-tooltip content="获取明文并复制" placement="top">
              <el-button link type="primary" :icon="CopyDocument" @click="onCopy(row)" />
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="provider_name" label="供应商" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">{{ row.provider_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="有效期" width="120">
          <template #default="{ row }">{{ row.expire_date || '永久' }}</template>
        </el-table-column>
        <el-table-column label="额度使用" min-width="200">
          <template #default="{ row }">
            <template v-if="row.max_quota && Number(row.max_quota) > 0">
              <el-progress
                :percentage="quotaPercent(row)"
                :status="quotaPercent(row) >= 90 ? 'exception' : undefined"
                :stroke-width="12"
              />
              <div class="quota-text">已用 ¥{{ Number(row.used_quota || 0).toFixed(2) }} / 限额 ¥{{ Number(row.max_quota).toFixed(2) }}</div>
            </template>
            <template v-else>
              <el-tag type="info" size="small">不限额度</el-tag>
              <div class="quota-text">已用 ¥{{ Number(row.used_quota || 0).toFixed(2) }}</div>
            </template>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center" >
          <template #default="{ row }">
            <el-switch :model-value="row.status === 1" :loading="row._switching" @change="(val: boolean) => onToggleStatus(row, val)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <!-- <el-button size="small" type="warning" @click="openAssign(row)">分配用户</el-button> -->
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
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑 API-KEY' : '新增 API-KEY'" width="600px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="备注名，如 OpenAI 主力 Key" style="width: 100%" />
        </el-form-item>
        <el-form-item label="所属供应商" prop="provider_id">
          <el-select v-model="form.provider_id" placeholder="选择启用中的供应商" style="width: 100%">
            <el-option v-for="p in providerOptions" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="API-KEY" prop="api_key">
          <el-input v-model="form.api_key" placeholder="sk-... 完整明文，仅本地落库" clearable />
          <div v-if="form.id" class="tip-text" style="margin-left: 0; width: 100%">留空表示不修改原 Key</div>
        </el-form-item>
        <el-form-item label="有效日期">
          <el-date-picker v-model="form.expire_date" type="date" value-format="YYYY-MM-DD" placeholder="空表示永久有效" style="width: 100%" />
        </el-form-item>
        <el-form-item label="最大限额(元)">
          <el-input-number v-model="form.max_quota" :min="0" :precision="2" :step="10" controls-position="right" style="width: 100%" />
          <div class="tip-text">0 表示不限额度</div>
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

    <!-- 分配用户弹窗 -->
    <el-dialog v-model="assignVisible" title="分配用户" width="480px">
      <el-alert title="勾选后将全量覆盖该 Key 的分配关系（PUT /api/keys/:id/users）" type="info" :closable="false" class="assign-tip" />
      <el-select v-model="selectedUserIds" multiple filterable placeholder="选择可使用该 Key 的用户" style="width: 100%">
        <el-option v-for="u in allUsers" :key="u.id" :label="`${u.display_name || u.username}（${u.username}）`" :value="u.id" />
      </el-select>
      <template #footer>
        <el-button @click="assignVisible = false">取消</el-button>
        <el-button type="primary" :loading="assigning" @click="onSaveAssign">保存分配</el-button>
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
  createKey,
  deleteKey,
  getKeyDetail,
  getKeyPage,
  getKeyUsers,
  saveKeyUsers,
  updateKey,
  updateKeyStatus
} from '@/api/keys'
import { getEnabledProviders } from '@/api/providers'
import { getUserPage } from '@/api/users'
import type { ApiKeyForm, ApiKeyItem, Provider, UserItem } from '@/types'
import { copyText } from '@/utils/clipboard'

type Row = ApiKeyItem & { _switching?: boolean }

const loading = ref(false)
const submitting = ref(false)
const tableData = ref<Row[]>([])
const total = ref(0)
const providerOptions = ref<Provider[]>([])

const query = reactive({ page: 1, page_size: 10, keyword: '', provider_id: '' as number | '' })

async function loadData() {
  loading.value = true
  try {
    const res = await getKeyPage({
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

/** 额度百分比；max_quota=0 不参与进度条 */
function quotaPercent(row: Row): number {
  const max = Number(row.max_quota || 0)
  if (max <= 0) return 0
  return Math.min(100, Math.round((Number(row.used_quota || 0) / max) * 100))
}

// ---------------- 表单弹窗 ----------------
const dialogVisible = ref(false)
const formRef = ref<FormInstance>()
const statusSwitch = ref(true)
const form = reactive<ApiKeyForm & { id?: number }>({
  id: undefined,
  name: '',
  provider_id: undefined,
  api_key: '',
  expire_date: null,
  max_quota: 0,
  status: 1,
  remark: ''
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  provider_id: [{ required: true, message: '请选择所属供应商', trigger: 'change' }],
  api_key: [{ required: true, message: '请输入 API-KEY', trigger: 'blur' }]
}

function openCreate() {
  dialogVisible.value = true
}

function openEdit(row: Row) {
  Object.assign(form, {
    id: row.id,
    name: row.name,
    provider_id: row.provider_id,
    api_key: '', // 编辑不回填明文，留空不修改
    expire_date: row.expire_date,
    max_quota: row.max_quota || 0,
    status: row.status,
    remark: row.remark || ''
  })
  statusSwitch.value = row.status === 1
  dialogVisible.value = true
}

async function onSubmit() {
  // 编辑时 api_key 非必填，动态校验
  if (!form.id) await formRef.value?.validate()
  else await formRef.value?.validateField(['name', 'provider_id'])
  submitting.value = true
  form.status = statusSwitch.value ? 1 : 0
  const payload: ApiKeyForm = { ...form }
  if (form.id && !payload.api_key) delete payload.api_key
  try {
    if (form.id) {
      await updateKey(form.id, payload)
      ElMessage.success('更新成功')
    } else {
      await createKey(payload)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

function resetForm() {
  formRef.value?.clearValidate()
  Object.assign(form, {
    id: undefined,
    name: '',
    provider_id: undefined,
    api_key: '',
    expire_date: null,
    max_quota: 0,
    status: 1,
    remark: ''
  })
  statusSwitch.value = true
}

async function onToggleStatus(row: Row, val: boolean) {
  row._switching = true
  try {
    await updateKeyStatus(row.id, val ? 1 : 0)
    row.status = val ? 1 : 0
    ElMessage.success(val ? '已启用' : '已停用')
  } finally {
    row._switching = false
  }
}

async function onDelete(row: Row) {
  await ElMessageBox.confirm(`确定删除 Key「${row.name}」吗？将同时清理用户分配关系。`, '删除确认', { type: 'warning' })
  await deleteKey(row.id)
  ElMessage.success('删除成功')
  loadData()
}

/** 复制完整明文：详情接口取完整 api_key */
async function onCopy(row: Row) {
  const detail = await getKeyDetail(row.id)
  if (!detail.api_key) {
    ElMessage.warning('该 Key 无明文内容')
    return
  }
  await copyText(detail.api_key, '完整 Key')
}

// ---------------- 分配用户 ----------------
const assignVisible = ref(false)
const assigning = ref(false)
const currentKeyId = ref<number | null>(null)
const selectedUserIds = ref<number[]>([])
const allUsers = ref<UserItem[]>([])

async function openAssign(row: Row) {
  currentKeyId.value = row.id
  assignVisible.value = true
  const [usersRes, assignedRes] = await Promise.all([getUserPage({ page: 1, page_size: 200 }), getKeyUsers(row.id)])
  allUsers.value = usersRes.list
  selectedUserIds.value = assignedRes.user_ids || []
}

async function onSaveAssign() {
  if (!currentKeyId.value) return
  assigning.value = true
  try {
    await saveKeyUsers(currentKeyId.value, selectedUserIds.value)
    ElMessage.success('分配关系已保存')
    assignVisible.value = false
    loadData()
  } finally {
    assigning.value = false
  }
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
.masked-key {
  font-family: monospace;
  color: #606266;
  margin-right: 6px;
}
.quota-text {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}
.tip-text {
  margin-left: 10px;
  font-size: 12px;
  color: #909399;
}
.assign-tip {
  margin-bottom: 14px;
}
</style>
