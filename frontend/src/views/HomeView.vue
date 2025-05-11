<script setup lang="ts">
import { ref, shallowRef, onMounted, onBeforeMount, onUnmounted, defineAsyncComponent } from 'vue'
import type { Component } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/authStore'

// Lazy load components
const DashboardComponent = defineAsyncComponent(() => import('../components/DashboardContent.vue'))
const CustomersComponent = defineAsyncComponent(() => import('../components/CustomerList.vue'))
const SalesOrdersComponent = defineAsyncComponent(() => import('../components/OrderList.vue'))
const InventoryComponent = defineAsyncComponent(() => import('../components/InventoryList.vue'))
const QuotationsComponent = defineAsyncComponent(() => import('../components/QuotationList.vue'))

// Initialize stores and router
const authStore = useAuthStore()
const router = useRouter()

// Types
interface NavigationItem {
  id: string;
  name: string;
  icon: string;
}

type ComponentMap = {
  [key: string]: Component;
}

// State
const isSidebarOpen = ref<boolean>(true) // Default to open
const isMobile = ref<boolean>(false)
const activeComponent = ref<string>('dashboard')
const currentComponent = shallowRef<Component>(DashboardComponent)
const darkMode = ref(false)

// Check for mobile screen
const checkScreenSize = (): void => {
  isMobile.value = window.innerWidth < 768
  // Auto close sidebar on small screens
  if (isMobile.value) {
    isSidebarOpen.value = false
  } else {
    // Auto open sidebar on larger screens
    isSidebarOpen.value = true
  }
}

// Methods
const toggleSidebar = (): void => {
  isSidebarOpen.value = !isSidebarOpen.value
}

const toggleDarkMode = () => {
  darkMode.value = !darkMode.value
  if (darkMode.value) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
}

// Logout function
const handleLogout = async () => {
  try {
    await authStore.logout()
    router.push('/login')
  } catch (error) {
    console.error('Logout failed:', error)
  }
}

const navigationItems: NavigationItem[] = [
  { id: 'dashboard', name: 'Dashboard', icon: 'home' },
  { id: 'customers', name: 'Customers', icon: 'users' },
  { id: 'orders', name: 'Orders', icon: 'clipboard-list' },
  { id: 'inventory', name: 'Inventory', icon: 'box' },
  { id: 'quotations', name: 'Quotations', icon: 'file-invoice-dollar' },
]

const componentMap: ComponentMap = {
  dashboard: DashboardComponent,
  customers: CustomersComponent,
  orders: SalesOrdersComponent,
  inventory: InventoryComponent,
  quotations: QuotationsComponent,
}

const switchComponent = (componentId: string): void => {
  // Close sidebar on mobile after selection
  if (isMobile.value) {
    isSidebarOpen.value = false
  }

  activeComponent.value = componentId
  currentComponent.value = componentMap[componentId]
}

onBeforeMount(() => {
  checkScreenSize()

  // Check authentication
  if (!authStore.isAuthenticated) {
    router.push('/login')
  }
})

onMounted(() => {
  // Set initial component
  currentComponent.value = componentMap[activeComponent.value]

  // Add window resize listener
  window.addEventListener('resize', checkScreenSize)
})

onUnmounted(() => {
  // Clean up resize listener
  window.removeEventListener('resize', checkScreenSize)
})
</script>

<template>
  <div :class="['min-h-screen flex flex-col md:flex-row', darkMode ? 'dark bg-bg-alt dark:bg-gray-900' : 'bg-white']">
    <!-- Overlay for mobile when sidebar is open -->
    <div
      v-if="isMobile && isSidebarOpen"
      @click="toggleSidebar"
      class="fixed inset-0 bg-black bg-opacity-50 z-20"
    ></div>

    <!-- Sidebar -->
    <aside
      :class="[
        'flex flex-col bg-primary text-text-primary transition-all duration-300 ease-in-out',
        'z-30 flex-shrink-0',
        isMobile ? 'fixed left-0 top-0 bottom-0 h-screen' : 'sticky top-0 max-h-screen overflow-y-auto',
        isSidebarOpen ? 'translate-x-0 w-64' : isMobile ? '-translate-x-full' : 'w-16',
        'dark:bg-gray-900 dark:text-gray-100'
      ]"
    >
      <!-- Sidebar header -->
      <div class="flex items-center justify-between h-16 px-4 border-b border-accent2 dark:border-gray-800">
        <div v-if="isSidebarOpen" class="flex items-center space-x-2">
          <span class="text-accent1 text-xl font-bold">SCMS</span>
        </div>
        <button
          @click="toggleSidebar"
          class="text-text-primary hover:text-accent1 transition-colors p-2"
          aria-label="Toggle sidebar"
        >
          <font-awesome-icon :icon="isSidebarOpen ? 'times' : 'bars'" class="h-5 w-5" />
        </button>
      </div>

      <!-- Navigation -->
      <nav class="flex-1 px-2 py-4 space-y-1 overflow-y-auto">
        <a
          v-for="item in navigationItems"
          :key="item.id"
          href="#"
          @click.prevent="switchComponent(item.id)"
          :class="[
            activeComponent === item.id ? 'bg-accent2 text-white dark:bg-gray-800 dark:text-white' : 'text-text-secondary hover:bg-accent2 hover:bg-opacity-25 hover:text-text-primary dark:text-gray-400 dark:hover:bg-gray-800 dark:hover:text-white',
            'flex items-center rounded-md text-sm font-medium transition-colors duration-150',
            isSidebarOpen ? 'px-4 py-3' : 'justify-center px-3 py-3'
          ]"
          :title="!isSidebarOpen ? item.name : ''"
        >
          <font-awesome-icon :icon="item.icon" :class="['h-5 w-5', isSidebarOpen && 'mr-3']" />
          <span v-if="isSidebarOpen">{{ item.name }}</span>
        </a>
      </nav>

      <!-- User profile and logout at the bottom -->
      <div class="mt-auto border-t border-accent2 dark:border-gray-800 p-4">
        <div :class="[isSidebarOpen ? 'flex items-center justify-between' : 'flex justify-center']">
          <div v-if="isSidebarOpen" class="flex items-center min-w-0">
            <div class="h-8 w-8 rounded-full bg-accent1 flex items-center justify-center text-primary flex-shrink-0">
              <font-awesome-icon icon="user" class="h-4 w-4" />
            </div>
            <div class="ml-3 min-w-0">
              <p class="text-sm font-medium text-text-primary dark:text-white truncate">
                {{ authStore.userFullName || 'User' }}
              </p>
              <p class="text-xs text-text-secondary dark:text-gray-400 truncate">
                {{ authStore.user?.email || authStore.userRole || 'User' }}
              </p>
            </div>
          </div>
          <button
            @click="handleLogout"
            class="text-text-secondary hover:text-accent1 transition-colors p-1 flex-shrink-0 dark:text-gray-400 dark:hover:text-accent1"
            aria-label="Logout"
          >
            <font-awesome-icon icon="sign-out-alt" class="h-5 w-5" />
          </button>
        </div>
      </div>
    </aside>

    <!-- Main content -->
    <div class="flex-1 flex flex-col w-full">
      <!-- Top navbar -->
      <header class="sticky top-0 bg-white dark:bg-gray-900 shadow-sm z-10">
        <div class="px-3 sm:px-4 lg:px-6 py-3 md:py-4 flex items-center justify-between">
          <!-- Mobile menu button -->
          <button
            @click="toggleSidebar"
            class="md:hidden text-primary dark:text-white mr-2"
            aria-label="Menu"
          >
            <font-awesome-icon icon="bars" class="h-5 w-5" />
          </button>

          <h1 class="text-base sm:text-lg md:text-xl font-semibold text-primary dark:text-white truncate">
            Sales & Customer Management System
          </h1>

          <div class="flex items-center space-x-4">
            <!-- Placeholder div to maintain flex layout -->
          </div>
        </div>
      </header>

      <!-- Page content -->
      <main class="flex-1 bg-bg-alt dark:bg-gray-900 p-2 sm:p-3 md:p-4 lg:p-6">
        <div class="max-w-full mx-auto w-full">
          <!-- Dynamic component with transition and loading indicator -->
          <transition
            name="component-fade"
            mode="out-in"
          >
            <div v-if="!currentComponent" class="flex justify-center items-center py-12">
              <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-accent1"></div>
            </div>
            <component v-else :is="currentComponent" :key="activeComponent"></component>
          </transition>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
/* Transition animations */
.component-fade-enter-active,
.component-fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.component-fade-enter-from {
  opacity: 0;
  transform: translateY(20px);
}

.component-fade-leave-to {
  opacity: 0;
  transform: translateY(-20px);
}

/* Sidebar transition */
aside {
  will-change: transform, width;
}

/* Mobile optimizations */
@media (max-width: 640px) {
  .component-fade-enter-active,
  .component-fade-leave-active {
    transition: opacity 0.2s ease;
  }

  .component-fade-enter-from,
  .component-fade-leave-to {
    transform: none;
  }
}

/* Ensure sidebar content is always visible */
aside > div {
  min-height: 0;
}

/* Prevent horizontal scrolling */
:deep(.overflow-x-auto) {
  max-width: 100%;
}
</style>
