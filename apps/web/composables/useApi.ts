// apps/web/composables/useApi.ts
export const useApi = () => {
  const config = useRuntimeConfig()
  const baseURL = config.public?.apiBase || 'http://localhost:8080/api/v1'

  const request = async <T = any>(
    endpoint: string,
    options: {
      method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
      body?: any
      headers?: Record<string, string>
      params?: Record<string, any>
    } = {}
  ): Promise<T> => {
    const url = `${baseURL}${endpoint.startsWith('/') ? endpoint : `/${endpoint}`}`
    
    // Retrieve auth token from localStorage if present
    let authHeader: Record<string, string> = {}
    if (import.meta.client) {
      const token = localStorage.getItem('treewiki_token')
      if (token) {
        authHeader['Authorization'] = `Bearer ${token}`
      }
    }

    try {
      const res = await $fetch<T>(url, {
        method: options.method || 'GET',
        body: options.body,
        params: options.params,
        headers: {
          'Content-Type': 'application/json',
          ...authHeader,
          ...options.headers,
        },
        credentials: 'include',
      })
      return res
    } catch (err: any) {
      console.warn(`[API] Error on ${options.method || 'GET'} ${url}:`, err.data || err.message)
      throw err
    }
  }

  return {
    baseURL,
    request,
    get: <T = any>(endpoint: string, params?: Record<string, any>, headers?: Record<string, string>) =>
      request<T>(endpoint, { method: 'GET', params, headers }),
    post: <T = any>(endpoint: string, body?: any, headers?: Record<string, string>) =>
      request<T>(endpoint, { method: 'POST', body, headers }),
    put: <T = any>(endpoint: string, body?: any, headers?: Record<string, string>) =>
      request<T>(endpoint, { method: 'PUT', body, headers }),
    delete: <T = any>(endpoint: string, headers?: Record<string, string>) =>
      request<T>(endpoint, { method: 'DELETE', headers }),
  }
}
