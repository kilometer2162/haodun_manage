<template>
  <div class="roles-page">
    <div class="page-header">
      <h2>角色管理</h2>
      <el-button type="primary" @click="handleAdd">新增角色</el-button>
    </div>

    <el-table :data="roles" v-loading="loading" style="margin-top: 20px">
      <el-table-column prop="name" label="角色名" width="160" />
      <el-table-column prop="description" label="描述" />
      <el-table-column label="权限数量">
        <template #default="{ row }">
          {{ row.permissions?.length || 0 }}
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatTime(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="250">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" @click="handleAssignPermissions(row)">分配权限</el-button>
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
      @size-change="loadRoles"
      @current-change="loadRoles"
      style="margin-top: 20px"
    />

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="角色名" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="permissionDialogVisible" title="分配权限" width="600px">
      <el-checkbox-group v-model="selectedPermissions">
        <el-checkbox
          v-for="permission in permissions"
          :key="permission.id"
          :label="permission.id"
        >
          {{ permission.name }} ({{ permission.code }})
        </el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="permissionDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handlePermissionSubmit">确定</el-button>
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
const roles = ref([])
const permissions = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const permissionDialogVisible = ref(false)
const dialogTitle = ref('新增角色')
const selectedRoleId = ref(null)
const selectedPermissions = ref([])
const formRef = ref(null)
const form = reactive({
  id: null,
  name: '',
  description: ''
})

const rules = {
  name: [{ required: true, message: '请输入角色名', trigger: 'blur' }]
}

const loadRoles = async () => {
  loading.value = true
  try {
    const response = await api.get(`/roles?page=${page.value}&page_size=${pageSize.value}`)
    roles.value = response.data.data
    total.value = response.data.total
  } catch (error) {
    ElMessage.error('加载角色列表失败')
  } finally {
    loading.value = false
  }
}

const loadPermissions = async () => {
  try {
    const response = await api.get('/permissions?page=1&page_size=1000')
    permissions.value = response.data.data
  } catch (error) {
    console.error('加载权限列表失败', error)
  }
}

const handleAdd = () => {
  dialogTitle.value = '新增角色'
  Object.assign(form, {
    id: null,
    name: '',
    description: ''
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑角色'
  Object.assign(form, {
    id: row.id,
    name: row.name,
    description: row.description
  })
  dialogVisible.value = true
}

const handleAssignPermissions = (row) => {
  selectedRoleId.value = row.id
  selectedPermissions.value = row.permissions?.map(p => p.id) || []
  permissionDialogVisible.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该角色吗？', '提示', {
      type: 'warning'
    })
    await api.delete(`/roles/${row.id}`)
    ElMessage.success('删除成功')
    loadRoles()
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
          await api.put(`/roles/${form.id}`, form)
          ElMessage.success('更新成功')
        } else {
          await api.post('/roles', form)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        loadRoles()
      } catch (error) {
        ElMessage.error(error.response?.data?.error || '操作失败')
      }
    }
  })
}

const handlePermissionSubmit = async () => {
  try {
    await api.post(`/roles/${selectedRoleId.value}/permissions`, {
      permission_ids: selectedPermissions.value
    })
    ElMessage.success('权限分配成功')
    permissionDialogVisible.value = false
    loadRoles()
  } catch (error) {
    ElMessage.error('权限分配失败')
  }
}

onMounted(() => {
  loadRoles()
  loadPermissions()
})
</script>

<style scoped>
.roles-page {
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

