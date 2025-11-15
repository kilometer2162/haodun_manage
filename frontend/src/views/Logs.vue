<template>
  <div class="logs-page">
    <div class="page-header">
      <h2>日志管理</h2>
    </div>

    <el-card style="margin-top: 20px">
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="用户名">
          <el-input v-model="searchForm.username" placeholder="请输入用户名" clearable />
        </el-form-item>
        <el-form-item label="操作类型">
          <el-select
            v-model="searchForm.action"
            placeholder="请选择"
            clearable
            filterable
            style="min-width: 140px"
          >
            <el-option
              v-for="item in actionOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="模块">
          <el-input v-model="searchForm.module" placeholder="请输入模块" clearable />
        </el-form-item>
        <el-form-item label="IP地址">
          <el-input v-model="searchForm.ip" placeholder="请输入IP地址" clearable />
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="searchForm.status"
            placeholder="请选择"
            clearable
            style="min-width: 140px"
          >
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="开始日期">
          <el-date-picker
            v-model="searchForm.start_date"
            type="date"
            placeholder="选择日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            clearable
          />
        </el-form-item>
        <el-form-item label="结束日期">
          <el-date-picker
            v-model="searchForm.end_date"
            type="date"
            placeholder="选择日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            clearable
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-table :data="logs" v-loading="loading" style="margin-top: 20px">
      <el-table-column prop="username" label="用户名" width="120" />
      <el-table-column prop="action" label="操作" width="100" />
      <el-table-column prop="module" label="模块" width="200" />
      <el-table-column prop="content" label="内容" show-overflow-tooltip />
      <el-table-column prop="ip" label="IP地址" width="150" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? '成功' : '失败' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="时间" width="180">
        <template #default="{ row }">
          {{ row.created_at ? formatTime(row.created_at) : '' }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button size="small" @click="handleView(row)">查看</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="page"
      v-model:page-size="pageSize"
      :total="total"
      :page-sizes="[10, 20, 50, 100]"
      layout="total, sizes, prev, pager, next, jumper"
      @size-change="loadLogs"
      @current-change="loadLogs"
      style="margin-top: 20px"
    />

    <el-dialog v-model="detailVisible" title="日志详情" width="600px">
      <el-descriptions :column="1" border v-if="currentLog">
        <el-descriptions-item label="ID">{{ currentLog.id }}</el-descriptions-item>
        <el-descriptions-item label="用户名">{{ currentLog.username }}</el-descriptions-item>
        <el-descriptions-item label="操作">{{ currentLog.action }}</el-descriptions-item>
        <el-descriptions-item label="模块">{{ currentLog.module }}</el-descriptions-item>
        <el-descriptions-item label="内容">{{ currentLog.content }}</el-descriptions-item>
        <el-descriptions-item label="IP地址">{{ currentLog.ip }}</el-descriptions-item>
        <el-descriptions-item label="用户代理">{{ currentLog.user_agent }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="currentLog.status === 1 ? 'success' : 'danger'">
            {{ currentLog.status === 1 ? '成功' : '失败' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="时间">{{ currentLog.created_at ? formatTime(currentLog.created_at) : '' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'
import { formatBeijingTime } from '@/utils/date'

const formatTime = (timeStr) => {
  if (!timeStr) return ''
  return formatBeijingTime(timeStr)
}

const loading = ref(false)
const logs = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const detailVisible = ref(false)
const currentLog = ref(null)
const actionOptions = ref([])
const statusOptions = ref([
  { label: '成功', value: 1 },
  { label: '失败', value: 0 }
])

const searchForm = reactive({
  username: '',
  action: '',
  module: '',
  ip: '',
  status: null,
  start_date: '',
  end_date: ''
})

const loadLogs = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      ...searchForm
    }
    // 移除空值
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === null) {
        delete params[key]
      }
    })
    
    const response = await api.get('/logs', { params })
    logs.value = response.data.data
    total.value = response.data.total
  } catch (error) {
    ElMessage.error('加载日志列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  page.value = 1
  loadLogs()
}

const handleReset = () => {
  Object.assign(searchForm, {
    username: '',
    action: '',
    module: '',
    ip: '',
    status: null,
    start_date: '',
    end_date: ''
  })
  handleSearch()
}

const loadLogOptions = async () => {
  try {
    const response = await api.get('/logs/options')
    actionOptions.value = (response.data.actions || []).map(action => ({
      label: action,
      value: action
    }))

    if (Array.isArray(response.data.statuses) && response.data.statuses.length > 0) {
      statusOptions.value = response.data.statuses
        .map(item => ({
          label: item.label,
          value: Number(item.value)
        }))
        .filter(item => !Number.isNaN(item.value))
    }
  } catch (error) {
    console.error('加载日志筛选项失败', error)
  }
}

const handleView = async (row) => {
  try {
    const response = await api.get(`/logs/${row.id}`)
    currentLog.value = response.data
    detailVisible.value = true
  } catch (error) {
    ElMessage.error('获取日志详情失败')
  }
}

onMounted(() => {
  loadLogOptions()
  loadLogs()
})
</script>

<style scoped>
.logs-page {
  background: #fff;
  padding: 20px;
  border-radius: 4px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-form {
  margin-bottom: 0;
}
</style>

