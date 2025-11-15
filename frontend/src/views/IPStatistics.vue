<template>
  <div class="ip-statistics-page">
    <div class="page-header">
      <h2>IP访问统计</h2>
    </div>

    <el-tabs v-model="activeTab" style="margin-top: 20px">
      <el-tab-pane label="访问记录" name="accesses">
        <el-card>
          <el-form :inline="true" :model="searchForm" class="search-form">
            <el-form-item label="IP地址">
              <el-input v-model="searchForm.ip" placeholder="请输入IP地址" clearable />
            </el-form-item>
            <el-form-item label="国家">
              <el-input v-model="searchForm.country" placeholder="请输入国家" clearable />
            </el-form-item>
            <el-form-item label="省份">
              <el-input v-model="searchForm.province" placeholder="请输入省份" clearable />
            </el-form-item>
            <el-form-item label="城市">
              <el-input v-model="searchForm.city" placeholder="请输入城市" clearable />
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

        <el-table :data="ipAccesses" v-loading="loading" style="margin-top: 20px">
          <el-table-column prop="ip" label="IP地址" width="150" />
          <el-table-column prop="date" label="日期" width="120">
            <template #default="{ row }">
              {{ row.date ? formatDate(row.date) : '' }}
            </template>
          </el-table-column>
          <el-table-column prop="country" label="国家" width="100">
            <template #default="{ row }">
              {{ row.country || '未知' }}
            </template>
          </el-table-column>
          <el-table-column prop="province" label="省份" width="120">
            <template #default="{ row }">
              {{ row.province || '未知' }}
            </template>
          </el-table-column>
          <el-table-column prop="city" label="城市" width="120">
            <template #default="{ row }">
              {{ row.city || '未知' }}
            </template>
          </el-table-column>
          <el-table-column prop="isp" label="运营商" width="120" />
          <el-table-column prop="access_count" label="访问次数" width="120">
            <template #default="{ row }">
              <el-tag type="primary">{{ row.access_count }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="180">
            <template #default="{ row }">
              {{ row.created_at ? formatTime(row.created_at) : '' }}
            </template>
          </el-table-column>
        </el-table>

        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadIPAccesses"
          @current-change="loadIPAccesses"
          style="margin-top: 20px"
        />
      </el-tab-pane>

      <el-tab-pane label="统计图表" name="statistics">
        <el-row :gutter="20" style="margin-top: 20px">
          <el-col :span="12">
            <el-card>
              <template #header>
                <span>按日期统计（最近30天）</span>
              </template>
              <el-table :data="dateStats" style="width: 100%">
                <el-table-column prop="date" label="日期">
                  <template #default="{ row }">
                    {{ row.date ? formatDate(row.date) : '' }}
                  </template>
                </el-table-column>
                <el-table-column prop="access_count" label="访问次数" />
                <el-table-column prop="ip_count" label="IP数量" />
              </el-table>
            </el-card>
          </el-col>
          <el-col :span="12">
            <el-card>
              <template #header>
                <span>按地区统计（Top 50）</span>
              </template>
              <el-table :data="regionStats" style="width: 100%">
                <el-table-column prop="country" label="国家" width="100">
                  <template #default="{ row }">
                    {{ row.country || '未知' }}
                  </template>
                </el-table-column>
                <el-table-column prop="province" label="省份" width="120">
                  <template #default="{ row }">
                    {{ row.province || '未知' }}
                  </template>
                </el-table-column>
                <el-table-column prop="city" label="城市" width="120">
                  <template #default="{ row }">
                    {{ row.city || '未知' }}
                  </template>
                </el-table-column>
                <el-table-column prop="access_count" label="访问次数" />
                <el-table-column prop="ip_count" label="IP数量" />
              </el-table>
            </el-card>
          </el-col>
        </el-row>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'
import { formatBeijingTime } from '@/utils/date'

const formatTime = (timeStr) => {
  if (!timeStr) return ''
  return formatBeijingTime(timeStr)
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  return formatBeijingTime(dateStr, 'YYYY-MM-DD')
}

const activeTab = ref('accesses')
const loading = ref(false)
const ipAccesses = ref([])
const dateStats = ref([])
const regionStats = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const searchForm = reactive({
  ip: '',
  country: '',
  province: '',
  city: '',
  start_date: '',
  end_date: ''
})

const loadIPAccesses = async () => {
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
    
    const response = await api.get('/ip-accesses', { params })
    ipAccesses.value = response.data.data
    total.value = response.data.total
  } catch (error) {
    ElMessage.error('加载IP访问记录失败')
  } finally {
    loading.value = false
  }
}

const loadStatistics = async () => {
  try {
    const response = await api.get('/ip-statistics')
    dateStats.value = response.data.date_stats
    regionStats.value = response.data.region_stats
  } catch (error) {
    ElMessage.error('加载统计信息失败')
  }
}

const handleSearch = () => {
  page.value = 1
  loadIPAccesses()
}

const handleReset = () => {
  Object.assign(searchForm, {
    ip: '',
    country: '',
    province: '',
    city: '',
    start_date: '',
    end_date: ''
  })
  handleSearch()
}

watch(activeTab, (newVal) => {
  if (newVal === 'statistics') {
    loadStatistics()
  }
})

onMounted(() => {
  loadIPAccesses()
})
</script>

<style scoped>
.ip-statistics-page {
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

