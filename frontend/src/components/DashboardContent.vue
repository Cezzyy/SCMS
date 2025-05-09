<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { Line, Bar } from 'vue-chartjs';
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, BarElement, Title, Tooltip, Legend } from 'chart.js';
import { useReportStore } from '@/stores/reportStore';

// Register Chart.js components
ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, BarElement, Title, Tooltip, Legend);

// Initialize the report store
const reportStore = useReportStore();

// Detect dark mode
const isDarkMode = ref(window.matchMedia('(prefers-color-scheme: dark)').matches);

// Update chart options based on dark/light mode
const chartOptions = computed(() => {
  return {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'top' as const,
        labels: {
          boxWidth: 15,
          usePointStyle: true,
          padding: 20,
          color: isDarkMode.value ? '#e2e8f0' : '#1f2937'
        }
      },
      tooltip: {
        backgroundColor: isDarkMode.value ? 'rgba(30, 41, 59, 0.8)' : 'rgba(0, 0, 0, 0.8)',
        titleFont: {
          size: 13
        },
        bodyFont: {
          size: 12
        },
        padding: 10,
        cornerRadius: 4,
        titleColor: '#fff',
        bodyColor: '#fff'
      }
    },
    scales: {
      y: {
        beginAtZero: true,
        grid: {
          display: true,
          drawBorder: false,
          color: isDarkMode.value ? 'rgba(75, 85, 99, 0.3)' : 'rgba(226, 232, 240, 0.5)'
        },
        ticks: {
          padding: 10,
          color: isDarkMode.value ? '#e2e8f0' : '#1f2937'
        }
      },
      x: {
        grid: {
          display: false,
          drawBorder: false
        },
        ticks: {
          padding: 10,
          color: isDarkMode.value ? '#e2e8f0' : '#1f2937'
        }
      }
    }
  };
});

// Chart data with dark mode compatible colors
const salesChartData = computed(() => {
  return {
    labels: reportStore.salesTrends.map(item => item.day),
    datasets: [
      {
        label: 'Sales',
        backgroundColor: isDarkMode.value ? 'rgba(59, 130, 246, 0.3)' : 'rgba(53, 162, 235, 0.5)',
        borderColor: isDarkMode.value ? 'rgba(59, 130, 246, 0.8)' : 'rgba(53, 162, 235, 1)',
        borderWidth: 2,
        tension: 0.4,
        pointBackgroundColor: isDarkMode.value ? 'rgba(59, 130, 246, 0.8)' : 'rgba(53, 162, 235, 1)',
        pointBorderColor: isDarkMode.value ? '#1e293b' : '#fff',
        pointBorderWidth: 1,
        pointRadius: 4,
        fill: true,
        data: reportStore.salesTrends.map(item => item.total_amount)
      }
    ]
  };
});

const inventoryChartData = computed(() => {
  return {
    labels: reportStore.lowStockItems.map(item => item.name),
    datasets: [
      {
        label: 'Current Stock',
        backgroundColor: isDarkMode.value ? 'rgba(252, 165, 165, 0.7)' : 'rgba(255, 99, 132, 0.7)',
        borderColor: isDarkMode.value ? 'rgba(239, 68, 68, 0.8)' : 'rgba(255, 99, 132, 1)',
        borderWidth: 1,
        borderRadius: 4,
        data: reportStore.lowStockItems.map(item => item.current_stock)
      },
      {
        label: 'Reorder Level',
        backgroundColor: isDarkMode.value ? 'rgba(103, 232, 249, 0.7)' : 'rgba(75, 192, 192, 0.7)',
        borderColor: isDarkMode.value ? 'rgba(8, 145, 178, 0.8)' : 'rgba(75, 192, 192, 1)',
        borderWidth: 1,
        borderRadius: 4,
        data: reportStore.lowStockItems.map(item => item.reorder_level)
      }
    ]
  };
});

// Top customers chart data
const customersChartData = computed(() => {
  return {
    labels: reportStore.topCustomersWithAvatars.map(customer => customer.name),
    datasets: [
      {
        label: 'Total Spent',
        backgroundColor: isDarkMode.value ? 'rgba(139, 92, 246, 0.7)' : 'rgba(124, 58, 237, 0.7)',
        borderColor: isDarkMode.value ? 'rgba(139, 92, 246, 0.9)' : 'rgba(124, 58, 237, 0.9)',
        borderWidth: 1,
        borderRadius: 4,
        data: reportStore.topCustomersWithAvatars.map(customer => customer.total_spent)
      }
    ]
  };
});

// Methods
const fetchAll = async () => {
  await reportStore.refreshAll({ days: reportStore.period });
};

const dismissError = (type: 'dashboard' | 'sales' | 'stock' | 'customers') => {
  reportStore.dismissError(type);
};

const downloadSalesCSV = () => {
  reportStore.downloadSalesCSV();
};

const downloadStockCSV = () => {
  reportStore.downloadStockCSV();
};

const downloadCustomersCSV = () => {
  reportStore.downloadCustomersCSV();
};

// Computed values to simplify template access
const totalSales = computed(() => reportStore.totalSales);
const numberOfOrders = computed(() => reportStore.orderCount);
const lowStockCount = computed(() => reportStore.lowStockCount);
const formattedPeriod = computed(() => reportStore.formattedPeriod);

// Loading and error state accessors
const loading = computed(() => reportStore.loading);
const error = computed(() => reportStore.error);

// Watch for system dark mode changes
onMounted(async () => {
  await fetchAll();
  
  const darkModeMediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
  const handleDarkModeChange = (e: MediaQueryListEvent) => {
    isDarkMode.value = e.matches;
  };
  
  darkModeMediaQuery.addEventListener('change', handleDarkModeChange);
});

// Refresh data when period changes
watch(() => reportStore.period, async (newPeriod) => {
  await fetchAll();
});
</script>

<template>
  <div class="dashboard-container bg-gray-100 dark:bg-gray-900 p-4 md:p-6">
    <!-- Dashboard Header -->
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-800 dark:text-gray-100">Dashboard</h1>
      <p class="text-sm text-gray-600 dark:text-gray-400 mt-1">Overview of business performance metrics</p>
    </div>

    <!-- Controls Panel -->
    <div class="flex justify-end mb-6">
      <button
        @click="fetchAll"
        class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors duration-200 shadow-sm flex items-center dark:bg-blue-700 dark:hover:bg-blue-600"
      >
        <span class="mr-2">Refresh All</span>
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
      </button>
    </div>

    <!-- KPI Cards Row -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-6 mb-6">
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md flex flex-col">
        <div class="px-6 py-4 border-b border-gray-100 dark:border-gray-700">
          <h3 class="text-sm font-medium text-gray-500 dark:text-gray-400">Total Sales</h3>
          <div class="text-3xl font-bold text-gray-800 dark:text-white mt-1">₱{{ totalSales.toLocaleString() }}</div>
        </div>
        <div class="px-6 py-3 bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 text-xs font-medium">
          Period: {{ formattedPeriod }}
        </div>
      </div>

      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md flex flex-col">
        <div class="px-6 py-4 border-b border-gray-100 dark:border-gray-700">
          <h3 class="text-sm font-medium text-gray-500 dark:text-gray-400">Number of Orders</h3>
          <div class="text-3xl font-bold text-gray-800 dark:text-white mt-1">{{ numberOfOrders }}</div>
        </div>
        <div class="px-6 py-3 bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 text-xs font-medium">
          Period: {{ formattedPeriod }}
        </div>
      </div>

      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md flex flex-col">
        <div class="px-6 py-4 border-b border-gray-100 dark:border-gray-700">
          <h3 class="text-sm font-medium text-gray-500 dark:text-gray-400">Low-Stock Items</h3>
          <div class="text-3xl font-bold text-gray-800 dark:text-white mt-1">{{ lowStockCount }}</div>
        </div>
        <div class="px-6 py-3 bg-red-50 dark:bg-red-900/30 text-red-600 dark:text-red-400 text-xs font-medium">
          Status: {{ lowStockCount > 0 ? 'Needs Attention' : 'All Good' }}
        </div>
      </div>
    </div>

    <!-- Main Content Panels - New Layout -->
    <div class="space-y-6">
      <!-- Sales Trends Panel - Full Width -->
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md flex flex-col">
        <div class="flex justify-between items-center p-4 border-b border-gray-100 dark:border-gray-700">
          <h3 class="text-lg font-semibold text-gray-800 dark:text-white">Sales Trends</h3>
          <button
            @click="downloadSalesCSV"
            class="px-3 py-1.5 bg-green-600 text-white text-sm rounded-md hover:bg-green-700 transition-colors duration-200 flex items-center dark:bg-green-700 dark:hover:bg-green-600"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
            Export CSV
          </button>
        </div>

        <!-- Error Alert -->
        <div v-if="error.sales" class="mx-4 mt-4 bg-red-50 dark:bg-red-900/20 border-l-4 border-red-500 rounded-md overflow-hidden">
          <div class="flex items-center justify-between p-3">
            <div class="flex items-center">
              <svg class="h-5 w-5 text-red-500 dark:text-red-400 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <p class="text-red-700 dark:text-red-400 text-sm font-medium">{{ error.sales }}</p>
            </div>
            <div class="flex items-center space-x-2">
              <button @click="fetchAll" class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 text-sm font-medium">Retry</button>
              <button @click="dismissError('sales')" class="text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>
        </div>

        <div class="p-4 flex-1 flex flex-col">
          <!-- Sales Chart -->
          <div class="chart-container h-64 mb-4">
            <div v-if="loading.sales" class="absolute inset-0 flex items-center justify-center bg-white dark:bg-gray-800 bg-opacity-80 dark:bg-opacity-80 z-10">
              <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-t-2 border-blue-600 dark:border-blue-500"></div>
            </div>
            <Line
              v-else-if="!error.sales && reportStore.salesTrends.length > 0"
              :data="salesChartData"
              :options="chartOptions"
              class="sales-chart"
            />
            <div v-else-if="!loading.sales && reportStore.salesTrends.length === 0" class="flex items-center justify-center h-full">
              <p class="text-gray-500 dark:text-gray-400">No sales data available for the selected period</p>
            </div>
          </div>

          <!-- Sales Data Table -->
          <div v-if="!loading.sales && !error.sales && reportStore.salesTrends.length > 0" class="border dark:border-gray-700 rounded-lg overflow-hidden bg-gray-50 dark:bg-gray-900 flex-1">
            <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
              <thead class="bg-gray-100 dark:bg-gray-800">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Date</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Sales Amount</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Trend</th>
                </tr>
              </thead>
              <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                <tr v-for="(item, index) in reportStore.salesTrends" :key="index" class="hover:bg-gray-50 dark:hover:bg-gray-700">
                  <td class="px-6 py-3 whitespace-nowrap text-sm font-medium text-gray-800 dark:text-gray-200">{{ item.day }}</td>
                  <td class="px-6 py-3 whitespace-nowrap text-sm text-gray-600 dark:text-gray-300">₱{{ item.total_amount.toLocaleString() }}</td>
                  <td class="px-6 py-3 whitespace-nowrap text-sm">
                    <div v-if="index > 0" class="flex items-center">
                      <span
                        :class="[
                          'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                          item.total_amount > reportStore.salesTrends[index-1].total_amount
                            ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
                            : item.total_amount < reportStore.salesTrends[index-1].total_amount
                              ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300'
                              : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
                        ]"
                      >
                        <svg
                          v-if="item.total_amount > reportStore.salesTrends[index-1].total_amount"
                          class="h-3 w-3 mr-1"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                        >
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
                        </svg>
                        <svg
                          v-else-if="item.total_amount < reportStore.salesTrends[index-1].total_amount"
                          class="h-3 w-3 mr-1"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                        >
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                        </svg>
                        <span v-if="item.total_amount > reportStore.salesTrends[index-1].total_amount">Up</span>
                        <span v-else-if="item.total_amount < reportStore.salesTrends[index-1].total_amount">Down</span>
                        <span v-else>No change</span>
                      </span>
                    </div>
                    <span v-else class="text-gray-500 dark:text-gray-400 text-xs">First day</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Two Column Layout for Inventory and Top Customers -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Inventory Status Panel -->
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md flex flex-col">
          <div class="flex justify-between items-center p-4 border-b border-gray-100 dark:border-gray-700">
            <h3 class="text-lg font-semibold text-gray-800 dark:text-white">Inventory Status</h3>
            <button
              @click="downloadStockCSV"
              class="px-3 py-1.5 bg-green-600 text-white text-sm rounded-md hover:bg-green-700 transition-colors duration-200 flex items-center dark:bg-green-700 dark:hover:bg-green-600"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              Export CSV
            </button>
          </div>

          <!-- Error Alert -->
          <div v-if="error.stock" class="mx-4 mt-4 bg-red-50 dark:bg-red-900/20 border-l-4 border-red-500 rounded-md overflow-hidden">
            <div class="flex items-center justify-between p-3">
              <div class="flex items-center">
                <svg class="h-5 w-5 text-red-500 dark:text-red-400 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <p class="text-red-700 dark:text-red-400 text-sm font-medium">{{ error.stock }}</p>
              </div>
              <div class="flex items-center space-x-2">
                <button @click="fetchAll" class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 text-sm font-medium">Retry</button>
                <button @click="dismissError('stock')" class="text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
          </div>

          <div class="p-4 flex-1 flex flex-col">
            <!-- Chart -->
            <div class="chart-container h-52 mb-4">
              <div v-if="loading.stock" class="absolute inset-0 flex items-center justify-center bg-white dark:bg-gray-800 bg-opacity-80 dark:bg-opacity-80 z-10">
                <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-t-2 border-blue-600 dark:border-blue-500"></div>
              </div>
              <Bar
                v-else-if="!error.stock && reportStore.lowStockItems.length > 0"
                :data="inventoryChartData"
                :options="chartOptions"
                class="inventory-chart"
              />
              <div v-else-if="!loading.stock && reportStore.lowStockItems.length === 0" class="flex items-center justify-center h-full">
                <p class="text-gray-500 dark:text-gray-400">No low stock items found</p>
              </div>
            </div>

            <!-- Inventory Table -->
            <div v-if="!loading.stock && !error.stock && reportStore.lowStockItems.length > 0" class="border dark:border-gray-700 rounded-lg overflow-hidden bg-gray-50 dark:bg-gray-900 flex-1">
              <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                <thead class="bg-gray-100 dark:bg-gray-800">
                  <tr>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Product Name</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Current Stock</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Reorder Level</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                  </tr>
                </thead>
                <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                  <tr v-for="item in reportStore.lowStockItems" :key="item.id" 
                      :class="{ 
                        'bg-red-50 dark:bg-red-900/20': item.current_stock < item.reorder_level,
                        'hover:bg-gray-50 dark:hover:bg-gray-700': item.current_stock >= item.reorder_level
                      }">
                    <td class="px-6 py-3 whitespace-nowrap text-sm font-medium text-gray-800 dark:text-gray-200">{{ item.name }}</td>
                    <td class="px-6 py-3 whitespace-nowrap text-sm text-gray-600 dark:text-gray-300">{{ item.current_stock }}</td>
                    <td class="px-6 py-3 whitespace-nowrap text-sm text-gray-600 dark:text-gray-300">{{ item.reorder_level }}</td>
                    <td class="px-6 py-3 whitespace-nowrap">
                      <span
                        :class="[
                          'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                          item.current_stock < item.reorder_level
                            ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300'
                            : 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
                        ]"
                      >
                        {{ item.current_stock < item.reorder_level ? 'Reorder' : 'OK' }}
                      </span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- Top Customers Panel -->
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md flex flex-col">
          <div class="flex justify-between items-center p-4 border-b border-gray-100 dark:border-gray-700">
            <h3 class="text-lg font-semibold text-gray-800 dark:text-white">Top Customers</h3>
            <button
              @click="downloadCustomersCSV"
              class="px-3 py-1.5 bg-green-600 text-white text-sm rounded-md hover:bg-green-700 transition-colors duration-200 flex items-center dark:bg-green-700 dark:hover:bg-green-600"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              Export CSV
            </button>
          </div>

          <!-- Error Alert -->
          <div v-if="error.customers" class="mx-4 mt-4 bg-red-50 dark:bg-red-900/20 border-l-4 border-red-500 rounded-md overflow-hidden">
            <div class="flex items-center justify-between p-3">
              <div class="flex items-center">
                <svg class="h-5 w-5 text-red-500 dark:text-red-400 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <p class="text-red-700 dark:text-red-400 text-sm font-medium">{{ error.customers }}</p>
              </div>
              <div class="flex items-center space-x-2">
                <button @click="fetchAll" class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 text-sm font-medium">Retry</button>
                <button @click="dismissError('customers')" class="text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
          </div>

          <div class="p-4 flex-1 flex flex-col">
            <!-- Chart -->
            <div class="chart-container h-52 mb-4">
              <div v-if="loading.customers" class="absolute inset-0 flex items-center justify-center bg-white dark:bg-gray-800 bg-opacity-80 dark:bg-opacity-80 z-10">
                <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-t-2 border-blue-600 dark:border-blue-500"></div>
              </div>
              <Bar
                v-else-if="!error.customers && reportStore.topCustomersWithAvatars.length > 0"
                :data="customersChartData"
                :options="chartOptions"
                class="customers-chart"
              />
              <div v-else-if="!loading.customers && reportStore.topCustomersWithAvatars.length === 0" class="flex items-center justify-center h-full">
                <p class="text-gray-500 dark:text-gray-400">No customer data available</p>
              </div>
            </div>

            <!-- Top Customers Table -->
            <div v-if="!loading.customers && !error.customers && reportStore.topCustomersWithAvatars.length > 0" class="border dark:border-gray-700 rounded-lg overflow-hidden bg-gray-50 dark:bg-gray-900 flex-1">
              <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                <thead class="bg-gray-100 dark:bg-gray-800">
                  <tr>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Customer</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Total Spent</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Orders</th>
                  </tr>
                </thead>
                <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                  <tr v-for="customer in reportStore.topCustomersWithAvatars" :key="customer.id" class="hover:bg-gray-50 dark:hover:bg-gray-700">
                    <td class="px-6 py-3 whitespace-nowrap">
                      <div class="flex items-center">
                        <div class="flex-shrink-0 h-8 w-8 rounded-full bg-blue-100 dark:bg-blue-900 flex items-center justify-center text-lg">
                          {{ customer.avatar }}
                        </div>
                        <div class="ml-4">
                          <div class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ customer.name }}</div>
                          <div v-if="customer.contact_name" class="text-xs text-gray-500 dark:text-gray-400">{{ customer.contact_name }}</div>
                        </div>
                      </div>
                    </td>
                    <td class="px-6 py-3 whitespace-nowrap text-sm text-gray-600 dark:text-gray-300">₱{{ customer.total_spent.toLocaleString() }}</td>
                    <td class="px-6 py-3 whitespace-nowrap">
                      <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300">
                        {{ customer.orders }}
                      </span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard-container {
  min-height: calc(100vh - 70px);
}

.chart-container {
  position: relative;
}

/* Improved chart styling */
:deep(.inventory-chart),
:deep(.sales-chart),
:deep(.customers-chart) {
  border-radius: 8px;
}

/* Dark mode canvas adjustment */
:deep(.dark canvas) {
  filter: brightness(0.9);
}

/* Table improvements */
:deep(table) {
  width: 100%;
  table-layout: fixed;
}

/* Responsive tweaks */
@media (max-width: 768px) {
  .dashboard-container {
    padding: 1rem;
  }
  
  .chart-container {
    height: 250px !important;
  }
}
</style>
