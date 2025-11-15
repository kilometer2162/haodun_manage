<template>
  <div class="login-container">
    <div class="login-box">
      <h2>{{ siteName }}系统</h2>
      <el-form
        :model="loginForm"
        :rules="rules"
        ref="loginFormRef"
        label-width="90px"
        size="large"
      >
        <el-form-item label="用户名" prop="username">
          <el-input v-model="loginForm.username" placeholder="请输入用户名" clearable />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="请输入密码"
            clearable
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <!-- 验证码暂时隐藏 -->
        <!-- <el-form-item label="验证码" prop="captchaCode">
          <div class="captcha-row">
            <el-input
              v-model="loginForm.captchaCode"
              placeholder="请输入验证码"
              style="width: 200px"
            />
            <el-button @click="getCaptcha" :loading="captchaLoading">
              {{ captchaCode || '获取验证码' }}
            </el-button>
          </div>
        </el-form-item> -->
        <el-form-item class="login-btn-item">
          <el-button
            type="primary"
            @click="handleLogin"
            :loading="loading"
            style="width: 100%"
            size="large"
            class="login-btn"
          >
            登录
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'
import { useAppConfigStore } from '@/stores/appConfig'

const router = useRouter()
const authStore = useAuthStore()
const appConfigStore = useAppConfigStore()

const loginFormRef = ref(null)
const loading = ref(false)
const captchaLoading = ref(false)
const captchaCode = ref('')
const captchaId = ref('')

const siteName = computed(() => appConfigStore.siteName || '管理系统')

watch(
  siteName,
  (val) => {
    appConfigStore.updateDocumentTitle(val)
  },
  { immediate: true }
)

const loginForm = reactive({
  username: '',
  password: '',
  captchaCode: ''
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
  // captchaCode: [{ required: true, message: '请输入验证码', trigger: 'blur' }] // 验证码暂时隐藏
}

const getCaptcha = async () => {
  captchaLoading.value = true
  try {
    const response = await api.get('/auth/captcha')
    captchaId.value = response.data.captcha_id
    captchaCode.value = response.data.captcha_code
    ElMessage.success('验证码已获取')
  } catch (error) {
    ElMessage.error('获取验证码失败')
  } finally {
    captchaLoading.value = false
  }
}

const handleLogin = async () => {
  if (!loginFormRef.value) return
  
  await loginFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        const result = await authStore.login(loginForm.username, loginForm.password)
        if (!result.success) {
          throw new Error(result.error || '登录失败')
        }

        await authStore.fetchUserInfo()

        ElMessage.success('登录成功')

        const permissions = authStore.userPermissions
        const targetRoute = getDefaultRoute(permissions)
        router.push(targetRoute)
      } catch (error) {
        ElMessage.error(error.message || error.response?.data?.error || '登录失败')
      } finally {
        loading.value = false
      }
    }
  })
}

const getDefaultRoute = (permissions) => {
  if (!permissions || permissions.length === 0) {
    return '/'
  }

  if (permissions.includes('menu:dashboard:view')) {
    return '/dashboard'
  }
  if (permissions.includes('menu:orders:view')) {
    return '/orders'
  }
  if (permissions.includes('menu:notices:view')) {
    return '/notices'
  }
  if (permissions.includes('menu:dicts:view')) {
    return '/dicts'
  }

  return '/'
}

// 页面加载时获取验证码 - 验证码暂时隐藏
// getCaptcha()
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: linear-gradient(135deg, #5a7bea 0%, #6f3b9d 55%, #2c2f66 100%);
  padding: 24px;
}

.login-box {
  width: 460px;
  padding: 52px 64px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 18px;
  box-shadow: 0 32px 70px rgba(22, 27, 45, 0.25);
  backdrop-filter: blur(12px);
}

.login-box h2 {
  text-align: center;
  margin-bottom: 40px;
  color: #1f2451;
  font-size: 32px;
  font-weight: 600;
  letter-spacing: 2px;
}

.login-box :deep(.el-form-item__label) {
  font-size: 16px;
  color: #394264;
  font-weight: 500;
}

.login-box :deep(.el-input__wrapper) {
  padding: 4px 16px;
  border-radius: 10px;
  box-shadow: none;
}

.login-box :deep(.el-input__inner) {
  font-size: 16px;
  line-height: 38px;
  font-weight: 500;
}

.login-btn-item {
  margin-top: 20px;
}

.login-box :deep(.login-btn) {
  height: 52px;
  font-size: 18px;
  border-radius: 12px;
  letter-spacing: 4px;
  font-weight: 600;
  box-shadow: 0 14px 28px rgba(79, 125, 247, 0.35);
}

.login-box :deep(.login-btn.is-loading) {
  box-shadow: none;
}

.captcha-row {
  display: flex;
  gap: 10px;
  align-items: center;
}
</style>

