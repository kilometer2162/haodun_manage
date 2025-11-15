<template>
  <div class="orders-page">
    <div class="page-header">
      <h2>订单管理</h2>
    </div>

    <el-form :model="filterForm" inline class="filter-form" label-width="100px">
      
      <el-form-item label="模糊查询">
        <el-select
          v-model="filterForm.fuzzyField"
          placeholder="选择字段"
          clearable
          style="width: 150px"  >
          <el-option
            v-for="item in fuzzyFieldOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
        <el-input
          v-model="filterForm.fuzzyKeyword"
          :placeholder="getOptionLabel(fuzzyFieldOptions, filterForm.fuzzyField) ? `请输入${getOptionLabel(fuzzyFieldOptions, filterForm.fuzzyField)}` : '请输入关键词'"
          :disabled="!filterForm.fuzzyField"
          clearable
          style="width: 220px; margin-left: 12px"
        />
      </el-form-item>
      <el-form-item label="精确查询">
        <el-select v-model="filterForm.exactField"
          placeholder="选择字段"  clearable style="width: 150px" >
          <el-option
            v-for="item in exactFieldOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
        <el-input
          v-model="filterForm.exactValue"
          :placeholder="getOptionLabel(exactFieldOptions, filterForm.exactField) ? `请输入${getOptionLabel(exactFieldOptions, filterForm.exactField)}` : '请输入'"
          :disabled="!filterForm.exactField"
          clearable
          style="width: 220px; margin-left: 12px"
        />
        <el-form-item label="时间范围">
        <el-select
          v-model="filterForm.timeField"
          placeholder="选择时间字段"
          clearable
          style="width: 180px"
        >
          <el-option
            v-for="item in timeFieldOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
        <el-date-picker
          v-model="filterForm.timeRange"
          type="datetimerange"
          value-format="YYYY-MM-DD HH:mm:ss"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          :disabled="!filterForm.timeField"
          style="width: 360px; margin-left: 12px"
          clearable
        />
      </el-form-item>
      
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>

      <el-form-item class="filter-actions" label-width="0">
        <el-button type="success" @click="handleImport">导入Excel</el-button>
        <el-button type="info" plain @click="openBatchUploadDialog">导入素材/面单</el-button>
        <el-button type="warning" @click="handleExport">导出Excel</el-button>
      </el-form-item>
    </el-form>

    <el-tabs v-model="activeTab" class="order-tabs" @tab-change="handleTabChange">
      <el-tab-pane label="平台面单" name="platform">
        <div class="table-container">
          <el-table :data="platformOrders" v-loading="loading" border style="width: 100%">
          <el-table-column label="素材图" width="120" fixed="left">
            <template #default="{ row }">
              <div class="material-cell">
                <el-image
                  v-if="row.material_image_url"
                  :src="getMaterialImageURL(row)"
                  fit="cover"
                  :preview-src-list="getMaterialPreviewList(row)"
                  :preview-teleported="true"
                />
                <el-button
                  v-else
                  circle
                  class="upload-button ghost"
                  @click="handleUploadAttachment(row, 'material_image')"
                >
                  <el-icon><PictureFilled /></el-icon>
                </el-button>
                <el-button
                  v-if="row.material_image_url"
                  type="danger"
                  circle
                  :icon="Delete"
                  size="small"
                  class="material-delete"
                  @click="handleRemoveAttachment(row, 'material_image')"
                />
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="gsp_order_no" label="GSP订单号" width="160" />
          <el-table-column prop="shipping_warehouse_code" label="发货仓库" width="110" />
          <el-table-column prop="shop_code" label="店铺编号" width="140" />
          <el-table-column prop="owner_name" label="负责人" width="110" />
          <el-table-column label="商品信息" min-width="280">
            <template #default="{ row }">
              <div class="cell-title">{{ row.product_name || '' }}</div>
              <div class="cell-sub">规格：{{ row.spec || '-' }}　货号：{{ row.item_no || '-' }}</div>
              <div class="cell-sub">卖家SKU：{{ row.seller_sku || '-' }}</div>
              <div class="cell-sub">平台SKU：{{ row.platform_sku || '-' }}</div>
              <div class="cell-sub">SKC：{{ row.platform_skc || '-' }}　SPU：{{ row.platform_spu || '-' }}</div>
              <div class="cell-sub">
                价格：{{ row.product_price !== null && row.product_price !== undefined && row.product_price !== '' ? `￥${formatCurrency(row.product_price)}` : '-' }}
              </div>
              <div class="cell-sub" v-if="row.special_product_note">备注：{{ row.special_product_note }}</div>
              <div class="cell-sub">应履约件数：{{ row.expected_fulfillment_qty ?? '-' }}</div>
            </template>
          </el-table-column>
          <el-table-column label="收货信息" min-width="340">
            <template #default="{ row }">
              <div class="cell-title">{{ row.customer_full_name }}</div>
              <!-- <div class="cell-sub">{{ buildFullName(row) }}</div> -->
              <div class="cell-sub">{{ buildRegion(row) }}</div>
              <div class="cell-sub">地址1：{{ formatAddress(row.address_line1) }}</div>
              <div class="cell-sub">地址2：{{ formatAddress(row.address_line2) }}</div>
            </template>
          </el-table-column>
          <el-table-column label="联系信息" width="250">
            <template #default="{ row }">
              <div class="cell-sub">电话：{{ row.phone_number || '-' }}</div>
              <div class="cell-sub">邮箱：{{ row.email || '-' }}</div>
              <div class="cell-sub">税号：{{ row.tax_number || '-' }}</div>
              <div class="cell-sub">邮编：{{ row.postal_code || '-' }}</div>
            </template>
          </el-table-column>
          <el-table-column label="面单" width="280" fixed="right">
            <template #default="{ row }">
              <div class="shipping-label-cell">
                <el-button
                  v-if="row.shipping_label_url"
                  type="primary"
                  link
                  :icon="Document"
                  @click="openShippingLabel(row)"
                >{{ row.shipping_label_file_name || extractFileName(row.shipping_label_url) }}</el-button>
                <el-button
                  v-else
                  circle
                  class="upload-button ghost"
                  @click="handleUploadAttachment(row, 'shipping_label')"
                >
                  <el-icon><DocumentAdd /></el-icon>
                </el-button>
                <el-button
                  v-if="row.shipping_label_url"
                  type="danger"
                  circle
                  :icon="Delete"
                  size="small"
                  class="material-delete"
                  @click="handleRemoveAttachment(row, 'shipping_label')"
                />
              </div>
            </template>
          </el-table-column>
          <el-table-column label="操作" fixed="right" width="200">
            <template #default="{ row }">
              <div class="operation-actions">
                <el-button type="primary" link @click="handleEditOrder(row)">编辑</el-button>
                <el-button type="primary" link @click="handleViewDetail(row)">查看</el-button>
                <el-button type="danger" link @click="handleDeleteOrder(row)">删除</el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
        </div>
        <div class="table-pagination">
          <el-pagination
            v-model:current-page="platformPage"
            v-model:page-size="platformPageSize"
            :total="platformTotal"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handlePlatformSizeChange"
            @current-change="handlePlatformPageChange"
          />
        </div>
      </el-tab-pane>

      <el-tab-pane label="工厂物流" name="factory">
        <div class="table-container">
          <el-table :data="factoryOrders" v-loading="loading" border style="width: 100%">
          <el-table-column label="素材图" width="120" fixed="left">
            <template #default="{ row }">
              <div class="material-cell">
                <el-image
                  v-if="row.material_image_url"
                  :src="getMaterialImageURL(row)"
                  fit="cover"
                  :preview-src-list="getMaterialPreviewList(row)"
                  :preview-teleported="true"
                />
                <el-button
                  v-else
                  circle
                  class="upload-button ghost"
                  @click="handleUploadAttachment(row, 'material_image')"
                >
                  <el-icon><PictureFilled /></el-icon>
                </el-button>
                <el-button
                  v-if="row.material_image_url"
                  type="danger"
                  circle
                  :icon="Delete"
                  size="small"
                  class="material-delete"
                  @click="handleRemoveAttachment(row, 'material_image')"
                />
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="gsp_order_no" label="GSP订单号" width="160" />
          <el-table-column prop="order_created_at" label="订单创建时间" width="180">
            <template #default="{ row }">
              {{ formatTime(row.order_created_at) }}
            </template>
          </el-table-column>
          <el-table-column prop="shop_code" label="店铺编号" width="140" />
          <el-table-column prop="owner_name" label="负责人" width="100" />
          <el-table-column label="商品信息" min-width="280">
            <template #default="{ row }">
              <div class="cell-title">{{ row.product_name || '' }}</div>
              <div class="cell-sub">规格：{{ row.spec || '-' }}　货号：{{ row.item_no || '-' }}</div>
              <div class="cell-sub">卖家SKU：{{ row.seller_sku || '-' }}</div>
              <div class="cell-sub">平台SKU：{{ row.platform_sku || '-' }}</div>
              <div class="cell-sub">SKC：{{ row.platform_skc || '-' }}　SPU：{{ row.platform_spu || '-' }}</div>
              <div class="cell-sub">价格：￥{{ formatCurrency(row.product_price) }}</div>
              <div class="cell-sub" v-if="row.special_product_note">备注：{{ row.special_product_note }}</div>
            </template>
          </el-table-column>
          <el-table-column label="收货信息" min-width="250">
            <template #default="{ row }">
              <div class="cell-title">{{ row.customer_full_name }}</div>
              <!-- <div class="cell-sub">{{ buildFullName(row) }}</div> -->
              <div class="cell-sub">{{ buildRegion(row) }}</div>
              <div class="cell-sub">地址1：{{ formatAddress(row.address_line1) }}</div>
              <div class="cell-sub">地址2：{{ formatAddress(row.address_line2) }}</div>
            </template>
          </el-table-column>
          <el-table-column label="联系信息" width="220">
            <template #default="{ row }">
              <div class="cell-sub">电话：{{ row.phone_number || '-' }}</div>
              <div class="cell-sub">邮箱：{{ row.email || '-' }}</div>
              <div class="cell-sub">税号：{{ row.tax_number || '-' }}</div>
              <div class="cell-sub">邮编：{{ row.postal_code || '-' }}</div>
            </template>
          </el-table-column>
          <el-table-column label="操作" fixed="right" width="200">
            <template #default="{ row }">
              <div class="operation-actions">
                <el-button type="primary" link @click="handleEditOrder(row)">编辑</el-button>
                <el-button type="primary" link @click="handleViewDetail(row)">查看</el-button>
                <el-button type="danger" link @click="handleDeleteOrder(row)">删除</el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
        </div>
        <div class="table-pagination">
          <el-pagination
            v-model:current-page="factoryPage"
            v-model:page-size="factoryPageSize"
            :total="factoryTotal"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleFactorySizeChange"
            @current-change="handleFactoryPageChange"
          />
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-drawer v-model="detailVisible" title="订单详情" size="50%" destroy-on-close>
      <el-descriptions :column="2" border v-if="detailOrder">
        <el-descriptions-item v-for="field in detailFields" :key="field.prop" :label="field.label">
          <template v-if="field.type === 'statusTag'">
            <template v-if="getOrderStatusLabel(detailOrder[field.prop])">
              <el-tag size="small">
                {{ getOrderStatusLabel(detailOrder[field.prop]) }}
              </el-tag>
            </template>
            <template v-else>
              {{ formatDetailField(field, detailOrder) }}
            </template>
          </template>
          <template v-else>
            {{ formatDetailField(field, detailOrder) }}
          </template>
        </el-descriptions-item>
      </el-descriptions>
      <el-empty v-else description="暂无数据" />
    </el-drawer>

    <el-dialog
      v-model="editDialogVisible"
      title="编辑订单"
      width="900px"
      class="order-edit-dialog"
      :close-on-click-modal="false"
      @closed="handleEditDialogClosed"
    >
      <el-form ref="editFormRef" :model="editForm" label-width="120px" class="order-edit-form">
        <template v-for="group in editableFieldGroups" :key="group.label">
          <el-divider content-position="left">{{ group.label }}</el-divider>
          <el-row :gutter="16">
            <el-col v-for="field in group.fields" :key="field.prop" :span="field.span || 12">
              <el-form-item :label="field.label">
                <template v-if="field.type === 'select'">
                  <el-select v-model="editForm[field.prop]" placeholder="请选择" style="width: 100%">
                    <el-option
                      v-for="item in getFieldOptions(field)"
                      :key="String(item.value)"
                      :label="item.label"
                      :value="item.value"
                    />
                  </el-select>
                </template>
                <template v-else-if="field.type === 'textarea'">
                  <el-input
                    v-model="editForm[field.prop]"
                    type="textarea"
                    :rows="3"
                    :placeholder="field.optional ? '可选填写' : ''"
                  />
                </template>
                <template v-else-if="field.type === 'float'">
                  <el-input-number
                    v-model="editForm[field.prop]"
                    :min="0"
                    :step="0.01"
                    :precision="2"
                    controls-position="right"
                    style="width: 100%"
                  />
                </template>
                <template v-else-if="field.type === 'int'">
                  <el-input-number
                    v-model="editForm[field.prop]"
                    :min="0"
                    :step="1"
                    controls-position="right"
                    style="width: 100%"
                  />
                </template>
                <template v-else>
                  <el-input
                    v-model="editForm[field.prop]"
                    :placeholder="field.optional ? '可选填写' : ''"
                    clearable
                  />
                </template>
              </el-form-item>
            </el-col>
          </el-row>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editLoading" @click="submitEdit">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="batchUploadDialogVisible"
      title="导入素材/面单文件"
      width="600px"
      @closed="handleBatchUploadDialogClosed"
    >
      <div class="upload-tip">
        <p>• 素材图文件名需与订单货号一致，例如 <code>ABC123.jpg</code></p>
        <p>• 面单文件需为 PDF，文件名需与订单 GSP 订单号一致，例如 <code>GSP0001.pdf</code></p>
      </div>
      <el-upload
        class="order-upload"
        drag
        :multiple="true"
        :auto-upload="false"
        :file-list="batchUploadFileList"
        :on-change="handleBatchUploadChange"
        :on-remove="handleBatchUploadRemove"
        :accept="batchUploadAccept"
      >
        <el-icon class="upload-icon"><Plus /></el-icon>
        <div class="el-upload__text">将文件拖到此处，或<em>点击上传</em></div>
        <div class="el-upload__tip">支持图片与 PDF 文件，可一次选择多个</div>
      </el-upload>
      <div v-if="batchUploadResult.success.length || batchUploadResult.failed.length" class="batch-upload-result">
        <el-alert
          v-if="batchUploadResult.success.length"
          title="成功上传"
          type="success"
          show-icon
          :closable="false"
        >
          <ul>
            <li v-for="item in batchUploadResult.success" :key="item.file_name">{{ item.message }}</li>
          </ul>
        </el-alert>
        <el-alert
          v-if="batchUploadResult.failed.length"
          title="上传失败"
          type="error"
          show-icon
          :closable="false"
          style="margin-top: 12px"
        >
          <ul>
            <li v-for="item in batchUploadResult.failed" :key="item.file_name">{{ item.message }}</li>
          </ul>
        </el-alert>
      </div>
      <template #footer>
        <el-button @click="batchUploadDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="batchUploading" @click="submitBatchUpload">开始上传</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="uploadDialogVisible" :title="uploadDialogTitle" width="520px" @closed="handleUploadDialogClosed">
      <el-form
        :model="uploadTarget"
        ref="uploadFormRef"
        :label-width="uploadTarget.fileType === 'import' ? 'auto' : '100px'"
        :label-position="uploadTarget.fileType === 'import' ? 'top' : 'right'"
        :class="['upload-form', uploadTarget.fileType === 'import' ? 'upload-form--import' : '']"
      >
        <template v-if="uploadTarget.fileType === 'import'">
          <el-form-item label="导入文件" required>
            <el-upload
              class="upload-block"
              drag
              action="#"
              :auto-upload="false"
              :show-file-list="true"
              :limit="1"
              :before-upload="() => false"
              :on-change="handleUploadChange"
              :on-remove="handleUploadRemove"
              :file-list="uploadFileList"
              :accept="uploadAccept"
            >
              <el-icon class="upload-icon"><UploadFilled /></el-icon>
              <div class="upload-text">{{ uploadTip }}</div>
            </el-upload>
          </el-form-item>
        </template>
        <template v-else>
          <el-form-item label="附件文件" required>
            <el-upload
              class="upload-block"
              drag
              action="#"
              :auto-upload="false"
              :show-file-list="true"
              :limit="1"
              :before-upload="() => false"
              :on-change="handleUploadChange"
              :on-remove="handleUploadRemove"
              :file-list="uploadFileList"
              :accept="uploadAccept"
            >
              <el-icon class="upload-icon"><UploadFilled /></el-icon>
              <div class="upload-text">{{ uploadTip }}</div>
            </el-upload>
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="submitUpload">开始上传</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="materialSelectorVisible"
      title="选择素材图"
      width="720px"
      @closed="handleMaterialSelectorClosed"
    >
      <div class="material-selector-toolbar">
        <el-input
          v-model="materialSelectorKeyword"
          placeholder="搜索编号/文件名/标题"
          clearable
          class="toolbar-search"
          @keyup.enter="handleMaterialSelectorSearch"
          @clear="handleMaterialSelectorSearch"
        >
          <template #append>
            <el-button @click="handleMaterialSelectorSearch">搜索</el-button>
          </template>
        </el-input>
        <el-button @click="fetchMaterialSelectorList" :loading="materialSelectorLoading">刷新</el-button>
      </div>
        <el-table
        ref="materialSelectorTableRef"
        :data="materialSelectorList"
        v-loading="materialSelectorLoading"
        border
        height="360"
        highlight-current-row
        @current-change="handleMaterialCurrentChange"
      >
        <el-table-column label="预览" width="110">
          <template #default="{ row }">
            <el-image
              v-if="getMaterialSelectorPreviewURL(row)"
              :src="getMaterialSelectorPreviewURL(row)"
              fit="cover"
              :preview-src-list="getMaterialSelectorPreviewList(row)"
              style="width: 72px; height: 72px"
              :preview-teleported="true"
            >
              <template #error>
                <span>--</span>
              </template>
            </el-image>
            <span v-else>--</span>
          </template>
        </el-table-column>
        <el-table-column label="编码 / 标题" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="material-selector-info">
              <div class="selector-code">编码：{{ row.code || '-' }}</div>
              <div class="selector-title">标题：{{ row.title || '-' }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="尺寸 / 规格" min-width="180">
          <template #default="{ row }">
            <div class="material-selector-spec">
              <div>尺寸：{{ formatDimensions(row) || '-' }}</div>
              <div>格式：{{ (row.format || '-').toUpperCase() }}</div>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <div class="material-selector-pagination">
        <el-pagination
          background
          layout="prev, pager, next, sizes, total"
          :current-page="materialSelectorPage"
          :page-size="materialSelectorPageSize"
          :page-sizes="[10, 20, 50]"
          :total="materialSelectorTotal"
          @current-change="handleMaterialSelectorPageChange"
          @size-change="handleMaterialSelectorSizeChange"
        />
      </div>
      <template #footer>
        <el-button @click="materialSelectorVisible = false">取消</el-button>
        <el-button type="primary" :disabled="!selectedMaterial" @click="submitMaterialSelection">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive, computed, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Document, DocumentAdd, PictureFilled, Search, UploadFilled } from '@element-plus/icons-vue'
import api from '@/utils/api'
import { formatBeijingTime } from '@/utils/date'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

const DEFAULT_ADDRESS_KEYWORDS = ['默认地址', '在【系统参数】修改']
const formatAddress = (value) => {
  if (!value) return '-'
  const normalized = String(value)
  if (DEFAULT_ADDRESS_KEYWORDS.some((keyword) => normalized.includes(keyword))) {
    return '-'
  }
  return normalized
}

const formatCurrency = (value) => {
  if (value === null || value === undefined || value === '') {
    return '-'
  }
  const num = Number(value)
  if (Number.isNaN(num)) {
    return value
  }
  return num.toFixed(2)
}

const formatTime = (value) => {
  if (!value) return '-'
  return formatBeijingTime(value)
}

const formatDimensions = (row) => {
  if (!row) return '-'
  if (row.dimensions) return row.dimensions
  const width = row.width || row.material_width || row.image_width || 0
  const height = row.height || row.material_height || row.image_height || 0
  if (width && height) {
    return `${width} × ${height}`
  }
  return '-'
}

const formatFileSize = (size) => {
  if (!size && size !== 0) return '-'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let value = Number(size)
  if (Number.isNaN(value) || value <= 0) return '-'
  let unitIndex = 0
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex++
  }
  return `${value.toFixed(unitIndex === 0 ? 0 : 2)} ${units[unitIndex]}`
}

const fuzzyFieldOptions = [
  { label: 'GSP订单号', value: 'gsp_order_no' },
  { label: '发货仓库', value: 'shipping_warehouse_code' },
  { label: '店铺编号', value: 'shop_code' },
  { label: '负责人', value: 'owner_name' },
  { label: '商品名称', value: 'product_name' },
  { label: '规格', value: 'spec' },
  { label: '货号', value: 'item_no' },
  { label: '卖家SKU', value: 'seller_sku' },
  { label: '平台SKU', value: 'platform_sku' },
  { label: '平台SKC', value: 'platform_skc' },
  { label: '平台SPU', value: 'platform_spu' },
  { label: '特殊产品备注', value: 'special_product_note' },
  { label: '币种', value: 'currency_code' },
  { label: '邮编', value: 'postal_code' },
  { label: '国家', value: 'country' },
  { label: '省份', value: 'province' },
  { label: '城市', value: 'city' },
  { label: '区', value: 'district' },
  { label: '用户地址1', value: 'address_line1' },
  { label: '用户地址2', value: 'address_line2' },
  { label: '用户全称', value: 'customer_full_name' },
  { label: '用户姓氏', value: 'customer_last_name' },
  { label: '用户名字', value: 'customer_first_name' },
  { label: '手机号', value: 'phone_number' },
  { label: '用户邮箱', value: 'email' },
  { label: '税号', value: 'tax_number' }
]

const exactFieldOptions = [
  { label: '商品价格', value: 'product_price' },
  { label: '商品预计收入', value: 'expected_revenue' },
  { label: '应履约件数', value: 'expected_fulfillment_qty' }
]

const filterForm = reactive({
  timeField: '',
  timeRange: [],
  exactField: '',
  exactValue: '',
  fuzzyField: '',
  fuzzyKeyword: ''
})

const timeFieldOptions = [
  { label: '订单创建时间', value: 'order_created_at' },
  { label: '完成时间', value: 'completed_at' },
  { label: '支付时间', value: 'payment_time' },
  { label: '要求签收时间', value: 'required_sign_at' }
]

const getOptionLabel = (options, value) => {
  const option = options.find((item) => item.value === value)
  return option ? option.label : ''
}

const buildRegion = (row) => {
  const parts = [row.country, row.province, row.city, row.district].filter(Boolean)
  return parts.length ? parts.join(' / ') : ''
}

const buildFullName = (row) => {
  const first = row.customer_first_name || ''
  const last = row.customer_last_name || ''
  if (!first && !last) {
    return '-'
  }
  return `${last} ${first}`.trim()
}

const isLocalStorage = (storage) => {
  if (!storage) return true
  return String(storage).toLowerCase() === 'local'
}

const appendTokenIfNeeded = (url, storage) => {
  if (!url) return ''
  if (isLocalStorage(storage)) {
    const token = authStore.token
    if (token) {
      const separator = url.includes('?') ? '&' : '?'
      return `${url}${separator}token=${encodeURIComponent(token)}`
    }
  }
  return url
}

const getMaterialImageURL = (row) => appendTokenIfNeeded(row.material_image_url, row.material_storage)

const getMaterialPreviewList = (row) => {
  const url = getMaterialImageURL(row)
  return url ? [url] : []
}

const getShippingLabelURL = (row) => appendTokenIfNeeded(row.shipping_label_url, row.shipping_label_storage)

const activeTab = ref('platform')
const loading = ref(false)
const platformOrders = ref([])
const factoryOrders = ref([])
const platformPage = ref(1)
const platformPageSize = ref(10)
const platformTotal = ref(0)
const factoryPage = ref(1)
const factoryPageSize = ref(10)
const factoryTotal = ref(0)
const detailVisible = ref(false)
const detailOrder = ref(null)

const uploadDialogVisible = ref(false)
const uploadFormRef = ref(null)
const uploadTarget = ref({ orderId: null, fileType: 'material_image', orderNo: '' })
const uploadFileList = ref([])
const uploading = ref(false)

const materialSelectorVisible = ref(false)
const materialSelectorLoading = ref(false)
const materialSelectorKeyword = ref('')
const materialSelectorList = ref([])
const materialSelectorPage = ref(1)
const materialSelectorPageSize = ref(10)
const materialSelectorTotal = ref(0)
const selectedMaterial = ref(null)
const materialSelectorOrder = ref(null)
const materialSelectorTableRef = ref(null)

const getMaterialSelectorPreviewURL = (row) => {
  if (!row) return ''
  let url = row.preview_url || row.download_url
  if (!url && row.id) {
    url = `/api/materials/${row.id}/download?disposition=inline`
  }
  return appendTokenIfNeeded(url, row.storage)
}

const getMaterialSelectorPreviewList = (row) => {
  const url = getMaterialSelectorPreviewURL(row)
  return url ? [url] : []
}

const batchUploadDialogVisible = ref(false)
const batchUploadFileList = ref([])
const batchUploading = ref(false)
const batchUploadResult = ref({ success: [], failed: [] })
const batchUploadAccept = '.pdf,image/*'

const uploadDialogTitle = computed(() =>
  uploadTarget.value.fileType === 'shipping_label'
    ? '上传面单文件'
    : uploadTarget.value.fileType === 'import'
      ? '导入订单 Excel'
      : '上传素材图'
)
const uploadAccept = computed(() =>
  uploadTarget.value.fileType === 'shipping_label'
    ? 'application/pdf'
    : uploadTarget.value.fileType === 'import'
      ? '.xlsx,.xls'
      : 'image/*'
)
const uploadTip = computed(() =>
  uploadTarget.value.fileType === 'shipping_label'
    ? '仅支持 PDF 格式文件'
    : uploadTarget.value.fileType === 'import'
      ? '请上传 Excel 文件（.xlsx）'
      : '支持常见图片格式（JPG/PNG）'
)

const orderTypeOptions = [
  { label: '平台面单', value: 'platform' },
  { label: '工厂物流', value: 'factory' }
]

const orderStatusOptions = ref([])
const orderStatusMap = computed(() => {
  const map = {}
  orderStatusOptions.value.forEach((item) => {
    const key = String(item.value)
    map[key] = item.label
  })
  return map
})

const getOrderStatusLabel = (value) => {
  if (value === null || value === undefined || value === '') {
    return ''
  }
  const key = String(value)
  return orderStatusMap.value[key] || ''
}

const getFieldOptions = (field) => {
  if (field.optionsKey === 'orderType') {
    return orderTypeOptions
  }
  if (field.optionsKey === 'status') {
    return orderStatusOptions.value
  }
  return field.options || []
}

const fetchOrderStatusOptions = async () => {
  try {
    const response = await api.get('/dict/order_status')
    const items = response.data?.items || []
    orderStatusOptions.value = items.map((item) => {
      const numeric = Number(item.value)
      const parsedValue = Number.isNaN(numeric) ? item.value : numeric
      return {
        label: item.label,
        value: parsedValue
      }
    })
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '加载订单状态失败')
  }
}

const ORDER_FIELD_DEFAULTS = {
  gsp_order_no: '',
  order_type: 'platform',
  order_created_at: '',
  status: 0,
  payment_time: '',
  completed_at: '',
  required_sign_at: '',
  shipping_warehouse_code: '',
  shop_code: '',
  product_id: '',
  owner_name: '',
  product_name: '',
  spec: '',
  item_no: '',
  seller_sku: '',
  platform_sku: '',
  platform_skc: '',
  platform_spu: '',
  product_price: null,
  expected_revenue: null,
  special_product_note: '',
  currency_code: '',
  expected_fulfillment_qty: null,
  item_count: null,
  postal_code: '',
  country: '',
  province: '',
  city: '',
  district: '',
  address_line1: '',
  address_line2: '',
  customer_full_name: '',
  customer_last_name: '',
  customer_first_name: '',
  phone_number: '',
  email: '',
  tax_number: ''
}

const ORDER_FIELD_KEYS = Object.keys(ORDER_FIELD_DEFAULTS)
const intFields = ['status', 'expected_fulfillment_qty', 'item_count']
const floatFields = ['product_price', 'expected_revenue']
const timeFields = ['order_created_at', 'payment_time', 'completed_at', 'required_sign_at']

const editableFieldGroups = [
  {
    label: '基础信息',
    fields: [
      { prop: 'gsp_order_no', label: 'GSP订单号', span: 12 },
      { prop: 'order_type', label: '订单类型', span: 12, type: 'select', optionsKey: 'orderType' },
      { prop: 'order_created_at', label: '订单创建时间', span: 12 },
      { prop: 'status', label: '状态', span: 12, type: 'select', optionsKey: 'status' },
      { prop: 'payment_time', label: '支付时间', span: 12 },
      { prop: 'completed_at', label: '完成时间', span: 12 },
      { prop: 'required_sign_at', label: '要求签收时间', span: 12 },
      { prop: 'shipping_warehouse_code', label: '发货仓库', span: 12 },
      { prop: 'shop_code', label: '店铺编号', span: 12 },
      { prop: 'product_id', label: '商品ID', span: 12 },
      { prop: 'owner_name', label: '负责人', span: 12 }
    ]
  },
  {
    label: '商品信息',
    fields: [
      { prop: 'product_name', label: '商品名称', span: 12 },
      { prop: 'spec', label: '规格', span: 12 },
      { prop: 'item_no', label: '货号', span: 12 },
      { prop: 'seller_sku', label: '卖家SKU', span: 12 },
      { prop: 'platform_sku', label: '平台SKU', span: 12 },
      { prop: 'platform_skc', label: '平台SKC', span: 12 },
      { prop: 'platform_spu', label: '平台SPU', span: 12 },
      { prop: 'product_price', label: '商品价格', span: 12, type: 'float' },
      { prop: 'expected_revenue', label: '商品预计收入', span: 12, type: 'float' },
      { prop: 'expected_fulfillment_qty', label: '应履约件数', span: 12, type: 'int' },
      { prop: 'item_count', label: '件数', span: 12, type: 'int', optional: true },
      { prop: 'currency_code', label: '币种', span: 12 },
      { prop: 'special_product_note', label: '特殊产品备注', span: 24, type: 'textarea', optional: true }
    ]
  },
  {
    label: '收货信息',
    fields: [
      { prop: 'postal_code', label: '邮编', span: 12 },
      { prop: 'country', label: '国家', span: 12 },
      { prop: 'province', label: '省份', span: 12 },
      { prop: 'city', label: '城市', span: 12 },
      { prop: 'district', label: '区', span: 12 },
      { prop: 'address_line1', label: '用户地址1', span: 24 },
      { prop: 'address_line2', label: '用户地址2', span: 24, optional: true }
    ]
  },
  {
    label: '联系人信息',
    fields: [
      { prop: 'customer_full_name', label: '用户全称', span: 12 },
      { prop: 'customer_last_name', label: '用户姓氏', span: 12 },
      { prop: 'customer_first_name', label: '用户名字', span: 12 },
      { prop: 'phone_number', label: '手机号', span: 12 },
      { prop: 'email', label: '用户邮箱', span: 12 },
      { prop: 'tax_number', label: '税号', span: 12, optional: true }
    ]
  }
]

const editDialogVisible = ref(false)
const editFormRef = ref(null)
const editLoading = ref(false)
const editingOrderId = ref(null)
const editForm = reactive({ ...ORDER_FIELD_DEFAULTS })

const openShippingLabel = (row) => {
  const url = getShippingLabelURL(row)
  if (!row?.shipping_label_url || !url) {
    ElMessage.warning('暂无面单文件，可点击 + 上传')
    return
  }
  window.open(url, '_blank')
}

const handleUploadAttachment = (row, type) => {
  if (type === 'material_image') {
    openMaterialSelector(row)
    return
  }
  uploadTarget.value = {
    orderId: row.id,
    fileType: type,
    orderNo: row.gsp_order_no || ''
  }
  uploadFileList.value = []
  uploadDialogVisible.value = true
}

const handleUploadChange = (file, fileList) => {
  if (fileList.length > 1) {
    fileList.splice(0, fileList.length - 1)
  }
  uploadFileList.value = fileList
}

const handleUploadRemove = (file, fileList) => {
  uploadFileList.value = fileList
}

const handleUploadDialogClosed = () => {
  uploadFileList.value = []
  uploadTarget.value = { orderId: null, fileType: 'material_image', orderNo: '' }
}

const submitUpload = async () => {
  if (!uploadFileList.value.length) {
    ElMessage.warning('请选择要上传的文件')
    return
  }
  if (uploadTarget.value.fileType !== 'import' && !uploadTarget.value.orderId) {
    ElMessage.error('未找到订单信息，无法上传')
    return
  }
  const formData = new FormData()
  formData.append('file', uploadFileList.value[0].raw)
  if (uploadTarget.value.fileType !== 'import') {
    formData.append('file_type', uploadTarget.value.fileType)
  }

  uploading.value = true
  try {
    if (uploadTarget.value.fileType === 'import') {
      const response = await api.post('/orders/import', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      ElMessage.success(response.data?.message || '导入成功')
      uploadDialogVisible.value = false
      const currentTab = activeTab.value
      await fetchOrders(currentTab)
      if (currentTab === 'platform') {
        await fetchOrders('factory', { setLoading: false })
      } else {
        await fetchOrders('platform', { setLoading: false })
      }
    } else {
      await api.post(`/orders/${uploadTarget.value.orderId}/attachments`, formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      ElMessage.success('上传成功')
      uploadDialogVisible.value = false
      const currentTab = activeTab.value
      await fetchOrders(currentTab)
      if (currentTab === 'platform') {
        await fetchOrders('factory', { setLoading: false })
      } else {
        await fetchOrders('platform', { setLoading: false })
      }
    }
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '上传失败')
  } finally {
    uploading.value = false
  }
}


const openBatchUploadDialog = () => {
  batchUploadDialogVisible.value = true
  batchUploadFileList.value = []
  batchUploadResult.value = { success: [], failed: [] }
}

const handleBatchUploadChange = (file, fileList) => {
  batchUploadFileList.value = [...fileList]
}

const handleBatchUploadRemove = (file, fileList) => {
  batchUploadFileList.value = [...fileList]
}

const handleBatchUploadDialogClosed = () => {
  batchUploadFileList.value = []
  batchUploadResult.value = { success: [], failed: [] }
}

const submitBatchUpload = async () => {
  if (!batchUploadFileList.value.length) {
    ElMessage.warning('请选择要上传的文件')
    return
  }
  const formData = new FormData()
  batchUploadFileList.value.forEach((item) => {
    if (item?.raw) {
      formData.append('files', item.raw)
    }
  })
  if (!formData.has('files')) {
    ElMessage.warning('没有有效的文件可上传')
    return
  }

  batchUploading.value = true
  try {
    const response = await api.post('/orders/batch-attachments', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    const data = response.data?.data || {}
    batchUploadResult.value = {
      success: data.success || [],
      failed: data.failed || []
    }
    const failedNames = batchUploadResult.value.failed.map((item) => item.file_name).join(', ')
    let message =
      response.data?.message ||
      `成功 ${batchUploadResult.value.success.length} 个，失败 ${batchUploadResult.value.failed.length} 个`
    if (batchUploadResult.value.failed.length) {
      message = `${message}。未匹配的文件：${failedNames}`
    }
    ElMessage.success(message)
    const currentTab = activeTab.value
    await fetchOrders(currentTab)
    if (currentTab === 'platform') {
      await fetchOrders('factory', { setLoading: false })
    } else {
      await fetchOrders('platform', { setLoading: false })
    }
  } catch (error) {
    const data = error.response?.data?.data
    if (data) {
      batchUploadResult.value = {
        success: data.success || [],
        failed: data.failed || []
      }
    }
    const failedNames = batchUploadResult.value.failed.map((item) => item.file_name)
    let message = error.response?.data?.error || '批量上传失败'
    if (failedNames.length) {
      message = `${message}。未匹配的文件：${failedNames.join(', ')}`
    }
    ElMessage.error(message)
  } finally {
    batchUploading.value = false
  }
}

const handleRemoveAttachment = async (row, type) => {
  const attachmentId =
    type === 'material_image' ? row.material_attachment_id : row.shipping_label_attachment_id
  if (!attachmentId) {
    ElMessage.warning('暂无可删除的附件')
    return
  }
  try {
    await ElMessageBox.confirm('确认删除该附件吗？', '提示', { type: 'warning' })
    await api.delete(`/orders/${row.id}/attachments/${attachmentId}`)
    ElMessage.success('删除成功')
    const currentTab = activeTab.value
    await fetchOrders(currentTab)
    if (currentTab === 'platform') {
      await fetchOrders('factory', { setLoading: false })
    } else {
      await fetchOrders('platform', { setLoading: false })
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '删除失败')
    }
  }
}

const resetEditForm = () => {
  Object.assign(editForm, { ...ORDER_FIELD_DEFAULTS })
}

const populateEditForm = (row) => {
  ORDER_FIELD_KEYS.forEach((key) => {
    const value = row?.[key]
    if (timeFields.includes(key)) {
      editForm[key] = value ? formatBeijingTime(value) : ''
    } else if (floatFields.includes(key) || intFields.includes(key)) {
      const numericValue = value === null || value === undefined ? null : Number(value)
      editForm[key] = Number.isNaN(numericValue) ? null : numericValue
    } else {
      editForm[key] = value === null || value === undefined ? '' : String(value)
    }
  })
  if (!editForm.order_type) {
    editForm.order_type = row?.order_type || (activeTab.value === 'factory' ? 'factory' : 'platform')
  }
  if (editForm.status === null || Number.isNaN(editForm.status)) {
    editForm.status = typeof row?.status === 'number' ? row.status : 0
  }
}

const toNullableString = (value) => {
  if (value === null || value === undefined) return null
  const str = String(value).trim()
  return str === '' ? null : str
}

const buildOrderPayload = () => {
  return {
    gsp_order_no: editForm.gsp_order_no.trim(),
    order_type: editForm.order_type,
    order_created_at: editForm.order_created_at.trim(),
    status: editForm.status === null ? null : Math.trunc(editForm.status),
    payment_time: toNullableString(editForm.payment_time),
    completed_at: toNullableString(editForm.completed_at),
    shipping_warehouse_code: editForm.shipping_warehouse_code.trim(),
    required_sign_at: toNullableString(editForm.required_sign_at),
    shop_code: editForm.shop_code.trim(),
    product_id: editForm.product_id.trim(),
    owner_name: editForm.owner_name.trim(),
    product_name: editForm.product_name.trim(),
    spec: editForm.spec.trim(),
    item_no: editForm.item_no.trim(),
    seller_sku: editForm.seller_sku.trim(),
    platform_sku: editForm.platform_sku.trim(),
    platform_skc: editForm.platform_skc.trim(),
    platform_spu: editForm.platform_spu.trim(),
    product_price: editForm.product_price === null ? null : Number(editForm.product_price),
    expected_revenue: editForm.expected_revenue === null ? null : Number(editForm.expected_revenue),
    special_product_note: toNullableString(editForm.special_product_note),
    currency_code: editForm.currency_code.trim(),
    expected_fulfillment_qty:
      editForm.expected_fulfillment_qty === null
        ? null
        : Math.trunc(editForm.expected_fulfillment_qty),
    item_count:
      editForm.item_count === null || editForm.item_count === ''
        ? null
        : Math.trunc(editForm.item_count),
    postal_code: editForm.postal_code.trim(),
    country: editForm.country.trim(),
    province: editForm.province.trim(),
    city: editForm.city.trim(),
    district: editForm.district.trim(),
    address_line1: editForm.address_line1.trim(),
    address_line2: editForm.address_line2.trim(),
    customer_full_name: editForm.customer_full_name.trim(),
    customer_last_name: editForm.customer_last_name.trim(),
    customer_first_name: editForm.customer_first_name.trim(),
    phone_number: editForm.phone_number.trim(),
    email: editForm.email.trim(),
    tax_number: toNullableString(editForm.tax_number)
  }
}

const handleEditOrder = (row) => {
  resetEditForm()
  populateEditForm(row)
  editingOrderId.value = row.id
  editDialogVisible.value = true
}

const handleEditDialogClosed = () => {
  resetEditForm()
  editingOrderId.value = null
}

const submitEdit = async () => {
  if (!editingOrderId.value) return

  const payload = buildOrderPayload()
  if (!payload.order_type) {
    payload.order_type = activeTab.value === 'factory' ? 'factory' : 'platform'
  }

  editLoading.value = true
  try {
    await api.put(`/orders/${editingOrderId.value}`, payload)
    ElMessage.success('更新成功')
    editDialogVisible.value = false
    const currentTab = activeTab.value
    await fetchOrders(currentTab)
    if (currentTab === 'platform') {
      await fetchOrders('factory', { setLoading: false })
    } else {
      await fetchOrders('platform', { setLoading: false })
    }
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '更新失败')
  } finally {
    editLoading.value = false
  }
}

const handleDeleteOrder = async (row) => {
  try {
    await ElMessageBox.confirm('确认删除该订单吗？删除后无法恢复。', '提示', {
      type: 'warning'
    })
    await api.delete(`/orders/${row.id}`)
    ElMessage.success('删除成功')
    const currentTab = activeTab.value
    await fetchOrders(currentTab)
    if (currentTab === 'platform') {
      await fetchOrders('factory', { setLoading: false })
    } else {
      await fetchOrders('platform', { setLoading: false })
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '删除失败')
    }
  }
}

const extractFileName = (url) => {
  if (!url) return '-'
  try {
    const decoded = decodeURIComponent(url.split('?')[0])
    return decoded.substring(decoded.lastIndexOf('/') + 1)
  } catch (err) {
    return url
  }
}

const detailFields = [
  { label: 'GSP订单号', prop: 'gsp_order_no' },
  { label: '订单创建时间', prop: 'order_created_at', type: 'time' },
  { label: '订单状态', prop: 'status', type: 'statusTag' },
  { label: '发货仓库', prop: 'shipping_warehouse_code' },
  { label: '店铺编号', prop: 'shop_code' },
  { label: '负责人', prop: 'owner_name' },
  { label: '商品名称', prop: 'product_name' },
  { label: '规格', prop: 'spec' },
  { label: '货号', prop: 'item_no' },
  { label: '卖家SKU', prop: 'seller_sku' },
  { label: '平台SKU', prop: 'platform_sku' },
  { label: '平台SKC', prop: 'platform_skc' },
  { label: '平台SPU', prop: 'platform_spu' },
  { label: '商品价格', prop: 'product_price', type: 'currency' },
  { label: '商品预计收入', prop: 'expected_revenue', type: 'currency' },
  { label: '特殊产品备注', prop: 'special_product_note' },
  { label: '币种', prop: 'currency_code' },
  { label: '应履约件数', prop: 'expected_fulfillment_qty' },
  { label: '邮编', prop: 'postal_code' },
  { label: '国家', prop: 'country' },
  { label: '省份', prop: 'province' },
  { label: '城市', prop: 'city' },
  { label: '区', prop: 'district' },
  { label: '用户地址1', prop: 'address_line1' },
  { label: '用户地址2', prop: 'address_line2' },
  { label: '用户全称', prop: 'customer_full_name' },
  { label: '用户姓氏', prop: 'customer_last_name' },
  { label: '用户名字', prop: 'customer_first_name' },
  { label: '手机号', prop: 'phone_number' },
  { label: '用户邮箱', prop: 'email' },
  { label: '税号', prop: 'tax_number' },
  { label: '支付时间', prop: 'payment_time', type: 'time' },
  { label: '完成时间', prop: 'completed_at', type: 'time' },
  { label: '要求签收时间', prop: 'required_sign_at', type: 'time' }
]

const formatDetailField = (field, order) => {
  const value = order[field.prop]
  if (value === null || value === undefined || value === '') {
    return '-'
  }
  if (field.type === 'statusTag') {
    const label = getOrderStatusLabel(value)
    return label || value
  }
  if (field.type === 'currency') {
    return formatCurrency(value)
  }
  if (field.type === 'time') {
    return formatTime(value)
  }
  return value
}

const fetchOrders = async (tab = activeTab.value, options = {}) => {
  const setLoading = options.setLoading !== false
  if (setLoading) {
    loading.value = true
  }
  try {
    const isFactory = tab === 'factory'
    const params = { tab }
    params.order_type = isFactory ? 'factory' : 'platform'
    params.page = isFactory ? factoryPage.value : platformPage.value
    params.page_size = isFactory ? factoryPageSize.value : platformPageSize.value
    if (filterForm.timeField && filterForm.timeRange.length === 2) {
      params.time_field = filterForm.timeField
      params.time_start = filterForm.timeRange[0]
      params.time_end = filterForm.timeRange[1]
    }
    if (filterForm.exactField && filterForm.exactValue !== '' && filterForm.exactValue !== null) {
      params.exact_field = filterForm.exactField
      params.exact_value = filterForm.exactValue
    }
    if (filterForm.fuzzyField && filterForm.fuzzyKeyword) {
      params.fuzzy_field = filterForm.fuzzyField
      params.fuzzy_keyword = filterForm.fuzzyKeyword
    }

    const response = await api.get('/orders', { params })
    const data = response.data?.data || {}
    const list = data.orders || []
    const total = Number(data.total || 0)
    if (isFactory) {
      factoryOrders.value = list
      factoryTotal.value = total
      if (typeof data.page === 'number') {
        factoryPage.value = data.page
      }
      if (typeof data.page_size === 'number') {
        factoryPageSize.value = data.page_size
      }
    } else {
      platformOrders.value = list
      platformTotal.value = total
      if (typeof data.page === 'number') {
        platformPage.value = data.page
      }
      if (typeof data.page_size === 'number') {
        platformPageSize.value = data.page_size
      }
    }
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '加载订单数据失败')
  } finally {
    if (setLoading) {
      loading.value = false
    }
  }
}

const handlePlatformPageChange = (page) => {
  platformPage.value = page
  if (activeTab.value === 'platform') {
    fetchOrders('platform')
  }
}

const handlePlatformSizeChange = (size) => {
  platformPageSize.value = size
  platformPage.value = 1
  if (activeTab.value === 'platform') {
    fetchOrders('platform')
  }
}

const handleFactoryPageChange = (page) => {
  factoryPage.value = page
  if (activeTab.value === 'factory') {
    fetchOrders('factory')
  }
}

const handleFactorySizeChange = (size) => {
  factoryPageSize.value = size
  factoryPage.value = 1
  if (activeTab.value === 'factory') {
    fetchOrders('factory')
  }
}

const handleTabChange = () => {
  fetchOrders(activeTab.value)
}

const handleSearch = () => {
  if (activeTab.value === 'factory') {
    factoryPage.value = 1
    fetchOrders('factory')
  } else {
    platformPage.value = 1
    fetchOrders('platform')
  }
}

const handleReset = () => {
  filterForm.timeField = ''
  filterForm.timeRange = []
  filterForm.exactField = ''
  filterForm.exactValue = ''
  filterForm.fuzzyField = ''
  filterForm.fuzzyKeyword = ''
  platformPage.value = 1
  factoryPage.value = 1
  fetchOrders(activeTab.value)
}

const handleImport = () => {
  uploadTarget.value = { orderId: null, fileType: 'import', orderNo: '' }
  uploadDialogVisible.value = true
}

const handleExport = async () => {
  try {
    const params = {}
    if (filterForm.timeField && filterForm.timeRange.length === 2) {
      params.time_field = filterForm.timeField
      params.time_start = filterForm.timeRange[0]
      params.time_end = filterForm.timeRange[1]
    }
    if (filterForm.exactField && filterForm.exactValue !== '' && filterForm.exactValue !== null) {
      params.exact_field = filterForm.exactField
      params.exact_value = filterForm.exactValue
    }
    if (filterForm.fuzzyField && filterForm.fuzzyKeyword) {
      params.fuzzy_field = filterForm.fuzzyField
      params.fuzzy_keyword = filterForm.fuzzyKeyword
    }

    const response = await api.get('/orders/export', {
      params,
      responseType: 'blob'
    })
    const blob = new Blob([response.data], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    const fileName = `orders_all_${Date.now()}.xlsx`
    link.setAttribute('download', fileName)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '导出失败')
  }
}

const handleViewDetail = (row) => {
  detailOrder.value = { ...row }
  detailVisible.value = true
}

const openMaterialSelector = (row) => {
  materialSelectorOrder.value = row
  selectedMaterial.value = null
  materialSelectorVisible.value = true
  materialSelectorKeyword.value = row.item_no || ''
  materialSelectorPage.value = 1
  fetchMaterialSelectorList()
}

const fetchMaterialSelectorList = async () => {
  materialSelectorLoading.value = true
  try {
    const params = {
      page: materialSelectorPage.value,
      page_size: materialSelectorPageSize.value
    }
    if (materialSelectorKeyword.value) {
      params.keyword = materialSelectorKeyword.value
    }
    const response = await api.get('/materials', { params })
    const payload = response.data?.data || {}
    materialSelectorList.value = payload.materials || []
    materialSelectorTotal.value = payload.total || 0
    await nextTick(() => {
      const table = materialSelectorTableRef.value
      if (!table) return
      table.setCurrentRow()
      const currentAssetId = materialSelectorOrder.value?.material_asset_id
      if (currentAssetId) {
        const target = materialSelectorList.value.find((item) => item.id === currentAssetId)
        if (target) {
          selectedMaterial.value = target
          table.setCurrentRow(target)
        }
      }
    })
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '加载素材失败')
  } finally {
    materialSelectorLoading.value = false
  }
}

const handleMaterialSelectorSearch = () => {
  materialSelectorPage.value = 1
  fetchMaterialSelectorList()
}

const handleMaterialSelectorPageChange = (page) => {
  materialSelectorPage.value = page
  fetchMaterialSelectorList()
}

const handleMaterialSelectorSizeChange = (size) => {
  materialSelectorPageSize.value = size
  materialSelectorPage.value = 1
  fetchMaterialSelectorList()
}

const handleMaterialCurrentChange = (row) => {
  selectedMaterial.value = row
}

const handleMaterialSelectorClosed = () => {
  materialSelectorOrder.value = null
  selectedMaterial.value = null
  materialSelectorKeyword.value = ''
  materialSelectorList.value = []
}

const submitMaterialSelection = async () => {
  if (!materialSelectorOrder.value?.id) {
    ElMessage.error('未找到订单信息')
    return
  }
  if (!selectedMaterial.value) {
    ElMessage.warning('请选择素材')
    return
  }
  try {
    await api.post(`/orders/${materialSelectorOrder.value.id}/attachments/link`, {
      material_id: selectedMaterial.value.id,
      file_type: 'material_image'
    })
    ElMessage.success('已关联素材')
    materialSelectorVisible.value = false
    const currentTab = activeTab.value
    await fetchOrders(currentTab)
    if (currentTab === 'platform') {
      await fetchOrders('factory', { setLoading: false })
    } else {
      await fetchOrders('platform', { setLoading: false })
    }
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '关联素材失败')
  }
}

onMounted(async () => {
  await fetchOrderStatusOptions()
  fetchOrders()
})
</script>

<style scoped>
.orders-page {
  padding: 20px;
  background: #fff;
  min-height: 100%;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.filter-form {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  margin-bottom: 12px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}

.filter-form :deep(.el-form-item) {
  margin-right: 24px;
  margin-bottom: 10px;
}

.upload-button {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
}

.upload-button.ghost {
  color: var(--el-color-primary);
}

.upload-button.ghost:hover {
  color: var(--el-color-primary-dark-2);
  background: rgba(64, 158, 255, 0.16);
}

.operation-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.order-edit-form {
  max-height: 55vh;
  overflow-y: auto;
  padding-right: 8px;
}

.order-edit-dialog :deep(.el-divider) {
  margin: 16px 0;
}

.filter-actions {
  margin-left: auto;
}

.filter-actions :deep(.el-form-item__content) {
  display: flex;
  align-items: center;
  gap: 12px;
}

.table-container {
  max-height: calc(100vh - 400px);
  overflow-y: auto;
  border: 1px solid #ebeef5;
  border-radius: 4px;
}

.order-tabs :deep(.el-table) {
  font-size: 13px;
}

.order-tabs :deep(.el-table th),
.order-tabs :deep(.el-table td) {
  padding: 8px 12px;
}

.cell-title {
  font-weight: 600;
  margin-bottom: 4px;
}

.cell-sub {
  color: #666;
  font-size: 12px;
  line-height: 1.5;
}

.material-cell,
.shipping-label-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.material-cell .el-image {
  width: 60px;
  height: 60px;
  border-radius: 6px;
  object-fit: cover;
}

.material-delete {
  margin-left: 4px;
}

.table-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
  padding: 12px 0;
}

.order-upload {
  width: 100%;
}

.order-upload :deep(.el-upload-dragger) {
  padding: 24px;
}

.upload-form--import :deep(.el-form-item__label) {
  font-weight: 600;
  padding-bottom: 8px;
  width: 100%;
  text-align: left;
}

.upload-form--import :deep(.el-form-item__content) {
  margin-left: 0 !important;
  width: 100%;
}

.upload-form--import :deep(.upload-block) {
  width: 100%;
}

.upload-form--import :deep(.upload-block .el-upload) {
  width: 100%;
}

.upload-form--import :deep(.upload-block .el-upload-dragger) {
  width: 100%;
  min-height: 100px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  padding: 10px;
}

.upload-tip {
  margin-bottom: 12px;
  color: #666;
}

.upload-icon {
  font-size: 40px;
  color: #409eff;
  margin-bottom: 12px;
}
.batch-upload-result {
  margin-top: 16px;
}

.batch-upload-result ul {
  margin: 8px 0 0;
  padding-left: 18px;
}

.batch-upload-result li {
  line-height: 1.6;
}

.material-selector-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.material-selector-toolbar .toolbar-search {
  flex: 1;
}

.material-selector-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  color: #606266;
}

.material-selector-info .selector-code {
  font-weight: 600;
  color: #303133;
}

.material-selector-spec {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  color: #606266;
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

.material-selector-pagination {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}

</style>

