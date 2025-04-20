<script setup lang="ts">
import { ref, defineAsyncComponent } from 'vue';
import { library } from '@fortawesome/fontawesome-svg-core';
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome';
import {
  faChartLine,
  faUsers,
  faShoppingCart,
  faBoxes,
  faFileInvoice,
  faBars,
  faBell,
  faSearch,
  faUser
} from '@fortawesome/free-solid-svg-icons';


const Dashboard = defineAsyncComponent(() => import('../components/Dashboard.vue'));
const Customers = defineAsyncComponent(() => import('../components/Customers.vue'));
const Sales = defineAsyncComponent(() => import('../components/Sales.vue'));
const Inventory = defineAsyncComponent(() => import('../components/Inventory.vue'));
const Quotation = defineAsyncComponent(() => import('../components/Quotation.vue'));

library.add(
  faChartLine,
  faUsers,
  faShoppingCart,
  faBoxes,
  faFileInvoice,
  faBars,
  faBell,
  faSearch,
  faUser
);

const isSidebarOpen = ref(true);
const toggleSidebar = () => {
  isSidebarOpen.value = !isSidebarOpen.value;
};

const activeMenuItem = ref('Dashboard');
const setActiveMenuItem = (menuItem: string) => {
  activeMenuItem.value = menuItem;
};


const activeComponent = ref(Dashboard);
const navigate = (menuItem: string) => {
  activeMenuItem.value = menuItem;
  switch(menuItem) {
    case 'Dashboard':
      activeComponent.value = Dashboard;
      break;
    case 'Customers':
      activeComponent.value = Customers;
      break;
    case 'Sales':
      activeComponent.value = Sales;
      break;
    case 'Inventory':
      activeComponent.value = Inventory;
      break;
    case 'Quotation':
      activeComponent.value = Quotation;
      break;
    default:
      activeComponent.value = Dashboard;
  }
};
</script>

<template>
  <div class="flex h-screen bg-gray-100">
    <!-- Sidebar -->
    <div :class="`bg-primary text-white ${isSidebarOpen ? 'w-64 md:w-55' : 'w-20'} transition-all duration-300 ease-in-out`">
      <!-- Logo -->
      <div class="flex items-center justify-between p-4 border-b border-gray-700">
        <div class="flex items-center space-x-2">
          <img src="../assets/CenterLogo.svg" alt="Logo" class="h-8 w-8" />
          <span v-if="isSidebarOpen" class="text-xl font-bold">ManagerPro</span>
        </div>
        <button @click="toggleSidebar" class="text-white hover:text-accent1 p-1">
          <font-awesome-icon icon="bars" />
        </button>
      </div>

      <!-- Navigation -->
      <nav class="mt-6">
        <div
          class="flex items-center px-4 py-3 cursor-pointer transition-colors duration-200"
          :class="activeMenuItem === 'Dashboard' ? 'bg-accent2 text-white' : 'text-gray-300 hover:bg-accent2 hover:text-white'"
          @click="navigate('Dashboard')"
        >
          <font-awesome-icon icon="chart-line" class="mr-3" />
          <span v-if="isSidebarOpen">Dashboard</span>
        </div>

        <div
          class="flex items-center px-4 py-3 cursor-pointer transition-colors duration-200 text-gray-300 hover:bg-accent2 hover:text-white"
          @click="navigate('Customers')"
        >
          <font-awesome-icon icon="users" class="mr-3" />
          <span v-if="isSidebarOpen">Customers</span>
        </div>

        <div
          class="flex items-center px-4 py-3 cursor-pointer transition-colors duration-200 text-gray-300 hover:bg-accent2 hover:text-white"
          @click="navigate('Sales')"
        >
          <font-awesome-icon icon="shopping-cart" class="mr-3" />
          <span v-if="isSidebarOpen">Sales</span>
        </div>

        <div
          class="flex items-center px-4 py-3 cursor-pointer transition-colors duration-200 text-gray-300 hover:bg-accent2 hover:text-white"
          @click="navigate('Inventory')"
        >
          <font-awesome-icon icon="boxes" class="mr-3" />
          <span v-if="isSidebarOpen">Inventory</span>
        </div>

        <div
          class="flex items-center px-4 py-3 cursor-pointer transition-colors duration-200 text-gray-300 hover:bg-accent2 hover:text-white"
          @click="navigate('Quotation')"
        >
          <font-awesome-icon icon="file-invoice" class="mr-3" />
          <span v-if="isSidebarOpen">Quotation</span>
        </div>
      </nav>
    </div>


    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- Header -->
      <header class="bg-white shadow-sm z-10">
        <div class="flex items-center justify-between p-4">
          <div class="flex items-center w-1/3">
            <div class="relative w-full max-w-md">
              <input
                type="text"
                placeholder="Search..."
                class="w-full pl-10 pr-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-accent1"
              />
              <font-awesome-icon icon="search" class="absolute left-3 top-3 text-gray-400" />
            </div>
          </div>

          <div class="flex items-center space-x-4">
            <button class="relative p-2 text-gray-500 hover:text-accent1 focus:outline-none">
              <font-awesome-icon icon="bell" />
              <span class="absolute top-0 right-0 h-2 w-2 rounded-full bg-red-500"></span>
            </button>

            <div class="flex items-center space-x-2">
              <div class="w-8 h-8 rounded-full bg-gray-300 flex items-center justify-center text-sm font-medium text-gray-700">
                <font-awesome-icon icon="user" />
              </div>
              <div class="text-sm">
                <p class="font-medium text-gray-700">Elgin Karl</p>
                <p class="text-gray-500 text-xs">Admin</p>
              </div>
            </div>
          </div>
        </div>
      </header>

      <!-- Main Content Area -->
      <main class="flex-1 overflow-y-auto bg-gray-100">

        <component :is="activeComponent" @navigate="setActiveMenuItem" />
      </main>
    </div>
  </div>
</template>
