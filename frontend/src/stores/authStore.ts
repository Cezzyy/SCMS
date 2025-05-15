import { defineStore } from 'pinia'
import axios from 'axios'
import { ref, computed } from 'vue'

// Hardcode the API URL and create a configured axios instance
const API_URL = 'http://localhost:8081/api'

// Create an axios instance with default config
const api = axios.create({
  baseURL: API_URL,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  }
})

export interface User {
  user_id: number
  email: string
  first_name: string
  last_name: string
  role: string
  session_id: string
  expires_at: string
}

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Get user from localStorage on initialization
  try {
    const storedUser = localStorage.getItem('user')
    if (storedUser) {
      user.value = JSON.parse(storedUser)
    }
  } catch (err) {
    console.error('Error loading user from localStorage', err)
    localStorage.removeItem('user')
  }

  // Computed
  const isAuthenticated = computed(() => !!user.value)
  const userFullName = computed(() => {
    if (!user.value) return ''
    return `${user.value.first_name} ${user.value.last_name}`
  })
  const userRole = computed(() => user.value?.role || '')

  // Actions
  const login = async (email: string, password: string) => {
    loading.value = true
    error.value = null

    console.log('Attempting login with API URL:', API_URL)

    try {
      const response = await api.post('/auth/login', { email, password })

      console.log('Login successful:', response.data)
      user.value = response.data
      localStorage.setItem('user', JSON.stringify(response.data))
      return response.data
    } catch (err: any) {
      console.error('Login error details:', err)

      if (err.response) {
        error.value = err.response.data || 'Invalid credentials'
        console.error('Server error response:', err.response.status, err.response.data)
      } else if (err.request) {
        error.value = 'Server not responding. Please try again later.'
        console.error('No response received:', err.request)
      } else {
        error.value = 'An unexpected error occurred'
        console.error('Error setting up request:', err.message)
      }
      throw err
    } finally {
      loading.value = false
    }
  }

  const logout = async () => {
    loading.value = true
    error.value = null

    try {
      await api.post('/auth/logout')
    } catch (err) {
      console.error('Error during logout', err)
    } finally {
      // Clear local data regardless of API result
      user.value = null
      localStorage.removeItem('user')
      loading.value = false
    }
  }

  const checkAuth = () => {
    // Check if the user is logged in and token is valid
    // This is a simplified version. In a real app, you might check token expiration
    return !!user.value
  }

  return {
    user,
    loading,
    error,
    isAuthenticated,
    userFullName,
    userRole,
    login,
    logout,
    checkAuth
  }
})
