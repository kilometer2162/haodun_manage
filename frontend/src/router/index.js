import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'

const routes = [
    {
        path: '/login',
        name: 'Login',
        component: () => import('@/views/Login.vue'),
        meta: { requiresAuth: false }
    },
    {
        path: '/',
        component: () => import('@/layouts/MainLayout.vue'),
        redirect: '/dashboard',
        meta: { requiresAuth: true },
        children: [
            {
                path: 'dashboard',
                name: 'Dashboard',
                component: () => import('@/views/Dashboard.vue'),
                meta: { requiresAuth: true, permission: 'menu:dashboard:view' }
            },
            {
                path: 'users',
                name: 'Users',
                component: () => import('@/views/Users.vue'),
                meta: { requiresAuth: true, permission: 'menu:users:view' }
            },
            {
                path: 'departments',
                name: 'Departments',
                component: () => import('@/views/Departments.vue'),
                meta: { requiresAuth: true, permission: 'menu:departments:view' }
            },
            {
                path: 'materials',
                name: 'Materials',
                component: () => import('@/views/Materials.vue'),
                meta: { requiresAuth: true, permission: 'menu:materials:view' }
            },
            {
                path: 'roles',
                name: 'Roles',
                component: () => import('@/views/Roles.vue'),
                meta: { requiresAuth: true, permission: 'menu:roles:view' }
            },
            {
                path: 'permissions',
                name: 'Permissions',
                component: () => import('@/views/Permissions.vue'),
                meta: { requiresAuth: true, permission: 'menu:permissions:view' }
            },
            {
                path: 'resources',
                name: 'Resources',
                component: () => import('@/views/Resources.vue'),
                meta: { requiresAuth: true, permission: 'menu:resources:view' }
            },
            {
                path: 'logs',
                name: 'Logs',
                component: () => import('@/views/Logs.vue'),
                meta: { requiresAuth: true, permission: 'menu:logs:view' }
            },
            {
                path: 'system-monitor',
                name: 'SystemMonitor',
                component: () => import('@/views/SystemMonitor.vue'),
                meta: { requiresAuth: true, permission: 'menu:system-monitor:view' }
            },
            {
                path: 'ip-statistics',
                name: 'IPStatistics',
                component: () => import('@/views/IPStatistics.vue'),
                meta: { requiresAuth: true, permission: 'menu:ip-statistics:view' }
            },
            {
                path: 'dicts',
                name: 'Dicts',
                component: () => import('@/views/Dicts.vue'),
                meta: { requiresAuth: true, permission: 'menu:dicts:view' }
            },
            {
                path: 'configs',
                name: 'Configs',
                component: () => import('@/views/Configs.vue'),
                meta: { requiresAuth: true, permission: 'menu:configs:view' }
            },
            {
                path: 'notices',
                name: 'Notices',
                component: () => import('@/views/Notices.vue'),
                meta: { requiresAuth: true, permission: 'menu:notices:view' }
            },
            {
                path: 'orders',
                name: 'Orders',
                component: () => import('@/views/Orders.vue'),
                meta: { requiresAuth: true, permission: 'menu:orders:view' }
            }
        ]
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

router.beforeEach(async (to, from, next) => {
    const authStore = useAuthStore()

    // If not authenticated and route requires auth, redirect to login
    if (to.meta.requiresAuth && !authStore.isAuthenticated) {
        next('/login')
        return
    }

    // If going to login and already authenticated, redirect to dashboard
    if (to.path === '/login' && authStore.isAuthenticated) {
        next('/')
        return
    }

    // For authenticated routes, ensure user info and permissions are loaded
    if (to.meta.requiresAuth && authStore.isAuthenticated) {
        // If permissions not loaded, fetch them
        if (!authStore.userPermissions.length) {
            const success = await authStore.fetchUserInfo()
            if (!success) {
                authStore.logout()
                next('/login')
                return
            }
        }

        // Check permission if route requires specific permission
        if (to.meta.permission) {
            const hasPermission = authStore.userPermissions.includes(to.meta.permission)
            if (!hasPermission) {
                ElMessage.warning('没有访问权限')
                // For operator role, redirect to orders or notices
                if (authStore.user.role_id === 2) {
                    next('/orders')
                } else {
                    next('/dashboard')
                }
                return
            }
        }
    }

    next()
})

export default router

