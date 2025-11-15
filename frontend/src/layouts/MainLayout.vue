<template>
  <el-container class="layout-container">
    <el-aside :width="isCollapse ? '64px' : '220px'" class="sidebar">
      <div class="logo">
        <h2 v-if="!isCollapse" :title="siteName">{{ siteName }}</h2>
        <h2 v-else :title="siteName">{{ siteNameAbbr }}</h2>
      </div>
      <div class="menu-tools" v-if="!isCollapse">
        <div class="menu-search">
          <el-input
            v-model="menuSearch"
            placeholder="搜索菜单"
            clearable
            size="default"
            class="menu-search-input"
          >
            <template #prefix>
              <el-icon class="menu-search-icon"><Search /></el-icon>
            </template>
          </el-input>
        </div>
        <div class="quick-menu" v-if="quickMenuItems.length">
          <div class="quick-menu-header" @click="quickMenuCollapsed = !quickMenuCollapsed">
            <div class="quick-menu-title">
              <el-icon><StarFilled /></el-icon>
              <span>快捷菜单</span>
            </div>
            <el-icon :class="['quick-menu-toggle', { expanded: !quickMenuCollapsed }]">
              <ArrowRight />
            </el-icon>
          </div>
          <div class="quick-menu-list" v-show="!quickMenuCollapsed">
            <div
              v-for="item in quickMenuItems"
              :key="item.path"
              :class="['quick-menu-item', { active: activeMenu === item.path }]"
              @click="handleQuickMenuClick(item.path)"
            >
              <el-icon v-if="item.icon">
                <component :is="item.icon" />
              </el-icon>
              <span>{{ item.title }}</span>
            </div>
          </div>
        </div>
      </div>
      <div v-if="filteredMenuTree.length === 0" class="menu-empty">
        暂无匹配菜单
      </div>
      <el-menu
        v-else
        :default-active="activeMenu"
        :collapse="isCollapse"
        router
        class="sidebar-menu"
      >
        <template v-for="menu in filteredMenuTree" :key="menu.path">
          <el-sub-menu v-if="menu.children && menu.children.length" :index="menu.path">
            <template #title>
              <el-icon v-if="menu.icon">
                <component :is="menu.icon" />
              </el-icon>
              <span>{{ menu.title }}</span>
            </template>
            <el-menu-item
              v-for="child in menu.children"
              :key="child.path"
              :index="child.path"
            >
              <el-icon v-if="child.icon">
                <component :is="child.icon" />
              </el-icon>
              <span>{{ child.title }}</span>
            </el-menu-item>
          </el-sub-menu>
          <el-menu-item v-else :index="menu.path">
            <el-icon v-if="menu.icon">
              <component :is="menu.icon" />
            </el-icon>
            <span>{{ menu.title }}</span>
          </el-menu-item>
        </template>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <!-- 标签页区域 -->
        <div class="tabs-container">
          <!-- 折叠按钮 - 在标签页最左边 -->
          <div class="collapse-btn" @click="toggleCollapse">
            <el-icon><Expand v-if="isCollapse" /><Fold v-else /></el-icon>
          </div>
          
          <!-- 标签页列表 -->
          <div class="tabs-wrapper">
            <div
              v-for="tab in tabs"
              :key="tab.path"
              :class="['tab-item', { active: tab.path === activeTab }]"
              @click="switchTab(tab.path)"
            >
              <span class="tab-title">{{ tab.title }}</span>
              <el-icon
                class="tab-close"
                @click.stop="closeTab(tab.path)"
                v-if="tabs.length > 1"
              >
                <Close />
              </el-icon>
            </div>
          </div>
          
          <!-- 分屏控制按钮 -->
          <div class="split-controls" v-if="tabs.length > 1">
            <el-dropdown @command="handleSplitCommand" trigger="click">
              <el-button size="default" text class="split-btn">
                <el-icon><Grid /></el-icon>
                分屏
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="1" :disabled="splitMode === 1">单屏</el-dropdown-item>
                  <el-dropdown-item command="2" :disabled="splitMode === 2">2分屏</el-dropdown-item>
                  <el-dropdown-item command="3" :disabled="splitMode === 3">3分屏</el-dropdown-item>
                  <el-dropdown-item command="4" :disabled="splitMode === 4">4分屏</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
        
        <!-- 右上角：消息通知和用户信息 -->
        <div class="header-right">
          <!-- 消息公告 -->
          <el-dropdown @command="handleNoticeCommand" trigger="click">
            <el-badge :value="unreadCount" :hidden="unreadCount === 0" class="notice-badge">
              <el-icon style="font-size: 20px; cursor: pointer"><Bell /></el-icon>
            </el-badge>
            <template #dropdown>
              <el-dropdown-menu class="notice-dropdown-menu">
                <div class="notice-dropdown-header">
                  <span class="notice-dropdown-title">消息通知</span>
                </div>
                <el-tabs v-model="noticeTab" @tab-change="handleNoticeTabChange" class="notice-tabs">
                  <el-tab-pane label="未读" name="unread">
                    <div class="notice-list">
                      <div
                        v-for="notice in unreadNotices"
                        :key="notice.id"
                        class="notice-item"
                        @click="handleReadNotice(notice)"
                      >
                        <div class="notice-title">{{ notice.title }}</div>
                        <div class="notice-time">{{ formatTime(notice.created_at) }}</div>
                      </div>
                      <div v-if="unreadNotices.length === 0" class="notice-empty">暂无未读消息</div>
                    </div>
                  </el-tab-pane>
                  <el-tab-pane label="已读" name="read">
                    <div class="notice-list">
                      <div
                        v-for="notice in readNotices"
                        :key="notice.id"
                        class="notice-item read"
                      >
                        <div class="notice-title">{{ notice.title }}</div>
                        <div class="notice-time">{{ formatTime(notice.created_at) }}</div>
                      </div>
                      <div v-if="readNotices.length === 0" class="notice-empty">暂无已读消息</div>
                    </div>
                  </el-tab-pane>
                </el-tabs>
                <el-dropdown-item divided>
                  <el-button text @click="router.push('/notices')" style="width: 100%">查看全部</el-button>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-icon><User /></el-icon>
              <div class="user-info-text">
                <span class="user-info-name">{{ authStore.user?.username }}</span>
                <span class="user-info-meta">{{ departmentName }} · {{ roleName }} · {{ employeeTypeLabel }}</span>
              </div>
              <el-icon class="el-icon--right"><arrow-down /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu class="user-dropdown-menu">
                <el-dropdown-item command="changePassword">修改密码</el-dropdown-item>
                <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="main-content">
        <!-- 单屏模式 -->
        <div v-if="splitMode === 1" class="split-container single">
        <router-view />
        </div>
        
        <!-- 2分屏模式 -->
        <div v-else-if="splitMode === 2" class="split-container double">
          <div class="split-panel" v-for="(panel, index) in splitPanels.slice(0, 2)" :key="`${panel.path}-${index}`">
            <div class="split-panel-header">
              <el-dropdown @command="(path) => switchPanel(index, path)" trigger="click">
                <span class="panel-title">
                  {{ panel.title }}
                  <el-icon class="el-icon--right" style="margin-left: 4px; font-size: 16px"><ArrowDown /></el-icon>
                </span>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item
                      v-for="tab in tabs"
                      :key="tab.path"
                      :command="tab.path"
                      :class="{ 'is-active': panel.path === tab.path }"
                    >
                      {{ tab.title }}
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <div class="panel-actions">
                <el-icon class="panel-action-btn" @click="minimizePanel(index)" title="最小化（隐藏当前面板）" v-if="splitPanels.length > 1">
                  <Minus />
                </el-icon>
                <el-icon class="panel-action-btn" @click="maximizePanel(index)" title="最大化（全屏显示）">
                  <FullScreen />
                </el-icon>
                <el-icon class="panel-close" @click="closePanel(index)" title="关闭（关闭面板和标签页）">
                  <Close />
                </el-icon>
              </div>
            </div>
            <div class="split-panel-content">
              <component :is="getComponent(panel.name)" v-if="panel.name" :key="`${panel.path}-${index}-${splitMode}`" />
            </div>
          </div>
        </div>
        
        <!-- 3分屏模式 -->
        <div v-else-if="splitMode === 3" class="split-container triple">
          <div class="split-panel" v-for="(panel, index) in splitPanels.slice(0, 3)" :key="`${panel.path}-${index}`">
            <div class="split-panel-header">
              <el-dropdown @command="(path) => switchPanel(index, path)" trigger="click">
                <span class="panel-title">
                  {{ panel.title }}
                  <el-icon class="el-icon--right" style="margin-left: 4px; font-size: 16px"><ArrowDown /></el-icon>
                </span>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item
                      v-for="tab in tabs"
                      :key="tab.path"
                      :command="tab.path"
                      :class="{ 'is-active': panel.path === tab.path }"
                    >
                      {{ tab.title }}
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <div class="panel-actions">
                <el-icon class="panel-action-btn" @click="minimizePanel(index)" title="最小化（隐藏当前面板）" v-if="splitPanels.length > 1">
                  <Minus />
                </el-icon>
                <el-icon class="panel-action-btn" @click="maximizePanel(index)" title="最大化（全屏显示）">
                  <FullScreen />
                </el-icon>
                <el-icon class="panel-close" @click="closePanel(index)" title="关闭（关闭面板和标签页）">
                  <Close />
                </el-icon>
              </div>
            </div>
            <div class="split-panel-content">
              <component :is="getComponent(panel.name)" v-if="panel.name" :key="`${panel.path}-${index}-${splitMode}`" />
            </div>
          </div>
        </div>
        
        <!-- 4分屏模式 -->
        <div v-else-if="splitMode === 4" class="split-container quadruple">
          <div class="split-panel" v-for="(panel, index) in splitPanels.slice(0, 4)" :key="`${panel.path}-${index}`">
            <div class="split-panel-header">
              <el-dropdown @command="(path) => switchPanel(index, path)" trigger="click">
                <span class="panel-title">
                  {{ panel.title }}
                  <el-icon class="el-icon--right" style="margin-left: 4px; font-size: 16px"><ArrowDown /></el-icon>
                </span>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item
                      v-for="tab in tabs"
                      :key="tab.path"
                      :command="tab.path"
                      :class="{ 'is-active': panel.path === tab.path }"
                    >
                      {{ tab.title }}
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <div class="panel-actions">
                <el-icon class="panel-action-btn" @click="minimizePanel(index)" title="最小化（隐藏当前面板）" v-if="splitPanels.length > 1">
                  <Minus />
                </el-icon>
                <el-icon class="panel-action-btn" @click="maximizePanel(index)" title="最大化（全屏显示）">
                  <FullScreen />
                </el-icon>
                <el-icon class="panel-close" @click="closePanel(index)" title="关闭（关闭面板和标签页）">
                  <Close />
                </el-icon>
              </div>
            </div>
            <div class="split-panel-content">
              <component :is="getComponent(panel.name)" v-if="panel.name" :key="`${panel.path}-${index}-${splitMode}`" />
            </div>
          </div>
        </div>
      </el-main>
    </el-container>
    
    <!-- 修改密码对话框 -->
    <el-dialog v-model="passwordDialogVisible" title="修改密码" width="500px">
      <el-form :model="passwordForm" :rules="passwordRules" ref="passwordFormRef" label-width="100px">
        <el-form-item label="原密码" prop="oldPassword">
          <el-input
            v-model="passwordForm.oldPassword"
            type="password"
            placeholder="请输入原密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input
            v-model="passwordForm.newPassword"
            type="password"
            placeholder="请输入新密码（至少6位）"
            show-password
          />
        </el-form-item>
        <el-form-item label="确认新密码" prop="confirmPassword">
          <el-input
            v-model="passwordForm.confirmPassword"
            type="password"
            placeholder="请再次输入新密码"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitPassword">确定</el-button>
      </template>
    </el-dialog>
  </el-container>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, h, defineAsyncComponent } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/utils/api'
import { formatBeijingTime } from '@/utils/date'
import { useAppConfigStore } from '@/stores/appConfig'
import {
  House,
  User,
  UserFilled,
  Lock,
  Document,
  List,
  DataAnalysis,
  ArrowDown,
  Monitor,
  ArrowRight,
  Setting,
  Tools,
  Expand,
  Fold,
  Bell,
  Close,
  Grid,
  FullScreen,
  Minus,
  Search,
  StarFilled,
  OfficeBuilding,
  Picture
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const appConfigStore = useAppConfigStore()

if (!appConfigStore.siteNameLoaded) {
  appConfigStore.loadSiteName().catch(() => {
    /* store 内已处理默认值 */
  })
}

const activeMenu = computed(() => route.path)
const isCollapse = ref(false)
const noticeTab = ref('unread')
const unreadNotices = ref([])
const readNotices = ref([])
const unreadCount = computed(() => unreadNotices.value.length)

const departmentName = computed(() => authStore.user?.department?.name || '未分配部门')
const roleName = computed(() => authStore.user?.role?.name || '未分配角色')
const employeeTypeLabel = computed(() => {
  const type = authStore.user?.employee_type
  if (type === 'external') {
    return '外部员工'
  }
  if (type === 'internal') {
    return '内部员工'
  }
  return '员工类型未设置'
})

const iconComponentMap = {
  House,
  Setting,
  User,
  UserFilled,
  Lock,
  Document,
  List,
  DataAnalysis,
  Monitor,
  Tools,
  Bell,
  OfficeBuilding,
  Picture,
}

const DEFAULT_MENU_TREE = [
  {
    path: '/dashboard',
    title: '仪表盘',
    icon: House,
    name: 'Dashboard',
    order: 10,
    permission: 'menu:dashboard:view'
  },
  {
    path: '/system',
    title: '系统管理',
    icon: Setting,
    name: 'System',
    order: 20,
    permission: 'menu:system:view',
    children: [
      { path: '/users', title: '用户管理', icon: User, name: 'Users', order: 21, permission: 'menu:users:view' },
      { path: '/departments', title: '部门管理', icon: OfficeBuilding, name: 'Departments', order: 22, permission: 'menu:departments:view' },
      { path: '/materials', title: '素材图库', icon: Picture, name: 'Materials', order: 23, permission: 'menu:materials:view' },
      { path: '/roles', title: '角色管理', icon: UserFilled, name: 'Roles', order: 24, permission: 'menu:roles:view' },
      { path: '/permissions', title: '权限管理', icon: Lock, name: 'Permissions', order: 25, permission: 'menu:permissions:view' },
      { path: '/resources', title: '资源管理', icon: Document, name: 'Resources', order: 26, permission: 'menu:resources:view' },
      { path: '/dicts', title: '系统字典', icon: List, name: 'Dicts', order: 27, permission: 'menu:dicts:view' },
      { path: '/configs', title: '系统参数', icon: Tools, name: 'Configs', order: 28, permission: 'menu:configs:view' }
    ]
  },
  {
    path: '/notices',
    title: '消息公告',
    icon: Bell,
    name: 'Notices',
    order: 30,
    permission: 'menu:notices:view'
  },
  {
    path: '/orders',
    title: '订单管理',
    icon: List,
    name: 'Orders',
    order: 35,
    permission: 'menu:orders:view'
  },
  {
    path: '/logs',
    title: '日志管理',
    icon: List,
    name: 'Logs',
    order: 40,
    permission: 'menu:logs:view'
  },
  {
    path: '/system-monitor',
    title: '系统监控',
    icon: Monitor,
    name: 'SystemMonitor',
    order: 45,
    permission: 'menu:system-monitor:view'
  }
]

const rawMenuResources = ref([])
const menuLoaded = ref(false)

const routeMapByPath = {}
router.getRoutes().forEach((routeRecord) => {
  const normalizedPath = routeRecord.path.startsWith('/') ? routeRecord.path : `/${routeRecord.path}`
  routeMapByPath[normalizedPath] = routeRecord
})

const normalizePath = (path) => {
  if (!path) return ''
  return path.startsWith('/') ? path : `/${path}`
}

const resolveIcon = (iconName) => {
  if (!iconName) return List
  return iconComponentMap[iconName] || List
}

const derivePermissionCode = (resource, routeRecord) => {
  if (routeRecord?.meta?.permission) {
    return routeRecord.meta.permission
  }
  const normalized = normalizePath(resource?.path)
  if (!normalized) return null
  const sanitized = normalized.replace(/^\//, '').replace(/\//g, '-')
  if (!sanitized) return null
  return `menu:${sanitized}:view`
}

const buildMenuTreeFromResources = (nodes = []) => {
  if (!Array.isArray(nodes) || nodes.length === 0) {
    return []
  }

  const result = []

  nodes.forEach((node) => {
    if (!node) return

    const childMenus = buildMenuTreeFromResources(node.children || [])

    if (node.type !== 'menu') {
      result.push(...childMenus)
      return
    }

    const normalizedPath = normalizePath(node.path)
    const routeRecord = normalizedPath ? (routeMapByPath[normalizedPath] ?? routeMapByPath[node.path]) : undefined
    const permissionCode = derivePermissionCode(node, routeRecord)
    const filteredChildren = childMenus.filter(Boolean)
    const hasPermission = !permissionCode || authStore.userPermissions.includes(permissionCode)

    if (!hasPermission && filteredChildren.length === 0) {
      return
    }

    result.push({
      path: normalizedPath || `/menu-${node.id}`,
      title: node.name || routeRecord?.meta?.title || routeRecord?.name || normalizedPath,
      icon: resolveIcon(node.icon),
      name: routeRecord?.name,
      order: typeof node.sort === 'number' ? node.sort : 0,
      permission: permissionCode,
      children: filteredChildren
    })
  })

  return result
    .filter(Boolean)
    .sort((a, b) => {
      const orderA = a.order ?? 0
      const orderB = b.order ?? 0
      if (orderA === orderB) {
        return (a.title || '').localeCompare(b.title || '', 'zh-CN')
      }
      return orderA - orderB
    })
}

const filterMenusByPermissions = (menus = []) => {
  if (!Array.isArray(menus) || menus.length === 0) {
    return []
  }

  return menus
    .map((menu) => {
      const filteredChildren = filterMenusByPermissions(menu.children || [])
      const hasPermission = !menu.permission || authStore.userPermissions.includes(menu.permission)

      if (!hasPermission && filteredChildren.length === 0) {
        return null
      }

      return {
        ...menu,
        children: filteredChildren
      }
    })
    .filter(Boolean)
    .sort((a, b) => {
      const orderA = a.order ?? 0
      const orderB = b.order ?? 0
      if (orderA === orderB) {
        return (a.title || '').localeCompare(b.title || '', 'zh-CN')
      }
      return orderA - orderB
    })
}

const baseMenuTree = computed(() => {
  const dynamicMenus = buildMenuTreeFromResources(rawMenuResources.value)
  if (dynamicMenus.length > 0) {
    return dynamicMenus
  }
  return filterMenusByPermissions(DEFAULT_MENU_TREE)
})

const loadMenuResources = async () => {
  try {
    menuLoaded.value = false
    const response = await api.get('/resources', { params: { tree: true } })
    const data = response.data?.data
    if (Array.isArray(data)) {
      rawMenuResources.value = data
    } else {
      rawMenuResources.value = []
    }
  } catch (error) {
    console.error('加载菜单资源失败', error)
    rawMenuResources.value = []
  } finally {
    menuLoaded.value = true
  }
}

const menuSearch = ref('')
const menuUsage = ref({})
const quickMenuCollapsed = ref(true)

const usageStorageKey = computed(() => {
  const userId = authStore.user?.id
  return userId ? `menu_usage_${userId}` : ''
})

const flattenMenus = (menus) => {
  const result = []
  const traverse = (items) => {
    items.forEach((item) => {
      if (item.children && item.children.length) {
        traverse(item.children)
      } else {
        result.push(item)
      }
    })
  }
  traverse(menus)
  return result
}

const flatMenuList = computed(() => flattenMenus(baseMenuTree.value))

const filterMenusByKeyword = (menus, keyword) => {
  return menus
    .map((menu) => {
      const titleMatch = menu.title.toLowerCase().includes(keyword)
      if (menu.children && menu.children.length) {
        const filteredChildren = filterMenusByKeyword(menu.children, keyword)
        if (titleMatch || filteredChildren.length) {
          return {
            ...menu,
            children: filteredChildren
          }
        }
        return null
      }
      return titleMatch ? { ...menu } : null
    })
    .filter(Boolean)
}

const filteredMenuTree = computed(() => {
  const keyword = menuSearch.value.trim().toLowerCase()
  if (!keyword) {
    return baseMenuTree.value
  }
  return filterMenusByKeyword(baseMenuTree.value, keyword)
})

const quickMenuItems = computed(() => {
  const items = flatMenuList.value.slice()
  items.sort((a, b) => {
    const countA = menuUsage.value[a.path] || 0
    const countB = menuUsage.value[b.path] || 0
    if (countA !== countB) {
      return countB - countA
    }
    return (a.order || 0) - (b.order || 0)
  })
  const used = items.filter((item) => (menuUsage.value[item.path] || 0) > 0)
  const unused = items.filter((item) => !used.includes(item))
  return [...used, ...unused].slice(0, 5)
})

const siteName = computed(() => appConfigStore.siteName || '管理系统')
const siteNameAbbr = computed(() => {
  const name = siteName.value
  if (!name) return '管理'
  if (name.length <= 2) return name
  return name.slice(0, 2)
})

// 修改密码对话框
const passwordDialogVisible = ref(false)
const passwordFormRef = ref(null)
const passwordForm = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const passwordRules = {
  oldPassword: [
    { required: true, message: '请输入原密码', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== passwordForm.value.newPassword) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

// 路由名称和标题的映射
const routeTitleMap = {
  'Dashboard': '仪表盘',
  'Users': '用户管理',
  'Departments': '部门管理',
  'Materials': '素材图库',
  'Roles': '角色管理',
  'Permissions': '权限管理',
  'Resources': '资源管理',
  'Logs': '日志管理',
  'SystemMonitor': '系统监控',
  'Dicts': '系统字典',
  'Configs': '系统参数',
  'Notices': '消息公告',
  'Orders': '订单管理'
}

const documentTitle = computed(() => {
  const baseTitle = siteName.value || '管理系统'
  const currentRouteName = route.name
  const routeLabel = currentRouteName ? routeTitleMap[currentRouteName] : ''
  return routeLabel ? `${routeLabel} - ${baseTitle}` : baseTitle
})

watch(
  documentTitle,
  (title) => {
    appConfigStore.updateDocumentTitle(title)
  },
  { immediate: true }
)

const loadMenuUsage = () => {
  const key = usageStorageKey.value
  if (!key) {
    menuUsage.value = {}
    return
  }
  try {
    const stored = localStorage.getItem(key)
    const parsed = stored ? JSON.parse(stored) : {}
    menuUsage.value = parsed && typeof parsed === 'object' ? parsed : {}
  } catch (error) {
    console.warn('加载菜单使用记录失败', error)
    menuUsage.value = {}
  }
}

const recordMenuUsage = (path) => {
  const key = usageStorageKey.value
  if (!key) return
  if (!flatMenuList.value.some((item) => item.path === path)) {
    return
  }
  const usage = {
    ...menuUsage.value,
    [path]: (menuUsage.value[path] || 0) + 1
  }
  menuUsage.value = usage
  try {
    localStorage.setItem(key, JSON.stringify(usage))
  } catch (error) {
    console.warn('记录菜单使用次数失败', error)
  }
}

const handleQuickMenuClick = (path) => {
  if (path) {
    router.push(path)
  }
}

watch(
  usageStorageKey,
  () => {
    loadMenuUsage()
  },
  { immediate: true }
)

watch(isCollapse, (collapsed) => {
  if (collapsed) {
    menuSearch.value = ''
  }
})

// 标签页管理
const tabs = ref([])
const activeTab = ref('')

// 分屏管理
const splitMode = ref(1) // 1: 单屏, 2: 2分屏, 3: 3分屏
const splitPanels = ref([]) // 分屏面板配置
const maximizedPanel = ref(null) // 最大化的面板索引，null表示没有最大化

// 组件映射 - 使用 defineAsyncComponent
const componentMap = {
  'Dashboard': defineAsyncComponent(() => import('@/views/Dashboard.vue')),
  'Users': defineAsyncComponent(() => import('@/views/Users.vue')),
  'Departments': defineAsyncComponent(() => import('@/views/Departments.vue')),
  'Materials': defineAsyncComponent(() => import('@/views/Materials.vue')),
  'Roles': defineAsyncComponent(() => import('@/views/Roles.vue')),
  'Permissions': defineAsyncComponent(() => import('@/views/Permissions.vue')),
  'Resources': defineAsyncComponent(() => import('@/views/Resources.vue')),
  'Logs': defineAsyncComponent(() => import('@/views/Logs.vue')),
  'SystemMonitor': defineAsyncComponent(() => import('@/views/SystemMonitor.vue')),
  'IPStatistics': defineAsyncComponent(() => import('@/views/IPStatistics.vue')),
  'Dicts': defineAsyncComponent(() => import('@/views/Dicts.vue')),
  'Configs': defineAsyncComponent(() => import('@/views/Configs.vue')),
  'Notices': defineAsyncComponent(() => import('@/views/Notices.vue')),
  'Orders': defineAsyncComponent(() => import('@/views/Orders.vue'))
}

// 获取组件
const getComponent = (name) => {
  if (!name || !componentMap[name]) return null
  return componentMap[name]
}

// 处理分屏命令
const handleSplitCommand = (command) => {
  const mode = parseInt(command)
  splitMode.value = mode
  
  if (mode === 1) {
    // 单屏模式，清空分屏面板
    splitPanels.value = []
  } else {
    // 多屏模式，初始化分屏面板
    updateSplitPanels()
  }
}

// 更新分屏面板
const updateSplitPanels = () => {
  const maxPanels = splitMode.value
  const currentPanels = splitPanels.value.length
  
  // 如果当前面板数少于需要的面板数，添加新面板
  if (currentPanels < maxPanels) {
    for (let i = currentPanels; i < maxPanels; i++) {
      // 查找一个不在当前面板中的标签页
      const availableTab = tabs.value.find(tab => 
        !splitPanels.value.some(panel => panel.path === tab.path)
      )
      if (availableTab) {
        splitPanels.value.push({
          path: availableTab.path,
          name: availableTab.name,
          title: availableTab.title
        })
      } else if (tabs.value[i]) {
        // 如果没有可用的标签页，使用索引对应的标签页
        splitPanels.value.push({
          path: tabs.value[i].path,
          name: tabs.value[i].name,
          title: tabs.value[i].title
        })
      }
    }
  } else if (currentPanels > maxPanels) {
    // 如果当前面板数多于需要的面板数，移除多余面板
    splitPanels.value = splitPanels.value.slice(0, maxPanels)
  }
  
  // 去重：确保没有重复的面板
  const seen = new Set()
  splitPanels.value = splitPanels.value.filter(panel => {
    if (seen.has(panel.path)) {
      return false
    }
    seen.add(panel.path)
    return true
  })
  
  // 如果去重后面板数不足，补充面板
  if (splitPanels.value.length < maxPanels) {
    const needed = maxPanels - splitPanels.value.length
    for (let i = 0; i < needed; i++) {
      const availableTab = tabs.value.find(tab => 
        !splitPanels.value.some(panel => panel.path === tab.path)
      )
      if (availableTab) {
        splitPanels.value.push({
          path: availableTab.path,
          name: availableTab.name,
          title: availableTab.title
        })
      }
    }
  }
}

// 关闭分屏面板（关闭面板并关闭对应的标签页）
const closePanel = (index) => {
  if (splitPanels.value[index]) {
    const panelPath = splitPanels.value[index].path
    // 关闭对应的标签页
    closeTab(panelPath)
    // 移除面板
    splitPanels.value.splice(index, 1)
    
    // 调整分屏模式
    if (splitPanels.value.length === 0) {
      splitMode.value = 1
    } else if (splitMode.value === 4 && splitPanels.value.length === 3) {
      splitMode.value = 3
    } else if (splitMode.value === 3 && splitPanels.value.length === 2) {
      splitMode.value = 2
    } else if (splitMode.value === 2 && splitPanels.value.length === 1) {
      splitMode.value = 1
      splitPanels.value = []
    }
  }
}

// 添加标签页
const addTab = (path, name) => {
  const title = routeTitleMap[name] || name
  const existingTab = tabs.value.find(tab => tab.path === path)
  if (!existingTab) {
    tabs.value.push({ path, name, title })
  }
  activeTab.value = path
  
  // 如果处于分屏模式，智能更新分屏面板
  if (splitMode.value > 1) {
    const tab = { path, name, title }
    const panelIndex = splitPanels.value.findIndex(p => p.path === path)
    
    if (panelIndex === -1) {
      // 如果该标签页不在分屏中
      // 检查是否有空面板（面板数少于最大面板数）
      if (splitPanels.value.length < splitMode.value) {
        // 有空面板，添加到第一个空位置
        splitPanels.value.push(tab)
      } else {
        // 没有空面板，替换第一个面板
        splitPanels.value[0] = tab
      }
    }
    // 如果已经在分屏中，不需要更新
  }
}

// 关闭标签页
const closeTab = (path) => {
  const index = tabs.value.findIndex(tab => tab.path === path)
  if (index > -1) {
    tabs.value.splice(index, 1)
    // 如果关闭的是当前激活的标签页，切换到其他标签页
    if (activeTab.value === path) {
      if (tabs.value.length > 0) {
        const newActiveTab = tabs.value[Math.min(index, tabs.value.length - 1)]
        router.push(newActiveTab.path)
      } else {
        // 如果没有标签页了，跳转到仪表盘
        router.push('/dashboard')
      }
    }
  }
}

// 切换标签页
const switchTab = (path) => {
  router.push(path)
  
  // 在分屏模式下，如果该标签页不在分屏中，则替换第一个面板
  if (splitMode.value > 1) {
    const tab = tabs.value.find(t => t.path === path)
    if (tab) {
      const panelIndex = splitPanels.value.findIndex(p => p.path === path)
      if (panelIndex === -1) {
        // 如果该标签页不在分屏中，替换第一个面板
        if (splitPanels.value.length > 0) {
          splitPanels.value[0] = {
            path: tab.path,
            name: tab.name,
            title: tab.title
          }
        } else {
          // 如果没有面板，添加一个
          splitPanels.value.push({
            path: tab.path,
            name: tab.name,
            title: tab.title
          })
        }
      }
    }
  }
}

// 切换分屏面板显示的标签页
const switchPanel = (panelIndex, path) => {
  const tab = tabs.value.find(t => t.path === path)
  if (tab && splitPanels.value[panelIndex]) {
    splitPanels.value[panelIndex] = {
      path: tab.path,
      name: tab.name,
      title: tab.title
    }
  }
}

// 最大化面板
const maximizePanel = (panelIndex) => {
  if (splitPanels.value[panelIndex]) {
    maximizedPanel.value = panelIndex
    // 切换到单屏模式，显示当前面板
    splitMode.value = 1
    // 更新路由到当前面板
    router.push(splitPanels.value[panelIndex].path)
  }
}

// 最小化面板（隐藏面板但不关闭标签页，自动调整分屏模式）
const minimizePanel = (panelIndex) => {
  if (splitPanels.value.length <= 1) {
    // 如果只有一个面板，最小化后切换到单屏模式
    splitMode.value = 1
    splitPanels.value = []
    maximizedPanel.value = null
    return
  }
  
  // 移除当前面板（不关闭标签页）
  splitPanels.value.splice(panelIndex, 1)
  
  // 调整分屏模式
  if (splitMode.value === 4 && splitPanels.value.length === 3) {
    // 4屏变成3屏
    splitMode.value = 3
  } else if (splitMode.value === 3 && splitPanels.value.length === 2) {
    // 3屏变成2屏
    splitMode.value = 2
  } else if (splitMode.value === 2 && splitPanels.value.length === 1) {
    // 2屏变成1屏（单屏模式）
    splitMode.value = 1
    splitPanels.value = []
  }
  
  maximizedPanel.value = null
}

// 使用时间格式化工具
const formatTime = formatBeijingTime

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

const handleCommand = async (command) => {
  if (command === 'logout') {
    authStore.logout()
    router.push('/login')
  } else if (command === 'changePassword') {
    handleChangePassword()
  }
}

const handleChangePassword = () => {
  // 重置表单
  passwordForm.value = {
    oldPassword: '',
    newPassword: '',
    confirmPassword: ''
  }
  passwordDialogVisible.value = true
  // 清除验证状态
  if (passwordFormRef.value) {
    passwordFormRef.value.clearValidate()
  }
}

const submitPassword = async () => {
  if (!passwordFormRef.value) return
  
  try {
    // 验证表单
    await passwordFormRef.value.validate()
    
    // 提交修改密码请求
    await api.put('/auth/change-password', {
      old_password: passwordForm.value.oldPassword,
      new_password: passwordForm.value.newPassword
    })
    
    ElMessage.success('密码修改成功')
    passwordDialogVisible.value = false
    // 重置表单
    passwordForm.value = {
      oldPassword: '',
      newPassword: '',
      confirmPassword: ''
    }
  } catch (error) {
    if (error !== false) { // 验证失败时 error 为 false
      ElMessage.error(error.response?.data?.error || '密码修改失败')
    }
  }
}

const loadNotices = async () => {
  try {
    const response = await api.get('/notices?page=1&page_size=10')
    const notices = response.data.data || []
    unreadNotices.value = notices.filter(n => n.is_read === 0 || !n.is_read)
    readNotices.value = notices.filter(n => n.is_read === 1)
  } catch (error) {
    console.error('加载消息失败', error)
  }
}

const handleNoticeTabChange = (tab) => {
  noticeTab.value = tab
}

const handleNoticeCommand = (command) => {
  // 处理消息相关命令
}

const handleReadNotice = async (notice) => {
  try {
    await api.put(`/notices/${notice.id}/read`)
    notice.is_read = 1
    unreadNotices.value = unreadNotices.value.filter(n => n.id !== notice.id)
    readNotices.value.unshift(notice)
    router.push(`/notices?id=${notice.id}`)
  } catch (error) {
    // 即使标记失败也跳转
    router.push(`/notices?id=${notice.id}`)
  }
}

// 定时刷新未读消息
let noticeTimer = null

// 确保展开状态下背景色透明
const ensureSubMenuBackground = () => {
  const subMenuTitles = document.querySelectorAll('.sidebar-menu .el-sub-menu.is-opened > .el-sub-menu__title')
  subMenuTitles.forEach((title) => {
    if (!title.matches(':hover')) {
      title.style.backgroundColor = 'transparent'
      title.style.background = 'transparent'
    }
  })
}

// 监听路由变化，自动添加标签页
router.afterEach((to) => {
  if (to.name && to.meta?.requiresAuth !== false) {
    addTab(to.path, to.name)
    recordMenuUsage(to.path)
  }
})

onMounted(() => {
  loadMenuResources()
  // 初始化当前路由的标签页
  if (route.name && route.meta?.requiresAuth !== false) {
    addTab(route.path, route.name)
    recordMenuUsage(route.path)
  }
  
  loadNotices()
  // 每30秒刷新一次未读消息
  noticeTimer = setInterval(() => {
    loadNotices()
  }, 30000)
  
  // 监听菜单展开事件，确保背景色透明
  const menu = document.querySelector('.sidebar-menu')
  if (menu) {
    menu.addEventListener('click', () => {
      setTimeout(ensureSubMenuBackground, 10)
    })
    // 使用 MutationObserver 监听 DOM 变化
    const observer = new MutationObserver(() => {
      ensureSubMenuBackground()
    })
    observer.observe(menu, { attributes: true, subtree: true, attributeFilter: ['class'] })
  }
})

onUnmounted(() => {
  if (noticeTimer) {
    clearInterval(noticeTimer)
  }
})
</script>

<style scoped>
/* 全局样式覆盖 - 确保子菜单箭头颜色 */
.layout-container {
  height: 100vh;
}

.sidebar {
  background-color: #304156;
  color: #fff;
  display: flex;
  flex-direction: column;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #434a56;
}

.logo h2 {
  color: #fff;
  font-size: 18px;
  font-weight: 600;
  padding: 0 12px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.menu-tools {
  padding: 16px 16px 0;
}

.menu-search {
  margin-bottom: 16px;
}

.menu-search-input :deep(.el-input__wrapper) {
  background-color: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.12);
  box-shadow: none;
}

.menu-search-input :deep(.el-input__inner) {
  color: #fff;
}

.menu-search-icon {
  color: #9aa8c3;
}

.quick-menu {
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  padding: 12px 14px;
  margin-bottom: 12px;
}

.quick-menu-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 16px;
  color: #d1ddf5;
  letter-spacing: 1px;
  cursor: pointer;
}

.quick-menu-title {
  display: flex;
  align-items: center;
  gap: 6px;
}

.quick-menu-toggle {
  transition: transform 0.2s ease;
  font-size: 16px;
  color: #d1ddf5;
}

.quick-menu-toggle.expanded {
  transform: rotate(90deg);
}

.quick-menu-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 12px;
}

.quick-menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 8px;
  cursor: pointer;
  color: #e0e7f8;
  font-size: 16px;
  transition: background-color 0.2s ease, transform 0.2s ease;
}

.quick-menu-item:hover {
  background-color: rgba(64, 158, 255, 0.15);
  transform: translateX(2px);
}

.quick-menu-item.active {
  background-color: rgba(64, 158, 255, 0.25);
  color: #fff;
  box-shadow: inset 0 0 0 1px rgba(64, 158, 255, 0.35);
}

.quick-menu-item .el-icon {
  font-size: 16px;
}

.menu-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9aa8c3;
  font-size: 16px;
  letter-spacing: 1px;
  padding: 20px;
}

:deep(.sidebar-menu) {
  border: none;
  background-color: #304156;
  flex: 1;
  overflow-y: auto;
  padding-bottom: 20px;
}

/* 菜单项基础样式 */
:deep(.sidebar-menu .el-menu-item) {
  color: #bfcbd9;
  height: 50px;
  line-height: 50px;
  font-size: 16px;
}

:deep(.sidebar-menu .el-menu-item:hover) {
  background-color: #263445 !important;
  color: #fff;
}

:deep(.sidebar-menu .el-menu-item.is-active) {
  background-color: #409eff !important;
  color: #fff;
}

/* 折叠按钮 - 在标签页最左边 */
.collapse-btn {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border-right: 1px solid #e4e7ed;
  transition: all 0.3s;
  color: #606266;
  flex-shrink: 0;
  background-color: #fff;
}

.collapse-btn:hover {
  background-color: #f5f7fa;
  color: #409eff;
}

.collapse-btn .el-icon {
  font-size: 18px;
  color: #606266;
  transition: all 0.3s;
}

.collapse-btn:hover .el-icon {
  color: #409eff;
}

/* 子菜单标题样式 - 与普通菜单项一致 */
:deep(.sidebar-menu .el-sub-menu__title) {
  color: #bfcbd9 !important;
  height: 50px;
  line-height: 50px;
  font-size: 16px;
  background-color: transparent !important;
}

/* 确保未展开状态下默认背景色透明 */
:deep(.sidebar-menu .el-sub-menu:not(.is-opened) > .el-sub-menu__title) {
  background-color: transparent !important;
}

:deep(.sidebar-menu .el-sub-menu__title span) {
  color: #bfcbd9 !important;
}

/* 未展开状态下的悬停样式 - 确保优先级最高 */
:deep(.sidebar-menu .el-sub-menu__title:hover),
:deep(.sidebar-menu .el-sub-menu:not(.is-opened) > .el-sub-menu__title:hover) {
  background-color: #263445 !important;
  color: #fff !important;
}

:deep(.sidebar-menu .el-sub-menu__title:hover span),
:deep(.sidebar-menu .el-sub-menu:not(.is-opened) > .el-sub-menu__title:hover span) {
  color: #fff !important;
}

:deep(.sidebar-menu .el-sub-menu__title:hover .el-icon),
:deep(.sidebar-menu .el-sub-menu:not(.is-opened) > .el-sub-menu__title:hover .el-icon) {
  color: #fff !important;
}

:deep(.sidebar-menu .el-sub-menu__title:hover .el-sub-menu__icon-arrow),
:deep(.sidebar-menu .el-sub-menu:not(.is-opened) > .el-sub-menu__title:hover .el-sub-menu__icon-arrow) {
  color: #fff !important;
}

/* 展开状态下背景色必须透明 */
:deep(.sidebar-menu .el-sub-menu.is-opened > .el-sub-menu__title) {
  color: #bfcbd9 !important;
  background: transparent !important;
  background-color: transparent !important;
}

:deep(.sidebar-menu .el-sub-menu.is-opened > .el-sub-menu__title:not(:hover)) {
  background: transparent !important;
  background-color: transparent !important;
}

:deep(.sidebar-menu .el-sub-menu.is-opened > .el-sub-menu__title span) {
  color: #bfcbd9 !important;
}

/* 子菜单展开箭头样式 - 与菜单文字颜色一致 */
.sidebar-menu .el-sub-menu__title .el-sub-menu__icon-arrow {
  color: #bfcbd9 !important;
  font-size: 12px;
  margin-top: -1px;
}

.sidebar-menu .el-sub-menu__title:hover .el-sub-menu__icon-arrow {
  color: #fff !important;
}

.sidebar-menu .el-sub-menu.is-opened > .el-sub-menu__title .el-sub-menu__icon-arrow {
  color: #bfcbd9 !important;
}

/* 确保展开状态下悬停样式与普通菜单项一致 */
.sidebar-menu .el-sub-menu.is-opened > .el-sub-menu__title:hover {
  background-color: #263445 !important;
  color: #fff !important;
}

.sidebar-menu .el-sub-menu.is-opened > .el-sub-menu__title:hover span {
  color: #fff !important;
}

.sidebar-menu .el-sub-menu.is-opened > .el-sub-menu__title:hover .el-icon {
  color: #fff !important;
}

.sidebar-menu .el-sub-menu.is-opened > .el-sub-menu__title:hover .el-sub-menu__icon-arrow {
  color: #fff !important;
}

/* 确保所有状态下的箭头颜色 */
.sidebar-menu .el-sub-menu .el-sub-menu__icon-arrow {
  color: #bfcbd9 !important;
}

.sidebar-menu .el-sub-menu .el-sub-menu__title:hover .el-sub-menu__icon-arrow {
  color: #fff !important;
}

/* 使用深度选择器确保样式生效 */
.sidebar-menu :deep(.el-sub-menu__icon-arrow) {
  color: #bfcbd9 !important;
}

.sidebar-menu :deep(.el-sub-menu__title:hover .el-sub-menu__icon-arrow) {
  color: #fff !important;
}

.sidebar-menu :deep(.el-sub-menu.is-opened .el-sub-menu__icon-arrow) {
  color: #bfcbd9 !important;
}

/* 子菜单项样式 */
.sidebar-menu .el-sub-menu .el-menu-item {
  background-color: #1f2d3d !important;
  color: #bfcbd9;
  padding-left: 50px !important;
  height: 45px;
  line-height: 45px;
  font-size: 16px;
  min-height: 45px;
}

.sidebar-menu .el-sub-menu .el-menu-item:hover {
  background-color: #263445 !important;
  color: #fff;
}

.sidebar-menu .el-sub-menu .el-menu-item.is-active {
  background-color: #409eff !important;
  color: #fff;
}

/* 子菜单容器样式 */
.sidebar-menu .el-sub-menu .el-menu {
  background-color: #1f2d3d;
}

/* 图标样式统一 */
.sidebar-menu .el-icon {
  font-size: 16px;
  margin-right: 8px;
}

/* 子菜单标题中的图标颜色 */
.sidebar-menu .el-sub-menu__title .el-icon {
  color: #bfcbd9 !important;
}

.sidebar-menu .el-sub-menu__title:hover .el-icon {
  color: #fff !important;
}

.sidebar-menu .el-sub-menu.is-opened > .el-sub-menu__title .el-icon {
  color: #bfcbd9 !important;
}

.header {
  background-color: #fff;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0;
  height: 60px;
  flex-direction: row;
}

.tabs-container {
  display: flex;
  align-items: center;
  flex: 1;
  border-bottom: 1px solid #e4e7ed;
  background-color: #fff;
  padding: 0;
  min-width: 0;
}

.tabs-wrapper {
  display: flex;
  align-items: center;
  flex: 1;
  overflow-x: auto;
  overflow-y: hidden;
}

.tabs-wrapper::-webkit-scrollbar {
  height: 4px;
}

.tabs-wrapper::-webkit-scrollbar-thumb {
  background-color: #c0c4cc;
  border-radius: 2px;
}

.tab-item {
  display: flex;
  align-items: center;
  padding: 8px 16px;
  cursor: pointer;
  border-right: 1px solid #e4e7ed;
  background-color: #fff;
  white-space: nowrap;
  transition: all 0.3s;
  position: relative;
  min-width: 100px;
  justify-content: space-between;
  gap: 8px;
}

.tab-item:hover {
  background-color: #f5f7fa;
}

.tab-item.active {
  background-color: #ecf5ff;
  color: #409eff;
  border-bottom: 2px solid #409eff;
}

.tab-title {
  font-size: 16px;
  color: #303133;
}

.tab-item.active .tab-title {
  color: #409eff;
  font-weight: 500;
}

.tab-close {
  font-size: 16px;
  color: #909399;
  cursor: pointer;
  padding: 2px;
  border-radius: 2px;
  transition: all 0.2s;
  flex-shrink: 0;
}

.tab-close:hover {
  background-color: #f56c6c;
  color: #fff;
}

.header-right {
  display: flex;
  align-items: center;
  padding: 0 20px;
  height: 60px;
  justify-content: flex-end;
  gap: 20px;
  flex-shrink: 0;
  border-bottom: 1px solid #e4e7ed;
  background-color: #fff;
}

.notice-badge {
  cursor: pointer;
}

.notice-dropdown-menu {
  padding: 0;
}

.notice-dropdown-header {
  padding: 12px 16px;
  border-bottom: 1px solid #e4e7ed;
  background-color: #f5f7fa;
}

.notice-dropdown-title {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.notice-tabs {
  padding: 0 12px;
}

.notice-tabs :deep(.el-tabs__header) {
  margin: 0;
  padding: 0;
}

.notice-tabs :deep(.el-tabs__nav-wrap) {
  padding: 0 8px;
}

.notice-tabs :deep(.el-tabs__item) {
  padding: 0 16px;
  height: 40px;
  line-height: 40px;
}

.notice-list {
  max-height: 400px;
  overflow-y: auto;
  min-width: 300px;
}

.notice-item {
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 1px solid #f0f0f0;
  transition: background-color 0.3s;
}

.notice-item:hover {
  background-color: #f5f7fa;
}

.notice-item.read {
  opacity: 0.6;
}

.notice-title {
  font-size: 16px;
  color: #303133;
  margin-bottom: 4px;
  font-weight: 500;
}

.notice-time {
  font-size: 16px;
  color: #909399;
}

.notice-empty {
  padding: 20px;
  text-align: center;
  color: #909399;
  font-size: 16px;
}

.sidebar {
  transition: width 0.3s;
}

.user-info {
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 5px;
}

.user-info-text {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  margin-left: 8px;
}

.user-info-name {
  font-size: 16px;
  color: #303133;
  font-weight: 500;
  line-height: 20px;
}

.user-info-meta {
  font-size: 13px;
  color: #909399;
  line-height: 18px;
}

.user-info .el-icon--right {
  color: #909399;
  font-size: 14px;
}

.user-dropdown-menu {
  padding: 10px 0;
  min-width: 240px;
}

.user-dropdown-info {
  padding: 12px 16px 10px;
  border-bottom: 1px solid #ebeef5;
}

.user-dropdown-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 6px;
}

.user-dropdown-meta {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: #606266;
}

.meta-item .el-icon {
  font-size: 16px;
  color: #909399;
}

.el-main {
  background-color: #f0f2f5;
  padding: 0;
  overflow: hidden;
}

.main-content {
  height: 100%;
  overflow: hidden;
}

.split-container {
  height: 100%;
  display: flex;
}

.split-container.single {
  flex-direction: column;
}

.split-container.double {
  flex-direction: row;
}

.split-container.triple {
  flex-direction: row;
}

.split-container.quadruple {
  flex-direction: column;
  display: grid;
  grid-template-columns: 1fr 1fr;
  grid-template-rows: 1fr 1fr;
}

.split-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  background-color: #fff;
  border-right: 1px solid #e4e7ed;
  border-bottom: 1px solid #e4e7ed;
  overflow: hidden;
}

.split-container.double .split-panel:last-child,
.split-container.triple .split-panel:last-child {
  border-right: none;
}

.split-container.quadruple .split-panel:nth-child(2n) {
  border-right: none;
}

.split-container.quadruple .split-panel:nth-child(n+3) {
  border-bottom: none;
}

.split-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background-color: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
  height: 40px;
}

.panel-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.panel-action-btn {
  font-size: 18px;
  color: #909399;
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
}

.panel-action-btn:hover {
  background-color: #ecf5ff;
  color: #409eff;
}

.panel-title {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
  cursor: pointer;
  display: flex;
  align-items: center;
  transition: color 0.2s;
}

.panel-title:hover {
  color: #409eff;
}

.panel-close {
  font-size: 18px;
  color: #909399;
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
}

.panel-close:hover {
  background-color: #f56c6c;
  color: #fff;
}

.split-panel-content {
  flex: 1;
  overflow: auto;
  padding: 20px;
}

.split-controls {
  display: flex;
  align-items: center;
  padding: 0 8px;
  border-left: 1px solid #e4e7ed;
  flex-shrink: 0;
}

.split-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #606266;
}

.split-btn:hover {
  color: #409eff;
}

/* 分屏面板下拉菜单激活状态 */
:deep(.el-dropdown-menu__item.is-active) {
  color: #409eff;
  background-color: #ecf5ff;
}
</style>

<style>
/* 全局样式 - 确保系统管理菜单的展开箭头颜色与文字一致 */
.sidebar-menu .el-sub-menu__icon-arrow {
  color: #bfcbd9 !important;
}

.sidebar-menu .el-sub-menu__title:hover .el-sub-menu__icon-arrow {
  color: #fff !important;
}

.sidebar-menu .el-sub-menu.is-opened .el-sub-menu__icon-arrow {
  color: #bfcbd9 !important;
}

.sidebar-menu .el-sub-menu.is-opened > .el-sub-menu__title:hover .el-sub-menu__icon-arrow {
  color: #fff !important;
}

/* 确保未展开状态下默认背景色透明 - 全局样式 */
.sidebar-menu .el-sub-menu:not(.is-opened) > .el-sub-menu__title {
  background-color: transparent !important;
  background: transparent !important;
}

/* 使用深度选择器确保覆盖Element Plus默认样式 */
.sidebar-menu :deep(.el-sub-menu:not(.is-opened) > .el-sub-menu__title) {
  background-color: transparent !important;
  background: transparent !important;
}

/* 确保未展开状态下悬停样式正确应用 - 全局样式优先级 */
.sidebar-menu .el-sub-menu:not(.is-opened) > .el-sub-menu__title:hover {
  background-color: #263445 !important;
  background: #263445 !important;
  color: #fff !important;
}

.sidebar-menu .el-sub-menu:not(.is-opened) > .el-sub-menu__title:hover span {
  color: #fff !important;
}

.sidebar-menu .el-sub-menu:not(.is-opened) > .el-sub-menu__title:hover .el-icon {
  color: #fff !important;
}

.sidebar-menu .el-sub-menu:not(.is-opened) > .el-sub-menu__title:hover .el-sub-menu__icon-arrow {
  color: #fff !important;
}

/* 使用深度选择器确保悬停样式覆盖 */
.sidebar-menu :deep(.el-sub-menu:not(.is-opened) > .el-sub-menu__title:hover) {
  background-color: #263445 !important;
  background: #263445 !important;
  color: #fff !important;
}

/* 确保展开状态下悬停样式与普通菜单项完全一致 */
.sidebar-menu .el-sub-menu.is-opened > .el-sub-menu__title:hover {
  background-color: #263445 !important;
  color: #fff !important;
}

.sidebar-menu .el-sub-menu.is-opened > .el-sub-menu__title:hover span {
  color: #fff !important;
}

.sidebar-menu .el-sub-menu.is-opened > .el-sub-menu__title:hover .el-icon {
  color: #fff !important;
}

/* 强制确保展开状态（点击后）下背景色保持透明 - 最高优先级 */
.sidebar-menu .el-sub-menu.is-opened > .el-sub-menu__title,
.sidebar-menu .el-sub-menu.is-opened .el-sub-menu__title {
  background: transparent !important;
  background-color: transparent !important;
}

.sidebar-menu .el-sub-menu.is-opened > .el-sub-menu__title:not(:hover),
.sidebar-menu .el-sub-menu.is-opened .el-sub-menu__title:not(:hover) {
  background: transparent !important;
  background-color: transparent !important;
}
</style>

