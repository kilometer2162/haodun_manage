<template>
  <div class="resources-page">
    <div class="page-header">
      <h2>资源管理</h2>
      <div>
        <el-radio-group v-model="viewMode" @change="handleViewModeChange" style="margin-right: 10px">
          <el-radio-button value="tree">树形视图</el-radio-button>
          <el-radio-button value="list">列表视图</el-radio-button>
        </el-radio-group>
      <el-button type="primary" @click="handleAdd">新增资源</el-button>
      </div>
    </div>

    <!-- 树形视图 -->
    <el-tree
      v-if="viewMode === 'tree'"
      :data="treeData"
      :props="{ children: 'children', label: 'name' }"
      node-key="id"
      default-expand-all
      v-loading="loading"
      style="margin-top: 20px; background: #fff; padding: 20px"
    >
      <template #default="{ node, data }">
        <span class="tree-node">
          <span class="tree-node-label">
            <span v-if="data.type === 'menu'" style="margin-right: 4px; color: #409eff">📁</span>
            <span v-else-if="data.type === 'button'" style="margin-right: 4px; color: #67c23a">🔘</span>
            <span v-else style="margin-right: 4px; color: #909399">🔗</span>
            <span>{{ data.name }}</span>
            <el-tag size="small" :type="getMethodType(data.method)" style="margin-left: 8px">
              {{ data.method }}
            </el-tag>
            <el-tag size="small" type="info" style="margin-left: 4px">
              {{ data.type }}
            </el-tag>
            <span style="margin-left: 8px; color: #909399; font-size: 12px">{{ data.path }}</span>
          </span>
          <span class="tree-node-actions">
            <el-button size="small" text type="primary" @click="handleAddChild(data)" v-if="data.type === 'menu'">添加子项</el-button>
            <el-button size="small" text @click="handleEdit(data)">编辑</el-button>
            <el-button size="small" text type="danger" @click="handleDelete(data)">删除</el-button>
          </span>
        </span>
      </template>
    </el-tree>

    <!-- 列表视图 -->
    <div v-else>
    <el-table :data="resources" v-loading="loading" style="margin-top: 20px">
      <el-table-column prop="name" label="资源名" />
      <el-table-column prop="path" label="路径" />
      <el-table-column prop="method" label="方法" width="100">
        <template #default="{ row }">
          <el-tag :type="getMethodType(row.method)">
            {{ row.method }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="type" label="类型" width="100" />
        <el-table-column prop="parent_id" label="父资源ID" width="100">
          <template #default="{ row }">
            <span v-if="row.parent_id">{{ row.parent_id }}</span>
            <span v-else style="color: #909399">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" />
      <el-table-column prop="description" label="描述" />
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
      @size-change="loadResources"
      @current-change="loadResources"
      style="margin-top: 20px"
    />
    </div>

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="600px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="资源名" prop="name">
          <el-input v-model="form.name" placeholder="请输入资源名" />
        </el-form-item>
        <el-form-item label="路径" prop="path">
          <el-input v-model="form.path" placeholder="请输入路径，如：/api/users" />
        </el-form-item>
        <el-form-item label="方法" prop="method">
          <el-select v-model="form.method" placeholder="请选择方法" style="width: 100%">
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type" placeholder="请选择类型" style="width: 100%">
            <el-option label="API" value="api" />
            <el-option label="菜单" value="menu" />
            <el-option label="按钮" value="button" />
          </el-select>
        </el-form-item>
        <el-form-item label="父资源" prop="parent_id">
          <el-cascader
            v-model="form.parent_id"
            :options="parentOptions"
            :props="{ value: 'id', label: 'name', children: 'children', checkStrictly: true, emitPath: false }"
            placeholder="请选择父资源（可选）"
            clearable
            style="width: 100%"
            filterable
          />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="图标" prop="icon">
          <el-input v-model="form.icon" placeholder="请输入图标名称（可选）" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入描述" />
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
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/utils/api'

const loading = ref(false)
const viewMode = ref('tree')
const resources = ref([])
const treeData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const dialogTitle = ref('新增资源')
const formRef = ref(null)
const parentResourceId = ref(null) // 用于记录添加子项时的父资源ID

const form = reactive({
  id: null,
  name: '',
  path: '',
  method: '',
  type: '',
  description: '',
  parent_id: null,
  sort: 0,
  icon: ''
})

const rules = {
  name: [{ required: true, message: '请输入资源名', trigger: 'blur' }],
  path: [{ required: true, message: '请输入路径', trigger: 'blur' }],
  method: [{ required: true, message: '请选择方法', trigger: 'change' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }]
}

// 构建父资源选项（排除当前编辑的资源及其子资源）
const parentOptions = computed(() => {
  const buildOptions = (items, excludeId = null) => {
    return items
      .filter(item => item.id !== excludeId)
      .map(item => {
        const option = {
          id: item.id,
          name: item.name,
          children: item.children && item.children.length > 0 ? buildOptions(item.children, excludeId) : []
        }
        return option
      })
  }
  return buildOptions(treeData.value, form.id)
})

const getMethodType = (method) => {
  const types = {
    GET: 'success',
    POST: 'primary',
    PUT: 'warning',
    DELETE: 'danger'
  }
  return types[method] || ''
}

const loadResources = async () => {
  loading.value = true
  try {
    if (viewMode.value === 'tree') {
      const response = await api.get('/resources?tree=true')
      treeData.value = response.data.data || []
    } else {
    const response = await api.get(`/resources?page=${page.value}&page_size=${pageSize.value}`)
      resources.value = response.data.data || []
      total.value = response.data.total || 0
    }
  } catch (error) {
    ElMessage.error('加载资源列表失败')
  } finally {
    loading.value = false
  }
}

const handleViewModeChange = () => {
  loadResources()
}

const handleAdd = () => {
  dialogTitle.value = '新增资源'
  parentResourceId.value = null
  Object.assign(form, {
    id: null,
    name: '',
    path: '',
    method: '',
    type: '',
    description: '',
    parent_id: null,
    sort: 0,
    icon: ''
  })
  dialogVisible.value = true
}

const handleAddChild = (parent) => {
  dialogTitle.value = '新增子资源'
  parentResourceId.value = parent.id
  Object.assign(form, {
    id: null,
    name: '',
    path: '',
    method: '',
    type: '',
    description: '',
    parent_id: parent.id,
    sort: 0,
    icon: ''
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑资源'
  parentResourceId.value = null
  Object.assign(form, {
    id: row.id,
    name: row.name,
    path: row.path,
    method: row.method,
    type: row.type,
    description: row.description || '',
    parent_id: row.parent_id || null,
    sort: row.sort || 0,
    icon: row.icon || ''
  })
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该资源吗？删除后其子资源也会被删除。', '提示', {
      type: 'warning'
    })
    await api.delete(`/resources/${row.id}`)
    ElMessage.success('删除成功')
    loadResources()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '删除失败')
    }
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        const submitData = {
          ...form,
          parent_id: form.parent_id || null
        }
        
        if (form.id) {
          await api.put(`/resources/${form.id}`, submitData)
          ElMessage.success('更新成功')
        } else {
          await api.post('/resources', submitData)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        loadResources()
      } catch (error) {
        ElMessage.error(error.response?.data?.error || '操作失败')
      }
    }
  })
}

onMounted(() => {
  loadResources()
})
</script>

<style scoped>
.resources-page {
  background: #fff;
  padding: 20px;
  border-radius: 4px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.tree-node {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 16px;
  padding-right: 8px;
}

.tree-node-label {
  display: flex;
  align-items: center;
}

.tree-node-actions {
  display: flex;
  gap: 8px;
}
</style>
