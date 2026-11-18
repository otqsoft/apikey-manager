<template>
  <div>
    <div class="toolbar">
      <el-form :inline="true" :model="query">
        <el-form-item label="关键字">
          <el-input
            v-model="query.keyword"
            placeholder="账号 / 昵称"
            clearable
            style="width: 220px"
            @keyup.enter="onSearch"
          />
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
      <el-table v-loading="loading" :data="tableData" border stripe empty-text="暂无登录日志">
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column label="登录账号" min-width="130" fixed="left" show-overflow-tooltip>
          <template #default="{ row }">
            <span>{{ row.username || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="昵称" min-width="130" show-overflow-tooltip>
          <template #default="{ row }">{{ row.display_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small" effect="light">
              {{ row.status === 1 ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="失败原因" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <span :class="{ 'fail-text': row.status === 0 }">{{ row.fail_reason || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="ip" label="IP" width="140" show-overflow-tooltip />
        <el-table-column label="客户端" min-width="240" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="ua-text">{{ row.user_agent || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="登录时间" width="170">
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
import { getLoginLogs } from '@/api/logs'
import type { LoginLogItem } from '@/types'
import { formatDateTime } from '@/utils/format'

const loading = ref(false)
const tableData = ref<LoginLogItem[]>([])
const total = ref(0)
const query = reactive({ page: 1, page_size: 10, keyword: '', status: undefined as number | undefined })

async function loadData() {
  loading.value = true
  try {
    const res = await getLoginLogs({
      page: query.page,
      page_size: query.page_size,
      keyword: query.keyword || undefined,
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
  query.status = undefined
  query.page = 1
  loadData()
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
.fail-text {
  color: #dc2626;
}
.ua-text {
  font-size: 12px;
  color: #64748b;
}
</style>
