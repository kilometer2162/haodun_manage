<template>
  <div class="permissions-page">
    <div class="page-header">
      <h2>权限管理</h2>
      <el-button type="primary" @click="handleAdd">新增权限</el-button>
    </div>

    <el-table :data="permissions" v-loading="loading" style="margin-top: 20px">
      <el-table-column prop="name" label="权限名" width="160" />
      <el-table-column prop="code" label="权限代码" />
      <el-table-column prop="description" label="描述" />
      <el-table-column prop="resource.name" label="资源" />
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatTime(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
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
      @size-change="loadPermissions"
      @current-change="loadPermissions"
      style="margin-top: 20px"
    />

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="权限名" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="权限代码" prop="code">
          <el-input v-model="form.code" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" />
        </el-form-item>
        <el-form-item label="资源" prop="resource_id">
          <el-select v-model="form.resource_id" placeholder="请选择资源">
            <el-option
              v-for="resource in resources"
              :key="resource.id"
              :label="resource.name"
              :value="resource.id"
            />
          </el-select>
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
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/utils/api'
import { formatBeijingTime } from '@/utils/date'

const formatTime = formatBeijingTime

const loading = ref(false)
const permissions = ref([])
const resources = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const dialogTitle = ref('新增权限')
const formRef = ref(null)
const form = reactive({
  id: null,
  name: '',
  code: '',
  description: '',
  resource_id: null
})

const rules = {
  name: [{ required: true, message: '请输入权限名', trigger: 'blur' }],
  code: [{ required: true, message: '请输入权限代码', trigger: 'blur' }],
  resource_id: [{ required: true, message: '请选择资源', trigger: 'change' }]
}

const loadPermissions = async () => {
  loading.value = true
  try {
    const response = await api.get(`/permissions?page=${page.value}&page_size=${pageSize.value}`)
    permissions.value = response.data.data
    total.value = response.data.total
  } catch (error) {
    ElMessage.error('加载权限列表失败')
  } finally {
    loading.value = false
  }
}

const loadResources = async () => {
  try {
    const response = await api.get('/resources?page=1&page_size=1000')
    resources.value = response.data.data
  } catch (error) {
    console.error('加载资源列表失败', error)
  }
}

const handleAdd = () => {
  dialogTitle.value = '新增权限'
  Object.assign(form, {
    id: null,
    name: '',
    code: '',
    description: '',
    resource_id: null
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑权限'
  Object.assign(form, {
    id: row.id,
    name: row.name,
    code: row.code,
    description: row.description,
    resource_id: row.resource_id
  })
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该权限吗？', '提示', {
      type: 'warning'
    })
    await api.delete(`/permissions/${row.id}`)
    ElMessage.success('删除成功')
    loadPermissions()
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
          await api.put(`/permissions/${form.id}`, form)
          ElMessage.success('更新成功')
        } else {
          await api.post('/permissions', form)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        loadPermissions()
      } catch (error) {
        ElMessage.error(error.response?.data?.error || '操作失败')
      }
    }
  })
}

onMounted(() => {
  loadPermissions()
  loadResources()
})
</script>

<style scoped>
.permissions-page {
  background: #fff;
  padding: 20px;
  border-radius: 4px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

