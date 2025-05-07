<script setup lang="ts">
import { ref, shallowRef, onMounted, onBeforeMount, onUnmounted, defineAsyncComponent } from 'vue'
import type { Component } from 'vue'

// Lazy load components
const DashboardComponent = defineAsyncComponent(() => import('../components/DashboardContent.vue'))
const CustomersComponent = defineAsyncComponent(() => import('../components/CustomerList.vue'))
const SalesOrdersComponent = defineAsyncComponent(() => import('../components/SalesOrdersPage.vue'))
const InventoryComponent = defineAsyncComponent(() => import('../components/InventoryList.vue'))
const QuotationsComponent = defineAsyncComponent(() => import('../components/QuotationList.vue'))
const ReportsComponent = defineAsyncComponent(() => import('../components/ReportsPage.vue'))
const SettingsComponent = defineAsyncComponent(() => import('../components/SettingsArea.vue'))

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

const navigationItems: NavigationItem[] = [
  { id: 'dashboard', name: 'Dashboard', icon: 'home' },
  { id: 'customers', name: 'Customers', icon: 'users' },
  { id: 'sales', name: 'Sales & Orders', icon: 'clipboard-list' },
  { id: 'inventory', name: 'Inventory', icon: 'box' },
  { id: 'quotations', name: 'Quotations', icon: 'file-invoice-dollar' },
  { id: 'reports', name: 'Reports', icon: 'chart-line' },
  { id: 'settings', name: 'Settings', icon: 'cog' },
]

const componentMap: ComponentMap = {
  dashboard: DashboardComponent,
  customers: CustomersComponent,
  sales: SalesOrdersComponent,
  inventory: InventoryComponent,
  quotations: QuotationsComponent,
  reports: ReportsComponent,
  settings: SettingsComponent
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
  <div class="min-h-screen bg-white flex flex-col md:flex-row overflow-hidden">
    <!-- Overlay for mobile when sidebar is open -->
    <div
      v-if="isMobile && isSidebarOpen"
      @click="toggleSidebar"
      class="fixed inset-0 bg-black bg-opacity-50 z-20"
    ></div>
    
    <!-- Sidebar -->
    <aside
      :class="[
        'bg-primary text-text-primary transition-all duration-300 ease-in-out',
        'z-30 h-screen flex-shrink-0',
        isMobile ? 'fixed left-0 top-0 bottom-0' : 'sticky top-0',
        isSidebarOpen ? 'translate-x-0 w-64' : isMobile ? '-translate-x-full' : 'w-16'
      ]"
    >
      <!-- Sidebar header -->
      <div class="flex items-center justify-between h-16 px-4 border-b border-accent2">
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
            activeComponent === item.id ? 'bg-accent2 text-white' : 'text-text-secondary hover:bg-accent2 hover:bg-opacity-25 hover:text-text-primary',
            'flex items-center rounded-md text-sm font-medium transition-colors duration-150',
            isSidebarOpen ? 'px-4 py-3' : 'justify-center px-3 py-3'
          ]"
          :title="!isSidebarOpen ? item.name : ''"
        >
          <font-awesome-icon :icon="item.icon" :class="['h-5 w-5', isSidebarOpen && 'mr-3']" />
          <span v-if="isSidebarOpen">{{ item.name }}</span>
        </a>
      </nav>

      <!-- User profile and logout -->
      <div class="border-t border-accent2 p-4">
        <div :class="[isSidebarOpen ? 'flex items-center justify-between' : 'flex justify-center']">
          <div v-if="isSidebarOpen" class="flex items-center min-w-0">
            <div class="h-8 w-8 rounded-full bg-accent1 flex items-center justify-center text-primary flex-shrink-0">
              <font-awesome-icon icon="user" class="h-4 w-4" />
            </div>
            <div class="ml-3 min-w-0">
              <p class="text-sm font-medium text-text-primary truncate">Admin User</p>
              <p class="text-xs text-text-secondary truncate">administrator</p>
            </div>
          </div>
          <button 
            class="text-text-secondary hover:text-accent1 transition-colors p-1 flex-shrink-0"
            aria-label="Logout"
          >
            <font-awesome-icon icon="sign-out-alt" class="h-5 w-5" />
          </button>
        </div>
      </div>
    </aside>

    <!-- Main content -->
    <div class="flex-1 flex flex-col overflow-hidden w-full">
      <!-- Top navbar -->
      <header class="bg-white shadow-sm z-10">
        <div class="px-3 sm:px-4 lg:px-6 py-3 md:py-4 flex items-center justify-between">
          <!-- Mobile menu button -->
          <button 
            @click="toggleSidebar"
            class="md:hidden text-primary mr-2"
            aria-label="Menu"
          >
            <font-awesome-icon icon="bars" class="h-5 w-5" />
          </button>
          
          <h1 class="text-base sm:text-lg md:text-xl font-semibold text-primary truncate">
            Sales & Customer Management System
          </h1>
          
          <div class="flex items-center space-x-4">
            <button 
              class="text-primary hover:text-accent1 transition-colors"
              aria-label="Settings"
            >
              <font-awesome-icon icon="cog" class="h-5 w-5" />
            </button>
          </div>
        </div>
      </header>

      <!-- Page content -->
      <main class="flex-1 overflow-y-auto overflow-x-hidden bg-bg-alt p-2 sm:p-3 md:p-4 lg:p-6">
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
