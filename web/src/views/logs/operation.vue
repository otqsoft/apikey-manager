<template>
  <div>
    <div class="toolbar">
      <el-form :inline="true" :model="query">
        <el-form-item label="关键字">
          <el-input
            v-model="query.keyword"
            placeholder="账号 / 昵称 / 路径 / 动作"
            clearable
            style="width: 240px"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item label="模块">
          <el-select v-model="query.module" placeholder="全部模块" clearable style="width: 150px">
            <el-option v-for="m in moduleOptions" :key="m" :label="m" :value="m" />
          </el-select>
        </el-form-item>
        <el-form-item label="结果">
          <el-select v-model="query.status" placeholder="全部" clearable style="width: 110px">
            <el-option label="成功" :value="1" />
            <el-option label="失败" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" plain :icon="Search" @click="onSearch">查询</el-button>
          <el-button plain :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <el-card shadow="never">
      <el-table v-loading="loading" :data="tableData" border stripe empty-text="暂无操作日志">
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column label="操作人" min-width="200" fixed="left" show-overflow-tooltip>
          <template #default="{ row }">
            <span>{{ row.display_name || row.username || '-' }}</span>
            <span v-if="row.display_name && row.username" class="sub-text">（{{ row.username }}）</span>
          </template>
        </el-table-column>
        <el-table-column prop="module" label="模块" width="120" show-overflow-tooltip />
        <el-table-column prop="action" label="动作" width="100" show-overflow-tooltip />
        <el-table-column label="请求" min-width="240" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tag size="small" :type="methodTagType(row.method)" effect="plain" class="method-tag">
              {{ row.method }}
            </el-tag>
            <span class="path-text">{{ row.path }}</span>
          </template>
        </el-table-column>
        <el-table-column label="详情" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="detail-text">{{ row.detail || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="ip" label="IP" width="130" show-overflow-tooltip />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small" effect="light">
              {{ row.status === 1 ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作时间" width="170">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
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
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { RefreshLeft, Search } from '@element-plus/icons-vue'
import { getOperationLogs } from '@/api/logs'
import type { OperationLogItem } from '@/types'
import { formatDateTime } from '@/utils/format'

const moduleOptions = ['供应商管理', '模型管理', 'API-KEY 管理', '用户管理', '对话测试', '认证', '其他']

const loading = ref(false)
const tableData = ref<OperationLogItem[]>([])
const total = ref(0)
const query = reactive({ page: 1, page_size: 10, keyword: '', module: '', status: undefined as number | undefined })

async function loadData() {
  loading.value = true
  try {
    const res = await getOperationLogs({
      page: query.page,
      page_size: query.page_size,
      keyword: query.keyword || undefined,
      module: query.module || undefined,
      status: query.status
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
  query.module = ''
  query.status = undefined
  query.page = 1
  loadData()
}
function methodTagType(method: string): 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  if (method === 'POST') return 'success'
  if (method === 'PUT') return 'warning'
  if (method === 'DELETE') return 'danger'
  return 'info'
}
onMounted(loadData)
</script>

<style scoped>
.toolbar {
  padding: 16px 16px 8px;
}
.pager {
  margin-top: 16px;
  justify-content: flex-end;
}
.sub-text {
  color: #94a3b8;
}
.method-tag {
  margin-right: 6px;
}
.path-text {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  color: #475569;
}
.detail-text {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  color: #64748b;
}
</style>
