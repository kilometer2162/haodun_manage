import { defineStore } from 'pinia'
import { jwtDecode } from 'jwt-decode'
import api from '@/utils/api'

const TOKEN_KEY = 'token'
const USER_KEY = 'user'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: JSON.parse(localStorage.getItem(USER_KEY) || 'null'),
    token: localStorage.getItem(TOKEN_KEY) || null,
    permissions: []
  }),

  getters: {
    isAuthenticated: (state) => !!state.token,
    userPermissions: (state) => state.permissions
  },

  actions: {
    setAuthData(token, user, permissions = []) {
      this.token = token
      this.user = user
      this.permissions = permissions

      if (token) {
        localStorage.setItem(TOKEN_KEY, token)
        api.defaults.headers.common['Authorization'] = `Bearer ${token}`
      } else {
        localStorage.removeItem(TOKEN_KEY)
        delete api.defaults.headers.common['Authorization']
      }

      if (user) {
        localStorage.setItem(USER_KEY, JSON.stringify(user))
      } else {
        localStorage.removeItem(USER_KEY)
      }
    },

    async login(username, password) {
      try {
        const response = await api.post('/auth/login', { username, password })
        const { token, user } = response.data
        const permissions = Array.isArray(user?.permissions) ? user.permissions : []

        this.setAuthData(token, user, permissions)
        return { success: true }
      } catch (error) {
        console.error('Login failed:', error)
        return { success: false, error: error.response?.data?.error || '登录失败' }
      }
    },

    logout() {
      this.setAuthData(null, null, [])
    },

    loadUserFromToken() {
      if (!this.token) {
        return false
      }

      try {
        const decoded = jwtDecode(this.token)
        this.user = {
          id: decoded.user_id,
          username: decoded.username,
          role_id: decoded.role_id
        }
        localStorage.setItem(USER_KEY, JSON.stringify(this.user))
        api.defaults.headers.common['Authorization'] = `Bearer ${this.token}`
        return true
      } catch (error) {
        console.error('Invalid token:', error)
        this.logout()
        return false
      }
    },

    async fetchUserInfo() {
      if (!this.token) {
        return false
      }

      try {
        const response = await api.get('/auth/current-user')
        const data = response.data
        const permissions = Array.isArray(data?.permissions) ? data.permissions : []

        this.setAuthData(this.token, data, permissions)
        return true
      } catch (error) {
        console.error('Failed to fetch user info:', error)
        this.logout()
        return false
      }
    }
  }
})

