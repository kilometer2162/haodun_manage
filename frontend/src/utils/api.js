import axios from 'axios'

const api = axios.create({
    baseURL: '/api',
    timeout: 10000
})

// 请求拦截器
api.interceptors.request.use(
    config => {
        const token = localStorage.getItem('token')
        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }
        return config
    },
    error => {
        return Promise.reject(error)
    }
)

// 响应拦截器
api.interceptors.response.use(
    response => {
        return response
    },
    error => {
        const status = error.response?.status
        const requestUrl = error.config?.url

        if (status === 401) {
            const isLoginRequest = requestUrl === '/auth/login'

            if (!isLoginRequest) {
                localStorage.removeItem('token')
                localStorage.removeItem('user')
                if (window.location.pathname !== '/login') {
                    window.location.href = '/login'
                }
            }
        }

        return Promise.reject(error)
    }
)

export default api

