<template>
  <div class="notices-page">
    <div class="page-header">
      <h2>消息公告</h2>
      <el-button type="primary" @click="handleAdd">新增公告</el-button>
    </div>

    <!-- 搜索栏 -->
    <el-card style="margin-top: 20px">
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="标题">
          <el-input
            v-model="searchForm.title"
            placeholder="请输入标题"
            clearable
            style="width: 200px"
            @keyup.enter="loadNotices"
          />
        </el-form-item>
        <el-form-item label="阅读状态">
          <el-select
            v-model="searchForm.is_read"
            placeholder="请选择"
            clearable
            style="width: 150px"
          >
            <el-option label="未读" value="0" />
            <el-option label="已读" value="1" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadNotices">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 消息列表 -->
    <el-table :data="notices" v-loading="loading" style="margin-top: 20px">
      <el-table-column prop="title" label="标题" min-width="200" />
      <el-table-column label="优先级" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.priority === 2" type="danger">紧急</el-tag>
          <el-tag v-else-if="row.priority === 1" type="warning">重要</el-tag>
          <el-tag v-else type="info">普通</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="阅读状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.is_read === 1 ? 'success' : 'danger'">
            {{ row.is_read === 1 ? '已读' : '未读' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="发布时间" width="180">
        <template #default="{ row }">
          {{ formatTime(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" @click="handleView(row)">查看</el-button>
          <el-button
            v-if="row.is_read !== 1"
            size="small"
            type="success"
            @click="handleMarkRead(row)"
          >
            标记已读
          </el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="page"
      v-model:page-size="pageSize"
      :total="total"
      :page-sizes="[10, 20, 50, 100]"
      layout="total, sizes, prev, pager, next, jumper"
      @size-change="loadNotices"
      @current-change="loadNotices"
      style="margin-top: 20px"
    />

    <!-- 查看详情对话框 -->
    <el-dialog v-model="viewDialogVisible" title="消息详情" width="800px">
      <div v-if="currentNotice">
        <h3>{{ currentNotice.title }}</h3>
        <div style="margin: 10px 0; color: #909399; font-size: 12px">
          发布时间：{{ formatTime(currentNotice.created_at) }}
        </div>
        <el-divider />
        <div class="notice-content" v-html="currentNotice.content"></div>
      </div>
    </el-dialog>

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="600px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入标题" />
        </el-form-item>
        <el-form-item label="内容" prop="content">
          <el-input v-model="form.content" type="textarea" :rows="6" placeholder="请输入内容" />
        </el-form-item>
        <el-form-item label="优先级" prop="priority">
          <el-radio-group v-model="form.priority">
            <el-radio :value="0">普通</el-radio>
            <el-radio :value="1">重要</el-radio>
            <el-radio :value="2">紧急</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">发布</el-radio>
            <el-radio :value="0">草稿</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/utils/api'
import { formatBeijingTime } from '@/utils/date'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const notices = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const viewDialogVisible = ref(false)
const dialogTitle = ref('新增公告')
const currentNotice = ref(null)
const formRef = ref(null)

const searchForm = reactive({
  title: '',
  is_read: ''
})

const form = reactive({
  id: null,
  title: '',
  content: '',
  priority: 0,
  status: 1
})

const rules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入内容', trigger: 'blur' }]
}

const formatTime = formatBeijingTime

const loadNotices = async () => {
  loading.value = true
  try {
    let url = `/notices?page=${page.value}&page_size=${pageSize.value}`
    if (searchForm.title) {
      url += `&title=${searchForm.title}`
    }
    if (searchForm.is_read !== '') {
      url += `&is_read=${searchForm.is_read}`
    }
    const response = await api.get(url)
    notices.value = response.data.data || []
    total.value = response.data.total
  } catch (error) {
    ElMessage.error('加载消息列表失败')
  } finally {
    loading.value = false
  }
}

const resetSearch = () => {
  searchForm.title = ''
  searchForm.is_read = ''
  loadNotices()
}

const handleAdd = () => {
  dialogTitle.value = '新增公告'
  Object.assign(form, {
    id: null,
    title: '',
    content: '',
    priority: 0,
    status: 1
  })
  dialogVisible.value = true
}

const handleView = async (row) => {
  try {
    const response = await api.get(`/notices/${row.id}`)
    currentNotice.value = response.data
    viewDialogVisible.value = true
    // 自动标记为已读
    if (row.is_read !== 1) {
      await handleMarkRead(row)
    }
  } catch (error) {
    ElMessage.error('加载消息详情失败')
  }
}

const handleMarkRead = async (row) => {
  try {
    await api.put(`/notices/${row.id}/read`)
    row.is_read = 1
    ElMessage.success('标记已读成功')
    loadNotices()
  } catch (error) {
    ElMessage.error('标记已读失败')
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该消息吗？', '提示', {
      type: 'warning'
    })
    await api.delete(`/notices/${row.id}`)
    ElMessage.success('删除成功')
    loadNotices()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        if (form.id) {
          await api.put(`/notices/${form.id}`, form)
          ElMessage.success('更新成功')
        } else {
          await api.post('/notices', form)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        loadNotices()
      } catch (error) {
        ElMessage.error(error.response?.data?.error || '操作失败')
      }
    }
  })
}

onMounted(() => {
  loadNotices()
  // 如果有ID参数，打开详情
  if (route.query.id) {
    handleView({ id: route.query.id })
  }
})
</script>

<style scoped>
.notices-page {
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
  margin: 0;
}

.notice-content {
  line-height: 1.8;
  color: #303133;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>

