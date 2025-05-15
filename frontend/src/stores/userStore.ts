import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/services/api'
import type { User } from '@/types/User'

export const useUserStore = defineStore('user', () => {
  // State
  const users = ref<User[]>([])
  const selectedUser = ref<User | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  
  // Computed properties
  const sortedUsers = computed(() => {
    return [...users.value].sort((a, b) => a.last_name.localeCompare(b.last_name))
  })

  // Actions
  const fetchUsers = async () => {
    loading.value = true
    error.value = null
    
    try {
      const response = await api.get('/api/users')
      users.value = response.data
      return response.data
    } catch (err: any) {
      handleError(err, 'Failed to fetch users')
      throw err
    } finally {
      loading.value = false
    }
  }

  const getUserById = async (id: number) => {
    loading.value = true
    error.value = null
    
    try {
      const response = await api.get(`/api/users/${id}`)
      selectedUser.value = response.data
      return response.data
    } catch (err: any) {
      handleError(err, 'Failed to fetch user')
      throw err
    } finally {
      loading.value = false
    }
  }

  const createUser = async (userData: Partial<User>) => {
    loading.value = true
    error.value = null
    
    try {
      const response = await api.post('/api/users', userData)
      users.value.push(response.data)
      return response.data
    } catch (err: any) {
      handleError(err, 'Failed to create user')
      throw err
    } finally {
      loading.value = false
    }
  }

  const updateUser = async (id: number, userData: Partial<User>) => {
    loading.value = true
    error.value = null
    
    try {
      const response = await api.put(`/api/users/${id}`, userData)
      
      // Update user in the users array
      const index = users.value.findIndex(u => u.user_id === id)
      if (index !== -1) {
        users.value[index] = { ...users.value[index], ...response.data }
      }
      
      return response.data
    } catch (err: any) {
      handleError(err, 'Failed to update user')
      throw err
    } finally {
      loading.value = false
    }
  }

  const deleteUser = async (id: number) => {
    loading.value = true
    error.value = null
    
    try {
      await api.delete(`/api/users/${id}`)
      users.value = users.value.filter(u => u.user_id !== id)
      return true
    } catch (err: any) {
      handleError(err, 'Failed to delete user')
      throw err
    } finally {
      loading.value = false
    }
  }

  const updatePassword = async (id: number, currentPassword: string, newPassword: string) => {
    loading.value = true
    error.value = null
    
    try {
      const response = await api.put(`/api/users/${id}/password`, {
        current_password: currentPassword,
        new_password: newPassword
      })
      return response.data
    } catch (err: any) {
      handleError(err, 'Failed to update password')
      throw err
    } finally {
      loading.value = false
    }
  }

  const searchUsers = async (term: string) => {
    loading.value = true
    error.value = null
    
    try {
      const response = await api.get(`/api/users/search?q=${encodeURIComponent(term)}`)
      return response.data
    } catch (err: any) {
      handleError(err, 'Failed to search users')
      throw err
    } finally {
      loading.value = false
    }
  }

  // Helper for error handling
  const handleError = (err: any, defaultMessage: string) => {
    if (err.response && err.response.data && err.response.data.error) {
      error.value = err.response.data.error
    } else {
      error.value = defaultMessage
    }
    console.error(defaultMessage, err)
  }

  return {
    users,
    selectedUser,
    loading,
    error,
    sortedUsers,
    fetchUsers,
    getUserById,
    createUser,
    updateUser,
    deleteUser,
    updatePassword,
    searchUsers
  }
}) 