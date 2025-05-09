<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import type { Order } from '../types/Order';
import { defineAsyncComponent } from 'vue';

const NewOrderModal = defineAsyncComponent(() => import('./NewOrderModal.vue'));
const ViewOrderModal = defineAsyncComponent(() => import('./ViewOrderModal.vue'));

// State
const isLoading = ref(false);
const searchQuery = ref('');
const orders = ref<Order[]>([]);
const orderToView = ref<Order | null>(null);
const showNewOrderModal = ref(false);

// Sample customer data (replace with API call in real app)
const customers = ref([
  { customer_id: 101, name: 'Alice Wonderland' },
  { customer_id: 102, name: 'Bob The Builder' },
  { customer_id: 103, name: 'Charlie Brown' },
  { customer_id: 104, name: 'Diana Prince' },
  { customer_id: 105, name: 'Edward Scissorhands' }
]);

// Computed properties for order counts
const pendingOrdersCount = computed(() => {
  return orders.value.filter(order => order.status === 'Pending').length;
});

const completedOrdersCount = computed(() => {
  return orders.value.filter(order => order.status === 'Delivered').length;
});

// Sample Data (replace with API call)
const sampleOrders: Order[] = [
  {
    order_id: 1,
    customer_id: 101,
    order_date: '2023-01-15',
    shipping_address: '123 Wonderland St',
    status: 'Delivered',
    total_amount: 150.75,
    created_at: '2023-01-15T10:30:00',
    updated_at: '2023-01-15T10:30:00'
  },
  {
    order_id: 2,
    customer_id: 102,
    order_date: '2023-02-20',
    shipping_address: '456 Builder Ave',
    status: 'Shipped',
    total_amount: 89.99,
    created_at: '2023-02-20T14:20:00',
    updated_at: '2023-02-20T14:20:00'
  },
  {
    order_id: 3,
    customer_id: 103,
    order_date: '2023-03-05',
    shipping_address: '789 Brown Rd',
    status: 'Pending',
    total_amount: 220.00,
    created_at: '2023-03-05T09:15:00',
    updated_at: '2023-03-05T09:15:00'
  },
  {
    order_id: 4,
    customer_id: 104,
    order_date: '2023-03-10',
    shipping_address: '101 Prince Blvd',
    status: 'Pending',
    total_amount: 45.50,
    created_at: '2023-03-10T16:45:00',
    updated_at: '2023-03-10T16:45:00'
  },
  {
    order_id: 5,
    customer_id: 105,
    order_date: '2023-03-12',
    shipping_address: '202 Scissorhands Ln',
    status: 'Cancelled',
    total_amount: 12.00,
    created_at: '2023-03-12T11:10:00',
    updated_at: '2023-03-12T11:10:00'
  },
];

// Sample order items data (in a real app, would be fetched from API)
const sampleOrderItems = [
  {
    order_item_id: 1,
    product_id: 1,
    product_name: 'Product A',
    quantity: 2,
    unit_price: 50.00,
    discount: 0,
    line_total: 100.00
  },
  {
    order_item_id: 2,
    product_id: 2,
    product_name: 'Product B',
    quantity: 1,
    unit_price: 75.00,
    discount: 10.00,
    line_total: 65.00
  }
];

// Load data function
const loadOrders = async () => {
  isLoading.value = true;
  // Simulate API call
  await new Promise(resolve => setTimeout(resolve, 1000));
  orders.value = sampleOrders;
  isLoading.value = false;
};

onMounted(() => {
  loadOrders();
});

// Filtered orders
const filteredOrders = computed(() => {
  if (!searchQuery.value) {
    return orders.value;
  }
  const query = searchQuery.value.toLowerCase();
  return orders.value.filter(order =>
    order.order_id.toString().includes(query) ||
    order.customer_id.toString().includes(query)
  );
});

// Pagination (basic example, can be enhanced)
const currentPage = ref(1);
const itemsPerPage = 5;
const totalPages = computed(() => Math.ceil(filteredOrders.value.length / itemsPerPage));

const paginatedOrders = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage;
  const end = start + itemsPerPage;
  return filteredOrders.value.slice(start, end);
});

const goToPage = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page;
  }
};

// Format money (example)
const formatMoney = (amount: number): string => {
  return '₱' + amount.toFixed(2);
};

// Get status class for styling
const getStatusClass = (status: Order['status']) => {
  switch (status) {
    case 'Pending': return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-700 dark:text-yellow-100';
    case 'Shipped': return 'bg-indigo-100 text-indigo-800 dark:bg-indigo-700 dark:text-indigo-100';
    case 'Delivered': return 'bg-green-100 text-green-800 dark:bg-green-700 dark:text-green-100';
    case 'Cancelled': return 'bg-red-100 text-red-800 dark:bg-red-700 dark:text-red-100';
    default: return 'bg-gray-100 text-gray-800 dark:bg-gray-600 dark:text-gray-200';
  }
};

const openNewOrder = () => {
  showNewOrderModal.value = true;
};

const viewOrder = (order: Order) => {
  orderToView.value = order;
};

// Event handler for saving a new order
const handleSaveOrder = (newOrder: Order) => {
  // In a real app, this would send the order to an API
  console.log('Saving new order:', newOrder);

  // Add the new order to the list with a generated ID
  const maxId = Math.max(...orders.value.map(o => o.order_id));
  const orderWithId = {
    ...newOrder,
    order_id: maxId + 1
  };

  orders.value.push(orderWithId);
  showNewOrderModal.value = false;
};

// Function to get customer name (since we only have customer_id in the Order interface)
const getCustomerName = (customerId: number): string => {
  const customer = customers.value.find(c => c.customer_id === customerId);
  return customer ? customer.name : `Customer ${customerId}`;
};

// Format date from ISO string to readable format
const formatDate = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });
};
</script>


<template>
  <div class="bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:justify-between md:items-center p-4 md:p-6 border-b border-gray-200 dark:border-gray-700 bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-gray-800 dark:to-gray-900">
      <div class="mb-4 md:mb-0">
        <h2 class="text-xl md:text-2xl font-bold text-gray-800 dark:text-white mb-1">Order Management</h2>
        <p class="text-gray-600 dark:text-gray-300 text-xs md:text-sm">Manage and track all customer orders efficiently.</p>
      </div>
      <div class="flex flex-wrap gap-3 sm:gap-4">
        <button
          @click="openNewOrder"
          class="bg-blue-600 text-white px-3 py-2 md:px-4 md:py-2 rounded-md hover:bg-blue-700 transition-colors duration-200 flex items-center shadow-sm text-sm md:text-base"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
          New Order
        </button>
      </div>
    </div>

    <!-- Stats summary cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3 p-4 bg-gray-50 dark:bg-gray-900">
      <div class="bg-white dark:bg-gray-800 p-3 md:p-4 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700">
        <div class="flex items-center">
          <div class="p-2 md:p-3 rounded-full bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-300 mr-3 md:mr-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
            </svg>
          </div>
          <div>
            <div class="text-xs md:text-sm font-medium text-gray-500 dark:text-gray-400">Total Orders</div>
            <div class="text-lg md:text-2xl font-semibold text-gray-800 dark:text-white">
              {{ orders.length }} <!-- Placeholder, update with actual data -->
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white dark:bg-gray-800 p-3 md:p-4 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700">
        <div class="flex items-center">
          <div class="p-2 md:p-3 rounded-full bg-yellow-100 dark:bg-yellow-900 text-yellow-600 dark:text-yellow-300 mr-3 md:mr-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </div>
          <div>
            <div class="text-xs md:text-sm font-medium text-gray-500 dark:text-gray-400">Pending Orders</div>
            <div class="text-lg md:text-2xl font-semibold text-gray-800 dark:text-white">
              {{ pendingOrdersCount }} <!-- Placeholder, update with actual data -->
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white dark:bg-gray-800 p-3 md:p-4 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700 sm:col-span-2 md:col-span-1">
        <div class="flex items-center">
          <div class="p-2 md:p-3 rounded-full bg-green-100 dark:bg-green-900 text-green-600 dark:text-green-300 mr-3 md:mr-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div>
            <div class="text-xs md:text-sm font-medium text-gray-500 dark:text-gray-400">Completed Orders</div>
            <div class="text-lg md:text-2xl font-semibold text-gray-800 dark:text-white">
              {{ completedOrdersCount }} <!-- Placeholder, update with actual data -->
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Search bar -->
    <div class="p-3 md:p-4 border-b border-gray-200 dark:border-gray-700">
      <div class="relative">
        <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
          <svg class="h-4 w-4 md:h-5 md:w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd" />
          </svg>
        </div>
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search by Order ID, Customer..."
          class="w-full pl-8 md:pl-10 px-3 md:px-4 py-2 md:py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all duration-200 dark:bg-gray-700 dark:text-white text-xs md:text-sm"
        />
      </div>
    </div>

    <!-- Loading Indicator -->
    <div v-if="isLoading" class="flex justify-center items-center p-12">
      <div class="relative">
        <div class="animate-spin rounded-full h-12 w-12 border-4 border-gray-200 dark:border-gray-600"></div>
        <div class="animate-spin rounded-full h-12 w-12 border-t-4 border-blue-600 dark:border-blue-500 absolute top-0 left-0"></div>
      </div>
      <div class="ml-4 text-gray-600 dark:text-gray-300 text-sm font-medium">Loading orders...</div>
    </div>

    <!-- Orders Table -->
    <div v-else class="p-4 overflow-x-auto">
      <div v-if="filteredOrders.length === 0 && searchQuery" class="flex flex-col items-center justify-center py-12 rounded-lg bg-gray-50 dark:bg-gray-700">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 text-gray-300 dark:text-gray-600 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <p class="text-gray-600 dark:text-gray-300 mb-2">No results found for "{{ searchQuery }}"</p>
        <p class="text-gray-500 dark:text-gray-400 text-sm mb-4">Try adjusting your search criteria.</p>
      </div>
      <div v-else-if="filteredOrders.length === 0" class="flex flex-col items-center justify-center py-12 rounded-lg bg-gray-50 dark:bg-gray-700">
         <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 text-gray-300 dark:text-gray-600 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
        </svg>
        <p class="text-gray-600 dark:text-gray-300 mb-2">No orders found</p>
        <p class="text-gray-500 dark:text-gray-400 text-sm mb-4">Create a new order to get started.</p>
        <button
          @click="openNewOrder"
          class="px-4 py-2 bg-blue-600 text-white text-sm rounded-md hover:bg-blue-700 transition-colors duration-200 flex items-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
          Create First Order
        </button>
      </div>

      <table v-else class="min-w-full divide-y divide-gray-200 dark:divide-gray-700 shadow-sm">
        <thead class="bg-gray-50 dark:bg-gray-700">
          <tr>
            <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider">Order ID</th>
            <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider">Customer</th>
            <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider">Order Date</th>
            <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider">Status</th>
            <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider">Total Amount</th>
            <th class="px-6 py-3.5 text-center text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider sticky right-0 bg-gray-50 dark:bg-gray-700 z-10">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
          <tr v-for="order in paginatedOrders" :key="order.order_id" class="hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors duration-150">
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-800 dark:text-gray-200">
              ORD{{ order.order_id.toString().padStart(3, '0') }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-800 dark:text-gray-200">
              {{ getCustomerName(order.customer_id) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-800 dark:text-gray-200">
              {{ formatDate(order.order_date) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span class="px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full" :class="getStatusClass(order.status)">
                {{ order.status }}
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-800 dark:text-gray-200">
              {{ formatMoney(order.total_amount) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium sticky right-0 bg-white dark:bg-gray-800 z-10">
              <button
                @click="viewOrder(order)"
                class="text-blue-600 dark:text-blue-400 hover:text-blue-900 dark:hover:text-blue-200 transition-colors duration-200"
              >
                View
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="flex justify-center mt-6">
        <nav class="inline-flex rounded-md shadow-sm -space-x-px" aria-label="Pagination">
          <button
            @click="goToPage(currentPage - 1)"
            :disabled="currentPage === 1"
            class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-sm font-medium text-gray-500 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span class="sr-only">Previous</span>
            <svg class="h-5 w-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
              <path fill-rule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clip-rule="evenodd" />
            </svg>
          </button>
          <button
            v-for="page in totalPages"
            :key="page"
            @click="goToPage(page)"
            :class="[
              'relative inline-flex items-center px-4 py-2 border text-sm font-medium',
              currentPage === page
                ? 'z-10 bg-blue-50 dark:bg-blue-900 border-blue-500 dark:border-blue-600 text-blue-600 dark:text-blue-300'
                : 'bg-white dark:bg-gray-700 border-gray-300 dark:border-gray-600 text-gray-500 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-600'
            ]"
          >
            {{ page }}
          </button>
          <button
            @click="goToPage(currentPage + 1)"
            :disabled="currentPage === totalPages"
            class="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-sm font-medium text-gray-500 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span class="sr-only">Next</span>
            <svg class="h-5 w-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
              <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
            </svg>
          </button>
        </nav>
      </div>
    </div>

    <!-- Modals -->
    <NewOrderModal
      v-model:show="showNewOrderModal"
      :customers="customers"
      @save="handleSaveOrder"
    />

    <ViewOrderModal
      v-if="orderToView"
      :show="!!orderToView"
      :order="orderToView"
      :order-items="sampleOrderItems"
      @update:show="(val) => { if (!val) orderToView = null }"
    />

  </div>
</template>

<style scoped>
/* Add any component-specific styles here if needed */
/* Tailwind utility classes should cover most styling */
.shadow-sticky {
  box-shadow: -3px 0 5px rgba(0, 0, 0, 0.07);
}

/* Ensure sticky columns stay visible on smaller table scroll */
@media (max-width: 768px) { /* md breakpoint */
  .sticky {
    position: sticky;
    right: 0;
    z-index: 10; /* Ensure it's above other content but below modals */
  }
  .bg-white.dark\:bg-gray-800.shadow-sticky {
    background-color: white; /* Ensure background for light mode */
  }
  .dark .bg-white.dark\:bg-gray-800.shadow-sticky {
    background-color: #1f2937; /* Ensure background for dark mode (gray-800) */
  }
}
</style>
