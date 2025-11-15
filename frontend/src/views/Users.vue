<template>
  <div class="users-page">
    <div class="page-header">
      <h2>用户管理</h2>
      <el-button type="primary" @click="handleAdd">新增用户</el-button>
    </div>

    <el-table :data="users" v-loading="loading" style="margin-top: 20px">
      <el-table-column prop="username" label="用户名" width="160" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column label="所属部门" width="180">
        <template #default="{ row }">
          {{ row.department?.name || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="员工类型" width="140">
        <template #default="{ row }">
          <el-tag :type="row.employee_type === 'internal' ? 'success' : 'warning'">
            {{ employeeTypeText(row.employee_type) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="role.name" label="角色" />
      <el-table-column prop="status" label="状态">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? '正常' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">
          {{ row.created_at ? formatTime(row.created_at) : '' }}
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
      @size-change="loadUsers"
      @current-change="loadUsers"
      style="margin-top: 20px"
    />

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="form.password" type="password" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="角色" prop="role_id">
          <el-select v-model="form.role_id" placeholder="请选择角色">
            <el-option
              v-for="role in roles"
              :key="role.id"
              :label="role.name"
              :value="role.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="所属部门" prop="department_id">
          <el-select v-model="form.department_id" placeholder="请选择部门">
            <el-option
              v-for="dept in departments"
              :key="dept.id"
              :label="dept.name"
              :value="dept.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="员工类型" prop="employee_type">
          <el-radio-group v-model="form.employee_type">
            <el-radio value="internal">内部员工</el-radio>
            <el-radio value="external">外部员工</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">正常</el-radio>
            <el-radio :value="0">禁用</el-radio>
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
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/utils/api'
import { formatBeijingTime } from '@/utils/date'

const formatTime = (timeStr) => {
  if (!timeStr) return ''
  return formatBeijingTime(timeStr)
}

const loading = ref(false)
const users = ref([])
const roles = ref([])
const departments = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const dialogTitle = ref('新增用户')
const formRef = ref(null)
const form = reactive({
  id: null,
  username: '',
  password: '',
  email: '',
  role_id: null,
  department_id: null,
  employee_type: 'internal',
  status: 1
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  email: [{ type: 'email', message: '请输入正确的邮箱', trigger: 'blur' }],
  role_id: [{ required: true, message: '请选择角色', trigger: 'change' }],
  department_id: [{ required: true, message: '请选择所属部门', trigger: 'change' }],
  employee_type: [{ required: true, message: '请选择员工类型', trigger: 'change' }]
}

const loadUsers = async () => {
  loading.value = true
  try {
    const response = await api.get(`/users?page=${page.value}&page_size=${pageSize.value}`)
    users.value = response.data.data
    total.value = response.data.total
  } catch (error) {
    ElMessage.error('加载用户列表失败')
  } finally {
    loading.value = false
  }
}

const loadRoles = async () => {
  try {
    const response = await api.get('/roles?page=1&page_size=100')
    roles.value = response.data.data
  } catch (error) {
    console.error('加载角色列表失败', error)
  }
}

const loadDepartments = async () => {
  try {
    const response = await api.get('/departments?page=1&page_size=200&status=1')
    departments.value = response.data.data || []
  } catch (error) {
    console.error('加载部门列表失败', error)
  }
}

const employeeTypeText = (value) => {
  if (value === 'external') return '外部员工'
  return '内部员工'
}

const handleAdd = () => {
  dialogTitle.value = '新增用户'
  Object.assign(form, {
    id: null,
    username: '',
    password: '',
    email: '',
    role_id: null,
    department_id: null,
    employee_type: 'internal',
    status: 1
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑用户'
  Object.assign(form, {
    id: row.id,
    username: row.username,
    password: '',
    email: row.email,
    role_id: row.role_id,
    department_id: row.department_id,
    employee_type: row.employee_type || 'internal',
    status: row.status
  })
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该用户吗？', '提示', {
      type: 'warning'
    })
    await api.delete(`/users/${row.id}`)
    ElMessage.success('删除成功')
    loadUsers()
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
          await api.put(`/users/${form.id}`, form)
          ElMessage.success('更新成功')
        } else {
          await api.post('/users', form)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        loadUsers()
      } catch (error) {
        ElMessage.error(error.response?.data?.error || '操作失败')
      }
    }
  })
}

onMounted(() => {
  loadUsers()
  loadRoles()
  loadDepartments()
})
</script>

<style scoped>
.users-page {
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

