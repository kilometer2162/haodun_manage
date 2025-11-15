<template>
  <div class="dicts-page">
    <div class="page-header">
      <h2>系统字典管理</h2>
      <el-button type="primary" @click="handleAddType">新增字典类型</el-button>
    </div>

    <el-tabs v-model="activeTab" @tab-change="handleTabChange">
      <el-tab-pane label="字典类型" name="types">
        <el-table :data="dictTypes" v-loading="typesLoading" style="margin-top: 20px">
          <el-table-column prop="code" label="代码" width="150" />
          <el-table-column prop="name" label="名称" />
          <el-table-column prop="description" label="描述" />
          <el-table-column label="字典项数量" width="120">
            <template #default="{ row }">
              {{ row.items?.length || 0 }}
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'danger'">
                {{ row.status === 1 ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="180">
            <template #default="{ row }">
              {{ formatTime(row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="250">
            <template #default="{ row }">
              <el-button size="small" @click="handleEditType(row)">编辑</el-button>
              <el-button size="small" @click="handleManageItems(row)">管理字典项</el-button>
              <el-button size="small" type="danger" @click="handleDeleteType(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-pagination
          v-model:current-page="typesPage"
          v-model:page-size="typesPageSize"
          :total="typesTotal"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadDictTypes"
          @current-change="loadDictTypes"
          style="margin-top: 20px"
        />
      </el-tab-pane>

      <el-tab-pane label="字典项" name="items">
        <div style="margin: 20px 0">
          <el-select
            v-model="selectedTypeId"
            placeholder="请选择字典类型"
            clearable
            @change="loadDictItems"
            style="width: 300px"
          >
            <el-option
              v-for="type in dictTypes"
              :key="type.id"
              :label="type.name"
              :value="type.id"
            />
          </el-select>
          <el-button
            type="primary"
            @click="handleAddItem"
            :disabled="!selectedTypeId"
            style="margin-left: 10px"
          >
            新增字典项
          </el-button>
        </div>

        <el-table :data="dictItems" v-loading="itemsLoading" style="margin-top: 20px">
          <el-table-column prop="sort" label="排序" width="80" />
          <el-table-column prop="type.name" label="字典类型" width="150" />
          <el-table-column prop="label" label="标签" />
          <el-table-column prop="value" label="值" width="120" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'danger'">
                {{ row.status === 1 ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="description" label="描述" />
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <el-button size="small" @click="handleEditItem(row)">编辑</el-button>
              <el-button size="small" type="danger" @click="handleDeleteItem(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-pagination
          v-model:current-page="itemsPage"
          v-model:page-size="itemsPageSize"
          :total="itemsTotal"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadDictItems"
          @current-change="loadDictItems"
          style="margin-top: 20px"
        />
      </el-tab-pane>
    </el-tabs>

    <!-- 字典类型对话框 -->
    <el-dialog v-model="typeDialogVisible" :title="typeDialogTitle" width="500px">
      <el-form :model="typeForm" :rules="typeRules" ref="typeFormRef" label-width="100px">
        <el-form-item label="代码" prop="code">
          <el-input v-model="typeForm.code" placeholder="如：user_status" />
        </el-form-item>
        <el-form-item label="名称" prop="name">
          <el-input v-model="typeForm.name" placeholder="如：用户状态" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="typeForm.description" type="textarea" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="typeForm.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="typeDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleTypeSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 字典项对话框 -->
    <el-dialog v-model="itemDialogVisible" :title="itemDialogTitle" width="500px">
      <el-form :model="itemForm" :rules="itemRules" ref="itemFormRef" label-width="100px">
        <el-form-item label="字典类型" prop="type_id">
          <el-select v-model="itemForm.type_id" placeholder="请选择字典类型" :disabled="!!itemForm.id">
            <el-option
              v-for="type in dictTypes"
              :key="type.id"
              :label="type.name"
              :value="type.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="标签" prop="label">
          <el-input v-model="itemForm.label" placeholder="如：正常" />
        </el-form-item>
        <el-form-item label="值" prop="value">
          <el-input v-model="itemForm.value" placeholder="如：1" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="itemForm.sort" :min="0" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="itemForm.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="itemForm.description" type="textarea" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="itemDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleItemSubmit">确定</el-button>
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

const activeTab = ref('types')
const typesLoading = ref(false)
const itemsLoading = ref(false)
const dictTypes = ref([])
const dictItems = ref([])
const selectedTypeId = ref(null)

// 分页
const typesPage = ref(1)
const typesPageSize = ref(10)
const typesTotal = ref(0)
const itemsPage = ref(1)
const itemsPageSize = ref(10)
const itemsTotal = ref(0)

// 对话框
const typeDialogVisible = ref(false)
const itemDialogVisible = ref(false)
const typeDialogTitle = ref('新增字典类型')
const itemDialogTitle = ref('新增字典项')
const typeFormRef = ref(null)
const itemFormRef = ref(null)

const typeForm = reactive({
  id: null,
  code: '',
  name: '',
  description: '',
  status: 1
})

const itemForm = reactive({
  id: null,
  type_id: null,
  label: '',
  value: '',
  sort: 0,
  status: 1,
  description: ''
})

const typeRules = {
  code: [{ required: true, message: '请输入代码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }]
}

const itemRules = {
  type_id: [{ required: true, message: '请选择字典类型', trigger: 'change' }],
  label: [{ required: true, message: '请输入标签', trigger: 'blur' }],
  value: [{ required: true, message: '请输入值', trigger: 'blur' }]
}

const loadDictTypes = async () => {
  typesLoading.value = true
  try {
    const response = await api.get(`/dict-types?page=${typesPage.value}&page_size=${typesPageSize.value}`)
    dictTypes.value = response.data.data
    typesTotal.value = response.data.total
  } catch (error) {
    ElMessage.error('加载字典类型列表失败')
  } finally {
    typesLoading.value = false
  }
}

const loadDictItems = async () => {
  itemsLoading.value = true
  try {
    let url = `/dict-items?page=${itemsPage.value}&page_size=${itemsPageSize.value}`
    if (selectedTypeId.value) {
      url += `&type_id=${selectedTypeId.value}`
    }
    const response = await api.get(url)
    dictItems.value = response.data.data
    itemsTotal.value = response.data.total
  } catch (error) {
    ElMessage.error('加载字典项列表失败')
  } finally {
    itemsLoading.value = false
  }
}

const handleTabChange = (tab) => {
  if (tab === 'items') {
    loadDictItems()
  }
}

const handleAddType = () => {
  typeDialogTitle.value = '新增字典类型'
  Object.assign(typeForm, {
    id: null,
    code: '',
    name: '',
    description: '',
    status: 1
  })
  typeDialogVisible.value = true
}

const handleEditType = (row) => {
  typeDialogTitle.value = '编辑字典类型'
  Object.assign(typeForm, {
    id: row.id,
    code: row.code,
    name: row.name,
    description: row.description,
    status: row.status
  })
  typeDialogVisible.value = true
}

const handleDeleteType = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该字典类型吗？', '提示', {
      type: 'warning'
    })
    await api.delete(`/dict-types/${row.id}`)
    ElMessage.success('删除成功')
    loadDictTypes()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '删除失败')
    }
  }
}

const handleManageItems = (row) => {
  activeTab.value = 'items'
  selectedTypeId.value = row.id
  loadDictItems()
}

const handleAddItem = () => {
  itemDialogTitle.value = '新增字典项'
  Object.assign(itemForm, {
    id: null,
    type_id: selectedTypeId.value,
    label: '',
    value: '',
    sort: 0,
    status: 1,
    description: ''
  })
  itemDialogVisible.value = true
}

const handleEditItem = (row) => {
  itemDialogTitle.value = '编辑字典项'
  Object.assign(itemForm, {
    id: row.id,
    type_id: row.type_id,
    label: row.label,
    value: row.value,
    sort: row.sort,
    status: row.status,
    description: row.description
  })
  itemDialogVisible.value = true
}

const handleDeleteItem = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该字典项吗？', '提示', {
      type: 'warning'
    })
    await api.delete(`/dict-items/${row.id}`)
    ElMessage.success('删除成功')
    loadDictItems()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleTypeSubmit = async () => {
  if (!typeFormRef.value) return
  
  await typeFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        if (typeForm.id) {
          await api.put(`/dict-types/${typeForm.id}`, typeForm)
          ElMessage.success('更新成功')
        } else {
          await api.post('/dict-types', typeForm)
          ElMessage.success('创建成功')
        }
        typeDialogVisible.value = false
        loadDictTypes()
      } catch (error) {
        ElMessage.error(error.response?.data?.error || '操作失败')
      }
    }
  })
}

const handleItemSubmit = async () => {
  if (!itemFormRef.value) return
  
  await itemFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        if (itemForm.id) {
          await api.put(`/dict-items/${itemForm.id}`, itemForm)
          ElMessage.success('更新成功')
        } else {
          await api.post('/dict-items', itemForm)
          ElMessage.success('创建成功')
        }
        itemDialogVisible.value = false
        loadDictItems()
      } catch (error) {
        ElMessage.error(error.response?.data?.error || '操作失败')
      }
    }
  })
}

onMounted(() => {
  loadDictTypes()
})
</script>

<style scoped>
.dicts-page {
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

