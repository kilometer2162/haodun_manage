<template>
  <div class="configs-page">
    <div class="page-header">
      <h2>系统参数管理</h2>
      <div>
        <el-select
          v-model="selectedGroup"
          placeholder="选择分组"
          clearable
          @change="loadConfigs"
          style="width: 200px; margin-right: 10px"
        >
          <el-option
            v-for="group in groups"
            :key="group"
            :label="group"
            :value="group"
          />
        </el-select>
        <el-button type="primary" @click="handleAdd">新增参数</el-button>
      </div>
    </div>

    <el-table :data="configs" v-loading="loading" style="margin-top: 20px">
      <el-table-column prop="key" label="参数键" width="200" />
      <el-table-column prop="label" label="参数标签" />
      <el-table-column prop="value" label="参数值" show-overflow-tooltip />
      <el-table-column prop="type" label="类型" width="100">
        <template #default="{ row }">
          <el-tag size="small">{{ row.type }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="group" label="分组" width="120" />
      <el-table-column prop="sort" label="排序" width="80" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="描述" show-overflow-tooltip />
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
      @size-change="loadConfigs"
      @current-change="loadConfigs"
      style="margin-top: 20px"
    />

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="600px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="参数键" prop="key">
          <el-input v-model="form.key" placeholder="如：site_name" />
        </el-form-item>
        <el-form-item label="参数标签" prop="label">
          <el-input v-model="form.label" placeholder="如：网站名称" />
        </el-form-item>
        <el-form-item label="参数值" prop="value">
          <el-input
            v-if="form.type === 'text'"
            v-model="form.value"
            placeholder="请输入参数值"
          />
          <el-input-number
            v-else-if="form.type === 'number'"
            v-model="form.value"
            style="width: 100%"
          />
          <el-switch
            v-else-if="form.type === 'boolean'"
            v-model="form.value"
            active-text="是"
            inactive-text="否"
          />
          <el-input
            v-else-if="form.type === 'textarea'"
            v-model="form.value"
            type="textarea"
            :rows="4"
            placeholder="请输入参数值"
          />
          <el-input
            v-else
            v-model="form.value"
            placeholder="请输入参数值"
          />
        </el-form-item>
        <el-form-item label="参数类型" prop="type">
          <el-select v-model="form.type" placeholder="请选择类型">
            <el-option label="文本" value="text" />
            <el-option label="数字" value="number" />
            <el-option label="布尔值" value="boolean" />
            <el-option label="多行文本" value="textarea" />
            <el-option label="选择" value="select" />
          </el-select>
        </el-form-item>
        <el-form-item label="分组" prop="group">
          <el-input v-model="form.group" placeholder="如：system, email, upload" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
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
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/utils/api'
import { formatBeijingTime } from '@/utils/date'

const formatTime = formatBeijingTime

const loading = ref(false)
const configs = ref([])
const groups = ref([])
const selectedGroup = ref('')
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const dialogTitle = ref('新增参数')
const formRef = ref(null)

const form = reactive({
  id: null,
  key: '',
  label: '',
  value: '',
  type: 'text',
  group: 'system',
  sort: 0,
  status: 1,
  description: ''
})

const rules = {
  key: [{ required: true, message: '请输入参数键', trigger: 'blur' }],
  label: [{ required: true, message: '请输入参数标签', trigger: 'blur' }]
}

const loadConfigs = async () => {
  loading.value = true
  try {
    let url = `/configs?page=${page.value}&page_size=${pageSize.value}`
    if (selectedGroup.value) {
      url += `&group=${encodeURIComponent(selectedGroup.value)}`
    }
    const response = await api.get(url)
    // 确保数据正确赋值
    if (response.data && response.data.data) {
      configs.value = response.data.data
      total.value = response.data.total || 0
    } else {
      // 如果响应结构不对，尝试直接使用response.data
      configs.value = Array.isArray(response.data) ? response.data : []
      total.value = configs.value.length
    }
  } catch (error) {
    console.error('加载参数列表失败:', error)
    ElMessage.error(error.response?.data?.error || '加载参数列表失败')
    configs.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const loadGroups = async () => {
  try {
    const response = await api.get('/configs/groups')
    groups.value = response.data.groups
  } catch (error) {
    console.error('加载分组列表失败', error)
  }
}

const handleAdd = () => {
  dialogTitle.value = '新增参数'
  Object.assign(form, {
    id: null,
    key: '',
    label: '',
    value: '',
    type: 'text',
    group: 'system',
    sort: 0,
    status: 1,
    description: ''
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑参数'
  // 处理布尔值类型
  let value = row.value
  if (row.type === 'boolean') {
    value = row.value === 'true' || row.value === '1' || row.value === true
  } else if (row.type === 'number') {
    value = Number(row.value) || 0
  }
  
  Object.assign(form, {
    id: row.id,
    key: row.key,
    label: row.label,
    value: value,
    type: row.type,
    group: row.group,
    sort: row.sort,
    status: row.status,
    description: row.description
  })
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该参数吗？', '提示', {
      type: 'warning'
    })
    await api.delete(`/configs/${row.id}`)
    ElMessage.success('删除成功')
    loadConfigs()
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
        // 处理不同类型的值
        let value = form.value
        if (form.type === 'boolean') {
          value = value ? '1' : '0'
        } else if (form.type === 'number') {
          value = String(value)
        } else {
          value = String(value || '')
        }

        const submitData = {
          ...form,
          value: value
        }

        if (form.id) {
          await api.put(`/configs/${form.id}`, submitData)
          ElMessage.success('更新成功')
        } else {
          await api.post('/configs', submitData)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        loadConfigs()
      } catch (error) {
        ElMessage.error(error.response?.data?.error || '操作失败')
      }
    }
  })
}

onMounted(() => {
  loadConfigs()
  loadGroups()
})
</script>

<style scoped>
.configs-page {
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

