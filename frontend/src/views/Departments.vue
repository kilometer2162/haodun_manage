<template>
  <div class="departments-page">
    <div class="page-header">
      <h2>部门管理</h2>
      <el-button type="primary" @click="handleAdd">新增部门</el-button>
    </div>

    <el-table :data="departments" v-loading="loading" style="margin-top: 20px">
      <el-table-column prop="sort" label="排序" width="100" />
      <el-table-column prop="name" label="部门名称" width="180" />
      <el-table-column label="上级部门" width="180">
        <template #default="{ row }">
          {{ row.parent?.name || '-' }}
        </template>
      </el-table-column>
    
      <el-table-column prop="status" label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="描述" />
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
      @size-change="loadDepartments"
      @current-change="loadDepartments"
      style="margin-top: 20px"
    />

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="520px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        
        <el-form-item label="部门名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="上级部门" prop="parent_id">
          <el-select v-model="form.parent_id" placeholder="请选择上级部门" clearable>
            <el-option label="无" :value="null" />
            <el-option
              v-for="dept in parentOptions"
              :key="dept.id"
              :label="dept.name"
              :value="dept.id"
              :disabled="dept.id === form.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" :max="9999" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" rows="3" />
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
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/utils/api'
import { formatBeijingTime } from '@/utils/date'

const formatTime = formatBeijingTime

const loading = ref(false)
const departments = ref([])
const allDepartments = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const dialogTitle = ref('新增部门')
const formRef = ref(null)
const form = reactive({
  id: null,
  name: '',
  parent_id: null,
  status: 1,
  sort: 0,
  description: ''
})

const rules = {
  name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }]
}

const parentOptions = computed(() => {
  if (!form.id) {
    return allDepartments.value
  }
  return allDepartments.value.filter((dept) => dept.id !== form.id)
})

const loadDepartments = async () => {
  loading.value = true
  try {
    const response = await api.get(`/departments?page=${page.value}&page_size=${pageSize.value}`)
    departments.value = response.data.data || []
    total.value = response.data.total || 0
  } catch (error) {
    ElMessage.error('加载部门列表失败')
  } finally {
    loading.value = false
  }
}

const loadAllDepartments = async () => {
  try {
    const response = await api.get('/departments?status=1&simple=1')
    allDepartments.value = response.data.data || []
  } catch (error) {
    console.error('加载全部部门失败', error)
  }
}

const handleAdd = () => {
  dialogTitle.value = '新增部门'
  Object.assign(form, {
    id: null,
    name: '',
    parent_id: null,
    status: 1,
    sort: 0,
    description: ''
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑部门'
  Object.assign(form, {
    id: row.id,
    name: row.name,
    parent_id: row.parent_id ?? null,
    status: row.status,
    sort: row.sort,
    description: row.description || ''
  })
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该部门吗？', '提示', {
      type: 'warning'
    })
    await api.delete(`/departments/${row.id}`)
    ElMessage.success('删除成功')
    loadDepartments()
    loadAllDepartments()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '删除失败')
    }
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      const payload = {
        name: form.name,
        parent_id: form.parent_id,
        status: form.status,
        sort: form.sort,
        description: form.description
      }
      if (form.id) {
        await api.put(`/departments/${form.id}`, payload)
        ElMessage.success('更新成功')
      } else {
        await api.post('/departments', payload)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      loadDepartments()
      loadAllDepartments()
    } catch (error) {
      ElMessage.error(error.response?.data?.error || '操作失败')
    }
  })
}

onMounted(() => {
  loadDepartments()
  loadAllDepartments()
})
</script>

<style scoped>
.departments-page {
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

