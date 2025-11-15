<template>
  <div class="materials-page">
    <el-row :gutter="16">
      <el-col :span="4" class="folder-column">
        <el-card shadow="never" class="folder-card" v-loading="folderLoading">
          <div class="folder-header">
            <span class="folder-title">素材文件夹</span>
            <div class="folder-header-actions">
              <el-button text size="small" @click="resetFolderFilter">全部素材</el-button>
              <el-button size="small" type="primary" @click="openFolderDialog('create')">
                <el-icon><Plus /></el-icon>
                新建
              </el-button>
            </div>
          </div>
          <div class="folder-actions">
            <el-button size="small" :disabled="!selectedFolder" @click="openFolderDialog('rename')">重命名</el-button>
            <el-button size="small" type="danger" plain :disabled="!selectedFolder" @click="confirmDeleteFolder">删除</el-button>
          </div>
          <el-tree
            ref="folderTreeRef"
            class="folder-tree"
            :data="folderTree"
            :props="treeProps"
            node-key="id"
            highlight-current
            :default-expand-all="true"
            :current-node-key="selectedFolderId"
            @node-click="handleFolderSelect"
            empty-text="暂无文件夹"
          >
            <template #default="{ data }">
              <span class="folder-node">
                <el-icon class="folder-node-icon"><Folder /></el-icon>
                <span class="folder-node-label">{{ data.name }}</span>
              </span>
            </template>
          </el-tree>
        </el-card>
      </el-col>
      <el-col :span="20">
        <el-card shadow="never" class="materials-card">
          <div class="materials-toolbar">
            <el-form :inline="true" :model="filters" class="materials-filter">
              <el-form-item label="关键词">
                <el-input
                  v-model="filters.keyword"
                  placeholder="编号/文件名/标题"
                  clearable
                  @keyup.enter="handleSearch"
                  @clear="handleSearch"
                />
              </el-form-item>
            <el-form-item>
              <template #label>
                <span>素材形状</span>
                <el-tooltip placement="top" effect="dark" :content="shapeTooltipHtml" raw-content>
                  <el-icon class="shape-tip-icon"><QuestionFilled /></el-icon>
                </el-tooltip>
              </template>
                <el-select v-model="filters.shape" placeholder="全部" clearable class="filter-select">
                  <el-option v-for="shape in shapeOptions" :key="shape" :label="shape" :value="shape" />
                </el-select>
              </el-form-item>
              <el-form-item label="格式">
                <el-select v-model="filters.format" placeholder="全部" clearable filterable class="filter-select">
                  <el-option v-for="fmt in formatOptions" :key="fmt" :label="fmt" :value="fmt" />
                </el-select>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="handleSearch">查询</el-button>
                <el-button @click="resetFilters">重置</el-button>
              </el-form-item>
            </el-form>
            <div class="toolbar-actions">
              <el-button type="primary" @click="openUploadDialog">
                <el-icon><UploadFilled /></el-icon>
                上传
              </el-button>
            </div>
          </div>

          <div class="materials-batch-bar" v-if="materials.length">
            <div class="batch-left">
              <el-checkbox
                :model-value="isAllMaterialsSelected"
                :indeterminate="isIndeterminateSelection"
                @change="handleToggleAllSelection"
              >
                全部文件 ({{ materials.length }})
              </el-checkbox>
              <span class="batch-selected">已选 {{ selectedMaterials.length }} 项</span>
              <el-button text type="primary" size="small" @click="clearMaterialSelection" :disabled="!selectedMaterials.length">
                清空已选
              </el-button>
            </div>
            <div class="batch-actions">
             
              <el-button  size="small" @click="openMoveDialog" :disabled="!selectedMaterials.length">移动到</el-button>
            
              <el-button size="small" type="danger" plain :disabled="!selectedMaterials.length" @click="handleBatchDelete">
                批量删除
              </el-button>
            </div>
          </div>

          <el-table ref="materialsTableRef" :data="materials" v-loading="materialsLoading" border size="small" class="materials-table" @selection-change="handleSelectionChange">
            <el-table-column type="selection" width="48" />
            <el-table-column label="素材" min-width="200">
              <template #default="{ row }">
                <div class="material-item">
                  <div class="material-thumb">
                    <el-image
                      :src="getMaterialPreview(row)"
                      fit="cover"
                      :preview-src-list="getMaterialPreviewList(row)"
                      :preview-teleported="true"
                    >
                      <template #error>
                        <div class="material-thumb__placeholder">无预览</div>
                      </template>
                    </el-image>
                  </div>
                  <div class="material-info">
                    <div class="material-code">编号：{{ row.code || '-' }}</div>
                    <div class="material-file">文件名：{{ row.file_name || '-' }}</div>
                    <div class="material-title">标题：{{ row.title || '-' }}</div>
                  </div>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="素材规格" min-width="80">
              <template #default="{ row }">
                <div class="material-spec">
                  <div>尺寸：{{ formatDimensions(row) || '-' }}</div>
                  <div>格式：{{ (row.format || '-').toUpperCase() }}</div>
                  <div>大小：{{ formatFileSize(row.file_size) }}</div>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="order_count" label="订单数" width="90" />
            <el-table-column label="创建人" width="150">
              <template #default="{ row }">
                <div>{{ row.created_by_name || row.created_by || '-' }}</div>
                <div>{{ formatTime(row.created_at) }}</div>
              </template>
            </el-table-column>
            <el-table-column label="更新人" width="150">
              <template #default="{ row }">
                <div>{{ row.updated_by_name || row.updated_by || '-' }}</div>
                <div>{{ formatTime(row.updated_at) }}</div>
              </template>
            </el-table-column>
            <el-table-column label="操作" fixed="right" width="220">
              <template #default="{ row }">
                <el-button size="small" text type="primary" @click="downloadMaterial(row)" :disabled="!row.download_url && !row.file_path">
                  <el-icon><Download /></el-icon>
                  下载
                </el-button>
                <el-button size="small" text type="primary" @click="openEditDialog(row)">
                  <el-icon><Edit /></el-icon>
                  编辑
                </el-button>
                <el-button size="small" text type="danger" @click="confirmDeleteMaterial(row)">
                  <el-icon><Delete /></el-icon>
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="materials-pagination">
            <el-pagination
              background
              layout="prev, pager, next, sizes, total"
              :current-page="pagination.page"
              :page-size="pagination.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="pagination.total"
              @current-change="handlePageChange"
              @size-change="handlePageSizeChange"
            />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 上传素材 -->
    <el-dialog v-model="uploadDialogVisible" title="上传素材" width="520px">
      <el-form :model="uploadForm" ref="uploadFormRef" label-width="100px" class="upload-form">
        <el-form-item label="素材文件" required>
          <el-upload
            class="upload-block"
            drag
            action="#"
            :auto-upload="false"
            :show-file-list="true"
            :multiple="true"
            :limit="10"
            accept="image/*"
            :before-upload="beforeUploadValidate"
            :on-change="handleUploadChange"
            :file-list="uploadFileList"
          >
            <el-icon class="upload-icon"><UploadFilled /></el-icon>
            <div class="upload-text">将文件拖到此处或点击上传</div>
            <div class="el-upload__tip">只支持图片，可一次上传多个文件</div>
          </el-upload>
        </el-form-item>
        <el-form-item label="归属文件夹">
          <el-tree-select
            v-model="uploadForm.folder_id"
            :data="folderSelectOptions"
            :props="{ children: 'children', label: 'label', value: 'value' }"
            check-strictly
            clearable
            placeholder="请选择文件夹"
            class="folder-select"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploadLoading" @click="submitUpload">开始上传</el-button>
      </template>
    </el-dialog>

    <!-- 移动素材 -->
    <el-dialog v-model="moveDialogVisible" title="移动素材" width="420px">
      <el-form :model="moveForm" label-width="100px">
        <el-form-item label="目标文件夹">
          <el-tree-select
            v-model="moveForm.folder_id"
            :data="folderSelectOptions"
            :props="{ children: 'children', label: 'label', value: 'value' }"
            check-strictly
            clearable
            placeholder="请选择目标文件夹"
            class="folder-select"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="moveDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitMove">确定</el-button>
      </template>
    </el-dialog>

    <!-- 编辑素材 -->
    <el-dialog v-model="materialDialogVisible" :title="editDialogTitle" width="560px">
      <el-form :model="materialForm" ref="materialFormRef" label-width="110px" class="material-form">
    
        <el-form-item label="标题">
          <el-input v-model="materialForm.title" placeholder="请输入标题" />
        </el-form-item>
        <el-form-item label="归属文件夹">
          <el-tree-select
            v-model="materialForm.folder_id"
            :data="folderSelectOptions"
            :props="{ children: 'children', label: 'label', value: 'value' }"
            check-strictly
            clearable
            placeholder="请选择文件夹"
            class="folder-select"
          />
        </el-form-item>
        
      </el-form>
      <template #footer>
        <el-button @click="materialDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="materialSaving" @click="submitMaterial">保存</el-button>
      </template>
    </el-dialog>

    <!-- 文件夹对话框 -->
    <el-dialog v-model="folderDialogVisible" :title="folderDialogMode === 'create' ? '新建文件夹' : '重命名文件夹'" width="400px">
      <el-form :model="folderForm" ref="folderFormRef" label-width="90px" class="folder-form">
        <el-form-item label="文件夹名" :rules="[{ required: true, message: '请输入文件夹名', trigger: 'blur' }]" prop="name">
          <el-input v-model="folderForm.name" placeholder="请输入名称" />
        </el-form-item>
        <el-form-item label="上级文件夹">
          <el-tree-select
            v-model="folderForm.parent_id"
            :data="folderSelectOptions"
            :props="{ children: 'children', label: 'label', value: 'value' }"
            check-strictly
            clearable
            placeholder="请选择上级文件夹"
            class="folder-select"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="folderDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="folderSaving" @click="submitFolder">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/utils/api'
import { formatBeijingTime } from '@/utils/date'
import { useAuthStore } from '@/stores/auth'
import {
  Plus,
  UploadFilled,
  Folder,
  Download,
  Edit,
  Delete,
  QuestionFilled
} from '@element-plus/icons-vue'

const authStore = useAuthStore()

const folderTree = ref([])
const folderLoading = ref(false)
const folderTreeRef = ref(null)
const selectedFolder = ref(null)
const treeProps = { children: 'children', label: 'name' }

const materials = ref([])
const materialsLoading = ref(false)
const materialsTableRef = ref(null)
const selectedMaterials = ref([])
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const filters = reactive({
  keyword: '',
  shape: '',
  format: ''
})

const shapeOptions = [
  '超横形',
  '横向形',
  '似方形',
  '竖向形',
  '超竖形'
]

const shapeTooltipHtml = [
  '超横形: 宽:高=1:0.2（含以下）',
  '横向形: 宽:高=1:0.2~1:0.9',
  '似方形: 宽:高=1:0.9~1:1.1',
  '竖向形: 宽:高=1:1.1~1:1.8',
  '超竖形: 宽:高=1:1.8（以上）'
].join('<br />')
const formatOptions = ref([])

const uploadDialogVisible = ref(false)
const uploadFormRef = ref(null)
const uploadForm = reactive({
  code: '',
  title: '',
  folder_id: null,
  order_count: 0
})
const uploadFileList = ref([])
const uploadFiles = ref([])
const uploadLoading = ref(false)
const allowedImageExtensions = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp']
const moveDialogVisible = ref(false)
const moveForm = reactive({
  folder_id: null
})

const materialDialogVisible = ref(false)
const materialFormRef = ref(null)
const materialForm = reactive({
  id: null,
  code: '',
  file_name: '',
  title: '',
  folder_id: null,
  order_count: 0,
  shape: '',
  width: 0,
  height: 0,
  format: '',
  file_size: 0,
  storage: '',
  file_path: ''
})
const materialSaving = ref(false)

const folderDialogVisible = ref(false)
const folderDialogMode = ref('create')
const folderFormRef = ref(null)
const folderForm = reactive({
  id: null,
  name: '',
  parent_id: null
})
const folderSaving = ref(false)

const selectedFolderId = computed(() => selectedFolder.value?.id ?? null)

const folderSelectOptions = computed(() => transformFolderOptions(folderTree.value))
const isAllMaterialsSelected = computed(
  () => selectedMaterials.value.length > 0 && selectedMaterials.value.length === materials.value.length
)
const isIndeterminateSelection = computed(
  () => selectedMaterials.value.length > 0 && !isAllMaterialsSelected.value
)
const batchEditQueue = ref([])
const currentEditIndex = ref(0)
const editDialogTitle = computed(() =>
  batchEditQueue.value.length
    ? `编辑素材 (${currentEditIndex.value + 1}/${batchEditQueue.value.length})`
    : '编辑素材'
)

function transformFolderOptions(items = []) {
  return items.map(item => ({
    value: item.id,
    label: item.name,
    children: item.children && item.children.length ? transformFolderOptions(item.children) : []
  }))
}

async function fetchFolders() {
  folderLoading.value = true
  try {
    const response = await api.get('/material-folders')
    folderTree.value = response.data?.data || []
    await nextTick()
    if (selectedFolderId.value) {
      folderTreeRef.value?.setCurrentKey(selectedFolderId.value)
    }
  } catch (error) {
    console.error('加载文件夹失败', error)
    ElMessage.error('加载文件夹失败')
  } finally {
    folderLoading.value = false
  }
}

async function fetchMaterials(resetPage = false) {
  if (resetPage) {
    pagination.page = 1
  }
  materialsLoading.value = true
  try {
    const response = await api.get('/materials', {
      params: {
        page: pagination.page,
        page_size: pagination.pageSize,
        keyword: filters.keyword || undefined,
        shape: filters.shape || undefined,
        format: filters.format || undefined,
        folder_id: selectedFolderId.value || undefined
      }
    })
    const payload = response.data?.data || {}
    materials.value = payload.materials || []
    pagination.total = payload.total || 0
    pagination.page = payload.page || pagination.page
    pagination.pageSize = payload.page_size || pagination.pageSize

    const formats = Array.from(
      new Set(materials.value.map(item => item.format).filter(Boolean))
    )
    formatOptions.value = formats

    await nextTick()
    clearMaterialSelection()
  } catch (error) {
    console.error('加载素材失败', error)
    ElMessage.error('加载素材失败')
  } finally {
    materialsLoading.value = false
  }
}

function handleFolderSelect(data) {
  selectedFolder.value = data
  fetchMaterials(true)
}

function resetFolderFilter() {
  selectedFolder.value = null
  folderTreeRef.value?.setCurrentKey(null)
  fetchMaterials(true)
}

function handleSearch() {
  fetchMaterials(true)
}

function resetFilters() {
  filters.keyword = ''
  filters.shape = ''
  filters.format = ''
  resetFolderFilter()
}

function handlePageChange(page) {
  pagination.page = page
  fetchMaterials()
}

function handlePageSizeChange(size) {
  pagination.pageSize = size
  fetchMaterials(true)
}

const isSingleUpload = computed(() => uploadFileList.value.length === 1)

function openUploadDialog() {
  uploadForm.code = ''
  uploadForm.title = ''
  uploadForm.folder_id = selectedFolderId.value
  uploadForm.order_count = 0
  uploadFileList.value = []
  uploadFiles.value = []
  uploadDialogVisible.value = true
}

function validateImageFile(rawFile) {
  if (!rawFile) return false
  const mimeType = (rawFile.type || '').toLowerCase()
  const ext = rawFile.name?.split('.').pop()?.toLowerCase()
  if (!ext) return false
  if (!allowedImageExtensions.includes(ext)) {
    return false
  }
  if (mimeType && !mimeType.startsWith('image/')) {
    return false
  }
  return true
}

function beforeUploadValidate(file) {
  if (!validateImageFile(file)) {
    ElMessage.error('仅支持上传 jpg/jpeg/png/gif/bmp/webp 图片文件')
  }
  // 阻止自动上传，统一在提交时处理
  return false
}

function handleUploadChange(file, fileList) {
  const filtered = fileList.filter(item => {
    if (validateImageFile(item.raw)) {
      return true
    }
    ElMessage.warning(`文件【${item.name}】不是支持的图片格式，已自动移除`)
    return false
  })
  uploadFileList.value = filtered
  uploadFiles.value = filtered.map(item => item.raw).filter(Boolean)
  if (!isSingleUpload.value) {
    uploadForm.code = ''
    uploadForm.title = ''
  }
}

async function submitUpload() {
  if (!uploadFiles.value.length) {
    ElMessage.warning('请先选择素材文件')
    return
  }
  uploadLoading.value = true
  try {
    for (const [index, file] of uploadFiles.value.entries()) {
      const formData = new FormData()
      formData.append('file', file)
      if (isSingleUpload.value) {
        if (uploadForm.code) formData.append('code', uploadForm.code)
        if (uploadForm.title) formData.append('title', uploadForm.title)
      }
      if (uploadForm.folder_id) formData.append('folder_id', uploadForm.folder_id)
      if (uploadForm.order_count != null) {
        formData.append('order_count', String(uploadForm.order_count))
      }
      await api.post('/materials/upload', formData)
      if (uploadFiles.value.length > 1) {
        ElMessage.success(`已上传 ${index + 1}/${uploadFiles.value.length} 个素材`)
      }
    }
    ElMessage.success('上传成功')
    uploadDialogVisible.value = false
    fetchMaterials(true)
    fetchFolders()
  } catch (error) {
    console.error('上传素材失败', error)
    ElMessage.error(error.response?.data?.error || '上传素材失败')
  } finally {
    uploadLoading.value = false
  }
}

function openEditDialog(row, fromBatch = false) {
  materialForm.id = row.id
  materialForm.code = row.code
  materialForm.file_name = row.file_name
  materialForm.title = row.title
  materialForm.folder_id = row.folder_id || null
  materialForm.order_count = row.order_count ?? 0
  materialForm.shape = row.shape || ''
  materialForm.width = row.width || 0
  materialForm.height = row.height || 0
  materialForm.format = row.format || ''
  materialForm.file_size = row.file_size || 0
  materialForm.storage = row.storage || ''
  materialForm.file_path = row.file_path || ''
  if (!fromBatch) {
    batchEditQueue.value = []
    currentEditIndex.value = 0
  }
  materialDialogVisible.value = true
}

async function submitMaterial() {
  if (!materialForm.id) return
  materialSaving.value = true
  try {
    const payload = {
      title: materialForm.title,
      folder_id: materialForm.folder_id,
      order_count: materialForm.order_count,
      shape: materialForm.shape,
      width: materialForm.width,
      height: materialForm.height,
      format: materialForm.format,
      file_size: materialForm.file_size
    }
    await api.put(`/materials/${materialForm.id}`, payload)
    ElMessage.success('保存成功')
    await fetchMaterials()
    fetchFolders()
    if (batchEditQueue.value.length) {
      const nextIndex = currentEditIndex.value + 1
      if (nextIndex < batchEditQueue.value.length) {
        currentEditIndex.value = nextIndex
        const nextId = batchEditQueue.value[nextIndex].id
        const nextMaterial =
          materials.value.find(item => item.id === nextId) || batchEditQueue.value[nextIndex]
        openEditDialog(nextMaterial, true)
        return
      }
      batchEditQueue.value = []
      currentEditIndex.value = 0
    }
    materialDialogVisible.value = false
  } catch (error) {
    console.error('保存素材失败', error)
    ElMessage.error(error.response?.data?.error || '保存素材失败')
  } finally {
    materialSaving.value = false
  }
}

async function confirmDeleteMaterial(row) {
  try {
    await ElMessageBox.confirm(`确定删除素材【${row.title || row.file_name}】吗？`, '提示', {
      type: 'warning'
    })
    await api.delete(`/materials/${row.id}`)
    ElMessage.success('删除成功')
    fetchMaterials()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除素材失败', error)
      ElMessage.error(error.response?.data?.error || '删除素材失败')
    }
  }
}

function handleSelectionChange(selection) {
  selectedMaterials.value = selection
}

function clearMaterialSelection() {
  materialsTableRef.value?.clearSelection()
  selectedMaterials.value = []
}

function handleToggleAllSelection(checked) {
  const table = materialsTableRef.value
  if (!table) return
  table.clearSelection()
  if (checked) {
    materials.value.forEach(row => {
      table.toggleRowSelection(row, true)
    })
  }
}

async function handleBatchDelete() {
  if (!selectedMaterials.value.length) return
  try {
    await ElMessageBox.confirm(`确定删除选中的 ${selectedMaterials.value.length} 个素材吗？`, '提示', {
      type: 'warning'
    })
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量删除取消或失败', error)
    }
    return
  }

  try {
    await Promise.all(selectedMaterials.value.map(item => api.delete(`/materials/${item.id}`)))
    ElMessage.success('批量删除成功')
    clearMaterialSelection()
    fetchMaterials()
  } catch (error) {
    console.error('批量删除素材失败', error)
    ElMessage.error(error.response?.data?.error || '批量删除素材失败')
  }
}

const defaultMaterialThumbnail =
  'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="120" height="120" viewBox="0 0 120 120"><rect width="120" height="120" rx="8" fill="%23f5f5f5"/><text x="50%" y="50%" dominant-baseline="middle" text-anchor="middle" font-size="12" fill="%23999">素材</text></svg>'

function withAuthToken(url) {
  const token = authStore.token || localStorage.getItem('token')
  if (!url || !token) return url
  const isApiUrl =
    url.startsWith('/api/') ||
    url.startsWith(`${window.location.origin}/api/`)
  if (!isApiUrl) return url
  const separator = url.includes('?') ? '&' : '?'
  return `${url}${separator}token=${encodeURIComponent(token)}`
}

function getMaterialPreviewUrl(row) {
  let url = row.preview_url || row.download_url
  if (!url && row.id) {
    url = `/api/materials/${row.id}/download?disposition=inline`
  }
  return withAuthToken(url)
}

function getMaterialDownloadUrl(row) {
  let url = row.download_url
  if (!url && row.id) {
    url = `/api/materials/${row.id}/download?disposition=attachment`
  }
  return withAuthToken(url)
}

function getMaterialPreview(row) {
  return getMaterialPreviewUrl(row) || defaultMaterialThumbnail
}

function getMaterialPreviewList(row) {
  const url = getMaterialPreviewUrl(row)
  return url ? [url] : []
}

function openMoveDialog() {
  if (!selectedMaterials.value.length) {
    ElMessage.warning('请先选择素材')
    return
  }
  moveForm.folder_id = selectedFolderId.value || null
  moveDialogVisible.value = true
}

function startBatchEdit() {
  if (!selectedMaterials.value.length) {
    ElMessage.warning('请先选择素材')
    return
  }
  batchEditQueue.value = selectedMaterials.value.map(item => ({ ...item }))
  currentEditIndex.value = 0
  openEditDialog(batchEditQueue.value[0], true)
}

async function submitMove() {
  if (!selectedMaterials.value.length) {
    ElMessage.warning('请选择要移动的素材')
    return
  }
  const payload = {
    folder_id: moveForm.folder_id || null
  }
  try {
    await Promise.all(selectedMaterials.value.map(item => api.put(`/materials/${item.id}`, payload)))
    ElMessage.success('移动成功')
    moveDialogVisible.value = false
    fetchMaterials()
    fetchFolders()
  } catch (error) {
    console.error('移动素材失败', error)
    ElMessage.error(error.response?.data?.error || '移动素材失败')
  }
}

function downloadMaterial(row) {
  const url = getMaterialDownloadUrl(row)
  if (!url) {
    ElMessage.warning('无法获取下载链接')
    return
  }
  window.open(url, '_blank')
}

function openFolderDialog(mode) {
  folderDialogMode.value = mode
  if (mode === 'create') {
    folderForm.id = null
    folderForm.name = ''
    folderForm.parent_id = selectedFolderId.value || null
  } else if (mode === 'rename' && selectedFolder.value) {
    folderForm.id = selectedFolder.value.id
    folderForm.name = selectedFolder.value.name
    folderForm.parent_id = selectedFolder.value.parent_id || null
  } else {
    return
  }
  folderDialogVisible.value = true
  nextTick(() => {
    folderFormRef.value?.clearValidate()
  })
}

async function submitFolder() {
  const payload = {
    name: folderForm.name.trim(),
    parent_id: folderForm.parent_id || null
  }
  if (!payload.name) {
    ElMessage.warning('文件夹名称不能为空')
    return
  }
  if (folderDialogMode.value === 'rename' && folderForm.id && payload.parent_id === folderForm.id) {
    ElMessage.warning('上级文件夹不能为自身')
    return
  }

  folderSaving.value = true
  try {
    if (folderDialogMode.value === 'create') {
      await api.post('/material-folders', payload)
      ElMessage.success('新增文件夹成功')
    } else if (folderDialogMode.value === 'rename' && folderForm.id) {
      await api.put(`/material-folders/${folderForm.id}`, payload)
      ElMessage.success('更新文件夹成功')
    }
    folderDialogVisible.value = false
    fetchFolders()
  } catch (error) {
    console.error('保存文件夹失败', error)
    ElMessage.error(error.response?.data?.error || '保存文件夹失败')
  } finally {
    folderSaving.value = false
  }
}

async function confirmDeleteFolder() {
  if (!selectedFolder.value) return
  try {
    await ElMessageBox.confirm(`确定删除文件夹【${selectedFolder.value.name}】吗？该操作不可恢复。`, '提示', {
      type: 'warning'
    })
    await api.delete(`/material-folders/${selectedFolder.value.id}`)
    ElMessage.success('删除成功')
    selectedFolder.value = null
    fetchFolders()
    fetchMaterials(true)
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除文件夹失败', error)
      ElMessage.error(error.response?.data?.error || '删除文件夹失败')
    }
  }
}

function formatDimensions(row) {
  if (row.dimensions) return row.dimensions
  const width = row.width || 0
  const height = row.height || 0
  if (!width || !height) return '-'
  return `${width} × ${height}`
}

function formatFileSize(size) {
  if (!size) return '-'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let value = size
  let unitIndex = 0
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex++
  }
  return `${value.toFixed(unitIndex === 0 ? 0 : 2)} ${units[unitIndex]}`
}

function formatTime(value) {
  return formatBeijingTime(value)
}

onMounted(() => {
  fetchFolders()
  fetchMaterials(true)
})
</script>

<style scoped>
.materials-page {
  padding: 16px;
}

.folder-column {
  min-height: 600px;
}

.folder-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.folder-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.folder-title {
  font-weight: 600;
  font-size: 16px;
}

.folder-header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.folder-actions {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.folder-tree {
  flex: 1;
  overflow: auto;
}

.folder-node {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.folder-node-icon {
  font-size: 16px;
}

.materials-card {
  min-height: 600px;
}

.materials-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
  gap: 12px;
}

.materials-filter {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.filter-select {
  width: 160px;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.materials-table {
  width: 100%;
}

.materials-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.materials-batch-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  margin-bottom: 12px;
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
}

.materials-batch-bar .batch-left {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 13px;
}

.materials-batch-bar .batch-selected {
  color: var(--el-text-color-secondary);
}

.materials-batch-bar .batch-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.upload-form .upload-block {
  width: 100%;
}

.upload-icon {
  font-size: 40px;
  color: var(--el-color-primary);
}

.upload-text {
  margin-top: 8px;
  font-size: 14px;
  color: var(--el-text-color-primary);
}

.material-form .folder-select,
.upload-form .folder-select,
.folder-form .folder-select {
  width: 100%;
}

.material-item {
  display: flex;
  align-items: center;
  gap: 16px;
}

.material-thumb {
  width: 120px;
  height: 120px;
  border-radius: 8px;
  overflow: hidden;
  background: var(--el-fill-color);
  border: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}

.material-thumb :deep(.el-image),
.material-thumb img {
  width: 100%;
  height: 100%;
}

.material-thumb__placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.material-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 13px;
}

.material-info .material-code {
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.material-info .material-file,
.material-info .material-title {
  color: var(--el-text-color-regular);
}

.material-spec {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 13px;
  color: var(--el-text-color-regular);
}

.shape-tip-icon {
  margin-left: 4px;
  cursor: pointer;
  color: var(--el-color-info);
  font-size: 14px;
}

.dimension-inputs {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dimension-separator {
  font-size: 16px;
  color: var(--el-text-color-regular);
}

:global(.el-image-viewer__wrapper) {
  z-index: 4000;
}

:global(.el-image-viewer__mask) {
  background-color: rgba(0, 0, 0, 0.85);
}

:global(.el-image-viewer__img) {
  max-width: 90vw;
  max-height: 90vh;
  object-fit: contain;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.45);
}
</style>
