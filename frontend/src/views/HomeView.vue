<script setup lang="ts">
import { ref, shallowRef, onMounted, defineAsyncComponent } from 'vue'
import type { Component } from 'vue'

// Lazy load components
const DashboardComponent = defineAsyncComponent(() => import('@/components/DashboardContent.vue'))
const CustomersComponent = defineAsyncComponent(() => import('@/components/CustomerList.vue'))
const SalesOrdersComponent = defineAsyncComponent(() => import('@/components/SalesOrdersPage.vue'))
const InventoryComponent = defineAsyncComponent(() => import('@/components/InventoryList.vue'))
const QuotationsComponent = defineAsyncComponent(() => import('@/components/QuotationList.vue'))
const ReportsComponent = defineAsyncComponent(() => import('@/components/ReportsPage.vue'))
const SettingsComponent = defineAsyncComponent(() => import('@/components/SettingsArea.vue'))

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
const isSidebarOpen = ref<boolean>(false)
const activeComponent = ref<string>('dashboard')
const currentComponent = shallowRef<Component>(DashboardComponent)

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
  if (window.innerWidth < 1024) {
    isSidebarOpen.value = false
  }
  
  activeComponent.value = componentId
  currentComponent.value = componentMap[componentId]
}

onMounted(() => {
  // Set initial component
  currentComponent.value = componentMap[activeComponent.value]
})
</script>

<template>
  <div class="min-h-screen bg-white flex">
    <!-- Main container with sidebar and content -->
    <div class="flex relative w-full overflow-hidden">
      <!-- Sidebar -->
      <div 
        :class="[
          'bg-primary text-text-primary transition-all duration-300 ease-in-out min-h-screen',
          isSidebarOpen ? 'w-64' : 'w-16'
        ]"
      >
        <!-- Expanded sidebar content -->
        <div class="h-full flex flex-col">
          <!-- Logo and close button -->
          <div :class="[
            'flex border-b border-accent2 p-4',
            isSidebarOpen ? 'items-center justify-between' : 'justify-center'
          ]">
            <div :class="[isSidebarOpen ? 'flex items-center space-x-2' : 'hidden']">
              <span class="text-accent1 text-2xl font-bold">SCMS</span>
            </div>
            <button 
              @click="toggleSidebar" 
              class="text-text-primary hover:text-accent1 transition-colors"
            >
              <font-awesome-icon :icon="isSidebarOpen ? 'times' : 'bars'" class="h-5 w-5" />
            </button>
          </div>

          <!-- Navigation -->
          <nav class="flex-1 px-2 py-4 space-y-1 overflow-y-auto">
            <a 
              v-for="item in navigationItems" 
              :key="item.name" 
              href="#"
              @click.prevent="switchComponent(item.id)"
              :class="[
                activeComponent === item.id ? 'bg-accent2 text-white' : 'text-text-secondary hover:bg-accent2 hover:bg-opacity-25 hover:text-text-primary',
                'flex items-center rounded-md text-sm font-medium transition-colors duration-150',
                isSidebarOpen ? 'px-4 py-3' : 'px-3 py-3 justify-center'
              ]"
              :title="!isSidebarOpen ? item.name : ''"
            >
              <font-awesome-icon :icon="item.icon" :class="['h-5 w-5', isSidebarOpen && 'mr-3']" />
              <span :class="[!isSidebarOpen && 'hidden']">{{ item.name }}</span>
            </a>
          </nav>

          <!-- User profile and logout -->
          <div class="p-4 border-t border-accent2">
            <div :class="[isSidebarOpen ? 'flex items-center justify-between' : 'flex justify-center']">
              <div :class="[isSidebarOpen ? 'flex items-center' : 'hidden']">
                <div class="h-8 w-8 rounded-full bg-accent1 flex items-center justify-center text-primary">
                  <font-awesome-icon icon="user" class="h-4 w-4" />
                </div>
                <div class="ml-3">
                  <p class="text-sm font-medium text-text-primary">Admin User</p>
                  <p class="text-xs text-text-secondary">administrator</p>
                </div>
              </div>
              <button class="text-text-secondary hover:text-accent1 transition-colors">
                <font-awesome-icon icon="sign-out-alt" class="h-5 w-5" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Main content -->
      <div class="flex-1 flex flex-col overflow-hidden transition-all duration-300">
        <!-- Top navbar -->
        <header class="bg-white shadow-sm z-10">
          <div class="px-4 sm:px-6 lg:px-8 py-4 flex items-center justify-between">
            <div class="text-xl font-semibold text-primary">Sales & Customer Management System</div>
            <div class="flex items-center space-x-4">
              <button class="text-primary hover:text-accent1 transition-colors">
                <font-awesome-icon icon="cog" class="h-5 w-5" />
              </button>
            </div>
          </div>
        </header>

        <!-- Page content -->
        <main class="flex-1 overflow-auto bg-bg-alt p-4 sm:p-6 lg:p-8">
          <div class="max-w-7xl mx-auto">
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
</style>
