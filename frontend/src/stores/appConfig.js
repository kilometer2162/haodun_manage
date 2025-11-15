import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/utils/api'

const DEFAULT_SITE_NAME = '管理系统'

export const useAppConfigStore = defineStore('appConfig', () => {
  const siteName = ref(DEFAULT_SITE_NAME)
  const siteNameLoaded = ref(false)
  const loading = ref(false)

  const updateDocumentTitle = (name) => {
    if (typeof document !== 'undefined') {
      document.title = name || DEFAULT_SITE_NAME
    }
  }

  const loadSiteName = async (force = false) => {
    if (loading.value || (siteNameLoaded.value && !force)) {
      return siteName.value
    }

    loading.value = true
    try {
      const response = await api.get('/config/site_name')
      const value = response.data?.value
      if (value) {
        siteName.value = value
      } else {
        siteName.value = DEFAULT_SITE_NAME
      }
    } catch (error) {
      console.warn('加载站点名称失败，将使用默认值', error)
      siteName.value = DEFAULT_SITE_NAME
    } finally {
      loading.value = false
      siteNameLoaded.value = true
      updateDocumentTitle(siteName.value)
    }

    return siteName.value
  }

  return {
    siteName,
    siteNameLoaded,
    loadSiteName,
    updateDocumentTitle
  }
})


