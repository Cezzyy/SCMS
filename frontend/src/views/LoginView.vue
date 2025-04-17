<script setup lang="ts">
import { ref } from 'vue'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faUser, faLock } from '@fortawesome/free-solid-svg-icons'

const username = ref('')
const password = ref('')

const handleLogin = () => {
  console.log('Login attempt:', { username: username.value, password: password.value })
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

      <!-- Login Form -->
      <form @submit.prevent="handleLogin" class="space-y-4">
        <!-- Username Field -->
        <div class="space-y-2">
          <label for="username" class="block text-sm font-semibold text-gray-700">Username</label>
          <div class="relative">
            <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <font-awesome-icon :icon="faUser" class="h-5 w-5 text-accent1" />
            </div>
            <input
              v-model="username"
              id="username"
              type="text"
              required
              class="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-accent1 focus:border-transparent text-gray-900 placeholder-gray-500"
              placeholder="Enter your username"
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

        <!-- Remember Me & Forgot Password -->
        <div class="flex items-center">
          <div class="flex items-center">
            <input
              id="remember-me"
              type="checkbox"
              class="h-4 w-4 text-accent1 focus:ring-accent1 border-gray-300 rounded"
            />
            <label for="remember-me" class="ml-2 block text-sm text-gray-600 font-medium">Remember me</label>
          </div>
        </div>

        <!-- Login Button -->
        <div class="flex items-center justify-end mb-4">
          <a href="#" class="text-sm font-medium text-accent1 hover:text-accent2">Contact IT Support</a>
        </div>
        <button
          type="submit"
          class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-accent1 hover:bg-accent2 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-accent1 transition-colors duration-200"
        >
          Sign in to Portal
        </button>
      </form>
    </div>
  </div>
</div>
</template>
