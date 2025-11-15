<template>
  <div class="system-monitor">
    <el-row :gutter="20">
      <el-col :xs="24" :sm="24" :md="12" :lg="12">
        <div class="stack-column">
          <el-card class="metric-card">
            <template #header>
              <div class="card-header">
                <span>内存使用情况</span>
              </div>
            </template>
            <div class="card-content vertical">
              <el-progress
                :percentage="metrics?.memory?.percent || 0"
                :stroke-width="18"
                color="#67C23A"
                class="linear-progress"
              />
              <div class="metric-summary">
                <div>
                  <span class="label">总计</span>
                  <span>{{ formatBytes(metrics?.memory?.total) }}</span>
                </div>
                <div>
                  <span class="label">已用</span>
                  <span>{{ formatBytes(metrics?.memory?.used) }}</span>
                </div>
                <div>
                  <span class="label">可用</span>
                  <span>{{ formatBytes(metrics?.memory?.free) }}</span>
                </div>
              </div>
            </div>
          </el-card>

          <el-card class="metric-card">
            <template #header>
              <div class="card-header">
                <span>磁盘使用率</span>
              </div>
            </template>
            <div class="card-content vertical">
              <el-progress
                :percentage="metrics?.disk?.percent || 0"
                :stroke-width="18"
                color="#E6A23C"
                class="linear-progress"
              />
              <div class="metric-summary">
                <div>
                  <span class="label">挂载点</span>
                  <span>{{ metrics?.disk?.path || '/' }}</span>
                </div>
                <div>
                  <span class="label">总计</span>
                  <span>{{ formatBytes(metrics?.disk?.total) }}</span>
                </div>
                <div>
                  <span class="label">已用</span>
                  <span>{{ formatBytes(metrics?.disk?.used) }}</span>
                </div>
                <div>
                  <span class="label">可用</span>
                  <span>{{ formatBytes(metrics?.disk?.free) }}</span>
                </div>
              </div>
            </div>
          </el-card>

          <el-card class="metric-card update-card stacked">
            <div class="update-info">
              <span>最近更新时间：{{ metrics?.updated_at || '-' }}</span>
              <el-button type="primary" text :loading="loading" @click="refresh">
                <el-icon><Refresh /></el-icon>
                立即刷新
              </el-button>
            </div>
          </el-card>
        </div>
      </el-col>

      <el-col :xs="24" :sm="24" :md="12" :lg="12">
        <div class="stack-column">
          <el-card class="metric-card cpu-card">
            <template #header>
              <div class="card-header">
                <span>CPU 使用率</span>
              </div>
            </template>
            <div class="card-content">
              <el-progress
                type="dashboard"
                :percentage="metrics?.cpu?.usage || 0"
                :width="180"
                :stroke-width="12"
              >
                <template #default="{ percentage }">
                  <span class="progress-value">{{ percentage.toFixed(1) }}%</span>
                </template>
              </el-progress>
              <ul class="metric-list">
                <li><span>核心数：</span>{{ metrics?.cpu?.cores || '-' }}</li>
                <li><span>用户态：</span>{{ formatPercent(metrics?.cpu?.user) }}</li>
                <li><span>内核态：</span>{{ formatPercent(metrics?.cpu?.system) }}</li>
                <li><span>空闲：</span>{{ formatPercent(metrics?.cpu?.idle) }}</li>
              </ul>
            </div>
            <div v-if="metrics?.cpu?.description" class="cpu-description">
              {{ metrics.cpu.description }}
            </div>
          </el-card>

        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import api from '@/utils/api'

const metrics = ref(null)
const loading = ref(false)
let timer = null

const fetchMetrics = async () => {
  loading.value = true
  try {
    const { data } = await api.get('/system/metrics')
    metrics.value = data
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '获取系统监控数据失败')
  } finally {
    loading.value = false
  }
}

const refresh = () => {
  fetchMetrics()
}

const formatPercent = (val) => {
  if (val === null || val === undefined || Number.isNaN(val)) return '-'
  return `${Number(val).toFixed(1)}%`
}

const formatBytes = (num) => {
  if (!num || Number.isNaN(num)) return '-'
  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  let value = Number(num)
  let unitIndex = 0
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex++
  }
  return `${value.toFixed(2)} ${units[unitIndex]}`
}

onMounted(() => {
  fetchMetrics()
  timer = setInterval(fetchMetrics, 5000)
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
})
</script>

<style scoped>
.system-monitor {
  padding: 20px;
  background: #f5f7fa;
  min-height: calc(100vh - 120px);
}

.metric-card {
  margin-bottom: 20px;
  border-radius: 14px;
  overflow: hidden;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 18px;
  font-weight: 600;
  letter-spacing: 1px;
  color: #303133;
}

.card-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.card-content.vertical {
  flex-direction: column;
  align-items: stretch;
}

.progress-value {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.metric-list {
  list-style: none;
  padding: 0;
  margin: 0;
  font-size: 16px;
  color: #606266;
}

.metric-list li {
  margin-bottom: 6px;
}

.metric-list li span {
  display: inline-block;
  min-width: 80px;
  color: #909399;
}

.metric-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 12px;
  width: 100%;
}

.metric-summary .label {
  display: block;
  font-size: 16px;
  color: #909399;
  margin-bottom: 4px;
}

.metric-summary span {
  font-size: 16px;
  color: #303133;
}

.linear-progress {
  width: 100%;
}

.stack-column {
  display: flex;
  flex-direction: column;
  gap:5px;
}

.cpu-card {
  flex: 1;
}

.cpu-description {
  margin-top: 12px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 10px;
  font-size: 14px;
  color: #909399;
}

.update-card {
  border-radius: 14px;
}

.update-card.stacked {
  flex: none;
}

.update-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 16px;
  color: #606266;
}

.update-info .el-button {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 16px;
}

@media (max-width: 768px) {
  .card-content {
    flex-direction: column;
    align-items: center;
  }

  .metric-list {
    width: 100%;
  }
}
</style>


