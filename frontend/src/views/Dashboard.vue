<template>
  <div class="dashboard">
    <h2>仪表盘</h2>
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-value">{{ stats.users }}</div>
            <div class="stat-label">用户总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-value">{{ stats.roles }}</div>
            <div class="stat-label">角色总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-value">{{ stats.logs }}</div>
            <div class="stat-label">日志总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-value">{{ stats.ipAccesses }}</div>
            <div class="stat-label">IP访问记录</div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '@/utils/api'

const stats = ref({
  users: 0,
  roles: 0,
  logs: 0,
  ipAccesses: 0
})

const loadStats = async () => {
  try {
    const [usersRes, rolesRes, logsRes, ipRes] = await Promise.all([
      api.get('/users?page=1&page_size=1'),
      api.get('/roles?page=1&page_size=1'),
      api.get('/logs?page=1&page_size=1'),
      api.get('/ip-accesses?page=1&page_size=1')
    ])
    
    stats.value = {
      users: usersRes.data.total,
      roles: rolesRes.data.total,
      logs: logsRes.data.total,
      ipAccesses: ipRes.data.total
    }
  } catch (error) {
    console.error('加载统计数据失败', error)
  }
}

onMounted(() => {
  loadStats()
})
</script>

<style scoped>
.dashboard {
  background: #fff;
  padding: 20px;
  border-radius: 4px;
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 10px;
}

.stat-label {
  font-size: 16px;
  color: #909399;
}
</style>

