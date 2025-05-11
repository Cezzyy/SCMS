<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faEnvelope, faLock } from '@fortawesome/free-solid-svg-icons'
import { useAuthStore } from '../stores/authStore'

const email = ref('')
const password = ref('')
const errorMessage = ref('')
const router = useRouter()
const authStore = useAuthStore()

// Clear any previous error messages on component mount
onMounted(() => {
  errorMessage.value = ''
  // For development/testing, you can set default values
  email.value = 'admin@example.com'
  password.value = 'password123'
})

const handleLogin = async () => {
  // Clear previous error messages
  errorMessage.value = ''

  if (!email.value || !password.value) {
    errorMessage.value = 'Email and password are required'
    return
  }

  try {
    console.log('Attempting login with:', { email: email.value })
    await authStore.login(email.value, password.value)

    // Redirect to home instead of dashboard
    router.push('/home')
  } catch (error: any) {
    console.error('Login failed:', error)
    // Display the error message from the store if available
    errorMessage.value = authStore.error || 'Login failed. Please check your credentials and try again.'
  }
}
</script>

<template>
  <div class="min-h-screen bg-primary flex items-center justify-center">
    <!-- Left Side - Company Branding -->
    <div class="hidden lg:flex flex-col items-center justify-center w-1/2 h-screen bg-accent1 p-8 text-white">
      <img src="@/assets/CenterLogo.svg" alt="Company Logo" class="w-32 h-32 text-white mb-6" />
      <h1 class="text-3xl font-bold mb-4 text-center">Center Industrial Supply Corporation</h1>
      <p class="text-xl text-center max-w-lg">
        Your trusted partner in welding machines & equipment supply. Delivering quality industrial solutions for your business.
      </p>
    </div>

    <!-- Right Side - Login Form -->
    <div class="w-full lg:w-1/2 flex items-center justify-center p-4">
      <div class="max-w-md w-full bg-bg-alt rounded-lg shadow-xl p-8 space-y-6">
      <!-- Logo/Header Section -->
      <div class="text-center">
        <h1 class="text-2xl font-bold text-gray-800 mb-2">Employee Login</h1>
        <p class="text-gray-600 font-medium">Please sign in with your company credentials</p>
      </div>

      <!-- Error message -->
      <div v-if="errorMessage" class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative" role="alert">
        <span class="block sm:inline">{{ errorMessage }}</span>
      </div>

      <!-- Login Form -->
      <form @submit.prevent="handleLogin" class="space-y-4">
        <!-- Email Field -->
        <div class="space-y-2">
          <label for="email" class="block text-sm font-semibold text-gray-700">Email</label>
          <div class="relative">
            <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <font-awesome-icon :icon="faEnvelope" class="h-5 w-5 text-accent1" />
            </div>
            <input
              v-model="email"
              id="email"
              type="email"
              required
              class="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent1 focus:border-transparent text-gray-900 placeholder-gray-500"
              placeholder="Enter your email"
            />
          </div>
        </div>

        <!-- Password Field -->
        <div class="space-y-2">
          <label for="password" class="block text-sm font-semibold text-gray-700">Password</label>
          <div class="relative">
            <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <font-awesome-icon :icon="faLock" class="h-5 w-5 text-accent1" />
            </div>
            <input
              v-model="password"
              id="password"
              type="password"
              required
              class="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent1 focus:border-transparent text-gray-900 placeholder-gray-500"
              placeholder="Enter your password"
            />
          </div>
        </div>

        <!-- Login Button -->
        <button
          type="submit"
          :disabled="authStore.loading"
          class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-accent1 hover:bg-accent2 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-accent1 transition-colors duration-200 disabled:opacity-50"
        >
          <span v-if="authStore.loading" class="flex items-center">
            <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Signing in...
          </span>
          <span v-else>Sign in to Portal</span>
        </button>
      </form>
    </div>
  </div>
</div>
</template>
