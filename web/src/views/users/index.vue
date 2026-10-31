<template>
  <div>
    <div class="toolbar">
      <el-form :inline="true" :model="query">
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" placeholder="用户名 / 显示名" clearable style="width: 220px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" plain :icon="Search" @click="onSearch">查询</el-button>
          <el-button plain :icon="RefreshLeft" @click="onReset">重置</el-button>
          <el-button type="primary" :icon="Plus" @click="openCreate">新增用户</el-button>
        </el-form-item>
      </el-form>
    </div>

    <el-card shadow="never">
      <el-table v-loading="loading" :data="tableData" border stripe empty-text="暂无用户数据，点击右上角「新增用户」录入">
        <el-table-column prop="id" label="ID" width="60" fixed="left" show-overflow-tooltip />
        <el-table-column prop="username" label="登录账号" min-width="100" fixed="left" show-overflow-tooltip />
        <el-table-column prop="display_name" label="显示名" min-width="150" show-overflow-tooltip />
        <el-table-column label="角色" width="120" align="left">
          <template #default="{ row }">
            <el-tag :type="row.role === 'admin' ? 'danger' : 'info'">{{ row.role === 'admin' ? '管理员' : '普通用户' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-switch
              :model-value="row.status === 1"
              :disabled="row.id === store.user?.id"
              :loading="row._switching"
              @change="(val: boolean) => onToggleStatus(row, val)"
            />
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="320" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="warning" @click="openResetPwd(row)">重置密码</el-button>
            <!-- <el-button size="small" type="info" @click="openUserKeys(row)">已分配 Key</el-button> -->
            <el-button
              size="small"
              type="danger"
              :disabled="row.id === store.user?.id"
              @click="onDelete(row)"
            >删除</el-button>
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

    <!-- 新增用户 -->
    <el-dialog v-model="createVisible" title="新增用户" width="460px" @closed="resetCreateForm">
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="92px">
        <el-form-item label="登录账号" prop="username">
          <el-input v-model="createForm.username" placeholder="唯一登录名" />
        </el-form-item>
        <el-form-item label="初始密码" prop="password">
          <el-input v-model="createForm.password" type="password" show-password placeholder="至少 6 位" />
        </el-form-item>
        <el-form-item label="显示名" prop="display_name">
          <el-input v-model="createForm.display_name" placeholder="可选" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-radio-group v-model="createForm.role">
            <el-radio value="user">普通用户</el-radio>
            <el-radio value="admin">管理员</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="createStatusSwitch" active-text="启用" inactive-text="停用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onCreateSubmit">保存</el-button>
      </template>
    </el-dialog>

    <!-- 编辑用户（不含密码） -->
    <el-dialog v-model="editVisible" title="编辑用户" width="460px">
      <el-form ref="editFormRef" :model="editForm" :rules="editRules" label-width="92px">
        <el-form-item label="登录账号">
          <el-input :model-value="editForm.username" disabled />
        </el-form-item>
        <el-form-item label="显示名" prop="display_name">
          <el-input v-model="editForm.display_name" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-radio-group v-model="editForm.role">
            <el-radio value="user">普通用户</el-radio>
            <el-radio value="admin">管理员</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="editStatusSwitch" active-text="启用" inactive-text="停用" :disabled="editForm.id === store.user?.id" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onEditSubmit">保存</el-button>
      </template>
    </el-dialog>

    <!-- 重置密码 -->
    <el-dialog v-model="resetVisible" title="重置密码" width="420px" @closed="resetPwdForm">
      <el-form ref="resetFormRef" :model="resetForm" :rules="resetRules" label-width="92px">
        <el-form-item label="目标用户">
          <el-input :model-value="resetTargetName" disabled />
        </el-form-item>
        <el-form-item label="新密码" prop="password">
          <el-input v-model="resetForm.password" type="password" show-password placeholder="至少 6 位" />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirm">
          <el-input v-model="resetForm.confirm" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onResetPwdSubmit">确定重置</el-button>
      </template>
    </el-dialog>

    <!-- 已分配 Key 查看 -->
    <el-dialog v-model="keysVisible" title="该用户已分配的 Key" width="640px">
      <el-table :data="userKeyList" border size="small">
        <el-table-column prop="name" label="名称" min-width="130" />
        <el-table-column label="掩码 Key" min-width="170">
          <template #default="{ row }">{{ row.api_key || '******' }}</template>
        </el-table-column>
        <el-table-column label="有效期" width="110">
          <template #default="{ row }">{{ row.expire_date || '永久' }}</template>
        </el-table-column>
        <el-table-column label="已用 / 限额" min-width="170">
          <template #default="{ row }">
            ¥{{ Number(row.used_quota || 0).toFixed(2) }} /
            {{ row.max_quota && Number(row.max_quota) > 0 ? `¥${Number(row.max_quota).toFixed(2)}` : '不限' }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="keysVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, RefreshLeft, Search } from '@element-plus/icons-vue'
import {
  createUser,
  deleteUser,
  getUserPage,
  getUserKeys,
  resetUserPassword,
  updateUser,
  updateUserStatus
} from '@/api/users'
import type { ApiKeyItem, RoleType, UserItem } from '@/types'
import { useUserStore } from '@/store/user'
import { formatDateTime } from '@/utils/format'

const store = useUserStore()

type Row = UserItem & { _switching?: boolean }

const loading = ref(false)
const submitting = ref(false)
const tableData = ref<Row[]>([])
const total = ref(0)
const query = reactive({ page: 1, page_size: 10, keyword: '' })

async function loadData() {
  loading.value = true
  try {
    const res = await getUserPage({ page: query.page, page_size: query.page_size, keyword: query.keyword || undefined })
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
  query.page = 1
  loadData()
}
onMounted(loadData)

// ---------------- 新增 ----------------
const createVisible = ref(false)
const createFormRef = ref<FormInstance>()
const createStatusSwitch = ref(true)
const createForm = reactive({
  username: '',
  password: '',
  display_name: '',
  role: 'user' as RoleType
})
const createRules: FormRules = {
  username: [{ required: true, message: '请输入登录账号', trigger: 'blur' }],
  password: [
    { required: true, message: '请输入初始密码', trigger: 'blur' },
    { min: 6, message: '密码至少 6 位', trigger: 'blur' }
  ],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }]
}

function openCreate() {
  createVisible.value = true
}

async function onCreateSubmit() {
  await createFormRef.value?.validate()
  submitting.value = true
  try {
    await createUser({
      username: createForm.username.trim(),
      password: createForm.password,
      role: createForm.role,
      display_name: createForm.display_name,
      status: createStatusSwitch.value ? 1 : 0
    })
    ElMessage.success('新增用户成功')
    createVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

function resetCreateForm() {
  createFormRef.value?.clearValidate()
  Object.assign(createForm, { username: '', password: '', display_name: '', role: 'user' })
  createStatusSwitch.value = true
}

// ---------------- 编辑 ----------------
const editVisible = ref(false)
const editFormRef = ref<FormInstance>()
const editStatusSwitch = ref(true)
const editForm = reactive({ id: 0, username: '', display_name: '', role: 'user' as RoleType })
const editRules: FormRules = {
  display_name: [{ required: true, message: '请输入显示名', trigger: 'blur' }],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }]
}

function openEdit(row: Row) {
  Object.assign(editForm, { id: row.id, username: row.username, display_name: row.display_name, role: row.role })
  editStatusSwitch.value = row.status === 1
  editVisible.value = true
}

async function onEditSubmit() {
  await editFormRef.value?.validate()
  submitting.value = true
  try {
    await updateUser(editForm.id, {
      role: editForm.role,
      display_name: editForm.display_name,
      status: editStatusSwitch.value ? 1 : 0
    })
    ElMessage.success('更新成功')
    editVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

// ---------------- 启停 ----------------
async function onToggleStatus(row: Row, val: boolean) {
  row._switching = true
  try {
    await updateUserStatus(row.id, val ? 1 : 0)
    row.status = val ? 1 : 0
    ElMessage.success(val ? '已启用' : '已停用')
  } finally {
    row._switching = false
  }
}

// ---------------- 删除（禁止删自己） ----------------
async function onDelete(row: Row) {
  if (row.id === store.user?.id) {
    ElMessage.warning('不能删除当前登录账号')
    return
  }
  await ElMessageBox.confirm(`确定删除用户「${row.username}」吗？将同时清理其 Key 分配关系。`, '删除确认', { type: 'warning' })
  await deleteUser(row.id)
  ElMessage.success('删除成功')
  loadData()
}

// ---------------- 重置密码 ----------------
const resetVisible = ref(false)
const resetFormRef = ref<FormInstance>()
const resetTargetId = ref(0)
const resetTargetName = ref('')
const resetForm = reactive({ password: '', confirm: '' })
const resetRules: FormRules = {
  password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码至少 6 位', trigger: 'blur' }
  ],
  confirm: [
    { required: true, message: '请再次确认密码', trigger: 'blur' },
    {
      validator: (_r: unknown, v: string, cb: (e?: Error) => void) => {
        if (v !== resetForm.password) cb(new Error('两次输入不一致'))
        else cb()
      },
      trigger: 'blur'
    }
  ]
}

function openResetPwd(row: Row) {
  resetTargetId.value = row.id
  resetTargetName.value = `${row.display_name || row.username}（${row.username}）`
  resetVisible.value = true
}

async function onResetPwdSubmit() {
  await resetFormRef.value?.validate()
  submitting.value = true
  try {
    await resetUserPassword(resetTargetId.value, resetForm.password)
    ElMessage.success('密码已重置')
    resetVisible.value = false
  } finally {
    submitting.value = false
  }
}

function resetPwdForm() {
  resetFormRef.value?.clearValidate()
  resetForm.password = ''
  resetForm.confirm = ''
}

// ---------------- 已分配 Key ----------------
const keysVisible = ref(false)
const userKeyList = ref<ApiKeyItem[]>([])

async function openUserKeys(row: Row) {
  const res = await getUserKeys(row.id)
  // 兼容数组或分页结构两种返回
  userKeyList.value = Array.isArray(res) ? res : res.list || []
  keysVisible.value = true
}
</script>

<style scoped>
.toolbar {
  padding: 16px 16px 8px;
}
.pager {
  margin-top: 16px;
  justify-content: flex-end;
}
</style>
